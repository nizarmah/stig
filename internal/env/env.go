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
	// GameURL is the URL of the game to play.
	GameURL string
	// LapTimeout is the timeout for a single lap (milliseconds).
	LapTimeout time.Duration
}

// NewEnv creates a new Env instance.
func NewEnv() (*Env, error) {
	browserWSURL, err := lookup("BROWSER_WS_URL")
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

	return &Env{
		BrowserWSURL: browserWSURL,
		GameURL:      gameURL,
		LapTimeout:   lapTimeout,
	}, nil
}

func lookup(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("env var %q not set", key)
	}

	return value, nil
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
