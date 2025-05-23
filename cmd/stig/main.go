// Package main is the entry point for stig.
package main

import (
	"context"
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

	// Create the menu, controller, and screen clients.
	menuClient := menu.NewClient(page)
	controllerClient := controller.NewClient(page)
	screenClient := screen.NewClient(page)

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

	// Start the training loop.
	go trainingLoop(
		ctx,
		menuClient,
		agentClient,
		baseBrain,
		// 1 minute 50 seconds.
		110*time.Second,
		brainPath,
	)

	// Wait for the context to be done.
	<-ctx.Done()
	return
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

func trainingLoop(
	ctx context.Context,
	menuClient *menu.Client,
	agentClient *agent.Client,
	bestBrain *brain.Brain,
	timeout time.Duration,
	savePath string,
) error {
	bestTime := math.MaxFloat64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			// Create a candidate brain by mutating the best so far.
			candidate := bestBrain.Mutate(0.02)

			// Use the candidate brain for this episode.
			agentClient.SetBrain(candidate)

			raceTimeStr, err := doTraining(
				ctx,
				menuClient,
				agentClient,
				timeout,
			)
			if err != nil {
				return err
			}

			raceMs, err := parseRaceTime(raceTimeStr)
			if err != nil {
				log.Printf("failed to parse race time %q: %v", raceTimeStr, err)
				continue
			}

			// Keep the candidate if it performed better.
			if raceMs < bestTime {
				bestTime = raceMs
				bestBrain = candidate
				log.Printf("🎉 new best time: %.0f ms", bestTime)

				if err := bestBrain.Save(savePath); err != nil {
					log.Printf("failed to save best brain: %v", err)
				}
			}
		}
	}
}

func doTraining(
	parentCtx context.Context,
	menuClient *menu.Client,
	agentClient *agent.Client,
	timeout time.Duration,
) (string, error) {
	// Start game
	if err := menuClient.StartGame(); err != nil {
		return "", fmt.Errorf("failed to start game: %w", err)
	}

	// Sleep until countdown is done.
	time.Sleep(2 * time.Second)

	// Game context.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Drive in game.
	go agentClient.Run(ctx, 100*time.Millisecond)

	// Wait for game to finish.
	if err := menuClient.WaitForFinish(ctx); err != nil {
		return "", fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	// Get final time.
	raceTime, err := menuClient.GetReplayTime()
	if err != nil {
		return "", fmt.Errorf("failed to get race time: %w", err)
	}

	// Log the final time.
	log.Printf("Finished in %q", raceTime)

	// Sleep until the replay is done.
	time.Sleep(2 * time.Second)

	return raceTime, nil
}

// parseRaceTime converts a time string in the format mm:ss:SSS to total
// milliseconds.
func parseRaceTime(s string) (float64, error) {
	var min, sec, ms int
	if _, err := fmt.Sscanf(s, "%02d:%02d:%03d", &min, &sec, &ms); err != nil {
		return 0, err
	}
	return float64(min*60000 + sec*1000 + ms), nil
}
