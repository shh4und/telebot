package ai

var AvailableModels = map[string]bool{
	"gemma3:1b":      true,
	"phi3:3.8b":      true,
	"phi4-mini:3.8b": true,
}

const DefaultModel = "phi4-mini:3.8b"

// Estruturas baseadas na API do Ollama
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	System string `json:"system"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
