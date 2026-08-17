package bot

import (
	"strings"
	"telegram-bot/internal/news"
	"telegram-bot/internal/quotes"
	"testing"
	"time"
)

func TestBuildNewsKeyboard(t *testing.T) {
	markup := buildNewsKeyboard("br")
	if len(markup.InlineKeyboard) != 2 {
		t.Fatalf("esperava 2 linhas de botões, obteve %d", len(markup.InlineKeyboard))
	}

	btnBR := markup.InlineKeyboard[0][0]
	if !strings.HasPrefix(btnBR.Text, "✅") {
		t.Errorf("esperava indicador selecionado no botão Brasil, obteve %s", btnBR.Text)
	}

	btnMundo := markup.InlineKeyboard[0][1]
	if strings.HasPrefix(btnMundo.Text, "✅") {
		t.Errorf("não esperava indicador selecionado no botão Mundo, obteve %s", btnMundo.Text)
	}
}

func TestFormatNewsMessage(t *testing.T) {
	articles := []news.Article{
		{
			Title:  "Nova tecnologia é lançada",
			URL:    "https://example.com/tech",
			Source: "Canaltech",
		},
	}
	info := &news.CategoryInfo{
		Title: "Tecnologia",
		Emoji: "💻",
	}

	msg := formatNewsMessage(articles, info)
	if !strings.Contains(msg, "💻 *Manchetes: Tecnologia*") {
		t.Errorf("título não encontrado na mensagem: %s", msg)
	}
	if !strings.Contains(msg, "[Nova tecnologia é lançada](https://example.com/tech)") {
		t.Errorf("link do artigo não formatado corretamente: %s", msg)
	}
	if !strings.Contains(msg, "Fonte: Canaltech") {
		t.Errorf("fonte não encontrada na mensagem: %s", msg)
	}
}

func TestFormatQuotesMessage(t *testing.T) {
	summary := &quotes.MarketSummary{
		Dollar: quotes.CurrencyQuote{
			Code:      "USD",
			Bid:       5.72,
			High:      5.75,
			Low:       5.68,
			PctChange: 0.35,
		},
		Euro: quotes.CurrencyQuote{
			Code:      "EUR",
			Bid:       6.24,
			High:      6.28,
			Low:       6.20,
			PctChange: -0.15,
		},
		Cryptos: []quotes.CryptoQuote{
			{
				ID:        "bitcoin",
				Symbol:    "BTC",
				Name:      "Bitcoin",
				PriceBRL:  548200.0,
				PriceUSD:  96150.0,
				Change24h: 2.45,
			},
		},
		UpdatedAt: time.Now(),
	}

	msg := formatQuotesMessage(summary)
	if !strings.Contains(msg, "💵 *COTAÇÃO DE MOEDAS & CRIPTO*") {
		t.Errorf("cabeçalho de cotação não encontrado: %s", msg)
	}
	if !strings.Contains(msg, "Dólar (USD):* R$ 5,72 (🟢 +0.35%)") {
		t.Errorf("cotação do dólar não formatada: %s", msg)
	}
	if !strings.Contains(msg, "Euro (EUR):* R$ 6,24 (🔴 -0.15%)") {
		t.Errorf("cotação do euro não formatada: %s", msg)
	}
	if !strings.Contains(msg, "Bitcoin (BTC):* 🟢 +2.45%") {
		t.Errorf("cotação do Bitcoin não formatada: %s", msg)
	}
}
