package inteligencia

type DespesaInterpretada struct {
	Descricao             string   `json:"descricao"`
	Valor                 float64  `json:"valor"`
	Categoria             string   `json:"categoria"`
	DataDespesa           string   `json:"dataDespesa"`
	PrecisaEsclarecimento bool     `json:"precisaEsclarecimento"`
	Pergunta              string   `json:"pergunta"`
	Opcoes                []string `json:"opcoes"`
}
