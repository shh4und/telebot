package bot

import (
	"fmt"
	"log"
	"strings"
	"telegram-bot/internal/ai"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// Dispatch analisa o update e decide o que fazer
func Dispatch(b *gotgbot.Bot, upd gotgbot.Update) {
	// Filtramos apenas mensagens de texto
	if upd.Message == nil || upd.Message.Text == "" {
		return
	}

	msg := upd.Message
	text := msg.Text

	args := strings.Fields(text)
	if len(args) == 0 {
		return
	}

	cmd := args[0]

	switch {
	case strings.HasPrefix(cmd, "/ping"):
		handlePing(b, msg)
	case strings.HasPrefix(cmd, "/ajuda"):
		handleAjuda(b, msg)
	case strings.HasPrefix(cmd, "/mimdiga"):
		log.Println("/mimdiga")
		handleMimDiga(b, msg, args)
	default:
		return
	}

}

func handlePing(b *gotgbot.Bot, msg *gotgbot.Message) {

	resposta := "pong! to funcionando!"

	//  opções de envio definindo o ID da mensagem que será respondida
	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId, // O ID da mensagem do usuário que enviou o comando
		},
	}

	b.SendMessage(msg.Chat.Id, resposta, opts)
}

func handleAjuda(b *gotgbot.Bot, msg *gotgbot.Message) {

	resposta := "Tem ajuda aqui não, pae"

	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId, // O ID da mensagem do usuário que enviou o comando
		},
	}
	// Enviar a resposta
	b.SendMessage(msg.Chat.Id, resposta, opts)

}

func handleMimDiga(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	query := ""
	if len(args) > 1 {
		query = strings.Join(args[1:], " ")
	} else {
		return
	}
	// // Menção ao usuário (Username ou Firstname se não houver username)
	// userName := msg.From.FirstName
	// if msg.From.Username != "" {
	// 	userName = "@" + msg.From.Username
	// }

	aiResponse, err := ai.AskOllama("", query)
	if err != nil {
		fmt.Printf("error at handling AI request: %v", err)
		return
	}

	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId, // O ID da mensagem do usuário que enviou o comando
		},
	}
	// Enviar a res
	b.SendMessage(msg.Chat.Id, aiResponse, opts)

}
