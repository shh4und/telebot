package stickers

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"telegram-bot/internal/config"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// Seu ID fixo de proprietário do pacote central
var OwnerUserID int64 = config.Envs.UserID

func newInputSticker(b *gotgbot.Bot, webpData []byte, mimeType string) (gotgbot.InputSticker, *gotgbot.File) {
	var format string
	var ext string
	if strings.HasPrefix(mimeType, "image/gif") ||
		strings.HasPrefix(mimeType, "video/mp4") ||
		strings.HasPrefix(mimeType, "video/webm") {
		format = "video"
		ext = "webm"
	} else {
		format = "static"
		ext = "webp"
	}

	// Primeiro faz upload do arquivo para obter um file_id
	uploadedFile, err := b.UploadStickerFile(OwnerUserID, gotgbot.InputFileByReader("sticker."+ext, bytes.NewReader(webpData)), format, nil)
	if err != nil || uploadedFile == nil {
		slog.Warn("upload failed", "error", err, "webpData size", len(webpData))
		return gotgbot.InputSticker{}, nil
	}

	slog.Info("upload realizado", "fileId", uploadedFile.FileId)

	inputSticker := gotgbot.InputSticker{
		Sticker:   uploadedFile.FileId,
		Format:    format,
		EmojiList: []string{"🖼️"},
	}
	return inputSticker, uploadedFile
}

// AddToCentralPack adiciona a figurinha ao pacote global mantido por você.
func AddToCentralPack(b *gotgbot.Bot, webpData []byte, mimeType string) (string, error) {
	// Nome fixo para o pacote central do bot
	setName := fmt.Sprintf("pack00_by_%s", b.User.Username)
	setTitle := setName

	// Primeiro faz upload do arquivo para obter um file_id
	inputSticker, uploadedFile := newInputSticker(b, webpData, mimeType)
	if uploadedFile == nil {
		return "", nil
	}

	slog.Info("upload realizado", "fileId", uploadedFile.FileId)

	// 1. Tenta adicionar ao pacote central existente (usando OwnerUserID)
	_, err := b.AddStickerToSet(OwnerUserID, setName, inputSticker, nil)
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
