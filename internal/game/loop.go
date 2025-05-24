package game

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RunInGameLoop runs a function in the game loop.
func (c *Client) RunInGameLoop(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	// Create interval in milliseconds based on the fps.
	interval := time.Second / time.Duration(c.fps)
	if c.debug {
		log.Println(fmt.Sprintf("running game loop with interval: %v", interval))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := fn(ctx); err != nil {
				return err
			}
		}
	}
}

// WaitForFinish waits for the game to finish and signals when it does.
func (c *Client) WaitForFinish(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			// Tie the DOM query to the same context so it respects timeouts.
			if el, _ := c.Page.Context(ctx).Element(`div[data-your-time="true"]`); el != nil {
				return nil
			}
		}
	}
}
