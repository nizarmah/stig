// Command stig plays the Horizon Drive game.
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

	"github.com/nizarmah/stig/game/internal/agent"
	"github.com/nizarmah/stig/game/internal/brain"
	"github.com/nizarmah/stig/game/internal/controller"
	"github.com/nizarmah/stig/game/internal/game"
	"github.com/nizarmah/stig/game/internal/screen"
)

func main() {
	// Environment.
	env, err := NewEnv()
	if err != nil {
		log.Fatalf("failed to create env: %v", err)
	}

	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Create the game client.
	gameClient, err := game.NewClient(ctx, game.ClientConfig{
		BrowserWSURL: env.BrowserWSURL,
		Debug:        env.GameDebug,
		FPS:          env.FramesPerSecond,
		GameURL:      env.GameURL,
	}, env.GameTimeout)
	if err != nil {
		log.Fatalf("failed to create game client: %v", err)
	}
	defer gameClient.Close()

	// Create the controller client.
	controllerClient := controller.NewClient(gameClient.Page)

	// Create the screen client.
	screenClient := screen.NewClient(screen.ClientConfiguration{
		Debug:      env.ScreenDebug,
		Page:       gameClient.Page,
		Resolution: env.ScreenResolution,
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
		gameClient,
		agentClient,
		baseBrain,
		brainPath,
		env.LapTimeout,
	)
}

func startTraining(
	ctx context.Context,
	gameClient *game.Client,
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
			gameClient,
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
	gameClient *game.Client,
	agentClient *agent.Client,
	timeout time.Duration,
) (float64, bool, error) {
	// Training context.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Reset the game.
	if err := gameClient.ResetGame(ctx); err != nil {
		return 0, false, fmt.Errorf("failed to reset game: %w", err)
	}

	// Start driving.
	go agentClient.Run(ctx, 100*time.Millisecond)

	// Wait for the game to finish.
	if err := gameClient.WaitForFinish(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return float64(timeout), false, nil
		}

		return 0, false, fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	// Get the final time.
	raceTime, err := gameClient.GetReplayTime(ctx)
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
