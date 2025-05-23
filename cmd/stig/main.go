// Package main is the entry point for stig.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-rod/rod"

	"github.com/nizarmah/stig/internal/agent"
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

	// Create the agent client.
	agentClient := agent.NewClient(
		controllerClient,
		screenClient,
	)

	// Start the training loop.
	go trainingLoop(
		ctx,
		menuClient,
		agentClient,
		// 1 minute 50 seconds.
		110*time.Second,
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
	timeout time.Duration,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			if _, err := doTraining(
				ctx,
				menuClient,
				agentClient,
				timeout,
			); err != nil {
				return err
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
