package inteligencia

// DespesaInterpretada representa a estrutura de dados retornada pela IA após interpretar a mensagem do usuário.
type DespesaInterpretada struct {
	Descricao             string   `json:"descricao"`
	Valor                 float64  `json:"valor"`
	Categoria             string   `json:"categoria"`
	DataDespesa           string   `json:"dataDespesa"`
	PrecisaEsclarecimento bool     `json:"precisaEsclarecimento"`
	Pergunta              string   `json:"pergunta"`
	Opcoes                []string `json:"opcoes"`
}
