package bot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"telegram-bot/internal/stickers"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func handlePhotoToSticker(b *gotgbot.Bot, msg *gotgbot.Message, photos []gotgbot.PhotoSize) {
	if len(photos) == 0 {
		return
	}

	// Pega o maior tamanho disponível da imagem
	photo := photos[len(photos)-1]

	file, err := b.GetFile(photo.FileId, nil)
	if err != nil {
		replyError(b, msg, "Erro ao obter dados da imagem no servidor do Telegram.")
		return
	}

	fileURL := file.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		replyError(b, msg, "Erro ao realizar o download da imagem.")
		return
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		replyError(b, msg, "Erro ao ler buffer da imagem.")
		return
	}

	webpData, err := stickers.ProcessImageToWebp(imgBytes)
	if err != nil {
		replyError(b, msg, "Erro ao converter imagem para sticker WEBP.")
		return
	}

	stickerFile := gotgbot.InputFileByReader("sticker.webp", bytes.NewReader(webpData))
	opts := &gotgbot.SendStickerOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId, // Responde à mensagem do comando
		},
	}

	_, err = b.SendSticker(msg.Chat.Id, stickerFile, opts)
	if err != nil {
		fmt.Printf("Erro ao enviar sticker: %v\n", err)
	}
}

func replyError(b *gotgbot.Bot, msg *gotgbot.Message, text string) {
	opts := &gotgbot.SendMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: msg.MessageId,
		},
	}
	b.SendMessage(msg.Chat.Id, text, opts)
}
