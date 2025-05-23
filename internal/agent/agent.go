// Package agent provides the agent that plays the game.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/nizarmah/stig/internal/controller"
	"github.com/nizarmah/stig/internal/game"
	"github.com/nizarmah/stig/internal/screen"
)

// Client is the agent that plays the game.
type Client struct {
	controller *controller.Client
	screen     *screen.Client
}

// NewClient creates a new client.
func NewClient(
	controller *controller.Client,
	screen *screen.Client,
) *Client {
	return &Client{
		controller: controller,
		screen:     screen,
	}
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

	action := generateNextAction(img)

	if err := c.controller.Apply(action); err != nil {
		fmt.Println("failed to send action")
		return fmt.Errorf("failed to send action: %w", err)
	}

	return nil
}

func generateNextAction(_ []byte) game.Action {
	return game.Action{
		Throttle: "accelerate",
		Steering: "",
	}
}
