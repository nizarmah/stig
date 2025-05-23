// Package screen provides the screen of the game.
package screen

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type ClientConfiguration struct {
	// Debug is whether to save the snapshot to a file.
	Debug bool
	// Page is the page of the game.
	Page *rod.Page
}

// Client is the screen of the game.
type Client struct {
	// Debug is whether to save the snapshot to a file.
	debug bool
	// Page is the page of the game.
	page *rod.Page
}

// NewClient creates a new client.
func NewClient(cfg ClientConfiguration) *Client {
	return &Client{
		debug: cfg.Debug,
		page:  cfg.Page,
	}
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

	if c.debug {
		if err := saveSnapshot(imgData); err != nil {
			log.Printf("failed to save snapshot: %v", err)
		}
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

func saveSnapshot(imgData []byte) error {
	// Create the directory "debug/screen" if not exists.
	dir := "debug/screen"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Check the frames in the directory.
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	// Keep only the last 60 frames.
	if len(files) > 60 {
		// Delete the oldest frame.
		if err := os.Remove(filepath.Join(dir, files[0].Name())); err != nil {
			return fmt.Errorf("failed to remove file %s: %v", files[0].Name(), err)
		}
	}

	// Save the snapshot to the directory.
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("frame_%d.jpg", time.Now().UnixMilli())), imgData, 0644); err != nil {
		return fmt.Errorf("failed to save snapshot to %s: %v", filepath.Join(dir, fmt.Sprintf("frame_%d.jpg", time.Now().UnixMilli())), err)
	}

	return nil
}
