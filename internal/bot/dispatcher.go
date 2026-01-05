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
		handlePing(b, msg, args)
	case strings.HasPrefix(cmd, "/ajuda"):
		handleAjuda(b, msg)
	case strings.HasPrefix(cmd, "/mimdiga"):
		log.Println("/mimdiga")
		handleMimDiga(b, msg, args)
	default:
		return
	}

}

func handlePing(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	// Dividir a string por espaços (Fields remove espaços extras automaticamente)

	palavra := ""
	if len(args) > 1 {
		palavra = args[1]
	}
	// Menção ao usuário (Username ou Firstname se não houver username)
	userName := msg.From.FirstName
	if msg.From.Username != "" {
		userName = "@" + msg.From.Username
	}

	resposta := fmt.Sprintf("pong %s %s to funcionando!", userName, palavra)

	// Enviar a resposta
	b.SendMessage(msg.Chat.Id, resposta, nil)
	// if err != nil {
	// 	fmt.Printf("Erro ao enviar mensagem: %v\n", err)
	// }
}

func handleAjuda(b *gotgbot.Bot, msg *gotgbot.Message) {

	resposta := "Tem ajuda aqui não, pae"

	// Enviar a resposta
	b.SendMessage(msg.Chat.Id, resposta, nil)

}

func handleMimDiga(b *gotgbot.Bot, msg *gotgbot.Message, args []string) {
	query := ""
	if len(args) > 1 {
		query = strings.Join(args[1:], " ")
	} else {
		return
	}
	// Menção ao usuário (Username ou Firstname se não houver username)
	userName := msg.From.FirstName
	if msg.From.Username != "" {
		userName = "@" + msg.From.Username
	}

	aiResponse, err := ai.AskOllama("", query)
	if err != nil {
		fmt.Printf("error at handling AI request: %v", err)
		return
	}

	res := fmt.Sprintf("Respondeno a @%s,\n%s", userName, aiResponse)
	// Enviar a res
	b.SendMessage(msg.Chat.Id, res, nil)
	// if err != nil {
	// 	fmt.Printf("Erro ao enviar mensagem: %v\n", err)
	// }
}
