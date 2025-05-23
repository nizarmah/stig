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
	// ScreenDebug is whether to debug the screen package.
	ScreenDebug bool
	// ScreenResolution is the resolution of the screen.
	ScreenResolution int
	// MotionThreshold is the minimum difference to consider as motion.
	MotionThreshold float64
	// MotionDebug is whether to debug the motion package.
	MotionDebug bool
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

	screenDebug, err := lookupBool("SCREEN_DEBUG")
	if err != nil {
		return nil, err
	}

	screenResolution, err := lookupInt("SCREEN_RESOLUTION")
	if err != nil {
		return nil, err
	}

	motionThreshold, err := lookupFloat("MOTION_THRESHOLD")
	if err != nil {
		return nil, err
	}

	motionDebug, err := lookupBool("MOTION_DEBUG")
	if err != nil {
		return nil, err
	}

	return &Env{
		BrowserWSURL:     browserWSURL,
		GameURL:          gameURL,
		LapTimeout:       lapTimeout,
		ScreenDebug:      screenDebug,
		ScreenResolution: screenResolution,
		MotionThreshold:  motionThreshold,
		MotionDebug:      motionDebug,
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

func lookupFloat(key string) (float64, error) {
	value, err := lookup(key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(value, 64)
}

func lookupDuration(key string, unitTime time.Duration) (time.Duration, error) {
	value, err := lookupInt(key)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * unitTime, nil
}
