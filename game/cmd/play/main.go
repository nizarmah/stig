// Command stig plays the Horizon Drive game.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/nizarmah/stig/game/internal/agent"
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

	// Create the agent client.
	agentClient := agent.NewClient(agent.ClientConfiguration{
		APIURL:  env.AgentURL,
		Debug:   env.AgentDebug,
		Timeout: env.AgentTimeout,
	})

	for {
		select {
		case <-ctx.Done():
			return

		default:
			if err := playLap(
				ctx,
				gameClient,
				agentClient,
				controllerClient,
				screenClient,
				env.LapTimeout,
			); err != nil {
				log.Println(fmt.Sprintf("failed to play lap: %v", err))
				continue
			}
		}
	}
}

// playLap plays a single lap of the game.
func playLap(
	parentCtx context.Context,
	gameClient *game.Client,
	agentClient *agent.Client,
	controllerClient *controller.Client,
	screenClient *screen.Client,
	timeout time.Duration,
) error {
	// Lap context.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Reset the game.
	if err := gameClient.ResetGame(ctx); err != nil {
		return fmt.Errorf("failed to reset game: %w", err)
	}

	// Wait for the countdown to finish.
	time.Sleep(3 * time.Second)

	go gameClient.RunInGameLoop(ctx,
		startGameplay(agentClient, controllerClient, screenClient))

	// Wait for the game to finish.
	if err := gameClient.WaitForFinish(ctx); err != nil {
		return fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	return nil
}

func startGameplay(
	agentClient *agent.Client,
	controllerClient *controller.Client,
	screenClient *screen.Client,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// Capture the frame.
		frame, err := screenClient.Peek(ctx)
		if err != nil {
			return fmt.Errorf("failed to capture screen: %w", err)
		}

		// Capture the action.
		action, err := agentClient.Act(frame)
		if err != nil {
			return fmt.Errorf("failed to predict action: %w", err)
		}

		// Apply the action.
		if err := controllerClient.Apply(action); err != nil {
			return fmt.Errorf("failed to apply action: %w", err)
		}

		return nil
	}
}
