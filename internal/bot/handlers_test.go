package bot

import (
	"strings"
	"telegram-bot/internal/news"
	"telegram-bot/internal/quotes"
	"testing"
	"time"
)

func TestBuildNewsKeyboard(t *testing.T) {
	markup := buildNewsKeyboard("br", 1, 3)
	if len(markup.InlineKeyboard) != 3 {
		t.Fatalf("esperava 3 linhas de botões (2 categorias + 1 navegação), obteve %d", len(markup.InlineKeyboard))
	}

	btnBR := markup.InlineKeyboard[0][0]
	if !strings.HasPrefix(btnBR.Text, "✅") {
		t.Errorf("esperava indicador selecionado no botão Brasil, obteve %s", btnBR.Text)
	}

	btnMundo := markup.InlineKeyboard[0][1]
	if strings.HasPrefix(btnMundo.Text, "✅") {
		t.Errorf("não esperava indicador selecionado no botão Mundo, obteve %s", btnMundo.Text)
	}

	navRow := markup.InlineKeyboard[2]
	// Na página 1 de 3, deve ter "🔄 Atualizar" e "Próxima ➡️"
	if len(navRow) != 2 {
		t.Fatalf("esperava 2 botões na linha de navegação, obteve %d", len(navRow))
	}
	if navRow[0].Text != "🔄 Atualizar" {
		t.Errorf("esperava botão Atualizar, obteve %s", navRow[0].Text)
	}
	if navRow[1].Text != "Próxima ➡️" {
		t.Errorf("esperava botão Próxima, obteve %s", navRow[1].Text)
	}
}

func TestFormatNewsMessage(t *testing.T) {
	pageResult := &news.PageResult{
		Articles: []news.Article{
			{
				Title:       "Nova tecnologia é lançada",
				Description: "Detalhes sobre a novidade no setor tecnológico",
				URL:         "https://example.com/tech",
				Source:      "Canaltech",
				PublishedAt: time.Now().Add(-15 * time.Minute),
			},
		},
		Category: &news.CategoryInfo{
			Title: "Tecnologia",
			Emoji: "💻",
		},
		CurrentPage: 1,
		TotalPages:  3,
		TotalItems:  6,
		FetchedAt:   time.Now().Add(-2 * time.Minute),
	}

	msg := formatNewsMessage(pageResult)
	if !strings.Contains(msg, "<b>💻 Manchetes: Tecnologia</b>") {
		t.Errorf("título não encontrado na mensagem: %s", msg)
	}
	if !strings.Contains(msg, `<a href="https://example.com/tech"><b>Nova tecnologia é lançada</b></a>`) {
		t.Errorf("link do artigo não formatado corretamente: %s", msg)
	}
	if !strings.Contains(msg, "<blockquote>Detalhes sobre a novidade no setor tecnológico</blockquote>") {
		t.Errorf("descrição em blockquote não encontrada: %s", msg)
	}
	if !strings.Contains(msg, "Canaltech") {
		t.Errorf("fonte não encontrada na mensagem: %s", msg)
	}
	if !strings.Contains(msg, "há 15 min") {
		t.Errorf("tempo relativo não encontrado na mensagem: %s", msg)
	}
	if !strings.Contains(msg, "Página 1 de 3") {
		t.Errorf("indicador de página não encontrado: %s", msg)
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
