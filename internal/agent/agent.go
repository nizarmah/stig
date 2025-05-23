// Package agent provides the agent that plays the game.
package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nizarmah/stig/internal/brain"
	"github.com/nizarmah/stig/internal/controller"
	"github.com/nizarmah/stig/internal/game"
	"github.com/nizarmah/stig/internal/motion"
	"github.com/nizarmah/stig/internal/screen"
)

// ClientConfiguration is the configuration for the agent client.
type ClientConfiguration struct {
	// Controller is the controller client.
	Controller *controller.Client
	// Screen is the screen client.
	Screen *screen.Client
	// Brain is the brain.
	Brain *brain.Brain
	// MotionThreshold is the minimum difference to consider as motion.
	MotionThreshold float64
	// MotionDebug is whether to debug motion detection.
	MotionDebug bool
}

// Client is the agent that plays the game.
type Client struct {
	controller     *controller.Client
	screen         *screen.Client
	brain          *brain.Brain
	motionDetector *motion.Detector
}

// NewClient creates a new client.
func NewClient(cfg ClientConfiguration) *Client {
	// Create motion detector with provided configuration
	motionDetector := motion.NewDetector(cfg.MotionThreshold, cfg.MotionDebug)

	return &Client{
		controller:     cfg.Controller,
		screen:         cfg.Screen,
		brain:          cfg.Brain,
		motionDetector: motionDetector,
	}
}

// SetBrain replaces the current brain with a new one.
func (c *Client) SetBrain(b *brain.Brain) {
	c.brain = b
}

// Run runs the agent.
func (c *Client) Run(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := c.doRun(ctx); err != nil {
				return err
			}
		}
	}
}

// doRun runs the agent.
func (c *Client) doRun(ctx context.Context) error {
	img, err := c.screen.Peek(ctx)
	if err != nil {
		return fmt.Errorf("failed to peek screen: %w", err)
	}

	// Process frame for motion detection
	motion, err := c.motionDetector.ProcessFrame(img)
	if err != nil {
		log.Printf("Warning: motion detection failed: %v", err)
		// Continue anyway, motion detection is not critical
	}

	// Get motion statistics
	totalDistance, frameCount, timeSinceMotion := c.motionDetector.GetStats()

	// Log motion information
	log.Printf("Frame %d - Motion: %.4f, Total distance: %.2f, Time since motion: %.1fs",
		frameCount, motion, totalDistance, timeSinceMotion.Seconds())

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
		return fmt.Errorf("failed to send action: %w", err)
	}

	return nil
}

// GetMotionStats returns the current motion statistics.
func (c *Client) GetMotionStats() (totalDistance float64, timeSinceMotion time.Duration) {
	return c.motionDetector.GetTotalDistance(), c.motionDetector.GetTimeSinceLastMotion()
}

// ResetMotionDetector resets the motion detector for a new run.
func (c *Client) ResetMotionDetector() {
	c.motionDetector.Reset()
}
