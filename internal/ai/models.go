package ai

var AvailableModels = map[string]bool{
	"ministral-3:8b": true,
}

const DefaultModel = "ministral-3:8b"

// Estruturas baseadas na API do Ollama
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
