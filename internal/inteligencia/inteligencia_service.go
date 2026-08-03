package inteligencia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Erros específicos para validação de despesas interpretadas pela IA.
var (
	ErrMensagemObrigatoria  = errors.New("mensagem obrigatória")
	ErrRespostaInvalida     = errors.New("resposta inválida da inteligência")
	ErrDescricaoObrigatoria = errors.New("descrição não identificada")
	ErrValorInvalido        = errors.New("valor não identificado ou inválido")
	ErrCategoriaInvalida    = errors.New("categoria inválida")
	ErrPerguntaObrigatoria  = errors.New("pergunta de esclarecimento obrigatória")
	ErrOpcoesInvalidas      = errors.New("opções de esclarecimento inválidas")
)

// categoriasPermitidas define as categorias válidas para despesas.
var categoriasPermitidas = map[string]bool{
	"Mercado":       true,
	"Alimentação":   true,
	"Transporte":    true,
	"Moradia":       true,
	"Saúde":         true,
	"Educação":      true,
	"Lazer":         true,
	"Compras":       true,
	"Contas":        true,
	"Investimentos": true,
	"Outros":        true,
}

// Service é responsável por interpretar mensagens de despesas usando a IA.
type Service struct {
	client *Client
}

// NewService cria uma nova instância do Service com o cliente de IA fornecido.
func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// InterpretarDespesa interpreta a mensagem do usuário e retorna uma estrutura DespesaInterpretada.
func (s *Service) InterpretarDespesa(
	ctx context.Context,
	mensagem string,
) (DespesaInterpretada, error) {
	mensagem = strings.TrimSpace(mensagem)

	if mensagem == "" {
		return DespesaInterpretada{}, ErrMensagemObrigatoria
	}

	prompt := montarPrompt(mensagem)

	respostaGemini, err := s.client.GerarResposta(ctx, prompt)
	if err != nil {
		return DespesaInterpretada{}, fmt.Errorf(
			"erro ao interpretar despesa: %w",
			err,
		)
	}

	respostaLimpa := limparRespostaJSON(respostaGemini)

	var despesa DespesaInterpretada

	if err := json.Unmarshal(
		[]byte(respostaLimpa),
		&despesa,
	); err != nil {
		return DespesaInterpretada{}, fmt.Errorf(
			"%w: %v",
			ErrRespostaInvalida,
			err,
		)
	}

	normalizarDespesa(&despesa)

	if err := validarDespesaInterpretada(despesa); err != nil {
		return DespesaInterpretada{}, err
	}

	return despesa, nil
}

// normalizarDespesa ajusta os campos da despesa interpretada para um formato consistente.
func normalizarDespesa(despesa *DespesaInterpretada) {
	despesa.Descricao = strings.TrimSpace(despesa.Descricao)
	despesa.Categoria = normalizarCategoria(despesa.Categoria)
	despesa.Pergunta = strings.TrimSpace(despesa.Pergunta)

	for i, opcao := range despesa.Opcoes {
		despesa.Opcoes[i] = normalizarCategoria(opcao)
	}
}

// validarDespesaInterpretada verifica se os campos da despesa interpretada estão corretos e completos.
func validarDespesaInterpretada(
	despesa DespesaInterpretada,
) error {
	if despesa.Descricao == "" {
		return ErrDescricaoObrigatoria
	}

	if despesa.Valor <= 0 {
		return ErrValorInvalido
	}

	if despesa.PrecisaEsclarecimento {
		if despesa.Pergunta == "" {
			return ErrPerguntaObrigatoria
		}

		if len(despesa.Opcoes) < 2 {
			return ErrOpcoesInvalidas
		}

		for _, opcao := range despesa.Opcoes {
			if !categoriaValida(opcao) {
				return fmt.Errorf(
					"%w: %s",
					ErrOpcoesInvalidas,
					opcao,
				)
			}
		}

		return nil
	}

	if !categoriaValida(despesa.Categoria) {
		return fmt.Errorf(
			"%w: %s",
			ErrCategoriaInvalida,
			despesa.Categoria,
		)
	}

	return nil
}

// categoriaValida verifica se a categoria fornecida é uma das categorias permitidas.
func categoriaValida(categoria string) bool {
	_, existe := categoriasPermitidas[categoria]
	return existe
}

// normalizarCategoria ajusta a categoria para um formato consistente, considerando variações de capitalização e acentuação.
func normalizarCategoria(categoria string) string {
	categoria = strings.TrimSpace(categoria)
	categoriaMinuscula := strings.ToLower(categoria)

	categorias := map[string]string{
		"mercado":       "Mercado",
		"alimentação":   "Alimentação",
		"alimentacao":   "Alimentação",
		"transporte":    "Transporte",
		"moradia":       "Moradia",
		"saúde":         "Saúde",
		"saude":         "Saúde",
		"educação":      "Educação",
		"educacao":      "Educação",
		"lazer":         "Lazer",
		"compras":       "Compras",
		"contas":        "Contas",
		"investimentos": "Investimentos",
		"outros":        "Outros",
	}

	categoriaNormalizada, existe := categorias[categoriaMinuscula]
	if existe {
		return categoriaNormalizada
	}

	return categoria
}

// limparRespostaJSON remove elementos indesejados da resposta JSON retornada pela IA, garantindo que apenas o conteúdo relevante seja processado.
func limparRespostaJSON(resposta string) string {
	resposta = strings.TrimSpace(resposta)

	resposta = strings.TrimPrefix(resposta, "```json")
	resposta = strings.TrimPrefix(resposta, "```")
	resposta = strings.TrimSuffix(resposta, "```")

	return strings.TrimSpace(resposta)
}
