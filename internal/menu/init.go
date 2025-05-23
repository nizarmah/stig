package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// NewGame opens the game page and waits for it to be ready.
func NewGame(ctx context.Context, browser *rod.Browser, gameURL string, timeout time.Duration) (*rod.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Open the game page.
	page, err := browser.
		Page(proto.TargetCreateTarget{URL: gameURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Wait for the page to load.
	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page to load: %w", err)
	}

	// Search for the "Start" button.
	startButton, err := page.ElementX(`//span[text()="Start"]`)
	if err != nil {
		return nil, fmt.Errorf("failed to find start button: %w", err)
	}

	// Wait until the "Start" button is interactable.
	if _, err := startButton.Context(ctx).WaitInteractable(); err != nil {
		return nil, fmt.Errorf("failed to wait for start button to be interactable: %w", err)
	}

	return page, nil
}
