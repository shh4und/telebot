package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"telegram-bot/internal/config"
)

var apiHost string = config.Envs.ApiHost

func AskOllama(model string, prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", apiHost)

	if model == "" {
		model = DefaultModel
	}

	if !(AvailableModels[model]) {
		return "", fmt.Errorf("Model %s not available", model)
	}
	payload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Message: []map[string]string{
			{
				"role":    "user",
				"content": "responda em pt-br",
			},
		},
		Format: "json",
	}

	jsonData, _ := json.Marshal(payload)
	slog.Info("payload", "payload", string(jsonData))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	slog.Info("response", "status", resp.Status)
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}
	slog.Info("ollama response", "response", ollamaResp.Response)
	return ollamaResp.Response, nil
}
