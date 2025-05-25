package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/nizarmah/stig/game/internal/game"
)

// WatcherConfiguration is the configuration for the watcher.
type WatcherConfiguration struct {
	// Debug is whether to print debug information.
	Debug bool
	// Page is the page of the game.
	Page *rod.Page
}

// Watcher is a watcher for controller actions.
type Watcher struct {
	// debug is whether to print debug information.
	debug bool
	// page is the page of the game.
	page *rod.Page
}

// NewWatcher creates a new watcher.
func NewWatcher(
	ctx context.Context,
	cfg WatcherConfiguration,
) (*Watcher, error) {
	if err := addKeyListener(ctx, cfg.Page, cfg.Debug); err != nil {
		return nil, fmt.Errorf("failed to add key listener: %w", err)
	}

	return &Watcher{
		debug: cfg.Debug,
		page:  cfg.Page,
	}, nil
}

// Peek returns the last action from the window.
func (w *Watcher) Peek() (game.Action, error) {
	action := game.Action{}

	// Get the last action from the window.
	res, err := w.page.Evaluate(&rod.EvalOptions{
		JS:      `() => window.lastAction`,
		ByValue: true,
	})
	if err != nil {
		return action, err
	}

	// Marshal the result to JSON.
	actionJSON, err := res.Value.MarshalJSON()
	if err != nil {
		return action, fmt.Errorf("failed to marshal action: %w", err)
	}

	// Unmarshal the JSON string into the action.
	if err := json.Unmarshal(actionJSON, &action); err != nil {
		return action, fmt.Errorf("failed to unmarshal action: %w", err)
	}

	if w.debug {
		log.Println(
			fmt.Sprintf(
				"peeked action: throttle: %q, steering: %q",
				action.Throttle,
				action.Steering,
			),
		)
	}

	return action, nil
}

// addKeyListener adds a key listener to the page.
func addKeyListener(
	ctx context.Context,
	page *rod.Page,
	debug bool,
) error {
	// Create the throttle key map.
	throttleKeyMapJSON, err := createKeyMapJSON(game.ThrottleStateMap)
	if err != nil {
		return fmt.Errorf("failed to create throttle key map: %w", err)
	}

	// Create the steering key map.
	steeringKeyMapJSON, err := createKeyMapJSON(game.SteeringStateMap)
	if err != nil {
		return fmt.Errorf("failed to create steering key map: %w", err)
	}

	// Create a keyboard event listener.
	listener := fmt.Sprintf(
		`
			() => {
				const throttleNeutral = %q
				const throttleKeyMap = JSON.parse(%q)

				const steeringNeutral = %q
				const steeringKeyMap = JSON.parse(%q)

				// we might have "accelerate" active,
				// then press "brake", without releasing "accelerate".
				// so we track if a state is active and when it was last pressed.
				// so if we release "brake" we go back to the last active state.
				const activeThrottleStates = {}
				Object.values(activeThrottleStates).forEach(
					(state) => activeThrottleStates[state] = null
				)

				// do the same for steering.
				const activeSteeringStates = {}
				Object.values(activeSteeringStates).forEach(
					(state) => activeSteeringStates[state] = null
				)

				// we want to easily return the last action upon request,
				// so we track it in a global variable.
				window.lastAction = {
					throttle: throttleNeutral,
					steering: steeringNeutral,
				}

				// whenever we get a key event, we update the action.
				const updateAction = (
					action,
					neutralState,
					activeStates,
					keyMap,
					code,
					type
				) => {
					const state = keyMap[code]
					if (!state) return

					// update the active states.
					activeStates[state] = Date.now()
					if (type === 'keyup') delete activeStates[state]

					// if no state is active, we return the neutral state.
					if (Object.keys(activeStates).length === 0) {
						window.lastAction[action] = neutralState
						return
					}

					// sort the active states by the most recent.
					const sortedActiveStates = Object
						.entries(activeStates)
						.sort((a, b) => b[1] - a[1])

					// update the last action with the most recent active state.
					window.lastAction[action] = sortedActiveStates[0][0]
				}

				const handler = (e) => {
					updateAction(
						"throttle",
						throttleNeutral,
						activeThrottleStates,
						throttleKeyMap,
						e.code,
						e.type
					)

					updateAction(
						"steering",
						steeringNeutral,
						activeSteeringStates,
						steeringKeyMap,
						e.code,
						e.type
					)
				}

				window.addEventListener('keydown', handler)
				window.addEventListener('keyup', handler)
			}
		`,
		game.ThrottleNeutral,
		throttleKeyMapJSON,
		game.SteeringStraight,
		steeringKeyMapJSON,
	)

	// Add the listener to the page.
	if _, err := page.
		Context(ctx).
		Evaluate(&rod.EvalOptions{JS: listener}); err != nil {
		if debug {
			log.Println(fmt.Sprintf("failed to add listener to page: %v", err))
		}

		return fmt.Errorf("failed to add listener to page: %w", err)
	}

	if debug {
		log.Println(fmt.Sprintf("added key listener: %s", strings.TrimSpace(listener)))
	}

	return nil
}

// createThrottleKeyMapJSON creates a JSON string for the throttle key map.
func createKeyMapJSON(
	stateMap map[string][]input.Key,
) (string, error) {
	// Create the key map as [keyCode]: state.
	keyMap := make(map[string]string)
	for state, keys := range stateMap {
		for _, key := range keys {
			keyMap[key.Info().Code] = state
		}
	}

	// Marshal the key map to JSON.
	mapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal key map: %w", err)
	}

	return string(mapJSON), nil
}
