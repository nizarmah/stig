// Package game provides game player logic.
package game

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

var (
	keyNil = input.NumLock
)

// Action represents a control state in the game.
// Only one throttle and one steering input are active at a time.
// The game will only recognize the last input, so we can only have one of the two.
// For example, if the player presses "accelerate" and then "brake", the game will only recognize "brake".
type Action struct {
	// Throttle can be "accelerate", "brake", or "" (neutral)
	Throttle string
	// Steering can be "left", "right", or "" (straight)
	Steering string
}

// Player plays the game.
type Player struct {
	// Page is the page of the game.
	page *rod.Page
}

// NewPlayer creates a new player.
func NewPlayer(page *rod.Page) (*Player, error) {
	return &Player{page: page}, nil
}

// Drive starts controlling the game until ctx ends (usually from `menu.WaitForFinish`).
func (p *Player) Drive(
	ctx context.Context,
	interval time.Duration,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	prev := Action{
		Throttle: "",
		Steering: "",
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			// Generate next action.
			action := p.generateNextAction()
			if err := p.applyAction(prev, action); err != nil {
				return err
			}

			prev = action
		}
	}
}

// GenerateNextAction generates the next action to take.
func (p *Player) generateNextAction() Action {
	// TODO: Implement this.
	return Action{
		Throttle: "accelerate",
		Steering: "",
	}
}

// ApplyAction applies an action to the game.
func (p *Player) applyAction(prev, curr Action) error {
	if err := p.applyThrottle(prev.Throttle, curr.Throttle); err != nil {
		return err
	}

	if err := p.applySteering(prev.Steering, curr.Steering); err != nil {
		return err
	}

	return nil
}

// ApplyThrottle applies a throttle action to the game.
func (p *Player) applyThrottle(prev, curr string) error {
	// If the key didn't change, do nothing.
	if prev == curr {
		return nil
	}

	if err := releaseKey(p.page, mapThrottleKey(prev)); err != nil {
		return err
	}

	if err := pressKey(p.page, mapThrottleKey(curr)); err != nil {
		return err
	}

	return nil
}

// ApplySteering applies a steering action to the game.
func (p *Player) applySteering(prev, curr string) error {
	// If the key didn't change, do nothing.
	if prev == curr {
		return nil
	}

	if err := releaseKey(p.page, mapSteeringKey(prev)); err != nil {
		return err
	}

	if err := pressKey(p.page, mapSteeringKey(curr)); err != nil {
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
