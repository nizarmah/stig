package game

import (
	"context"
	"fmt"

	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// StartGame starts the game by clicking the "Start" button.
func (c *Client) StartGame(ctx context.Context) error {
	// Search for the "Start" button.
	startButton, err := c.Page.Context(ctx).ElementX(`//span[text()="Start"]`)
	if err != nil {
		return fmt.Errorf("failed to find start button: %w", err)
	}

	// Click the "Start" button.
	if err := startButton.Context(ctx).Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click on start button: %w", err)
	}

	return nil
}

// ResetGame resets the game before the current one ends.
func (c *Client) ResetGame(ctx context.Context) error {
	// Press "Escape" to go back to the main menu.
	pressEscape := c.Page.Context(ctx).KeyActions().Type(input.Escape)
	if err := pressEscape.Do(); err != nil {
		return fmt.Errorf("failed to press escape: %w", err)
	}

	// Start the game.
	if err := c.StartGame(ctx); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

// GetReplayTime retrieves the final game time shown on the Replay screen.
func (c *Client) GetReplayTime(ctx context.Context) (string, error) {
	// Target the deepest child with the time value (has aria-label)
	el, err := c.Page.Context(ctx).Element(`div[data-your-time="true"] div[aria-label]`)
	if err != nil {
		return "", fmt.Errorf("failed to find time element: %w", err)
	}

	// aria-label has the full time (e.g., "01:49:214")
	timeStr, err := el.Context(ctx).Attribute("aria-label")
	if err != nil || timeStr == nil {
		return "", fmt.Errorf("failed to extract time from aria-label")
	}

	return *timeStr, nil
}
