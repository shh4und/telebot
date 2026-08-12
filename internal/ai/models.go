package ai

var AvailableModels = map[string]bool{
	"gemma3:latest": true,
}

const DefaultModel = "gemma3:latest"

// Estruturas baseadas na API do Ollama
type OllamaRequest struct {
	Model   string              `json:"model"`
	Prompt  string              `json:"prompt"`
	Stream  bool                `json:"stream"`
	Message []map[string]string `json:"message,omitempty"`
	Format  string              `json:"format,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
