package main

import (
	"time"

	"github.com/nizarmah/stig/game/internal/env"
)

// Env represents the environment variables for the application.
type Env struct {
	// AgentDebug is whether to debug the agent client.
	AgentDebug bool
	// AgentURL is the URL of the agent to use.
	AgentURL string
	// AgentTimeout is the timeout for the agent to act.
	AgentTimeout time.Duration
	// BrowserWSURL is the URL of the browser to control.
	BrowserWSURL string
	// FramesPerSecond is the frames per second of the game loop.
	FramesPerSecond int
	// GameDebug is whether to debug the game client.
	GameDebug bool
	// GameTimeout is the timeout for starting the game client (seconds).
	GameTimeout time.Duration
	// GameURL is the URL of the game to play.
	GameURL string
	// LapTimeout is the timeout for a single lap (seconds).
	LapTimeout time.Duration
	// ScreenDebug is whether to debug the screen package.
	ScreenDebug bool
	// ScreenResolution is the resolution of the screen.
	ScreenResolution int
	// WindowHeight is the height of the window.
	WindowHeight int
	// WindowWidth is the width of the window.
	WindowWidth int
}

// NewEnv creates a new Env instance.
func NewEnv() (*Env, error) {
	agentDebug, err := env.LookupBool("AGENT_DEBUG")
	if err != nil {
		return nil, err
	}

	agentURL, err := env.Lookup("AGENT_URL")
	if err != nil {
		return nil, err
	}

	agentTimeout, err := env.LookupDuration("AGENT_TIMEOUT", time.Second)
	if err != nil {
		return nil, err
	}

	browserWSURL, err := env.Lookup("BROWSER_WS_URL")
	if err != nil {
		return nil, err
	}

	framesPerSecond, err := env.LookupInt("FRAMES_PER_SECOND")
	if err != nil {
		return nil, err
	}

	gameDebug, err := env.LookupBool("GAME_DEBUG")
	if err != nil {
		return nil, err
	}

	gameTimeout, err := env.LookupDuration("GAME_TIMEOUT", time.Second)
	if err != nil {
		return nil, err
	}

	gameURL, err := env.Lookup("GAME_URL")
	if err != nil {
		return nil, err
	}

	lapTimeout, err := env.LookupDuration("LAP_TIMEOUT", time.Second)
	if err != nil {
		return nil, err
	}

	screenDebug, err := env.LookupBool("SCREEN_DEBUG")
	if err != nil {
		return nil, err
	}

	screenResolution, err := env.LookupInt("SCREEN_RESOLUTION")
	if err != nil {
		return nil, err
	}

	windowHeight, err := env.LookupInt("WINDOW_HEIGHT")
	if err != nil {
		return nil, err
	}

	windowWidth, err := env.LookupInt("WINDOW_WIDTH")
	if err != nil {
		return nil, err
	}

	return &Env{
		AgentDebug:       agentDebug,
		AgentURL:         agentURL,
		AgentTimeout:     agentTimeout,
		BrowserWSURL:     browserWSURL,
		FramesPerSecond:  framesPerSecond,
		GameDebug:        gameDebug,
		GameTimeout:      gameTimeout,
		GameURL:          gameURL,
		LapTimeout:       lapTimeout,
		ScreenDebug:      screenDebug,
		ScreenResolution: screenResolution,
		WindowHeight:     windowHeight,
		WindowWidth:      windowWidth,
	}, nil
}
