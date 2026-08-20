package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"telegram-bot/internal/quotes"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// buildQuotesKeyboard cria o botão inline para atualizar cotações.
func buildQuotesKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "🔄 Atualizar Cotações",
					CallbackData: "refresh_quotes",
				},
			},
		},
	}
}

// formatQuotesMessage monta a resposta visual para o Telegram.
func formatQuotesMessage(summary *quotes.MarketSummary) string {
	var sb strings.Builder

	sb.WriteString("💵 *COTAÇÃO DE MOEDAS & CRIPTO*\n\n")

	// Câmbio Comercial
	sb.WriteString("🇧🇷 *Câmbio Comercial (BRL):*\n")
	if summary.Dollar.Bid > 0 {
		var suffix string
		if summary.Dollar.PctChange != 0 {
			suffix = fmt.Sprintf(" (%s)", quotes.VariationIndicator(summary.Dollar.PctChange))
		}
		sb.WriteString(fmt.Sprintf("• *Dólar (USD):* R$ %s%s\n",
			quotes.FormatBRL(summary.Dollar.Bid),
			suffix,
		))
		if summary.Dollar.High > 0 && summary.Dollar.Low > 0 && summary.Dollar.High != summary.Dollar.Low {
			sb.WriteString(fmt.Sprintf("  ↳ _Mín: R$ %s | Máx: R$ %s_\n",
				quotes.FormatBRL(summary.Dollar.Low),
				quotes.FormatBRL(summary.Dollar.High),
			))
		}
	} else {
		sb.WriteString("• *Dólar (USD):* _Indisponível no momento_\n")
	}

	if summary.Euro.Bid > 0 {
		var suffix string
		if summary.Euro.PctChange != 0 {
			suffix = fmt.Sprintf(" (%s)", quotes.VariationIndicator(summary.Euro.PctChange))
		}
		sb.WriteString(fmt.Sprintf("• *Euro (EUR):* R$ %s%s\n",
			quotes.FormatBRL(summary.Euro.Bid),
			suffix,
		))
		if summary.Euro.High > 0 && summary.Euro.Low > 0 && summary.Euro.High != summary.Euro.Low {
			sb.WriteString(fmt.Sprintf("  ↳ _Mín: R$ %s | Máx: R$ %s_\n",
				quotes.FormatBRL(summary.Euro.Low),
				quotes.FormatBRL(summary.Euro.High),
			))
		}
	} else {
		sb.WriteString("• *Euro (EUR):* _Indisponível no momento_\n")
	}

	sb.WriteString("\n🪙 *Criptomoedas (Tempo Real):*\n")
	if len(summary.Cryptos) == 0 {
		sb.WriteString("_Cotações de criptomoedas indisponíveis no momento_\n")
	} else {
		for _, c := range summary.Cryptos {
			sb.WriteString(fmt.Sprintf("• *%s (%s):* %s\n",
				c.Name,
				c.Symbol,
				quotes.VariationIndicator(c.Change24h),
			))
			sb.WriteString(fmt.Sprintf("  ↳ R$ %s | US$ %s\n",
				quotes.FormatBRL(c.PriceBRL),
				quotes.FormatUSD(c.PriceUSD),
			))
		}
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local
	}
	formattedTime := summary.UpdatedAt.In(loc).Format("02/01/2006 15:04:05")
	sb.WriteString(fmt.Sprintf("\n🕒 _Atualizado em: %s_", formattedTime))

	return sb.String()
}

// handleMoedas trata os comandos /moedas e /cotacao.
func handleMoedas(b *gotgbot.Bot, msg *gotgbot.Message) {
	slog.Info("comando moedas recebido", "user_id", msg.From.Id, "chat_id", msg.Chat.Id)

	_, _ = b.SendChatAction(msg.Chat.Id, "typing", nil)

	ctx := context.Background()
	summary, err := quotes.GetMarketSummary(ctx, false)
	if err != nil {
		slog.Error("falha ao buscar cotacoes", "error", err, "user_id", msg.From.Id)
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "❌ Não foi possível carregar as cotações no momento. Tente novamente em instantes.", opts)
		return
	}

	text := formatQuotesMessage(summary)
	markup := buildQuotesKeyboard()

	opts := &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: markup,
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}

	_, err = b.SendMessage(msg.Chat.Id, text, opts)
	if err != nil {
		slog.Error("falha ao enviar mensagem de cotacoes", "error", err, "chat_id", msg.Chat.Id)
	}
}

var quotesRefreshCooldown = NewCooldownTracker(15 * time.Second)

// handleQuotesCallback trata o clique no botão "Atualizar Cotações".
func handleQuotesCallback(b *gotgbot.Bot, cb *gotgbot.CallbackQuery) {
	userID := cb.From.Id
	slog.Info("callback de atualizacao de cotacoes acionado", "user_id", userID)

	if allowed, remaining := quotesRefreshCooldown.Allow(userID); !allowed {
		secs := int(remaining.Seconds()) + 1
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text:      fmt.Sprintf("⏳ Aguarde %ds para atualizar novamente.", secs),
			ShowAlert: false,
		})
		return
	}

	ctx := context.Background()
	summary, err := quotes.GetMarketSummary(ctx, true) // forçar refresh
	if err != nil {
		slog.Error("falha ao atualizar cotacoes no callback", "error", err, "user_id", userID)
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Erro ao atualizar cotações. Tente novamente.",
			ShowAlert: false,
		})
		return
	}

	_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Cotações atualizadas!",
	})

	if cb.Message != nil {
		text := formatQuotesMessage(summary)
		markup := buildQuotesKeyboard()

		_, _, err = b.EditMessageText(text, &gotgbot.EditMessageTextOpts{
			ChatId:      cb.Message.GetChat().Id,
			MessageId:   cb.Message.GetMessageId(),
			ParseMode:   "Markdown",
			ReplyMarkup: markup,
		})
		if err != nil {
			slog.Debug("falha ao editar mensagem de cotacoes (conteudo identico)", "error", err)
		}
	}
}
