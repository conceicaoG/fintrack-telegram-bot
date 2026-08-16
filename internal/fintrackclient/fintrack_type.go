package fintrackclient

// UsuarioRequest representa os dados do usuário do Telegram.
type UsuarioRequest struct {
	TelegramUserID int64  `json:"telegramUserId"`
	Nome           string `json:"nome"`
	Username       string `json:"username"`
}

// DespesaRequest representa os dados da despesa.
type DespesaRequest struct {
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
}

// CriarDespesaRequest representa o contrato enviado pelo Bot para o BFA.
type CriarDespesaRequest struct {
	Usuario UsuarioRequest `json:"usuario"`
	Despesa DespesaRequest `json:"despesa"`
}

// DespesaResponse representa a resposta recebida após criar uma despesa.
type DespesaResponse struct {
	ID          int     `json:"id"`
	Descricao   string  `json:"descricao"`
	Valor       float64 `json:"valor"`
	Categoria   string  `json:"categoria"`
	DataDespesa string  `json:"dataDespesa"`
	UsuarioID   int64   `json:"usuarioId"`
}
