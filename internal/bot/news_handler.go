package bot

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"telegram-bot/internal/news"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const newsPageSize = 2

// buildNewsKeyboard constrói o teclado inline com categorias e navegação por páginas.
func buildNewsKeyboard(activeCat string, currentPage, totalPages int) gotgbot.InlineKeyboardMarkup {
	buttons := []struct {
		key   string
		label string
	}{
		{key: "br", label: "🇧🇷 Brasil"},
		{key: "mundo", label: "🌍 Mundo"},
		{key: "eco", label: "📈 Economia"},
		{key: "tech", label: "💻 Tecnologia"},
	}

	var row1 []gotgbot.InlineKeyboardButton
	var row2 []gotgbot.InlineKeyboardButton

	for i, b := range buttons {
		text := b.label
		if b.key == activeCat {
			text = "✅ " + text
		}
		btn := gotgbot.InlineKeyboardButton{
			Text:         text,
			CallbackData: "news_cat:" + b.key,
		}

		if i < 2 {
			row1 = append(row1, btn)
		} else {
			row2 = append(row2, btn)
		}
	}

	var navRow []gotgbot.InlineKeyboardButton
	if currentPage > 1 {
		navRow = append(navRow, gotgbot.InlineKeyboardButton{
			Text:         "⬅️ Anterior",
			CallbackData: fmt.Sprintf("news_page:%s:%d", activeCat, currentPage-1),
		})
	}

	navRow = append(navRow, gotgbot.InlineKeyboardButton{
		Text:         "🔄 Atualizar",
		CallbackData: fmt.Sprintf("news_refresh:%s:%d", activeCat, currentPage),
	})

	if currentPage < totalPages {
		navRow = append(navRow, gotgbot.InlineKeyboardButton{
			Text:         "Próxima ➡️",
			CallbackData: fmt.Sprintf("news_page:%s:%d", activeCat, currentPage+1),
		})
	}

	keyboard := [][]gotgbot.InlineKeyboardButton{row1, row2}
	if len(navRow) > 0 {
		keyboard = append(keyboard, navRow)
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

func cleanSnippet(desc string, maxLen int) string {
	trimmed := strings.TrimSpace(desc)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return trimmed
}

// formatNewsMessage formata as notícias com HTML, blockquotes, badges de tempo e paginação.
func formatNewsMessage(res *news.PageResult) string {
	if res == nil {
		return "Nenhuma notícia encontrada no momento."
	}

	var sb strings.Builder

	emoji := "📰"
	title := "Notícias"
	if res.Category != nil {
		if res.Category.Emoji != "" {
			emoji = res.Category.Emoji
		}
		if res.Category.Title != "" {
			title = res.Category.Title
		}
	}

	sb.WriteString(fmt.Sprintf("<b>%s Manchetes: %s</b>\n", emoji, html.EscapeString(title)))
	if !res.FetchedAt.IsZero() {
		rel := news.FormatRelativeTime(res.FetchedAt)
		if rel != "" {
			sb.WriteString(fmt.Sprintf("<i>Atualizado %s</i>\n", html.EscapeString(rel)))
		}
	}
	sb.WriteString("\n")

	if len(res.Articles) == 0 {
		sb.WriteString("Nenhuma notícia encontrada no momento. Tente novamente mais tarde.")
		return sb.String()
	}

	for _, a := range res.Articles {
		escapedTitle := html.EscapeString(a.Title)
		escapedURL := html.EscapeString(a.URL)
		escapedSource := html.EscapeString(a.Source)

		sb.WriteString(fmt.Sprintf("📌 <a href=\"%s\"><b>%s</b></a>\n", escapedURL, escapedTitle))

		desc := cleanSnippet(a.Description, 180)
		if desc != "" {
			sb.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", html.EscapeString(desc)))
		}

		timeBadge := ""
		if !a.PublishedAt.IsZero() {
			relTime := news.FormatRelativeTime(a.PublishedAt)
			if relTime != "" {
				timeBadge = fmt.Sprintf("  •  ⏱ <i>%s</i>", html.EscapeString(relTime))
			}
		}

		sb.WriteString(fmt.Sprintf("🏷 <i>%s</i>%s\n\n", escapedSource, timeBadge))
	}

	if res.TotalPages > 1 {
		sb.WriteString(fmt.Sprintf("<i>📄 Página %d de %d (%d notícias)</i>\n", res.CurrentPage, res.TotalPages, res.TotalItems))
	}
	sb.WriteString("💡 <i>Selecione uma categoria ou navegue pelas páginas:</i>")

	return sb.String()
}

// handleNoticias trata o comando /noticias.
func handleNoticias(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	slog.Info("comando noticias recebido", "user_id", msg.From.Id, "chat_id", msg.Chat.Id, "args", args)

	categoryKey := "br"
	if len(args) > 0 {
		normalized := news.NormalizeCategory(args[0])
		if normalized != "" {
			categoryKey = normalized
		}
	}

	_, _ = b.SendChatAction(msg.Chat.Id, "typing", nil)

	ctx := context.Background()
	res, err := news.GetPagedNews(ctx, categoryKey, 1, newsPageSize, false)
	if err != nil {
		slog.Error("falha ao obter noticias", "error", err, "categoria", categoryKey, "user_id", msg.From.Id)
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "❌ Não foi possível carregar as notícias no momento. Tente novamente em instantes.", opts)
		return
	}

	text := formatNewsMessage(res)
	markup := buildNewsKeyboard(categoryKey, res.CurrentPage, res.TotalPages)

	opts := &gotgbot.SendMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: markup,
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true, // Desativa pré-visualização para manter mensagem limpa
		},
	}

	_, err = b.SendMessage(msg.Chat.Id, text, opts)
	if err != nil {
		slog.Error("falha ao enviar mensagem de noticias", "error", err, "chat_id", msg.Chat.Id)
	}
}

func parseNewsCallbackData(data string) (action, category string, page int) {
	page = 1
	category = "br"
	parts := strings.Split(data, ":")
	if len(parts) >= 2 {
		action = parts[0]
		category = parts[1]
	}
	if len(parts) >= 3 {
		if p, err := strconv.Atoi(parts[2]); err == nil && p > 0 {
			page = p
		}
	}
	return action, category, page
}

// handleNewsCallback trata navegação, troca de categoria e atualização via botões inline.
func handleNewsCallback(b *gotgbot.Bot, cb *gotgbot.CallbackQuery) {
	action, categoryKey, page := parseNewsCallbackData(cb.Data)
	forceRefresh := action == "news_refresh"

	normalized := news.NormalizeCategory(categoryKey)
	if normalized != "" {
		categoryKey = normalized
	}

	slog.Info("callback de noticias acionado", "user_id", cb.From.Id, "action", action, "categoria", categoryKey, "page", page, "refresh", forceRefresh)

	if forceRefresh {
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Atualizando feed de notícias...",
		})
	}

	ctx := context.Background()
	res, err := news.GetPagedNews(ctx, categoryKey, page, newsPageSize, forceRefresh)
	if err != nil {
		slog.Error("falha ao obter noticias no callback", "error", err, "categoria", categoryKey, "user_id", cb.From.Id)
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Erro ao carregar notícias. Tente novamente.",
			ShowAlert: false,
		})
		return
	}

	if !forceRefresh {
		title := categoryKey
		if res.Category != nil {
			title = res.Category.Title
		}
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("Notícias de %s (Pág. %d/%d)", title, res.CurrentPage, res.TotalPages),
		})
	}

	if cb.Message != nil {
		text := formatNewsMessage(res)
		markup := buildNewsKeyboard(categoryKey, res.CurrentPage, res.TotalPages)

		_, _, err = b.EditMessageText(text, &gotgbot.EditMessageTextOpts{
			ChatId:      cb.Message.GetChat().Id,
			MessageId:   cb.Message.GetMessageId(),
			ParseMode:   "HTML",
			ReplyMarkup: markup,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
		if err != nil {
			slog.Debug("falha ao editar mensagem de noticias", "error", err)
		}
	}
}

