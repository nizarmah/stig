// Package agent provides the agent that plays the game.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/nizarmah/stig/internal/brain"
	"github.com/nizarmah/stig/internal/controller"
	"github.com/nizarmah/stig/internal/game"
	"github.com/nizarmah/stig/internal/screen"
)

// Client is the agent that plays the game.
type Client struct {
	controller *controller.Client
	screen     *screen.Client
	brain      *brain.Brain
}

// NewClient creates a new client.
func NewClient(
	controller *controller.Client,
	screen *screen.Client,
	brain *brain.Brain,
) *Client {
	return &Client{
		controller: controller,
		screen:     screen,
		brain:      brain,
	}
}

// SetBrain replaces the current brain with a new one.
func (c *Client) SetBrain(b *brain.Brain) {
	c.brain = b
}

// Run runs the agent.
func (c *Client) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.doRun(ctx)
		}
	}
}

// doRun runs the agent.
func (c *Client) doRun(ctx context.Context) error {
	img, err := c.screen.Peek(ctx)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to peek screen: %w", err))
		return fmt.Errorf("failed to peek screen: %w", err)
	}

	// Decide the next action using the brain.
	throttle, steering, err := c.brain.Predict(img)
	if err != nil {
		return fmt.Errorf("brain prediction failed: %w", err)
	}

	action := game.Action{
		Throttle: throttle,
		Steering: steering,
	}

	if err := c.controller.Apply(action); err != nil {
		fmt.Println("failed to send action")
		return fmt.Errorf("failed to send action: %w", err)
	}

	return nil
}

// generateNextAction is now handled by the brain, but we keep an empty stub to
// avoid breaking other parts of the codebase that might depend on it. It is
// deprecated and will be removed in the future.
// Deprecated: use brain.Predict instead.
func generateNextAction(_ []byte) game.Action { return game.Action{} }
