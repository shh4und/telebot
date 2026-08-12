package bot

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// SetBotCommands atualiza a lista nos escopos Global, Privado e Grupos
func SetBotCommands(b *gotgbot.Bot) error {
	commands := []gotgbot.BotCommand{
		{
			Command:     "ping",
			Description: "Verifica se o bot está online",
		},
		{
			Command:     "ajuda",
			Description: "Exibe informações de ajuda",
		},
		{
			Command:     "fig",
			Description: "Converte uma imagem enviada ou respondida em sticker",
		},
		{
			Command:     "addfig",
			Description: "Adiciona uma figurinha ao pacote",
		},
		{
			Command:     "gif",
			Description: "Converte um GIF enviado ou respondido em sticker",
		},
		{
			Command:     "addgif",
			Description: "Adiciona um GIF ao pacote",
		},
		{
			Command:     "pergunta",
			Description: "Faz uma pergunta a IA do bot",
		},
	}

	// 1. Aplica para o escopo Default (fallback geral)
	_, err := b.SetMyCommands(commands, &gotgbot.SetMyCommandsOpts{
		Scope: gotgbot.BotCommandScopeDefault{},
	})
	if err != nil {
		return fmt.Errorf("erro no escopo default: %w", err)
	}

	// 2. Aplica para Chats Privados
	_, err = b.SetMyCommands(commands, &gotgbot.SetMyCommandsOpts{
		Scope: gotgbot.BotCommandScopeAllPrivateChats{},
	})
	if err != nil {
		return fmt.Errorf("erro no escopo private: %w", err)
	}

	// 3. Aplica para Todos os Grupos
	_, err = b.SetMyCommands(commands, &gotgbot.SetMyCommandsOpts{
		Scope: gotgbot.BotCommandScopeAllGroupChats{},
	})
	if err != nil {
		return fmt.Errorf("erro no escopo grupos: %w", err)
	}

	fmt.Println("Comandos atualizados em todos os escopos com sucesso!")
	return nil
}

// ClearBotCommands apaga os comandos registrados em todos os escopos possíveis
func ClearBotCommands(b *gotgbot.Bot) error {
	scopes := []gotgbot.BotCommandScope{
		gotgbot.BotCommandScopeDefault{},
		gotgbot.BotCommandScopeAllPrivateChats{},
		gotgbot.BotCommandScopeAllGroupChats{},
		gotgbot.BotCommandScopeAllChatAdministrators{},
	}

	for _, scope := range scopes {
		_, err := b.DeleteMyCommands(&gotgbot.DeleteMyCommandsOpts{
			Scope: scope,
		})
		if err != nil {
			return fmt.Errorf("falha ao deletar comandos no escopo %T: %w", scope, err)
		}
	}

	fmt.Println("Comandos removidos de todos os escopos!")
	return nil
}
