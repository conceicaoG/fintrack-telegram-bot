package telegram

type Etapa string

const (
	EtapaAguardarEsclarecimento Etapa = "aguardar_esclarecimento"
	EtapaAguardarConfirmacao    Etapa = "aguardar_confirmacao"
)

type EstadoConversa struct {
	Etapa       Etapa
	Categoria   string
	Descricao   string
	Valor       float64
	DataDespesa string
	Opcoes      []string
}
