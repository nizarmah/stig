package main

import (
	"github.com/nizarmah/stig/game/internal/env"
)

// Env represents the environment variables for the application.
type Env struct {
	// BrowserWSURL is the URL of the browser to control.
	BrowserWSURL string
	// ControllerDebug is whether to debug the controller package.
	ControllerDebug bool
	// FramesPerSecond is the frames per second of the game loop.
	FramesPerSecond int
	// GameDebug is whether to debug the game client.
	GameDebug bool
	// GameURL is the URL of the game to play.
	GameURL string
	// LapsNum is the number of laps to record.
	LapsNum int
	// RecordingsDir is the directory to output the recordings.
	RecordingsDir string
	// ScreenDebug is whether to debug the screen package.
	ScreenDebug bool
	// ScreenResolution is the resolution of the screen.
	ScreenResolution int
}

// NewEnv creates a new Env instance.
func NewEnv() (*Env, error) {
	browserWSURL, err := env.Lookup("BROWSER_WS_URL")
	if err != nil {
		return nil, err
	}

	controllerDebug, err := env.LookupBool("CONTROLLER_DEBUG")
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

	gameURL, err := env.Lookup("GAME_URL")
	if err != nil {
		return nil, err
	}

	lapsNum, err := env.LookupInt("LAPS_NUM")
	if err != nil {
		return nil, err
	}

	recordingsDir, err := env.Lookup("RECORDINGS_DIR")
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

	return &Env{
		BrowserWSURL:     browserWSURL,
		ControllerDebug:  controllerDebug,
		FramesPerSecond:  framesPerSecond,
		GameDebug:        gameDebug,
		GameURL:          gameURL,
		LapsNum:          lapsNum,
		RecordingsDir:    recordingsDir,
		ScreenDebug:      screenDebug,
		ScreenResolution: screenResolution,
	}, nil
}
