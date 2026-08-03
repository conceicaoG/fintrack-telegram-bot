package inteligencia

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client
	model  string
}

// NewClient cria uma nova instância do Client para interagir com a API Gemini.
func NewClient() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.6-flash"
	}

	client, err := genai.NewClient(
		context.Background(),
		&genai.ClientConfig{
			APIKey: apiKey,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar client do Gemini: %w", err)
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

// GerarResposta envia um prompt para a API Gemini e retorna a resposta gerada.
func (c *Client) GerarResposta(
	ctx context.Context,
	prompt string,
) (string, error) {

	resp, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)

	if err != nil {
		return "", fmt.Errorf(
			"erro ao chamar Gemini: %w",
			err,
		)
	}

	return resp.Text(), nil
}
