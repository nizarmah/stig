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
		WindowHeight: env.WindowHeight,
		WindowWidth:  env.WindowWidth,
	}, env.GameTimeout)
	if err != nil {
		log.Fatalf("failed to create game client: %v", err)
	}
	defer gameClient.Close()

	// Create the controller watcher.
	controllerWatcher, err := controller.NewWatcher(ctx, controller.WatcherConfiguration{
		Debug: env.ControllerDebug,
		Page:  gameClient.Page,
	})
	if err != nil {
		log.Fatalf("failed to create controller watcher: %v", err)
	}

	// Create the screen client.
	screenClient := screen.NewClient(screen.ClientConfiguration{
		Debug:        env.ScreenDebug,
		Page:         gameClient.Page,
		Resolution:   env.ScreenResolution,
		WindowHeight: env.WindowHeight,
		WindowWidth:  env.WindowWidth,
	})

	// Create the session.
	sessionTime := time.Now().Format(time.RFC3339)
	session := fmt.Sprintf("session_%s", sessionTime)

	// Create the session directory.
	sessionDir := filepath.Join(env.RecordingsDir, session)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		log.Fatalf("failed to create session directory: %v", err)
	}

	// Start the recorder loop.
	for lap := range env.LapsNum {
		select {
		case <-ctx.Done():
			return

		default:
			// Create the lap directory.
			lapDir := filepath.Join(sessionDir, fmt.Sprintf("lap_%d", lap))
			if err := os.MkdirAll(lapDir, 0755); err != nil {
				log.Println(fmt.Sprintf("failed to create lap %d directory: %v", lap, err))
				continue
			}

			// Record the lap.
			if err := recordLap(
				ctx,
				gameClient,
				controllerWatcher,
				screenClient,
				lapDir,
			); err != nil {
				log.Println(fmt.Sprintf("failed to record lap %d: %v", lap, err))
				continue
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
				"frame_%d_%s_%s.jpeg",
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
