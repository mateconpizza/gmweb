// Package qr contains functions for generating QR-Codes.
package qr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// QRCode represents a QR-Code.
type QRCode struct {
	QR   *qrcode.QRCode
	From string
	Size int
}

// Generate generates a QR-Code from a given string.
func (q *QRCode) Generate() error {
	var err error

	q.QR, err = qrcode.New(q.From, qrcode.Highest)
	if err != nil {
		return fmt.Errorf("generating qr-code: %w", err)
	}

	return nil
}

func (q *QRCode) ImageBase64() (string, error) {
	var buf bytes.Buffer

	if err := q.QR.Write(q.Size, &buf); err != nil {
		return "", fmt.Errorf("writing qr-code: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (q *QRCode) ImagePNG(bg, fg string) (bytes.Buffer, error) {
	var buf bytes.Buffer

	q.QR.BackgroundColor = darkHex(bg)
	q.QR.ForegroundColor = lightHex(fg)

	if err := q.QR.Write(q.Size, &buf); err != nil {
		return buf, fmt.Errorf("writing qr-code: %w", err)
	}

	return buf, nil
}

// New creates a new QR-Code.
func New(s string, size int) *QRCode {
	return &QRCode{
		From: s,
		Size: size,
	}
}

func hexToRGBA(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		log.Fatalf("Invalid hex color: %s", hex)
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func darkHex(hex string) color.RGBA { return hexToRGBA(hex) }

func lightHex(hex string) color.RGBA { return hexToRGBA(hex) }
