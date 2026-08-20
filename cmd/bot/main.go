package main

import (
	"log/slog"
	"net/http"
	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/internal/logger"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func main() {
	// Inicializa o logger estruturado
	logger.Init(config.Envs.LogLevel, config.Envs.LogFormat)

	// Get token from the environment variable
	token := config.Envs.BotToken
	if token == "" {
		slog.Error("variável de ambiente BOT_TK não informada ou vazia")
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
		slog.Error("falha ao instanciar bot", "error", err)
		panic("failed to create new bot: " + err.Error())
	}

	if err := bot.SetBotCommands(b); err != nil {
		slog.Error("falha ao registrar comandos do bot", "error", err)
		panic("failed to set bot commands: " + err.Error())
	}

	slog.Info("bot iniciado com sucesso", "username", b.User.Username, "log_level", config.Envs.LogLevel, "log_format", config.Envs.LogFormat)

	// Create updater and dispatcher.
	var offset int64
	for {
		// Long Polling de 30 segundos
		updates, err := b.GetUpdates(&gotgbot.GetUpdatesOpts{
			Offset:  offset,
			Timeout: 30,
		})
		if err != nil {
			slog.Error("erro ao buscar updates (long polling)", "error", err)
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
