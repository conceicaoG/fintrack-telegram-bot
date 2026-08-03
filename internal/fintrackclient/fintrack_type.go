package fintrackclient

// CriarDespesaRequest representa a estrutura de dados para criar uma nova despesa.
type CriarDespesaRequest struct {
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
}

// DespesaResponse representa a resposta recebida após criar uma despesa.
type DespesaResponse struct {
	ID          int     `json:"id"`
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
}
