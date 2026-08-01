package fintrackclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NovoClient() *Client {
	baseURL := os.Getenv("FINTRACK_BFA_URL")

	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) CriarDespesa(
	despesa CriarDespesaRequest,
) (*DespesaResponse, error) {
	body, err := json.Marshal(despesa)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao converter despesa para JSON: %w",
			err,
		)
	}

	url := c.baseURL + "/despesas"

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao criar requisição: %w",
			err,
		)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao chamar API FinTrack: %w",
			err,
		)
	}
	defer resp.Body.Close()

	respostaBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"erro ao ler resposta da API: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"API FinTrack retornou status %d: %s",
			resp.StatusCode,
			string(respostaBody),
		)
	}

	var resposta DespesaResponse

	if err := json.Unmarshal(respostaBody, &resposta); err != nil {
		return nil, fmt.Errorf(
			"erro ao converter resposta da API: %w",
			err,
		)
	}

	return &resposta, nil
}
