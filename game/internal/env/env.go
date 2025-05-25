// Package env provides environment variables for the application.
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Lookup looks up a string from the environment variable.
func Lookup(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("env var %q not set", key)
	}

	return value, nil
}

// LookupBool looks up a boolean from the environment variable.
func LookupBool(key string) (bool, error) {
	value, err := Lookup(key)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(value)
}

// LookupInt looks up an integer from the environment variable.
func LookupInt(key string) (int, error) {
	value, err := Lookup(key)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

// LookupDuration looks up a duration from the environment variable.
func LookupDuration(key string, unitTime time.Duration) (time.Duration, error) {
	value, err := LookupInt(key)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * unitTime, nil
}
