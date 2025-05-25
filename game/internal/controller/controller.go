// Package controller provides functions to control the game.
package controller

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"github.com/nizarmah/stig/game/internal/game"
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
	if err := c.applyKey(
		mapInputKey(game.ThrottleStateMap, c.action.Throttle),
		mapInputKey(game.ThrottleStateMap, action.Throttle),
	); err != nil {
		return err
	}

	if err := c.applyKey(
		mapInputKey(game.SteeringStateMap, c.action.Steering),
		mapInputKey(game.SteeringStateMap, action.Steering),
	); err != nil {
		return err
	}

	return nil
}

// ApplyKey applies a key action to the game.
func (c *Client) applyKey(prev, curr input.Key) error {
	// If the key didn't change, do nothing.
	if prev == curr {
		return nil
	}

	if err := releaseKey(c.page, prev); err != nil {
		return err
	}

	if err := pressKey(c.page, curr); err != nil {
		return err
	}

	return nil
}

// MapThrottleKey maps a throttle action to a keyboard key.
func mapInputKey(stateMap map[string][]input.Key, state string) input.Key {
	keys, ok := stateMap[state]
	if ok {
		return keys[0]
	}

	return keyNil
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
