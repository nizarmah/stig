// Package menu provides functions to interact with the menu.
package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// Manager manages the game menu.
type Manager struct {
	// Page is the page of the game.
	page *rod.Page
}

// NewManager creates a new manager.
func NewManager(page *rod.Page) (*Manager, error) {
	return &Manager{page: page}, nil
}

// StartGame starts the game by clicking the "Start" button.
func (m *Manager) StartGame() error {
	// Press "Enter" to start the game.
	pressEnter := m.page.KeyActions().Type(input.Enter)
	if err := pressEnter.Do(); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

// ResetGame resets the game before the current one ends.
func (m *Manager) ResetGame() error {
	// Press "Escape" to open the initial menu again.
	pressEscape := m.page.KeyActions().Type(input.Escape)
	if err := pressEscape.Do(); err != nil {
		return fmt.Errorf("failed to press escape: %w", err)
	}

	// Start the game.
	if err := m.StartGame(); err != nil {
		return fmt.Errorf("failed to reset game: %w", err)
	}

	return nil
}

// GetReplayTime retrieves the final game time shown on the Replay screen.
func (m *Manager) GetReplayTime() (string, error) {
	// Target the deepest child with the time value (has aria-label)
	el, err := m.page.Element(`div[data-your-time="true"] div[aria-label]`)
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
func (m *Manager) WaitForFinish(
	ctx context.Context,
	interval time.Duration,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // timeout, cancel, or interrupt
		case <-ticker.C:
			if el, _ := m.page.Element(`div[data-your-time="true"]`); el != nil {
				return nil // game finished
			}
		}
	}
}
