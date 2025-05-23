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
	"github.com/go-rod/rod/lib/proto"

	"github.com/nizarmah/stig/internal/env"
	"github.com/nizarmah/stig/internal/game"
	"github.com/nizarmah/stig/internal/menu"
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
	page, err := openGame(ctx, browser, e.GameURL, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	// Create the menu manager.
	menu, err := menu.NewManager(page)
	if err != nil {
		log.Fatal(err)
	}

	// Create the game player.
	player, err := game.NewPlayer(page)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down gracefully.")
			return
		default:
		}

		raceTime, err := playGameLoop(ctx, menu, player)
		if err != nil {
			log.Println("Game error:", err)
			continue
		}

		log.Println(fmt.Sprintf("Finished race in %q", raceTime))
	}
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

func openGame(
	ctx context.Context,
	browser *rod.Browser,
	gameURL string,
	timeout time.Duration,
) (*rod.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Open the game page.
	page, err := browser.
		Page(proto.TargetCreateTarget{URL: gameURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Wait for the page to load.
	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page to load: %w", err)
	}

	// Search for the "Start" button.
	startButton, err := page.ElementX(`//span[text()="Start"]`)
	if err != nil {
		return nil, fmt.Errorf("failed to find start button: %w", err)
	}

	// Wait until the "Start" button is interactable.
	if _, err := startButton.Context(ctx).WaitInteractable(); err != nil {
		return nil, fmt.Errorf("failed to wait for start button to be interactable: %w", err)
	}

	return page, nil
}

func playGameLoop(
	parentCtx context.Context,
	menu *menu.Manager,
	player *game.Player,
) (string, error) {
	// Start game
	if err := menu.StartGame(); err != nil {
		return "", fmt.Errorf("failed to start game: %w", err)
	}

	// Game context.
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Drive in game.
	go player.Drive(ctx, 100*time.Millisecond)

	// Wait for game to finish.
	if err := menu.WaitForFinish(ctx, 1*time.Second); err != nil {
		return "", fmt.Errorf("failed to wait for game to finish: %w", err)
	}

	// Get final time.
	raceTime, err := menu.GetReplayTime()
	if err != nil {
		return "", fmt.Errorf("failed to get race time: %w", err)
	}

	return raceTime, nil
}
