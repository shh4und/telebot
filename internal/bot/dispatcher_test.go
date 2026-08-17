package bot

import "testing"

func TestShouldPublishToTelegraph(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Short plain text",
			input:    "Go é uma linguagem compilada e tipada desenvolvida pela Google.",
			expected: false,
		},
		{
			name:     "Short text with code block",
			input:    "Exemplo:\n```go\nfmt.Println(\"Oi\")\n```",
			expected: true,
		},
		{
			name: "Long text over 300 characters",
			input: "Go (também conhecida como Golang) é uma linguagem de programação de código aberto criada pelo Google. " +
				"Ela foi projetada para ser simples, eficiente, rápida de compilar e ter suporte nativo a concorrência " +
				"através de goroutines e canais. É amplamente utilizada no desenvolvimento de sistemas back-end, microsserviços, " +
				"ferramentas de linha de comando e infraestrutura de nuvem moderna.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldPublishToTelegraph(tt.input)
			if result != tt.expected {
				t.Errorf("shouldPublishToTelegraph() = %v, want %v", result, tt.expected)
			}
		})
	}
}
