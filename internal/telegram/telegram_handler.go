package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/conceicaoG/fintrack-telegram-bot/internal/fintrackclient"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var conversas = make(map[int64]*EstadoConversa)

func IniciarBot() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN não foi encontrado")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Erro ao conectar com o Telegram:", err)
	}

	log.Printf("Bot conectado: @%s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		texto := strings.TrimSpace(update.Message.Text)

		log.Printf(
			"Mensagem recebida de %s: %s",
			update.Message.From.UserName,
			texto,
		)

		if update.Message.IsCommand() {
			processarComando(
				bot,
				chatID,
				update.Message.Command(),
			)
			continue
		}

		processarResposta(bot, chatID, texto)
	}
}

func processarComando(
	bot *tgbotapi.BotAPI,
	chatID int64,
	comando string,
) {
	switch comando {
	case "start":
		conversas[chatID] = &EstadoConversa{
			Etapa: EtapaEscolherCategoria,
		}

		enviarMensagem(
			bot,
			chatID,
			`Olá! Vamos registrar um novo gasto 💰

Escolha uma categoria:

/mercado
/alimentacao
/transporte
/moradia
/saude
/educacao
/lazer
/compras
/contas
/investimentos
/outros`,
		)

	case "mercado":
		selecionarCategoria(bot, chatID, "Mercado")

	case "alimentacao":
		selecionarCategoria(bot, chatID, "Alimentação")

	case "transporte":
		selecionarCategoria(bot, chatID, "Transporte")

	case "moradia":
		selecionarCategoria(bot, chatID, "Moradia")

	case "saude":
		selecionarCategoria(bot, chatID, "Saúde")

	case "educacao":
		selecionarCategoria(bot, chatID, "Educação")

	case "lazer":
		selecionarCategoria(bot, chatID, "Lazer")

	case "compras":
		selecionarCategoria(bot, chatID, "Compras")

	case "contas":
		selecionarCategoria(bot, chatID, "Contas")

	case "investimentos":
		selecionarCategoria(bot, chatID, "Investimentos")

	case "outros":
		selecionarCategoria(bot, chatID, "Outros")

	case "confirmar":
		confirmarDespesa(bot, chatID)

	case "cancelar":
		delete(conversas, chatID)

		enviarMensagem(
			bot,
			chatID,
			"Cadastro cancelado. Digite /start para começar novamente.",
		)

	default:
		enviarMensagem(
			bot,
			chatID,
			"Comando não reconhecido. Digite /start para registrar um gasto.",
		)
	}
}

func selecionarCategoria(
	bot *tgbotapi.BotAPI,
	chatID int64,
	categoria string,
) {
	estado, existe := conversas[chatID]

	if !existe {
		estado = &EstadoConversa{}
		conversas[chatID] = estado
	}

	estado.Categoria = categoria
	estado.Etapa = EtapaAguardarDescricao

	enviarMensagem(
		bot,
		chatID,
		fmt.Sprintf(
			"Categoria selecionada: %s.\n\nO que você comprou?",
			categoria,
		),
	)
}

func processarResposta(
	bot *tgbotapi.BotAPI,
	chatID int64,
	texto string,
) {
	estado, existe := conversas[chatID]

	if !existe {
		enviarMensagem(
			bot,
			chatID,
			"Digite /start para começar o cadastro de uma despesa.",
		)
		return
	}

	switch estado.Etapa {
	case EtapaEscolherCategoria:
		enviarMensagem(
			bot,
			chatID,
			"Escolha uma categoria usando um dos comandos disponíveis.",
		)

	case EtapaAguardarDescricao:
		processarDescricao(bot, chatID, estado, texto)

	case EtapaAguardarValor:
		processarValor(bot, chatID, estado, texto)

	case EtapaAguardarConfirmacao:
		enviarMensagem(
			bot,
			chatID,
			"Digite /confirmar para salvar ou /cancelar para desistir.",
		)

	default:
		enviarMensagem(
			bot,
			chatID,
			"Não consegui identificar a etapa atual. Digite /start novamente.",
		)
	}
}

func processarDescricao(
	bot *tgbotapi.BotAPI,
	chatID int64,
	estado *EstadoConversa,
	descricao string,
) {
	if descricao == "" {
		enviarMensagem(
			bot,
			chatID,
			"A descrição não pode ficar vazia. Digite o nome do que voce comprou.",
		)
		return
	}

	estado.Descricao = descricao
	estado.Etapa = EtapaAguardarValor

	enviarMensagem(
		bot,
		chatID,
		"Qual foi o valor?\nExemplo: 20,80",
	)
}

func processarValor(
	bot *tgbotapi.BotAPI,
	chatID int64,
	estado *EstadoConversa,
	texto string,
) {
	valorTexto := strings.ReplaceAll(texto, ",", ".")

	valor, err := strconv.ParseFloat(valorTexto, 64)
	if err != nil || valor <= 0 {
		enviarMensagem(
			bot,
			chatID,
			"Valor inválido. Digite apenas o valor, por exemplo: 20,80",
		)
		return
	}

	estado.Valor = valor
	estado.Etapa = EtapaAguardarConfirmacao

	resumo := fmt.Sprintf(
		`Confirma este gasto?

Descrição: %s
Categoria: %s
Valor: R$ %.2f

/confirmar
/cancelar`,
		estado.Descricao,
		estado.Categoria,
		estado.Valor,
	)

	enviarMensagem(bot, chatID, resumo)
}

func confirmarDespesa(
	bot *tgbotapi.BotAPI,
	chatID int64,
) {
	estado, existe := conversas[chatID]

	if !existe || estado.Etapa != EtapaAguardarConfirmacao {
		enviarMensagem(
			bot,
			chatID,
			"Não existe uma despesa aguardando confirmação.",
		)
		return
	}

	client := fintrackclient.NovoClient()

	dataAtual := time.Now().
		UTC().
		Format("2006-01-02T00:00:00Z")

	request := fintrackclient.CriarDespesaRequest{
		Descricao:   estado.Descricao,
		Valor:       estado.Valor,
		Categoria:   estado.Categoria,
		DataDespesa: dataAtual,
	}

	despesaCriada, err := client.CriarDespesa(request)
	if err != nil {
		log.Println("Erro ao cadastrar despesa:", err)

		enviarMensagem(
			bot,
			chatID,
			"Não consegui registrar a despesa. Tente novamente.",
		)
		return
	}

	mensagem := fmt.Sprintf(
		`✅ Despesa registrada com sucesso!

ID: %d
Descrição: %s
Categoria: %s
Valor: R$ %.2f`,
		despesaCriada.ID,
		despesaCriada.Descricao,
		despesaCriada.Categoria,
		despesaCriada.Valor,
	)

	enviarMensagem(bot, chatID, mensagem)

	delete(conversas, chatID)
}

func enviarMensagem(
	bot *tgbotapi.BotAPI,
	chatID int64,
	texto string,
) {
	mensagem := tgbotapi.NewMessage(chatID, texto)

	if _, err := bot.Send(mensagem); err != nil {
		log.Println("Erro ao enviar mensagem:", err)
	}
}
