package fintrackclient

type CriarDespesaRequest struct {
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
}

type DespesaResponse struct {
	ID          int     `json:"id"`
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
}