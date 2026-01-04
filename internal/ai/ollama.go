package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func askOllama(model string, prompt string) (string, error) {
	url := "http://127.0.0.1:11434/api/generate"

	if !(AvailableModels[model]) {
		return "", fmt.Errorf("Model %s not available", model)
	}

	payload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
