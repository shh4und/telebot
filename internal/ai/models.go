package ai

const DefaultModel = "gemma3:latest"

// Estruturas baseadas na API do Ollama
type OllamaRequest struct {
	Model   string              `json:"model"`
	Prompt  string              `json:"prompt"`
	Stream  bool                `json:"stream"`
	Message []map[string]string `json:"message,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type OllamaTagsResponse struct {
	Models []OllamaModelInfo `json:"models"`
}

type OllamaModelInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt string       `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}
