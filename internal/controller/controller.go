// Package controller provides functions to control the game.
package controller

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"github.com/nizarmah/stig/internal/game"
)

var (
	keyNil = input.NumLock
)

// Client controls the game.
type Client struct {
	// Page is the page of the game.
	page *rod.Page
	// Action is the last action done by the controller.
	action game.Action
}

// NewClient creates a new client.
func NewClient(page *rod.Page) *Client {
	return &Client{
		page:   page,
		action: game.Action{},
	}
}

// Apply applies an action in the game.
func (c *Client) Apply(action game.Action) error {
	if err := c.applyThrottle(c.action.Throttle, action.Throttle); err != nil {
		return err
	}

	if err := c.applySteering(c.action.Steering, action.Steering); err != nil {
		return err
	}

	return nil
}

// ApplyThrottle applies a throttle action to the game.
func (c *Client) applyThrottle(prev, curr string) error {
	// If the key didn't change, do nothing.
	if prev == curr {
		return nil
	}

	if err := releaseKey(c.page, mapThrottleKey(prev)); err != nil {
		return err
	}

	if err := pressKey(c.page, mapThrottleKey(curr)); err != nil {
		return err
	}

	return nil
}

// ApplySteering applies a steering action to the game.
func (c *Client) applySteering(prev, curr string) error {
	// If the key didn't change, do nothing.
	if prev == curr {
		return nil
	}

	if err := releaseKey(c.page, mapSteeringKey(prev)); err != nil {
		return err
	}

	if err := pressKey(c.page, mapSteeringKey(curr)); err != nil {
		return err
	}

	return nil
}

// MapThrottleKey maps a throttle action to a keyboard key.
func mapThrottleKey(throttle string) input.Key {
	switch throttle {
	case "accelerate":
		return input.ArrowUp
	case "brake":
		return input.ArrowDown
	default:
		return keyNil
	}
}

// MapSteeringKey maps a steering action to a keyboard key.
func mapSteeringKey(steering string) input.Key {
	switch steering {
	case "left":
		return input.ArrowLeft
	case "right":
		return input.ArrowRight
	default:
		return keyNil
	}
}

// PressKey presses and holds a key on the keyboard.
func pressKey(page *rod.Page, key input.Key) error {
	if key == keyNil {
		return nil
	}

	return page.Keyboard.Press(key)
}

// ReleaseKey releases a key on the keyboard.
func releaseKey(page *rod.Page, key input.Key) error {
	if key == keyNil {
		return nil
	}

	return page.Keyboard.Release(key)
}
