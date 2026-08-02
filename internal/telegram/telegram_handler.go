package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/conceicaoG/fintrack-telegram-bot/internal/fintrackclient"
	"github.com/conceicaoG/fintrack-telegram-bot/internal/inteligencia"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot                 *tgbotapi.BotAPI
	inteligenciaService *inteligencia.Service
	fintrackClient      *fintrackclient.Client
}

var conversas = make(map[int64]*EstadoConversa)

func NovoHandler(
	bot *tgbotapi.BotAPI,
	inteligenciaService *inteligencia.Service,
	fintrackClient *fintrackclient.Client,
) *Handler {
	return &Handler{
		bot:                 bot,
		inteligenciaService: inteligenciaService,
		fintrackClient:      fintrackClient,
	}
}

func IniciarBot(
	inteligenciaService *inteligencia.Service,
	fintrackClient *fintrackclient.Client,
) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN não foi encontrado")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("erro ao conectar com o Telegram:", err)
	}

	handler := NovoHandler(
		bot,
		inteligenciaService,
		fintrackClient,
	)

	log.Printf("[BOT] Conectado: @%s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		handler.processarUpdate(update)
	}
}

func (h *Handler) processarUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	texto := strings.TrimSpace(update.Message.Text)

	log.Printf(
		"[BOT] Mensagem recebida de %s: %s",
		update.Message.From.UserName,
		texto,
	)

	if update.Message.IsCommand() {
		h.processarComando(
			chatID,
			update.Message.Command(),
		)
		return
	}

	h.processarMensagem(chatID, texto)
}

func (h *Handler) processarComando(
	chatID int64,
	comando string,
) {
	switch comando {
	case "start", "ajuda":
		h.enviarApresentacao(chatID)

	case "confirmar":
		h.confirmarDespesa(chatID)

	case "cancelar":
		h.cancelarDespesa(chatID)

	default:
		h.enviarMensagem(
			chatID,
			"Comando não reconhecido. Digite /ajuda para ver como usar o FinTrack.",
		)
	}
}

func (h *Handler) enviarApresentacao(chatID int64) {
	h.enviarMensagem(
		chatID,
		`Olá! 👋 Eu sou o agente do FinTrack.

Envie seu gasto em uma frase, por exemplo:

• Papel higiênico 20,80
• Conta de luz 150
• Uber 35 ontem
• Comprei uma camisa por 89,90 na Shopee

Eu vou interpretar as informações e pedir sua confirmação antes de salvar.`,
	)
}

func (h *Handler) processarMensagem(
	chatID int64,
	texto string,
) {
	if texto == "" {
		h.enviarMensagem(
			chatID,
			"Envie uma mensagem com a descrição e o valor do gasto.",
		)
		return
	}

	estado, existe := conversas[chatID]

	if existe {
		switch estado.Etapa {
		case EtapaAguardarEsclarecimento:
			h.processarEsclarecimento(
				chatID,
				estado,
				texto,
			)
			return

		case EtapaAguardarConfirmacao:
			h.enviarMensagem(
				chatID,
				"Existe uma despesa aguardando confirmação.\n\nDigite /confirmar ou /cancelar.",
			)
			return
		}
	}

	if ehCumprimento(texto) {
		h.enviarApresentacao(chatID)
		return
	}

	h.interpretarMensagemComIA(chatID, texto)
}

func (h *Handler) interpretarMensagemComIA(
	chatID int64,
	texto string,
) {
	h.enviarMensagem(
		chatID,
		"Estou analisando seu gasto... 🤖",
	)

	despesa, err := h.inteligenciaService.InterpretarDespesa(
		context.Background(),
		texto,
	)
	if err != nil {
		log.Printf(
			"[BOT] Erro ao interpretar mensagem com Gemini: %v",
			err,
		)

		delete(conversas, chatID)

		h.enviarMensagem(
			chatID,
			`Não consegui interpretar esse gasto.

Tente novamente escrevendo, por exemplo:

"Uber 20,80"`,
		)
		return
	}

	estado := &EstadoConversa{
		Descricao:   despesa.Descricao,
		Valor:       despesa.Valor,
		Categoria:   despesa.Categoria,
		DataDespesa: despesa.DataDespesa,
		Opcoes:      despesa.Opcoes,
	}

	conversas[chatID] = estado

	if despesa.PrecisaEsclarecimento {
		estado.Etapa = EtapaAguardarEsclarecimento

		mensagem := despesa.Pergunta

		if len(despesa.Opcoes) > 0 {
			mensagem += "\n\nEscolha uma opção:\n"

			for _, opcao := range despesa.Opcoes {
				mensagem += fmt.Sprintf("• %s\n", opcao)
			}
		}

		h.enviarMensagem(chatID, mensagem)
		return
	}

	estado.Etapa = EtapaAguardarConfirmacao
	h.enviarResumoConfirmacao(chatID, estado)
}

func (h *Handler) processarEsclarecimento(
	chatID int64,
	estado *EstadoConversa,
	resposta string,
) {
	categoriaEscolhida := encontrarOpcao(
		resposta,
		estado.Opcoes,
	)

	if categoriaEscolhida == "" {
		mensagem := "Não reconheci essa opção. Escolha uma destas categorias:\n\n"

		for _, opcao := range estado.Opcoes {
			mensagem += fmt.Sprintf("• %s\n", opcao)
		}

		h.enviarMensagem(chatID, mensagem)
		return
	}

	estado.Categoria = categoriaEscolhida
	estado.Opcoes = nil
	estado.Etapa = EtapaAguardarConfirmacao

	h.enviarResumoConfirmacao(chatID, estado)
}

func encontrarOpcao(
	resposta string,
	opcoes []string,
) string {
	resposta = strings.TrimSpace(resposta)

	for _, opcao := range opcoes {
		if strings.EqualFold(resposta, opcao) {
			return opcao
		}
	}

	return ""
}

func (h *Handler) enviarResumoConfirmacao(
	chatID int64,
	estado *EstadoConversa,
) {
	mensagem := fmt.Sprintf(
		`Entendi seu gasto 📝

Descrição: %s
Categoria: %s
Valor: R$ %.2f

Está tudo certo?

/confirmar
/cancelar`,
		estado.Descricao,
		estado.Categoria,
		estado.Valor,
	)

	h.enviarMensagem(chatID, mensagem)
}

func (h *Handler) confirmarDespesa(chatID int64) {
	estado, existe := conversas[chatID]

	if !existe || estado.Etapa != EtapaAguardarConfirmacao {
		h.enviarMensagem(
			chatID,
			"Não existe uma despesa aguardando confirmação.",
		)
		return
	}

	dataDespesa := estado.DataDespesa

	if dataDespesa == "" {
		dataDespesa = dataAtualRFC3339()
	}

	request := fintrackclient.CriarDespesaRequest{
		Descricao:   estado.Descricao,
		Valor:       estado.Valor,
		Categoria:   estado.Categoria,
		DataDespesa: dataDespesa,
	}

	log.Printf(
		"[BOT] Enviando despesa para o BFA: %+v",
		request,
	)

	despesaCriada, err := h.fintrackClient.CriarDespesa(request)
	if err != nil {
		log.Printf(
			"[BOT] Erro ao cadastrar despesa: %v",
			err,
		)

		h.enviarMensagem(
			chatID,
			"Não consegui registrar a despesa. Digite /confirmar para tentar novamente ou /cancelar.",
		)
		return
	}

	mensagem := fmt.Sprintf(
		`✅ Despesa registrada com sucesso!

🛒 Descrição: %s
📦 Categoria: %s
💰 Valor: R$ %.2f

Seu gasto foi salvo no FinTrack.`,
		despesaCriada.Descricao,
		despesaCriada.Categoria,
		despesaCriada.Valor,
	)
	h.enviarMensagem(chatID, mensagem)

	delete(conversas, chatID)
}

func (h *Handler) cancelarDespesa(chatID int64) {
	_, existe := conversas[chatID]

	if !existe {
		h.enviarMensagem(
			chatID,
			"Não existe uma despesa em andamento.",
		)
		return
	}

	delete(conversas, chatID)

	h.enviarMensagem(
		chatID,
		"Despesa cancelada. Envie outro gasto quando quiser.",
	)
}

func dataAtualRFC3339() string {
	return time.Now().
		UTC().
		Format("2006-01-02T00:00:00Z")
}

func (h *Handler) enviarMensagem(
	chatID int64,
	texto string,
) {
	mensagem := tgbotapi.NewMessage(chatID, texto)

	if _, err := h.bot.Send(mensagem); err != nil {
		log.Printf(
			"[BOT] Erro ao enviar mensagem: %v",
			err,
		)
	}
}

func ehCumprimento(texto string) bool {
	texto = strings.ToLower(strings.TrimSpace(texto))

	cumprimentos := map[string]bool{
		"oi":        true,
		"olá":       true,
		"ola":       true,
		"e aí":      true,
		"e ai":      true,
		"bom dia":   true,
		"boa tarde": true,
		"boa noite": true,
		"tudo bem":  true,
		"tudo bem?": true,
		"como vai":  true,
		"como vai?": true,
		"fala":      true,
		"opa":       true,
		"salve":     true,
	}

	return cumprimentos[texto]
}
