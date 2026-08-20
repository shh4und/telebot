package bot

import (
	"fmt"
	"log/slog"
	"strings"
	"telegram-bot/internal/ai"
	"telegram-bot/internal/telegraph"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func Dispatch(b *gotgbot.Bot, upd gotgbot.Update) {
	// Trata cliques em botões inline (Callback Queries)
	if upd.CallbackQuery != nil {
		handleCallbackQuery(b, upd.CallbackQuery)
		return
	}

	if upd.Message == nil {
		return
	}

	msg := upd.Message
	// 1. Processamento quando o comando está na legenda (Caption) de uma foto enviada diretamente
	if msg.ReplyToMessage == nil && len(msg.Photo) > 0 && msg.Caption != "" {
		if isCommand(msg.Caption, "/fig") || isCommand(msg.Caption, "/sticker") {
			handlePhotoToSticker(b, msg, msg.Photo, false)
			return
		} else if isCommand(msg.Caption, "/addfig") || isCommand(msg.Caption, "/addsticker") {
			handlePhotoToSticker(b, msg, msg.Photo, true)
			return
		}
	}
	if msg != nil && msg.Animation != nil && msg.Animation.FileSize > 0 && msg.Caption != "" {
		if isCommand(msg.Caption, "/gif") {
			handleGifToSticker(b, msg, msg.Animation, false)
			return

		} else if isCommand(msg.Caption, "/addgif") {
			handleGifToSticker(b, msg, msg.Animation, true)
			return
		}
	}
	// 2. Processamento de comandos de texto
	if msg.Text == "" {
		return
	}

	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}

	cmd := args[0]

	switch {
	case isCommand(cmd, "/ping"):
		handlePing(b, msg)

	case isCommand(cmd, "/ajuda"):
		handleAjuda(b, msg)

	case isCommand(cmd, "/modelo") || isCommand(cmd, "/modelos"):
		handleModelo(b, msg)

	case isCommand(cmd, "/noticias") || isCommand(cmd, "/noticia"):
		handleNoticias(b, msg, args[1:])

	case isCommand(cmd, "/moedas") || isCommand(cmd, "/moeda") || isCommand(cmd, "/cotacao") || isCommand(cmd, "/cotacoes"):
		handleMoedas(b, msg)

	case isCommand(cmd, "/pergunta"):
		handlePergunta(b, msg, args[1:])

	case isCommand(cmd, "/fig") || isCommand(cmd, "/sticker"):
		// Caso 2: Comando /fig enviado como RESPOSTA a uma foto
		if msg.ReplyToMessage != nil && len(msg.ReplyToMessage.Photo) > 0 {
			handlePhotoToSticker(b, msg, msg.ReplyToMessage.Photo, false)
		}
	case isCommand(cmd, "/addfig") || isCommand(cmd, "/addsticker"):
		// Caso 3: Comando /addfig enviado como RESPOSTA a uma foto
		if msg.ReplyToMessage != nil && len(msg.ReplyToMessage.Photo) > 0 {
			handlePhotoToSticker(b, msg, msg.ReplyToMessage.Photo, true)
		}

	case isCommand(cmd, "/gif"):
		// Caso 3: Comando /addfig enviado como RESPOSTA a uma foto
		if msg.ReplyToMessage != nil && msg.ReplyToMessage.Animation != nil && msg.ReplyToMessage.Animation.FileSize > 0 {
			handleGifToSticker(b, msg, msg.ReplyToMessage.Animation, false)
		}
	case isCommand(cmd, "/addgif"):
		// Caso 4: Comando /addgif enviado como RESPOSTA a uma gif
		if msg.ReplyToMessage != nil && msg.ReplyToMessage.Animation != nil && msg.ReplyToMessage.Animation.FileSize > 0 {
			handleGifToSticker(b, msg, msg.ReplyToMessage.Animation, true)
		}
	default:
		return
	}
}

// Helper para validar comandos ignorando username do bot (ex: /fig@MeuBot)
func isCommand(input, cmd string) bool {
	if strings.HasPrefix(input, cmd) {
		// Garante que /figura não dê match com /fig
		rest := input[len(cmd):]
		return rest == "" || strings.HasPrefix(rest, "@") || strings.HasPrefix(rest, " ")
	}
	return false
}

func handlePing(b *gotgbot.Bot, msg *gotgbot.Message) {
	resposta := "pong! to funcionando!"
	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}
	slog.Info("ping", "message_id", msg.MessageId)
	b.SendMessage(msg.Chat.Id, resposta, opts)
}

func handleAjuda(b *gotgbot.Bot, msg *gotgbot.Message) {
	resposta := "Tem ajuda aqui não, pae"
	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}
	slog.Info("ajuda", "message_id", msg.MessageId)
	b.SendMessage(msg.Chat.Id, resposta, opts)
}

func buildModelKeyboard(models []ai.OllamaModelInfo, currentModel string) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton
	for _, m := range models {
		label := m.Name
		if m.Details.ParameterSize != "" {
			label = fmt.Sprintf("%s (%s)", m.Name, m.Details.ParameterSize)
		}
		if m.Name == currentModel {
			label = "✅ " + label
		}
		btn := gotgbot.InlineKeyboardButton{
			Text:         label,
			CallbackData: "set_model:" + m.Name,
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{btn})
	}
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func handleModelo(b *gotgbot.Bot, msg *gotgbot.Message) {
	slog.Info("handleModelo", "user_id", msg.From.Id)
	models, err := ai.GetInstalledModels()
	if err != nil {
		slog.Error("failed to get installed models", "error", err)
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "❌ Não foi possível consultar os modelos do Ollama no momento.", opts)
		return
	}

	if len(models) == 0 {
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "⚠️ Nenhum modelo encontrado no servidor Ollama.", opts)
		return
	}

	currentModel := GetUserModel(msg.From.Id)
	markup := buildModelKeyboard(models, currentModel)
	text := fmt.Sprintf("🤖 *Escolha o modelo de IA para suas consultas:*\n\nModelo atual: `%s`", currentModel)

	opts := &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: markup,
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}
	b.SendMessage(msg.Chat.Id, text, opts)
}

func handleCallbackQuery(b *gotgbot.Bot, cb *gotgbot.CallbackQuery) {
	if strings.HasPrefix(cb.Data, "set_model:") {
		modelName := strings.TrimPrefix(cb.Data, "set_model:")
		SetUserModel(cb.From.Id, modelName)
		slog.Info("model changed for user", "user_id", cb.From.Id, "model", modelName)

		_, _ = b.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("Modelo alterado para %s!", modelName),
		})

		if cb.Message != nil {
			models, err := ai.GetInstalledModels()
			if err == nil {
				markup := buildModelKeyboard(models, modelName)
				text := fmt.Sprintf("🤖 *Escolha o modelo de IA para suas consultas:*\n\nModelo atual: `%s`", modelName)

				_, _, _ = b.EditMessageText(text, &gotgbot.EditMessageTextOpts{
					ChatId:      cb.Message.GetChat().Id,
					MessageId:   cb.Message.GetMessageId(),
					ParseMode:   "Markdown",
					ReplyMarkup: markup,
				})
			}
		}
		return
	}

	if strings.HasPrefix(cb.Data, "news_") {
		handleNewsCallback(b, cb)
		return
	}

	if cb.Data == "refresh_quotes" {
		handleQuotesCallback(b, cb)
		return
	}
}

const shortMessageThreshold = 300

// shouldPublishToTelegraph checks if the response is long enough or contains complex formatting (like code blocks) to warrant a Telegraph Instant View page.
func shouldPublishToTelegraph(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len([]rune(trimmed)) > shortMessageThreshold || strings.Contains(trimmed, "```") {
		return true
	}
	return false
}

func sendDirectResponse(b *gotgbot.Bot, chatID int64, replyToID int64, text string) {
	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: replyToID,
		},
	}
	_, err := b.SendMessage(chatID, text, opts)
	if err != nil {
		// Fallback without parse mode if Markdown rendering fails
		opts.ParseMode = ""
		b.SendMessage(chatID, text, opts)
	}
}

func handlePergunta(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	slog.Info("pergunta", "message_id", msg.MessageId, "args", args)
	if len(args) == 0 {
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, "Por favor, envie uma pergunta após o comando. Exemplo: `/pergunta o que é Go?`", opts)
		return
	}

	query := strings.Join(args, " ")
	_, _ = b.SendChatAction(msg.Chat.Id, "typing", nil)

	selectedModel := GetUserModel(msg.From.Id)
	aiResponse, err := ai.AskOllama(selectedModel, query)
	if err != nil {
		slog.Error("error at handling AI request", "error", err, "model", selectedModel)
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, fmt.Sprintf("Desculpe, ocorreu um erro ao consultar o modelo `%s`.", selectedModel), opts)
		return
	}

	// Respostas curtas são enviadas diretamente no chat
	if !shouldPublishToTelegraph(aiResponse) {
		sendDirectResponse(b, msg.Chat.Id, msg.MessageId, aiResponse)
		return
	}

	// Respostas longas (> 300 caracteres ou com código) são publicadas via Telegraph
	title := query
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	pageURL, err := telegraph.PublishMarkdown(title, aiResponse)
	if err != nil {
		slog.Error("error publishing to telegraph, falling back to direct message", "error", err)
		sendDirectResponse(b, msg.Chat.Id, msg.MessageId, aiResponse)
		return
	}

	respostaMsg := fmt.Sprintf("📄 *Resposta da IA (Instant View):*\n%s", pageURL)
	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}
	b.SendMessage(msg.Chat.Id, respostaMsg, opts)
}

