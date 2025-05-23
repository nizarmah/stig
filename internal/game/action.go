// Package game provides entities for the game.
package game

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
