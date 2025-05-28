// Package game provides a client for the game.
package game

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ClientConfig is the configuration for the browser client.
type ClientConfig struct {
	// BrowserWSURL is the websocket URL of the browser.
	BrowserWSURL string
	// Debug is whether to print debug information.
	Debug bool
	// FPS is the frames per second of the game loop.
	FPS int
	// GameURL is the URL of the game.
	GameURL string
	// WindowHeight is the height of the window.
	WindowHeight int
	// WindowWidth is the width of the window.
	WindowWidth int
}

// Client is a client for the browser.
type Client struct {
	// browser is the browser instance.
	browser *rod.Browser
	// debug is whether to print debug information.
	debug bool
	// fps is the frames per second of the game loop.
	fps int
	// Page is the Page of the game.
	Page *rod.Page
}

// NewClient creates a new game client.
func NewClient(
	ctx context.Context,
	config ClientConfig,
	timeout time.Duration,
) (*Client, error) {
	browser := rod.New().
		Context(ctx).
		ControlURL(config.BrowserWSURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	page, err := openGamePage(
		ctx,
		browser,
		config.GameURL,
		config.WindowHeight,
		config.WindowWidth,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open game page: %w", err)
	}

	return &Client{
		browser: browser,
		debug:   config.Debug,
		fps:     config.FPS,
		Page:    page,
	}, nil
}

// Close closes the game client.
func (c *Client) Close() {
	if err := c.Page.Close(); err != nil {
		log.Println(fmt.Sprintf("failed to close page: %v", err))
	}

	// Commenting because this causes friction while testing.
	// if err := c.browser.Close(); err != nil {
	// 	log.Println(fmt.Sprintf("failed to close browser: %v", err))
	// }
}

// openGamePage opens the game page.
func openGamePage(
	ctx context.Context,
	browser *rod.Browser,
	gameURL string,
	windowHeight int,
	windowWidth int,
	timeout time.Duration,
) (*rod.Page, error) {
	// Open the game page.
	page, err := browser.
		Context(ctx).
		Page(proto.TargetCreateTarget{URL: gameURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set the viewport.
	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Viewport: &proto.PageViewport{
			Width:  float64(windowWidth),
			Height: float64(windowHeight),
			Scale:  1,
		},
		Width:             windowWidth,
		Height:            windowHeight,
		ScreenWidth:       &windowWidth,
		ScreenHeight:      &windowHeight,
		DeviceScaleFactor: 1,
		Mobile:            false,
	})

	loadCtx, loadCancel := context.WithTimeout(ctx, timeout)
	defer loadCancel()

	// Wait for the page to load.
	if err := page.Context(loadCtx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page to load: %w", err)
	}

	// Search for the "Start" button.
	startButton, err := page.ElementX(`//span[text()="Start"]`)
	if err != nil {
		return nil, fmt.Errorf("failed to find start button: %w", err)
	}

	// Wait until the "Start" button is interactable.
	if _, err := startButton.Context(loadCtx).WaitInteractable(); err != nil {
		return nil, fmt.Errorf("failed to wait for start button to be interactable: %w", err)
	}

	return page, nil
}
