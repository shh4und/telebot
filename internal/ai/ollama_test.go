package ai

import (
	"encoding/json"
	"testing"
)

func TestDecodeOllamaTags(t *testing.T) {
	rawJSON := `{
		"models": [
			{
				"name": "qwen3.5:0.8b",
				"model": "qwen3.5:0.8b",
				"modified_at": "2026-08-16T22:25:58.175744464-03:00",
				"size": 1036046583,
				"digest": "f3817196d142eaf72ce79dfebe53dcb20bd21da87ce13e138a8f8e10a866b3a4",
				"details": {
					"parent_model": "",
					"format": "gguf",
					"family": "qwen35",
					"families": ["qwen35"],
					"parameter_size": "873.44M",
					"quantization_level": "Q8_0"
				}
			},
			{
				"name": "deepseek-r1:1.5b",
				"model": "deepseek-r1:1.5b",
				"modified_at": "2026-08-16T22:23:55.981176707-03:00",
				"size": 1117322768,
				"digest": "e0979632db5a88d1a53884cb2a941772d10ff5d055aabaa6801c4e36f3a6c2d7",
				"details": {
					"parent_model": "",
					"format": "gguf",
					"family": "qwen2",
					"families": ["qwen2"],
					"parameter_size": "1.8B",
					"quantization_level": "Q4_K_M"
				}
			}
		]
	}`

	var resp OllamaTagsResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	if len(resp.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(resp.Models))
	}

	if resp.Models[0].Name != "qwen3.5:0.8b" {
		t.Errorf("expected qwen3.5:0.8b, got %s", resp.Models[0].Name)
	}
	if resp.Models[0].Details.ParameterSize != "873.44M" {
		t.Errorf("expected 873.44M, got %s", resp.Models[0].Details.ParameterSize)
	}

	if resp.Models[1].Name != "deepseek-r1:1.5b" {
		t.Errorf("expected deepseek-r1:1.5b, got %s", resp.Models[1].Name)
	}
	if resp.Models[1].Details.ParameterSize != "1.8B" {
		t.Errorf("expected 1.8B, got %s", resp.Models[1].Details.ParameterSize)
	}
}
