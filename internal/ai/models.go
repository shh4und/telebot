package ai

var AvailableModels = map[string]bool{
	"gemma3n:e4b": true,
	"gemma3:4b":   true,
}

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
