package telegraph

import (
	"encoding/json"
	"testing"
)

func TestMarkdownToNodes(t *testing.T) {
	md := `# Título Principal
Este é um **parágrafo** com texto em *itálico* e um [link](https://telegram.org).

- Item 1 da lista
- Item 2 da lista

```go
func Hello() {
    fmt.Println("World")
}
```
`

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
