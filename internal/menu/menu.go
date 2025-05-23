// Package menu provides functions to interact with the menu.
package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Manager manages the game menu.
type Manager struct {
	page *rod.Page
}

// NewManager creates a new manager.
func NewManager(
	ctx context.Context,
	browser *rod.Browser,
	gameURL string,
	timeout time.Duration,
) (*Manager, error) {
	page, err := openInitialMenu(ctx, browser, gameURL, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	return &Manager{page: page}, nil
}

// StartGame starts the game by clicking the "Start" button.
func (m *Manager) StartGame() error {
	// Search for the "Start" button.
	startButton, err := m.page.ElementX(`//span[text()="Start"]`)
	if err != nil {
		return fmt.Errorf("failed to find start button: %w", err)
	}

	// Click on the "Start" button.
	if err := startButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click on start button: %w", err)
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

// ReplayGame starts a new game after the current one ended.
func (m *Manager) ReplayGame() error {
	btn, err := m.page.Element(`button[data-results-button="true"]`)
	if err != nil {
		return fmt.Errorf("failed to find replay button: %w", err)
	}

	if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click replay: %w", err)
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

// WaitForFinish blocks until the game ends or the context is canceled.
func (m *Manager) WaitForFinish(
	ctx context.Context,
	timeout time.Duration,
) error {
	finishCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return m.doWaitForFinish(finishCtx)
}

// doWaitForFinish waits for the game to finish.
func (m *Manager) doWaitForFinish(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
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

// Close closes the manager and the game page.
func (m *Manager) Close() error {
	if err := m.page.Close(); err != nil {
		return fmt.Errorf("failed to close page: %w", err)
	}

	return nil
}

// openInitialMenu opens the game initial menu.
func openInitialMenu(
	ctx context.Context,
	browser *rod.Browser,
	gameURL string,
	timeout time.Duration,
) (*rod.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Open the game page.
	page, err := browser.Page(proto.TargetCreateTarget{URL: gameURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Wait for the page to load.
	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page to load: %w", err)
	}

	return page, nil
}
