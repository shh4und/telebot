package bot

import (
	"telegram-bot/internal/ai"
	"testing"
)

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

func TestBuildModelKeyboard(t *testing.T) {
	models := []ai.OllamaModelInfo{
		{
			Name: "gemma3:latest",
			Details: ai.ModelDetails{
				ParameterSize: "4.3B",
			},
		},
		{
			Name: "deepseek-r1:1.5b",
			Details: ai.ModelDetails{
				ParameterSize: "1.8B",
			},
		},
	}

	keyboard := buildModelKeyboard(models, "gemma3:latest")
	if len(keyboard.InlineKeyboard) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(keyboard.InlineKeyboard))
	}

	btn1 := keyboard.InlineKeyboard[0][0]
	if btn1.Text != "✅ gemma3:latest (4.3B)" {
		t.Errorf("expected selected indicator on btn1, got %s", btn1.Text)
	}
	if btn1.CallbackData != "set_model:gemma3:latest" {
		t.Errorf("expected callback data set_model:gemma3:latest, got %s", btn1.CallbackData)
	}

	btn2 := keyboard.InlineKeyboard[1][0]
	if btn2.Text != "deepseek-r1:1.5b (1.8B)" {
		t.Errorf("expected unselected label on btn2, got %s", btn2.Text)
	}
}
