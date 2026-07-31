package stickers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

// ProcessImageToWebp redimensiona a imagem mantendo o aspect ratio (max 512x512) e converte para WEBP.
func ProcessImageToWebp(inputData []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(inputData))
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar imagem: %w", err)
	}

	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newWidth, newHeight uint
	if width > height {
		newWidth = 512
		newHeight = 0
	} else {
		newWidth = 0
		newHeight = 512
	}

	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Bilinear)

	var buf bytes.Buffer
	err = webp.Encode(&buf, resizedImg, &webp.Options{Lossless: false, Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar para webp: %w", err)
	}

	return buf.Bytes(), nil
}

func ProcessGifToWebm(inputData []byte) ([]byte, error) {
	// Cria diretório temporário isolado para o processamento
	tmpDir, err := os.MkdirTemp("", "sticker_conv_*")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar dir temp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input.gif")
	outputPath := filepath.Join(tmpDir, "output.webm")

	if err := os.WriteFile(inputPath, inputData, 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo temp: %w", err)
	}

	// Comando FFmpeg com as restrições estritas do Telegram:
	// - t 3: limita a 3 segundos
	// - vf: redimensiona mantendo proporção até 512px e força fps em 30
	// - c:v libvpx-vp9: codec VP9
	// - an: remove áudio
	// - b:v 256k: limita bitrate para ficar abaixo de 256KB
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inputPath,
		"-t", "3",
		"-vf", "scale='if(gt(iw,ih),512,-1)':'if(gt(ih,iw),512,-1)',fps=30",
		"-c:v", "libvpx-vp9",
		"-crf", "30",
		"-b:v", "256k",
		"-an",
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("falha no ffmpeg (%w): %s", err, stderr.String())
	}

	webmBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo final: %w", err)
	}

	return webmBytes, nil
}
