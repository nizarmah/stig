// Package env provides environment variables for the application.
package env

import (
	"fmt"
	"os"
)

// Env represents the environment variables for the application.
type Env struct {
	// BrowserWSURL is the URL of the browser to control.
	BrowserWSURL string
	// GameURL is the URL of the game to play.
	GameURL string
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

	return &Env{
		BrowserWSURL: browserWSURL,
		GameURL:      gameURL,
	}, nil
}

func lookup(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("env var %q not set", key)
	}

	return value, nil
}
