// Package menu provides functions to interact with the game menu.
package menu

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
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
	// Press "Enter" to start the game.
	pressEnter := c.page.KeyActions().Type(input.Enter)
	if err := pressEnter.Do(); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

// ResetGame resets the game before the current one ends.
func (c *Client) ResetGame() error {
	// Press "Delete" to instantly reset the game.
	pressDelete := c.page.KeyActions().Type(input.Delete)
	if err := pressDelete.Do(); err != nil {
		return fmt.Errorf("failed to reset game: %w", err)
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
			if el, _ := c.page.Element(`div[data-your-time="true"]`); el != nil {
				return nil
			}
		}
	}
}
