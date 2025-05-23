// Package screen provides the screen of the game.
package screen

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ClientConfiguration is the configuration for the screen client.
type ClientConfiguration struct {
	// Debug is whether to save the snapshot to a file.
	Debug bool
	// Page is the page of the game.
	Page *rod.Page
	// Resolution is the resolution of the screen (0 to 100).
	Resolution int
}

// Client is the screen of the game.
type Client struct {
	// Debug is whether to save the snapshot to a file.
	debug bool
	// Page is the page of the game.
	page *rod.Page
	// Resolution is the resolution of the screen (0 to 100).
	resolution int
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
	imageData, err := c.page.
		Context(ctx).
		Screenshot(true, &proto.PageCaptureScreenshot{
			Format:                proto.PageCaptureScreenshotFormatJpeg,
			Quality:               &[]int{c.resolution}[0],
			OptimizeForSpeed:      true,
			FromSurface:           true,
			CaptureBeyondViewport: false,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to take snapshot: %w", err)
	}

	if c.debug {
		if err := saveSnapshot(imageData); err != nil {
			log.Printf("failed to save snapshot: %v", err)
		}
	}

	return imageData, nil
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
