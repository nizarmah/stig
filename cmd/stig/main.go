// Package main is the entry point for stig.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-rod/rod"

	"github.com/nizarmah/stig/internal/agent"
	"github.com/nizarmah/stig/internal/brain"
	"github.com/nizarmah/stig/internal/controller"
	"github.com/nizarmah/stig/internal/env"
	"github.com/nizarmah/stig/internal/menu"
	"github.com/nizarmah/stig/internal/screen"
)

func main() {
	// Environment.
	e, err := env.NewEnv()
	if err != nil {
		log.Fatalf("failed to create env: %v", err)
	}

	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Connect to the browser.
	browser, err := connectToBrowser(ctx, e.BrowserWSURL)
	if err != nil {
		log.Fatal(err)
	}
	// defer browser.Close()

	// Open the game.
	page, err := menu.NewGame(ctx, browser, e.GameURL, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	// Create the menu client.
	menuClient := menu.NewClient(page)

	// Create the controller client.
	controllerClient := controller.NewClient(page)

	// Create the screen client.
	screenClient := screen.NewClient(screen.ClientConfiguration{
		Debug:      e.ScreenDebug,
		Page:       page,
		Resolution: e.ScreenResolution,
	})

	// Brain persistence file.
	const brainPath = "brain.gob"

	// Try loading an existing brain; otherwise create a new one.
	var baseBrain *brain.Brain

	if _, err := os.Stat(brainPath); err == nil {
		b, err := brain.Load(brainPath)
		if err != nil {
			log.Fatalf("failed to load brain: %v", err)
		}
		log.Printf("loaded brain from %s", brainPath)
		baseBrain = b
	} else {
		// Create the initial brain with an approximate input size.
		const (
			approxInputSize = 80 * 60 // matches the 0.1 scale in screen.Peek
			hiddenSize      = 64
		)
		baseBrain = brain.NewBrain(approxInputSize, hiddenSize)
		if err := baseBrain.Save(brainPath); err != nil {
			log.Printf("failed to save initial brain: %v", err)
		}
	}

	// Create the agent client.
	agentClient := agent.NewClient(agent.ClientConfiguration{
		Controller:      controllerClient,
		Screen:          screenClient,
		Brain:           baseBrain,
		MotionThreshold: e.MotionThreshold,
		MotionDebug:     e.MotionDebug,
	})

	startTraining(
		ctx,
		menuClient,
		agentClient,
		baseBrain,
		brainPath,
		e.LapTimeout,
	)
}

func connectToBrowser(
	ctx context.Context,
	browserWSURL string,
) (*rod.Browser, error) {
	browser := rod.New().
		Context(ctx).
		ControlURL(browserWSURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	return browser, nil
}

func startTraining(
	ctx context.Context,
	menuClient *menu.Client,
	agentClient *agent.Client,
	baseBrain *brain.Brain,
	brainPath string,
	timeout time.Duration,
) error {
	bestBrain := baseBrain
	bestScore := -math.MaxFloat64

	for {
		// Mutate the best brain and use it for this training run.
		candidate := bestBrain.Mutate(0.02)
		agentClient.SetBrain(candidate)

		// Run training.
		score, err := runTraining(
			ctx,
			menuClient,
			agentClient,
			timeout,
		)
		if err != nil {
			return err
		}

		if score > bestScore {
			bestScore = score
			bestBrain = candidate
			log.Printf("🎉 new best score: %.2f", bestScore)

			if err := bestBrain.Save(brainPath); err != nil {
				return fmt.Errorf("failed to save best brain: %v", err)
			}
		}
	}
}

func runTraining(
	parentCtx context.Context,
	menuClient *menu.Client,
	agentClient *agent.Client,
	timeout time.Duration,
) (float64, error) {
	// Training context.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Reset the game.
	if err := menuClient.ResetGame(); err != nil {
		return 0, fmt.Errorf("failed to reset game: %w", err)
	}

	// Wait 3 seconds for the game to start.
	time.Sleep(3 * time.Second)

	// Reset motion detector for new run
	agentClient.ResetMotionDetector()

	// Start monitoring for stuck situations
	stuckCtx, stuckCancel := context.WithCancel(ctx)
	defer stuckCancel()

	go monitorStuckState(stuckCtx, stuckCancel, agentClient)

	// Start driving.
	go agentClient.Run(stuckCtx, 100*time.Millisecond)

	// Wait for the game to finish or get stuck.
	err := menuClient.WaitForFinish(stuckCtx)

	// Check if we got stuck (no motion timeout)
	if errors.Is(err, context.Canceled) {
		totalDistance, timeSinceMotion := agentClient.GetMotionStats()
		log.Printf("❌ Run canceled - stuck for %.1fs, total distance: %.2f",
			timeSinceMotion.Seconds(), totalDistance)
		// Return negative score based on distance traveled
		return totalDistance - 1000, nil // Penalty for getting stuck
	}

	// Check if we hit the overall timeout
	if errors.Is(err, context.DeadlineExceeded) {
		totalDistance, _ := agentClient.GetMotionStats()
		log.Printf("⏱️ Timeout reached, total distance: %.2f", totalDistance)
		// Return score based on distance traveled
		return totalDistance, nil
	}

	if err != nil {
		return 0, fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	// Game finished successfully - get the final time.
	raceTime, err := menuClient.GetReplayTime()
	if err != nil {
		return 0, fmt.Errorf("failed to get replay time: %w", err)
	}

	raceMs, err := parseRaceTime(raceTime)
	if err != nil {
		return 0, fmt.Errorf("failed to parse race time %q: %v", raceTime, err)
	}

	log.Printf("✅ Finished race in %.0f ms", raceMs)

	// For finished races, use negative time as score (lower time = higher score)
	return -raceMs, nil
}

// monitorStuckState monitors the agent's motion and cancels if stuck for too long.
func monitorStuckState(ctx context.Context, cancel context.CancelFunc, agentClient *agent.Client) {
	const stuckTimeout = 5 * time.Second
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, timeSinceMotion := agentClient.GetMotionStats()
			if timeSinceMotion > stuckTimeout {
				log.Printf("🛑 No motion detected for %.1fs - canceling run", timeSinceMotion.Seconds())
				cancel()
				return
			}
		}
	}
}

// parseRaceTime converts a time string in the format mm:ss:SSS to total milliseconds.
func parseRaceTime(s string) (float64, error) {
	var min, sec, ms int
	if _, err := fmt.Sscanf(s, "%02d:%02d:%03d", &min, &sec, &ms); err != nil {
		return 0, err
	}
	return float64(min*60000 + sec*1000 + ms), nil
}
