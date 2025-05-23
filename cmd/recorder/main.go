// Command recorder records the gameplay for supervised learning.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nizarmah/stig/internal/controller"
	"github.com/nizarmah/stig/internal/env"
	"github.com/nizarmah/stig/internal/game"
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

	// Create the game client.
	gameClient, err := game.NewClient(ctx, game.ClientConfig{
		BrowserWSURL: e.BrowserWSURL,
		Debug:        false,
		FPS:          e.FramesPerSecond,
		GameURL:      e.GameURL,
	}, 10*time.Second)
	if err != nil {
		log.Fatalf("failed to create game client: %v", err)
	}
	defer gameClient.Close()

	// Create the controller watcher.
	controllerWatcher, err := controller.NewWatcher(ctx, controller.WatcherConfiguration{
		Debug: true,
		Page:  gameClient.Page,
	})
	if err != nil {
		log.Fatalf("failed to create controller watcher: %v", err)
	}

	// Create the screen client.
	screenClient := screen.NewClient(screen.ClientConfiguration{
		Debug:      false,
		Page:       gameClient.Page,
		Resolution: 100,
	})

	// Create the session.
	sessionTime := time.Now().Format(time.RFC3339)
	session := fmt.Sprintf("session_%s", sessionTime)

	// Create the session directory.
	sessionDir := filepath.Join(e.RecorderOutputDir, session)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		log.Fatalf("failed to create session directory: %v", err)
	}

	// Start the recorder loop.
	for lap := range e.RecorderLapsNum {
		select {
		case <-ctx.Done():
			return

		default:
			// Create the lap directory.
			lapDir := filepath.Join(sessionDir, fmt.Sprintf("lap_%d", lap))
			if err := os.MkdirAll(lapDir, 0755); err != nil {
				log.Fatalf("failed to create lap %d directory: %v", lap, err)
			}

			// Record the lap.
			if err := recordLap(
				ctx,
				gameClient,
				controllerWatcher,
				screenClient,
				lapDir,
			); err != nil {
				log.Fatalf("failed to record lap %d: %v", lap, err)
			}
		}
	}
}

// recordLap records a single lap of the game.
func recordLap(
	parentCtx context.Context,
	gameClient *game.Client,
	controllerWatcher *controller.Watcher,
	screenClient *screen.Client,
	outputDir string,
) error {
	// Lap context.
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	// Reset the game.
	if err := gameClient.ResetGame(ctx); err != nil {
		return fmt.Errorf("failed to reset game: %w", err)
	}

	// Wait for the countdown to finish.
	time.Sleep(3 * time.Second)

	go gameClient.RunInGameLoop(ctx,
		recordGameplay(controllerWatcher, screenClient, outputDir))

	// Wait for the game to finish.
	if err := gameClient.WaitForFinish(ctx); err != nil {
		return fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	return nil
}

func recordGameplay(
	controllerWatcher *controller.Watcher,
	screenClient *screen.Client,
	outputDir string,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// Capture the controller input.
		action, err := controllerWatcher.Peek()
		if err != nil {
			log.Println(fmt.Sprintf("failed to capture controller action: %v", err))
			return fmt.Errorf("failed to capture controller action: %w", err)
		}

		// Capture the frame.
		frame, err := screenClient.Peek(ctx)
		if err != nil {
			log.Println(fmt.Sprintf("failed to capture screen: %v", err))
			return fmt.Errorf("failed to capture screen: %w", err)
		}

		// Save the frame to the output directory.
		framePath := filepath.Join(
			outputDir,
			fmt.Sprintf(
				"frame_%d_%s_%s.png",
				time.Now().UnixNano(),
				action.Throttle,
				action.Steering,
			),
		)
		if err := os.WriteFile(framePath, frame, 0644); err != nil {
			return fmt.Errorf("failed to write frame: %w", err)
		}

		return nil
	}
}
