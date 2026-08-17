package quotes

import (
	"testing"
)

func TestFormatBRL(t *testing.T) {
	tests := []struct {
		val      float64
		expected string
	}{
		{5.72, "5,72"},
		{548200.50, "548.200,50"},
		{1234567.89, "1.234.567,89"},
		{0.05, "0,05"},
	}

	for _, tt := range tests {
		got := FormatBRL(tt.val)
		if got != tt.expected {
			t.Errorf("FormatBRL(%f) = %s; esperado %s", tt.val, got, tt.expected)
		}
	}
}

func TestFormatUSD(t *testing.T) {
	tests := []struct {
		val      float64
		expected string
	}{
		{96150.00, "96,150.00"},
		{2680.50, "2,680.50"},
		{196.25, "196.25"},
	}

	for _, tt := range tests {
		got := FormatUSD(tt.val)
		if got != tt.expected {
			t.Errorf("FormatUSD(%f) = %s; esperado %s", tt.val, got, tt.expected)
		}
	}
}

func TestVariationIndicator(t *testing.T) {
	if got := VariationIndicator(2.45); got != "🟢 +2.45%" {
		t.Errorf("esperado '🟢 +2.45%%', obteve '%s'", got)
	}
	if got := VariationIndicator(-1.20); got != "🔴 -1.20%" {
		t.Errorf("esperado '🔴 -1.20%%', obteve '%s'", got)
	}
	if got := VariationIndicator(0.00); got != "⚪ 0.00%" {
		t.Errorf("esperado '⚪ 0.00%%', obteve '%s'", got)
	}
}
