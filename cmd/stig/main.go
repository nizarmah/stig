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
		Debug: e.DebugScreen,
		Page:  page,
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
	agentClient := agent.NewClient(
		controllerClient,
		screenClient,
		baseBrain,
	)

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
	bestTime := math.MaxFloat64

	for {
		// Mutate the best brain and use it for this training run.
		candidate := bestBrain.Mutate(0.02)
		agentClient.SetBrain(candidate)

		// Run training.
		raceMs, didFinish, err := runTraining(
			ctx,
			menuClient,
			agentClient,
			timeout,
		)
		if err != nil {
			return err
		}

		if didFinish {
			return nil
		}

		if raceMs < bestTime {
			bestTime = raceMs
			bestBrain = candidate
			log.Printf("🎉 new best time: %.0f ms", bestTime)

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
) (float64, bool, error) {
	// Training context.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Reset the game.
	if err := menuClient.ResetGame(); err != nil {
		return 0, false, fmt.Errorf("failed to reset game: %w", err)
	}

	// Start driving.
	go agentClient.Run(ctx, 100*time.Millisecond)

	// Wait for the game to finish.
	if err := menuClient.WaitForFinish(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return float64(timeout), false, nil
		}

		return 0, false, fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	// Get the final time.
	raceTime, err := menuClient.GetReplayTime()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get replay time: %w", err)
	}

	raceMs, err := parseRaceTime(raceTime)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse race time %q: %v", raceTime, err)
	}

	return raceMs, true, nil
}

// parseRaceTime converts a time string in the format mm:ss:SSS to total milliseconds.
func parseRaceTime(s string) (float64, error) {
	var min, sec, ms int
	if _, err := fmt.Sscanf(s, "%02d:%02d:%03d", &min, &sec, &ms); err != nil {
		return 0, err
	}
	return float64(min*60000 + sec*1000 + ms), nil
}
