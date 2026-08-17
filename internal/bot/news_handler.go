package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"telegram-bot/internal/news"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// buildNewsKeyboard constrói o teclado inline com as categorias de notícias.
func buildNewsKeyboard(activeCat string) gotgbot.InlineKeyboardMarkup {
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

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{row1, row2},
	}
}

// formatNewsMessage formata as notícias com limite e links.
func formatNewsMessage(articles []news.Article, info *news.CategoryInfo) string {
	var sb strings.Builder

	emoji := "📰"
	title := "Notícias"
	if info != nil {
		if info.Emoji != "" {
			emoji = info.Emoji
		}
		if info.Title != "" {
			title = info.Title
		}
	}

	sb.WriteString(fmt.Sprintf("%s *Manchetes: %s*\n\n", emoji, title))

	if len(articles) == 0 {
		sb.WriteString("Nenhuma notícia encontrada no momento. Tente novamente mais tarde.")
		return sb.String()
	}

	for i, a := range articles {
		// Formatação: 1. [Título da Notícia](URL) (Fonte)
		// Escapando caracteres para Markdown simples
		cleanTitle := strings.ReplaceAll(a.Title, "[", "(")
		cleanTitle = strings.ReplaceAll(cleanTitle, "]", ")")
		cleanTitle = strings.ReplaceAll(cleanTitle, "*", "")

		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, cleanTitle, a.URL))
		sb.WriteString(fmt.Sprintf("   _Fonte: %s_\n\n", a.Source))
	}

	sb.WriteString("💡 _Selecione uma categoria abaixo para alternar:_")
	return sb.String()
}

// handleNoticias trata o comando /noticias.
func handleNoticias(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	slog.Info("handleNoticias", "user_id", msg.From.Id, "args", args)

	categoryKey := "br"
	if len(args) > 0 {
		normalized := news.NormalizeCategory(args[0])
		if normalized != "" {
			categoryKey = normalized
		}
	}

	_, _ = b.SendChatAction(msg.Chat.Id, "typing", nil)

	ctx := context.Background()
	articles, info, err := news.GetNews(ctx, categoryKey)
	if err != nil {
		slog.Error("erro ao buscar notícias", "categoria", categoryKey, "error", err)
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "❌ Não foi possível carregar as notícias no momento. Tente novamente em instantes.", opts)
		return
	}

	text := formatNewsMessage(articles, info)
	markup := buildNewsKeyboard(categoryKey)

	opts := &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
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
		slog.Error("falha ao enviar mensagem de notícias", "error", err)
	}
}

// handleNewsCallback trata a troca de categoria via botões inline.
func handleNewsCallback(b *gotgbot.Bot, cb *gotgbot.CallbackQuery) {
	categoryKey := strings.TrimPrefix(cb.Data, "news_cat:")
	slog.Info("news category changed", "user_id", cb.From.Id, "categoria", categoryKey)

	ctx := context.Background()
	articles, info, err := news.GetNews(ctx, categoryKey)
	if err != nil {
		slog.Error("erro ao buscar notícias no callback", "categoria", categoryKey, "error", err)
		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Erro ao carregar notícias. Tente novamente.",
			ShowAlert: false,
		})
		return
	}

	_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
		Text: fmt.Sprintf("Carregando notícias de %s...", info.Title),
	})

	if cb.Message != nil {
		text := formatNewsMessage(articles, info)
		markup := buildNewsKeyboard(categoryKey)

		_, _, err = b.EditMessageText(text, &gotgbot.EditMessageTextOpts{
			ChatId:      cb.Message.GetChat().Id,
			MessageId:   cb.Message.GetMessageId(),
			ParseMode:   "Markdown",
			ReplyMarkup: markup,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		})
		if err != nil {
			slog.Debug("falha ao editar mensagem de notícias", "error", err)
		}
	}
}
