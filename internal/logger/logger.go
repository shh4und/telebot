package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// ParseLevel converte uma string de nível em slog.Level correspondente.
func ParseLevel(lvl string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New cria e configura um slog.Logger com o nível e formato especificados.
func New(levelStr, formatStr string, out io.Writer) *slog.Logger {
	if out == nil {
		out = os.Stdout
	}

	level := ParseLevel(levelStr)
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(formatStr)) {
	case "json":
		handler = slog.NewJSONHandler(out, opts)
	default:
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}

// Init inicializa o logger padrão do slog com as opções configuradas.
func Init(levelStr, formatStr string) *slog.Logger {
	l := New(levelStr, formatStr, os.Stdout)
	slog.SetDefault(l)
	return l
}
