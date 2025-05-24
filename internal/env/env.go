// Package env provides environment variables for the application.
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Env represents the environment variables for the application.
type Env struct {
	// BrowserWSURL is the URL of the browser to control.
	BrowserWSURL string
	// FramesPerSecond is the frames per second of the game loop.
	FramesPerSecond int
	// GameURL is the URL of the game to play.
	GameURL string
	// LapTimeout is the timeout for a single lap (milliseconds).
	LapTimeout time.Duration
	// RecorderOutputDir is the directory to output the recorder recordings.
	RecorderOutputDir string
	// RecorderLapsNum is the number of laps to record.
	RecorderLapsNum int
	// ScreenDebug is whether to debug the screen package.
	ScreenDebug bool
	// ScreenResolution is the resolution of the screen.
	ScreenResolution int
}

// NewEnv creates a new Env instance.
func NewEnv() (*Env, error) {
	browserWSURL, err := lookup("BROWSER_WS_URL")
	if err != nil {
		return nil, err
	}

	framesPerSecond, err := lookupInt("FRAMES_PER_SECOND")
	if err != nil {
		return nil, err
	}

	gameURL, err := lookup("GAME_URL")
	if err != nil {
		return nil, err
	}

	lapTimeout, err := lookupDuration("LAP_TIMEOUT", time.Millisecond)
	if err != nil {
		return nil, err
	}

	recorderLapsNum, err := lookupInt("RECORDER_LAPS_NUM")
	if err != nil {
		return nil, err
	}

	recorderOutputDir, err := lookup("RECORDER_OUTPUT_DIR")
	if err != nil {
		return nil, err
	}

	screenDebug, err := lookupBool("SCREEN_DEBUG")
	if err != nil {
		return nil, err
	}

	screenResolution, err := lookupInt("SCREEN_RESOLUTION")
	if err != nil {
		return nil, err
	}

	return &Env{
		BrowserWSURL:      browserWSURL,
		FramesPerSecond:   framesPerSecond,
		GameURL:           gameURL,
		LapTimeout:        lapTimeout,
		RecorderLapsNum:   recorderLapsNum,
		RecorderOutputDir: recorderOutputDir,
		ScreenDebug:       screenDebug,
		ScreenResolution:  screenResolution,
	}, nil
}

func lookup(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("env var %q not set", key)
	}

	return value, nil
}

func lookupBool(key string) (bool, error) {
	value, err := lookup(key)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(value)
}

func lookupInt(key string) (int, error) {
	value, err := lookup(key)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

func lookupDuration(key string, unitTime time.Duration) (time.Duration, error) {
	value, err := lookupInt(key)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * unitTime, nil
}
