package bot

import (
	"log/slog"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func Dispatch(b *gotgbot.Bot, upd gotgbot.Update) {
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

// func handleMimDiga(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
// 	query := ""
// 	if len(args) > 1 {
// 		query = strings.Join(args[1:], " ")
// 	} else {
// 		return
// 	}
// 	// // Menção ao usuário (Username ou Firstname se não houver username)
// 	// userName := msg.From.FirstName
// 	// if msg.From.Username != "" {
// 	// 	userName = "@" + msg.From.Username
// 	// }

// 	aiResponse, err := ai.AskOllama("", query)
// 	if err != nil {
// 		fmt.Printf("error at handling AI request: %v", err)
// 		return
// 	}

// 	opts := &gotgbot.SendMessageOpts{
// 		ReplyParameters: &gotgbot.ReplyParameters{
// 			MessageId: msg.MessageId, // O ID da mensagem do usuário que enviou o comando
// 		},
// 	}
// 	// Enviar a res
// 	b.SendMessage(msg.Chat.Id, aiResponse, opts)

// }
