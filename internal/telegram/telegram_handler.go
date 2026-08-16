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

// Handler é responsável por gerenciar a interação com o bot do Telegram, processando mensagens e comandos dos usuários.
type Handler struct {
	bot                 *tgbotapi.BotAPI
	inteligenciaService *inteligencia.Service
	fintrackClient      *fintrackclient.Client
}

var conversas = make(map[int64]*EstadoConversa)

// EstadoConversa mantém o estado atual da conversa com um usuário específico, incluindo informações sobre a despesa em andamento e a etapa do processo.
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

// IniciarBot inicia o bot do Telegram, configurando o cliente, o serviço de IA e o cliente do FinTrack, e começa a processar atualizações recebidas.
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

// processarUpdate processa cada atualização recebida do Telegram, determinando se é uma mensagem ou comando e chamando os métodos apropriados para lidar com ela.
func (h *Handler) processarUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	texto := strings.TrimSpace(update.Message.Text)

	telegramUserID := update.Message.From.ID

	nome := strings.TrimSpace(
		update.Message.From.FirstName + " " +
			update.Message.From.LastName,
	)

	username := update.Message.From.UserName

	log.Printf(
		"[BOT] Mensagem recebida | telegramUserId=%d | nome=%s | username=%s | texto=%s",
		telegramUserID,
		nome,
		username,
		texto,
	)

	if update.Message.IsCommand() {
		h.processarComando(
			chatID,
			update.Message.Command(),
		)
		return
	}

	h.processarMensagem(
		chatID,
		texto,
		telegramUserID,
		nome,
		username,
	)
}

// processarComando lida com os comandos recebidos do usuário, como /start, /ajuda, /confirmar e /cancelar, chamando os métodos correspondentes para cada comando.
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

// enviarApresentacao envia uma mensagem de apresentação ao usuário, explicando como o bot funciona e fornecendo exemplos de como enviar gastos para interpretação.
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

// processarMensagem lida com mensagens de texto recebidas do usuário, determinando se é uma saudação, uma despesa ou se há necessidade de esclarecimento, e chamando os métodos apropriados para cada caso.
func (h *Handler) processarMensagem(
	chatID int64,
	texto string,
	telegramUserID int64,
	nome string,
	username string,
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

	h.interpretarMensagemComIA(
		chatID,
		texto,
		telegramUserID,
		nome,
		username,
	)
}

// interpretarMensagemComIA envia a mensagem do usuário para o serviço de IA para interpretação, processa a resposta e atualiza o estado da conversa, solicitando esclarecimentos ou confirmação conforme necessário.
func (h *Handler) interpretarMensagemComIA(
	chatID int64,
	texto string,
	telegramUserID int64,
	nome string,
	username string,
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
		Descricao:      despesa.Descricao,
		Valor:          despesa.Valor,
		Categoria:      despesa.Categoria,
		DataDespesa:    despesa.DataDespesa,
		Opcoes:         despesa.Opcoes,
		TelegramUserID: telegramUserID,
		Nome:           nome,
		Username:       username,
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

// processarEsclarecimento lida com a resposta do usuário quando a IA solicitou esclarecimentos sobre a categoria da despesa, atualizando o estado da conversa e solicitando confirmação.
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

// encontrarOpcao verifica se a resposta do usuário corresponde a uma das opções fornecidas pela IA, retornando a opção correspondente ou uma string vazia se não houver correspondência.
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

// EstadoConversa mantém o estado atual da conversa com um usuário específico, incluindo informações sobre a despesa em andamento e a etapa do processo.
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

// confirmarDespesa envia a despesa interpretada para o FinTrack, registrando-a no sistema, e envia uma mensagem de confirmação ao usuário. Caso haja algum erro, informa o usuário e mantém a despesa em estado de espera.
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
		Usuario: fintrackclient.UsuarioRequest{
			TelegramUserID: estado.TelegramUserID,
			Nome:           estado.Nome,
			Username:       estado.Username,
		},
		Despesa: fintrackclient.DespesaRequest{
			Descricao:   estado.Descricao,
			Valor:       estado.Valor,
			Categoria:   estado.Categoria,
			DataDespesa: dataDespesa,
		},
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

// cancelarDespesa cancela a despesa em andamento, removendo o estado da conversa e informando ao usuário que a despesa foi cancelada.
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

// dataAtualRFC3339 retorna a data atual no formato RFC3339, que é utilizado para registrar a data da despesa no FinTrack.
func dataAtualRFC3339() string {
	return time.Now().
		UTC().
		Format("2006-01-02T00:00:00Z")
}

// enviarMensagem envia uma mensagem de texto para o usuário no Telegram, lidando com possíveis erros de envio e registrando-os no log.
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

// ehCumprimento verifica se a mensagem recebida é uma saudação comum, retornando true se for, e false caso contrário.
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
