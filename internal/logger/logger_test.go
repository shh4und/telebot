package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
	}
}

func TestNew_TextHandler(t *testing.T) {
	var buf bytes.Buffer
	l := New("debug", "text", &buf)

	l.Debug("mensagem de depuracao", "chave", "valor")
	output := buf.String()

	if !strings.Contains(output, "level=DEBUG") {
		t.Errorf("saída esperada contendo level=DEBUG, obtido: %s", output)
	}
	if !strings.Contains(output, "msg=\"mensagem de depuracao\"") && !strings.Contains(output, "msg=mensagem de depuracao") {
		t.Errorf("saída esperada contendo mensagem, obtido: %s", output)
	}
	if !strings.Contains(output, "chave=valor") {
		t.Errorf("saída esperada contendo chave=valor, obtido: %s", output)
	}
}

func TestNew_JSONHandler(t *testing.T) {
	var buf bytes.Buffer
	l := New("info", "json", &buf)

	l.Info("mensagem json", "user_id", 12345)
	output := buf.String()

	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("saída esperada contendo level JSON INFO, obtido: %s", output)
	}
	if !strings.Contains(output, `"msg":"mensagem json"`) {
		t.Errorf("saída esperada contendo msg JSON, obtido: %s", output)
	}
	if !strings.Contains(output, `"user_id":12345`) {
		t.Errorf("saída esperada contendo user_id JSON, obtido: %s", output)
	}
}

func TestInit(t *testing.T) {
	l := Init("info", "text")
	if l == nil {
		t.Fatal("esperado logger não nulo retornado por Init")
	}
	if slog.Default() != l {
		t.Error("esperado slog.Default() atualizado com o novo logger")
	}
}
