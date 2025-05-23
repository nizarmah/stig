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

	"github.com/nizarmah/stig/internal/env"
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

	// Open the game and its menu.
	menu, err := menu.NewManager(ctx, browser, e.GameURL, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer menu.Close()

	// Start the game.
	if err := menu.StartGame(); err != nil {
		log.Fatal(err)
	}

	// Wait until the game finishes.
	if err := menu.WaitForFinish(ctx, 3*time.Minute); err != nil {
		log.Fatal(err)
	}

	// Log the race time.
	if timeStr, err := menu.GetReplayTime(); err == nil {
		log.Println("Time:", timeStr)
	}

	// Replay the game.
	if err := menu.ReplayGame(); err != nil {
		log.Fatal(err)
	}

	// Wait for the game to replay.
	time.Sleep(5 * time.Second)
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
