package barcode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

const (
	minCode128Width  = 500
	minCode128Height = 100
	minQRSize        = 250
)

func calcScaleDims(intrinsicW, intrinsicH, targetW, targetH int) (int, int) {
	scaleX := 1
	if intrinsicW > 0 {
		scaleX = (targetW + intrinsicW - 1) / intrinsicW
		if scaleX < 1 {
			scaleX = 1
		}
	}

	scaleY := 1
	if intrinsicH > 0 {
		scaleY = (targetH + intrinsicH - 1) / intrinsicH
		if scaleY < 1 {
			scaleY = 1
		}
	}

	return intrinsicW * scaleX, intrinsicH * scaleY
}

// GenerateCode128Base64 creates a Code128 barcode png image encoded in Base64 Data URI
func GenerateCode128Base64(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("empty content for barcode")
	}

	bc, err := code128.Encode(content)
	if err != nil {
		return "", fmt.Errorf("failed to encode code128: %w", err)
	}

	intrinsicW := bc.Bounds().Dx()
	intrinsicH := bc.Bounds().Dy()
	targetW, targetH := calcScaleDims(intrinsicW, intrinsicH, minCode128Width, minCode128Height)

	scaled, err := barcode.Scale(bc, targetW, targetH)
	if err != nil {
		return "", fmt.Errorf("failed to scale barcode: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("failed to encode png: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

// GenerateQRCodeBase64 creates a QR Code png image encoded in Base64 Data URI
func GenerateQRCodeBase64(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("empty content for qr code")
	}

	qrCode, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("failed to encode qr code: %w", err)
	}

	intrinsicW := qrCode.Bounds().Dx()
	intrinsicH := qrCode.Bounds().Dy()
	targetW, targetH := calcScaleDims(intrinsicW, intrinsicH, minQRSize, minQRSize)

	scaled, err := barcode.Scale(qrCode, targetW, targetH)
	if err != nil {
		return "", fmt.Errorf("failed to scale qr code: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("failed to encode png: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}
