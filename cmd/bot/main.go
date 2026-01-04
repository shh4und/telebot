package main

import (
	"fmt"
	"log"
	"net/http"
	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func main() {

	// Get token from the environment variable
	token := config.Envs.BotToken
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	// Create bot from environment value.
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: http.Client{Timeout: 45 * time.Second},
			DefaultRequestOpts: &gotgbot.RequestOpts{
				Timeout: 45 * time.Second,      // Customise the default request timeout here
				APIURL:  gotgbot.DefaultAPIURL, // As well as the Default API URL here (in case of using local bot API servers)
			},
		},
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}
	fmt.Printf("Bot online: @%s\n", b.User.Username)
	// Create updater and dispatcher.
	var offset int64
	for {
		// Long Polling de 30 segundos
		updates, err := b.GetUpdates(&gotgbot.GetUpdatesOpts{
			Offset:  offset,
			Timeout: 30,
		})
		if err != nil {
			log.Printf("Error at GetUpdates: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, upd := range updates {
			// Chamada concorrente: O loop continua livre para o próximo update
			go bot.Dispatch(b, upd)

			// Atualiza o offset para confirmar o recebimento
			offset = upd.UpdateId + 1
		}
	}
}
