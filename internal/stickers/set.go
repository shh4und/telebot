package stickers

import (
	"bytes"
	"fmt"
	"log/slog"
	"telegram-bot/internal/config"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// Seu ID fixo de proprietário do pacote central
var OwnerUserID int64 = config.Envs.UserID

// AddToCentralPack adiciona a figurinha ao pacote global mantido por você.
func AddToCentralPack(b *gotgbot.Bot, webpData []byte) (string, error) {
	// Nome fixo para o pacote central do bot
	setName := fmt.Sprintf("pack00_by_%s", b.User.Username)
	setTitle := setName

	// Primeiro faz upload do arquivo para obter um file_id
	uploadedFile, err := b.UploadStickerFile(OwnerUserID, gotgbot.InputFileByReader("sticker.webp", bytes.NewReader(webpData)), "static", nil)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer upload da figurinha: %w", err)
	}
	slog.Info("upload realizado", "fileId", uploadedFile.FileId)

	inputSticker := gotgbot.InputSticker{
		Sticker:   uploadedFile.FileId,
		Format:    "static",
		EmojiList: []string{"🖼️"},
	}

	// 1. Tenta adicionar ao pacote central existente (usando OwnerUserID)
	_, err = b.AddStickerToSet(OwnerUserID, setName, inputSticker, nil)
	if err == nil {
		slog.Info("adicionado ao pacote central existente", "setName", setName, "fileId", uploadedFile.FileId)
		return setName, nil
	}
	slog.Warn("pacote central não encontrado, criando novo", "setName", setName, "fileId", uploadedFile.FileId)
	// 2. Se o pacote central não existir ainda, cria utilizando seu OwnerUserID
	_, createErr := b.CreateNewStickerSet(OwnerUserID, setName, setTitle, []gotgbot.InputSticker{inputSticker}, &gotgbot.CreateNewStickerSetOpts{
		StickerType: "regular",
	})
	if createErr != nil {
		return "", fmt.Errorf("erro ao criar pacote central: %w", createErr)
	}
	slog.Info("pacote central criado", "setName", setName, "fileId", uploadedFile.FileId)

	return setName, nil
}
