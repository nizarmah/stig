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
	defer browser.Close()

	// Start the game.
	page, err := startGame(ctx, browser, e.GameURL, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

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

func startGame(
	ctx context.Context,
	browser *rod.Browser,
	gameURL string,
	timeout time.Duration,
) (*rod.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Open the game page.
	page, err := browser.Page(proto.TargetCreateTarget{URL: gameURL})
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

	// Click on the "Start" button.
	if err := startButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return nil, fmt.Errorf("failed to click on start button: %w", err)
	}

	return page, nil
}
