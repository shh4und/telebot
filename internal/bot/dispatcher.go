package bot

import (
	"fmt"
	"strings"

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

	// Rota: /ping <palavra>
	if strings.HasPrefix(text, "/ping") {
		handlePing(b, msg)
	}
}

func handlePing(b *gotgbot.Bot, msg *gotgbot.Message) {
	// Dividir a string por espaços (Fields remove espaços extras automaticamente)
	args := strings.Fields(msg.Text)

	palavra := ""
	if len(args) > 1 {
		palavra = args[1]
	} else {
		palavra = "(sem palavra)"
	}

	// Menção ao usuário (Username ou Firstname se não houver username)
	userName := msg.From.FirstName
	if msg.From.Username != "" {
		userName = "@" + msg.From.Username
	}

	resposta := fmt.Sprintf("pong %s %s to funcionando!", userName, palavra)

	// Enviar a resposta
	_, err := b.SendMessage(msg.Chat.Id, resposta, nil)
	if err != nil {
		fmt.Printf("Erro ao enviar mensagem: %v\n", err)
	}
}
