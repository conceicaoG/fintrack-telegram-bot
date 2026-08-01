package telegram

type Etapa string

const (
	EtapaEscolherCategoria   Etapa = "escolher_categoria"
	EtapaAguardarDescricao   Etapa = "aguardar_descricao"
	EtapaAguardarValor       Etapa = "aguardar_valor"
	EtapaAguardarConfirmacao Etapa = "aguardar_confirmacao"
)

type EstadoConversa struct {
	Etapa     Etapa
	Categoria string
	Descricao string
	Valor     float64
}
