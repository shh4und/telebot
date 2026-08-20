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
	slog.Info("processando foto para sticker", "user_id", msg.From.Id, "chat_id", msg.Chat.Id, "add_to_pack", addToPack, "file_id", photo.FileId)

	file, err := b.GetFile(photo.FileId, nil)
	if err != nil {
		slog.Error("falha ao obter metadados da imagem no telegram", "error", err, "file_id", photo.FileId)
		replyError(b, msg, "Erro ao obter dados da imagem no servidor do Telegram.")
		return
	}

	fileURL := file.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		slog.Error("falha ao baixar imagem do telegram", "error", err, "file_url", fileURL)
		replyError(b, msg, "Erro ao realizar o download da imagem.")
		return
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("falha ao ler buffer da imagem", "error", err)
		replyError(b, msg, "Erro ao ler buffer da imagem.")
		return
	}

	webpData, err := stickers.ProcessImageToWebp(imgBytes)
	if err != nil {
		slog.Error("falha ao converter imagem para webp", "error", err, "bytes_len", len(imgBytes))
		replyError(b, msg, "Erro ao converter imagem para sticker WEBP.")
		return
	}

	if addToPack {
		setName, err := stickers.AddToCentralPack(b, webpData, "image/webp")
		if err != nil {
			slog.Error("falha ao adicionar sticker ao pacote central", "error", err, "bytes_len", len(webpData))
			replyError(b, msg, fmt.Sprintf("Erro ao atualizar pacote de figurinhas: %v", err))
			return
		}

		slog.Info("sticker adicionado ao pacote central com sucesso", "user_id", msg.From.Id, "set_name", setName)
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
			slog.Error("falha ao enviar sticker estático", "error", err, "chat_id", msg.Chat.Id)
		} else {
			slog.Info("sticker enviado com sucesso", "user_id", msg.From.Id, "chat_id", msg.Chat.Id)
		}
	}
}

func handleGifToSticker(b *gotgbot.Bot, msg *gotgbot.Message, gif *gotgbot.Animation, addToPack bool) {
	slog.Info("processando gif para sticker animado", "user_id", msg.From.Id, "chat_id", msg.Chat.Id, "add_to_pack", addToPack, "file_id", gif.FileId)

	file, err := b.GetFile(gif.FileId, nil)
	if err != nil {
		slog.Error("falha ao obter metadados do gif no telegram", "error", err, "file_id", gif.FileId)
		replyError(b, msg, "Erro ao obter dados da gif no servidor do Telegram.")
		return
	}

	fileURL := file.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		slog.Error("falha ao baixar gif do telegram", "error", err, "file_url", fileURL)
		replyError(b, msg, "Erro ao realizar o download da gif.")
		return
	}
	defer resp.Body.Close()

	gifBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("falha ao ler buffer do gif", "error", err)
		replyError(b, msg, "Erro ao ler buffer da gif.")
		return
	}

	// 1. Processa o GIF/Vídeo via FFmpeg
	webmData, err := stickers.ProcessGifToWebm(gifBytes)
	if err != nil {
		slog.Error("falha ao processar gif para webm via ffmpeg", "error", err)
		replyError(b, msg, "Erro ao processar GIF/Vídeo para sticker.")
		return
	}

	if addToPack && gif != nil && gif.MimeType != "" {
		setName, err := stickers.AddToCentralPack(b, webmData, gif.MimeType)
		if err != nil {
			slog.Error("falha ao adicionar gif ao pacote central", "error", err, "mime_type", gif.MimeType)
			replyError(b, msg, fmt.Sprintf("Erro ao atualizar pacote de figurinhas: %v", err))
			return
		}

		slog.Info("sticker de gif adicionado ao pacote central com sucesso", "user_id", msg.From.Id, "set_name", setName)
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

		_, err = b.SendSticker(msg.Chat.Id, stickerFile, opts)
		if err != nil {
			slog.Error("falha ao enviar sticker de vídeo", "error", err, "chat_id", msg.Chat.Id)
		} else {
			slog.Info("sticker de vídeo enviado com sucesso", "user_id", msg.From.Id, "chat_id", msg.Chat.Id)
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
