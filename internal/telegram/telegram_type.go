package telegram

type Etapa string

// Constantes que representam as etapas do processo de registro de uma despesa.
const (
	EtapaAguardarEsclarecimento Etapa = "aguardar_esclarecimento"
	EtapaAguardarConfirmacao    Etapa = "aguardar_confirmacao"
)

// EstadoConversa mantém o estado atual da conversa com um usuário específico, incluindo informações sobre a despesa em andamento e a etapa do processo.
type EstadoConversa struct {
	Etapa       Etapa
	Categoria   string
	Descricao   string
	Valor       float64
	DataDespesa string
	Opcoes      []string
}
