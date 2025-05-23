// Package menu provides functions to interact with the game menu.
package menu

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Client controls the game menu.
type Client struct {
	// Page is the page of the game.
	page *rod.Page
}

// NewClient creates a new manager.
func NewClient(page *rod.Page) *Client {
	return &Client{page: page}
}

// StartGame starts the game by clicking the "Start" button.
func (c *Client) StartGame() error {
	// Search for the "Start" button.
	startButton, err := c.page.ElementX(`//span[text()="Start"]`)
	if err != nil {
		return fmt.Errorf("failed to find start button: %w", err)
	}

	// Click the "Start" button.
	if err := startButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click on start button: %w", err)
	}

	return nil
}

// ResetGame resets the game before the current one ends.
func (c *Client) ResetGame() error {
	// Press "Escape" to go back to the main menu.
	pressEscape := c.page.KeyActions().Type(input.Escape)
	if err := pressEscape.Do(); err != nil {
		return fmt.Errorf("failed to press escape: %w", err)
	}

	// Start the game.
	if err := c.StartGame(); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

// GetReplayTime retrieves the final game time shown on the Replay screen.
func (c *Client) GetReplayTime() (string, error) {
	// Target the deepest child with the time value (has aria-label)
	el, err := c.page.Element(`div[data-your-time="true"] div[aria-label]`)
	if err != nil {
		return "", fmt.Errorf("failed to find time element: %w", err)
	}

	// aria-label has the full time (e.g., "01:49:214")
	timeStr, err := el.Attribute("aria-label")
	if err != nil || timeStr == nil {
		return "", fmt.Errorf("failed to extract time from aria-label")
	}

	return *timeStr, nil
}

// WaitForFinish waits for the game to finish and signals when it does.
func (c *Client) WaitForFinish(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			// Tie the DOM query to the same context so it respects timeouts.
			if el, _ := c.page.Context(ctx).Element(`div[data-your-time="true"]`); el != nil {
				return nil
			}
		}
	}
}
