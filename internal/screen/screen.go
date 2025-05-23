// Package screen provides the screen of the game.
package screen

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
)

// Client is the screen of the game.
type Client struct {
	// Page is the page of the game.
	page *rod.Page
}

// NewClient creates a new client.
func NewClient(page *rod.Page) *Client {
	return &Client{page: page}
}

// Peek takes a snapshot of the screen.
func (c *Client) Peek(ctx context.Context) ([]byte, error) {
	script := `() => {
		const canvas = document.querySelector("canvas");
		if (!canvas) throw new Error("missing canvas");

		const scaled = document.createElement("canvas");
		const ctx = scaled.getContext("2d");
		if (!ctx) throw new Error("missing ctx");

		scaled.width = canvas.width * 0.1;
		scaled.height = canvas.height * 0.1;

		ctx.drawImage(canvas, 0, 0, scaled.width, scaled.height);
		return scaled.toDataURL("image/jpeg", 0.6);
	}`

	dataURL, err := c.page.
		Context(ctx).
		Evaluate(&rod.EvalOptions{JS: script, ByValue: true})
	if err != nil {
		return nil, fmt.Errorf("failed to take snapshot: %w", err)
	}

	imgData, err := extractBase64Image(dataURL.Value.Str())
	if err != nil {
		return nil, fmt.Errorf("failed to extract base64 image: %w", err)
	}

	return imgData, nil
}

// ExtractBase64Image strips the prefix and decodes the base64 image.
func extractBase64Image(dataURL string) ([]byte, error) {
	const prefix = "data:image/jpeg;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		return nil, fmt.Errorf("invalid data URL format: %s", dataURL)
	}

	return base64.StdEncoding.DecodeString(
		strings.TrimPrefix(dataURL, prefix),
	)
}
