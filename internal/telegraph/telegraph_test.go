package telegraph

import (
	"encoding/json"
	"testing"
)

func TestMarkdownToNodes(t *testing.T) {
	md := "# Título Principal\n" +
		"Este é um **parágrafo** com texto em *itálico* e um [link](https://telegram.org).\n\n" +
		"- Item 1 da lista\n" +
		"- Item 2 da lista\n\n" +
		"```go\n" +
		"func Hello() {\n" +
		"    fmt.Println(\"World\")\n" +
		"}\n" +
		"```\n"

	nodes, err := MarkdownToNodes(md)
	if err != nil {
		t.Fatalf("MarkdownToNodes falhou: %v", err)
	}

	if len(nodes) == 0 {
		t.Fatalf("Esperava nós retornados, recebeu 0")
	}

	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		t.Fatalf("Erro ao serializar nós para JSON: %v", err)
	}

	t.Logf("Nós gerados (JSON):\n%s", string(nodesJSON))
}
