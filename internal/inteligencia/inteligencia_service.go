package inteligencia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMensagemObrigatoria  = errors.New("mensagem obrigatória")
	ErrRespostaInvalida     = errors.New("resposta inválida da inteligência")
	ErrDescricaoObrigatoria = errors.New("descrição não identificada")
	ErrValorInvalido        = errors.New("valor não identificado ou inválido")
	ErrCategoriaInvalida    = errors.New("categoria inválida")
	ErrPerguntaObrigatoria  = errors.New("pergunta de esclarecimento obrigatória")
	ErrOpcoesInvalidas      = errors.New("opções de esclarecimento inválidas")
)

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

type Service struct {
	client *Client
}

func NovoService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

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

func normalizarDespesa(despesa *DespesaInterpretada) {
	despesa.Descricao = strings.TrimSpace(despesa.Descricao)
	despesa.Categoria = normalizarCategoria(despesa.Categoria)
	despesa.Pergunta = strings.TrimSpace(despesa.Pergunta)

	for i, opcao := range despesa.Opcoes {
		despesa.Opcoes[i] = normalizarCategoria(opcao)
	}
}

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

func categoriaValida(categoria string) bool {
	_, existe := categoriasPermitidas[categoria]
	return existe
}

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

func limparRespostaJSON(resposta string) string {
	resposta = strings.TrimSpace(resposta)

	resposta = strings.TrimPrefix(resposta, "```json")
	resposta = strings.TrimPrefix(resposta, "```")
	resposta = strings.TrimSuffix(resposta, "```")

	return strings.TrimSpace(resposta)
}
