package stickers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

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

	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = webp.Encode(&buf, resizedImg, &webp.Options{Lossless: false, Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar para webp: %w", err)
	}

	return buf.Bytes(), nil
}
