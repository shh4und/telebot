package bot

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"telegram-bot/internal/stickers"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func handlePhotoToSticker(b *gotgbot.Bot, msg *gotgbot.Message, photos []gotgbot.PhotoSize, addToPack bool) {
	if len(photos) == 0 {
		return
	}

	// Pega o maior tamanho disponível da imagem
	photo := photos[len(photos)-1]

	file, err := b.GetFile(photo.FileId, nil)
	if err != nil {
		slog.Error("b.GetFile", "error", err, "photo", photo.FileId)
		replyError(b, msg, "Erro ao obter dados da imagem no servidor do Telegram.")
		return
	}

	fileURL := file.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		slog.Error("http.Get", "error", err, "fileURL", fileURL)
		replyError(b, msg, "Erro ao realizar o download da imagem.")
		return
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("io.ReadAll", "error", err, "imgBytes", len(imgBytes))
		replyError(b, msg, "Erro ao ler buffer da imagem.")
		return
	}

	webpData, err := stickers.ProcessImageToWebp(imgBytes)
	if err != nil {
		slog.Error("stickers.ProcessImageToWebp", "error", err, "imgBytes", len(imgBytes))
		replyError(b, msg, "Erro ao converter imagem para sticker WEBP.")
		return
	}
	if addToPack {

		setName, err := stickers.AddToCentralPack(b, webpData, "image/webp")
		if err != nil {
			slog.Error("stickers.AddToCentralPack", "error", err, "webpData", len(webpData))
			replyError(b, msg, fmt.Sprintf("Erro ao atualizar pacote de figurinhas: %v", err))
			return
		}

		packURL := fmt.Sprintf("https://t.me/addstickers/%s", setName)
		resposta := fmt.Sprintf("Figurinha adicionada ao pacote!\nAcesse seu pacote aqui: %s", packURL)

		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, resposta, opts)
	} else {
		stickerFile := gotgbot.InputFileByReader("sticker.webp", bytes.NewReader(webpData))
		opts := &gotgbot.SendStickerOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}

		_, err = b.SendSticker(msg.Chat.Id, stickerFile, opts)
		if err != nil {
			slog.Error("b.SendSticker", "error", err, "stickerFile", stickerFile)
			fmt.Printf("Erro ao enviar sticker: %v\n", err)
		}
	}

}

func handleGifToSticker(b *gotgbot.Bot, msg *gotgbot.Message, gif *gotgbot.Animation, addToPack bool) {

	// Pega o maior tamanho disponível da gif

	file, err := b.GetFile(gif.FileId, nil)
	if err != nil {
		slog.Error("b.GetFile", "error", err, "gif", gif.FileId)
		replyError(b, msg, "Erro ao obter dados da gif no servidor do Telegram.")
		return
	}

	fileURL := file.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		slog.Error("http.Get", "error", err, "fileURL", fileURL)
		replyError(b, msg, "Erro ao realizar o download da gif.")
		return
	}
	defer resp.Body.Close()
	slog.Info("Download GIF", "fileURL", fileURL, "gif", gif.FileId, "resp.StatusCode", resp.StatusCode)

	gifBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("io.ReadAll", "error", err, "gifBytes", len(gifBytes))
		replyError(b, msg, "Erro ao ler buffer da gif.")
		return
	}
	slog.Info("Convert GIF em Bytes", "gifBytes size", len(gifBytes))
	// 1. Processa o GIF/Vídeo via FFmpeg
	webmData, err := stickers.ProcessGifToWebm(gifBytes)
	if err != nil {
		replyError(b, msg, "Erro ao processar GIF/Vídeo para sticker.")
		return
	}
	slog.Info("Process GIF Bytes to Webm", "webmData size", len(webmData))

	slog.Info("gif.Animation", "gif.FileId", len(gif.FileId), "gif.MimeType", gif.MimeType)
	if addToPack && gif != nil && gif.MimeType != "" {
		setName, err := stickers.AddToCentralPack(b, webmData, gif.MimeType)
		if err != nil {
			slog.Error("stickers.AddToCentralPack", "error", err, "webmData", len(webmData), "mimeType", gif.MimeType)
			replyError(b, msg, fmt.Sprintf("Erro ao atualizar pacote de figurinhas: %v", err))
			return
		}

		packURL := fmt.Sprintf("https://t.me/addstickers/%s", setName)
		resposta := fmt.Sprintf("Figurinha adicionada ao pacote!\nAcesse seu pacote aqui: %s", packURL)

		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		b.SendMessage(msg.Chat.Id, resposta, opts)

	} else {

		// 2. Envia especificando o nome do arquivo com extensão .webm
		stickerFile := gotgbot.InputFileByReader("sticker.webm", bytes.NewReader(webmData))
		opts := &gotgbot.SendStickerOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: msg.MessageId,
			},
		}
		slog.Info("Send Sticker", "stickerFile", stickerFile, "opts", opts)

		_, err = b.SendSticker(msg.Chat.Id, stickerFile, opts)
		if err != nil {
			fmt.Printf("Erro ao enviar sticker de vídeo: %v\n", err)
		}
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
