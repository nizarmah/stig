// Package game provides entities for the game.
package game

import "github.com/go-rod/rod/lib/input"

// Action represents a control state in the game.
// Only one throttle and one steering input are active at a time.
// The game will only recognize the last input, so we can only have one of the two.
// For example, if the player presses "accelerate" and then "brake", the game will only recognize "brake".
type Action struct {
	// Throttle is the throttle state.
	Throttle Throttle `json:"throttle"`
	// Steering is the steering state.
	Steering Steering `json:"steering"`
}

// Throttle represents a throttle state in the game.
type Throttle = string

const (
	// ThrottleNeutral is the throttle action for neutral.
	ThrottleNeutral Throttle = ""
	// ThrottleAccelerate is the throttle action for accelerating.
	ThrottleAccelerate Throttle = "accelerate"
	// ThrottleBrake is the throttle action for braking.
	ThrottleBrake Throttle = "brake"
)

var (
	// ThrottleAccelerateKeys is the keys for accelerating.
	ThrottleAccelerateKeys = []input.Key{
		input.ArrowUp,
		input.KeyW,
	}
	// ThrottleBrakeKeys is the keys for braking.
	ThrottleBrakeKeys = []input.Key{
		input.ArrowDown,
		input.KeyS,
		input.Space,
	}
)

var (
	// ThrottleStateMap is the map of throttle states to their keys.
	ThrottleStateMap = map[Throttle][]input.Key{
		ThrottleAccelerate: ThrottleAccelerateKeys,
		ThrottleBrake:      ThrottleBrakeKeys,
	}
)

// Steering represents a steering state in the game.
type Steering = string

const (
	// SteeringStraight is the steering action for straight.
	SteeringStraight Steering = ""
	// SteeringLeft is the steering action for turning left.
	SteeringLeft Steering = "left"
	// SteeringRight is the steering action for turning right.
	SteeringRight Steering = "right"
)

var (
	// SteeringLeftKeys is the keys for turning left.
	SteeringLeftKeys = []input.Key{
		input.ArrowLeft,
		input.KeyA,
	}
	// SteeringRightKeys is the keys for turning right.
	SteeringRightKeys = []input.Key{
		input.ArrowRight,
		input.KeyD,
	}
	// SteeringKeys is the keys for all steering actions.
	SteeringKeys = append(
		SteeringLeftKeys,
		SteeringRightKeys...,
	)
)

var (
	// SteeringStateMap is the map of steering states to their keys.
	SteeringStateMap = map[Steering][]input.Key{
		SteeringLeft:  SteeringLeftKeys,
		SteeringRight: SteeringRightKeys,
	}
)
