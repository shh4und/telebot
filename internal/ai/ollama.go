package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"telegram-bot/internal/config"
	"time"
)

var (
	modelsCache      []OllamaModelInfo
	modelsCacheMu    sync.RWMutex
	lastCacheFetch   time.Time
	cacheTTL         = 1 * time.Minute
	httpClient       = &http.Client{Timeout: 60 * time.Second}
)

// GetInstalledModels consulta o endpoint /api/tags do Ollama para obter a lista de modelos instalados.
// Possui um cache de curta duração para otimizar chamadas consecutivas.
func GetInstalledModels() ([]OllamaModelInfo, error) {
	modelsCacheMu.RLock()
	if len(modelsCache) > 0 && time.Since(lastCacheFetch) < cacheTTL {
		cached := make([]OllamaModelInfo, len(modelsCache))
		copy(cached, modelsCache)
		modelsCacheMu.RUnlock()
		return cached, nil
	}
	modelsCacheMu.RUnlock()

	apiHost := config.Envs.ApiHost
	if apiHost == "" {
		return nil, fmt.Errorf("API_HOST is not configured")
	}

	url := fmt.Sprintf("%s/api/tags", apiHost)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models from ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama /api/tags returned status %d", resp.StatusCode)
	}

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama tags response: %w", err)
	}

	modelsCacheMu.Lock()
	modelsCache = tagsResp.Models
	lastCacheFetch = time.Now()
	cached := make([]OllamaModelInfo, len(modelsCache))
	copy(cached, modelsCache)
	modelsCacheMu.Unlock()

	return cached, nil
}

func AskOllama(model string, prompt string) (string, error) {
	apiHost := config.Envs.ApiHost
	if apiHost == "" {
		return "", fmt.Errorf("API_HOST is not configured")
	}

	url := fmt.Sprintf("%s/api/generate", apiHost)

	if model == "" {
		model = DefaultModel
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
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	slog.Info("sending request to ollama", "model", model, "payload_len", len(jsonData))
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	slog.Info("ollama response received", "status", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama generate failed with status: %s", resp.Status)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
