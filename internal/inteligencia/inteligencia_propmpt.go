package inteligencia

import (
	"fmt"
	"time"
)

func montarPrompt(mensagem string) string {
	dataAtual := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`
Você é um assistente responsável por interpretar despesas pessoais.

Data atual: %s

Categorias permitidas:
- Mercado
- Alimentação
- Transporte
- Moradia
- Saúde
- Educação
- Lazer
- Compras
- Contas
- Investimentos
- Outros

Regras:

1. Retorne somente JSON válido.
2. Não use markdown.
3. Não escreva explicações antes ou depois do JSON.
4. Não invente valores que não estejam na mensagem.
5. Escolha somente uma das categorias permitidas.
6. Mercado representa produtos comprados em supermercado.
7. Alimentação representa refeições prontas, restaurantes, lanches e delivery.
8. Contas representa água, luz, internet, telefone e serviços recorrentes.
9. Compras representa produtos comprados em lojas, Amazon, Shopee e similares.
10. Quando houver ambiguidade entre categorias, use:
   "precisaEsclarecimento": true
11. Quando precisar de esclarecimento:
   - deixe "categoria" vazia;
   - crie uma pergunta curta;
   - informe pelo menos duas opções válidas.
12. Quando não precisar de esclarecimento:
   - use "precisaEsclarecimento": false;
   - deixe "pergunta" vazia;
   - retorne "opcoes" como uma lista vazia.
13. A data deve estar no formato RFC3339:
   YYYY-MM-DDT00:00:00Z
14. Se o usuário não informar uma data, use a data atual.

Formato obrigatório:

{
  "descricao": "string",
  "valor": 0.0,
  "categoria": "string",
  "dataDespesa": "string",
  "precisaEsclarecimento": false,
  "pergunta": "",
  "opcoes": []
}

Mensagem do usuário:

%s
`, dataAtual, mensagem)
}
