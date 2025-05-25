// Package brain provides the brain of the agent.
package brain

import (
	"bytes"
	"image/jpeg"
	"math"
	"math/rand"
	"time"
)

// Brain is a very small fully-connected neural network that maps raw pixels
// to throttle and steering commands.
// The network topology is:
//  input  -> hidden (ReLU) -> 2 softmax heads (throttle, steering)
//  * throttle head has three classes: accelerate, brake, neutral
//  * steering head has three classes: left, right, straight
// The weights are optimised through a simple evolutionary strategy: every
// iteration we sample a mutated copy, keep it if it performs better.
// This keeps the implementation lightweight while still allowing the policy
// to improve over time.
// NOTE: This is **not** state-of-the-art, but is sufficient for a PoC.

// Brain is the brain of the agent.
type Brain struct {
	inputSize  int
	hiddenSize int

	// First layer.
	w1 [][]float64 // [input][hidden]
	b1 []float64   // [hidden]

	// Throttle head.
	w2T [][]float64 // [hidden][3]
	b2T []float64   // [3]

	// Steering head.
	w2S [][]float64 // [hidden][3]
	b2S []float64   // [3]
}

// NewBrain creates a new randomly initialised Brain.
func NewBrain(inputSize, hiddenSize int) *Brain {
	rand.Seed(time.Now().UnixNano())

	b := &Brain{
		inputSize:  inputSize,
		hiddenSize: hiddenSize,
		w1:         make([][]float64, inputSize),
		b1:         make([]float64, hiddenSize),
		w2T:        make([][]float64, hiddenSize),
		b2T:        make([]float64, 3),
		w2S:        make([][]float64, hiddenSize),
		b2S:        make([]float64, 3),
	}

	for i := 0; i < inputSize; i++ {
		b.w1[i] = make([]float64, hiddenSize)
		for j := 0; j < hiddenSize; j++ {
			b.w1[i][j] = rand.Float64()*0.2 - 0.1 // ~U(-0.1, 0.1)
		}
	}

	for j := 0; j < hiddenSize; j++ {
		b.w2T[j] = make([]float64, 3)
		b.w2S[j] = make([]float64, 3)
		for k := 0; k < 3; k++ {
			b.w2T[j][k] = rand.Float64()*0.2 - 0.1
			b.w2S[j][k] = rand.Float64()*0.2 - 0.1
		}
	}

	return b
}

// Clone makes a deep copy of the brain.
func (b *Brain) Clone() *Brain {
	clone := &Brain{
		inputSize:  b.inputSize,
		hiddenSize: b.hiddenSize,
		w1:         make([][]float64, len(b.w1)),
		b1:         append([]float64(nil), b.b1...),
		w2T:        make([][]float64, len(b.w2T)),
		b2T:        append([]float64(nil), b.b2T...),
		w2S:        make([][]float64, len(b.w2S)),
		b2S:        append([]float64(nil), b.b2S...),
	}

	for i := range b.w1 {
		clone.w1[i] = append([]float64(nil), b.w1[i]...)
	}
	for i := range b.w2T {
		clone.w2T[i] = append([]float64(nil), b.w2T[i]...)
	}
	for i := range b.w2S {
		clone.w2S[i] = append([]float64(nil), b.w2S[i]...)
	}

	return clone
}

// Mutate returns a mutated copy of the brain. The weights are perturbed with
// Gaussian noise of standard deviation `scale`.
func (b *Brain) Mutate(scale float64) *Brain {
	m := b.Clone()

	for i := range m.w1 {
		for j := range m.w1[i] {
			m.w1[i][j] += rand.NormFloat64() * scale
		}
	}

	for j := range m.w2T {
		for k := range m.w2T[j] {
			m.w2T[j][k] += rand.NormFloat64() * scale
			m.w2S[j][k] += rand.NormFloat64() * scale
		}
	}

	for i := range m.b1 {
		m.b1[i] += rand.NormFloat64() * scale
	}
	for i := range m.b2T {
		m.b2T[i] += rand.NormFloat64() * scale
		m.b2S[i] += rand.NormFloat64() * scale
	}

	return m
}

// Predict converts the raw JPEG bytes into a feature vector, feeds it through
// the network, and returns the chosen throttle and steering actions.
func (b *Brain) Predict(imgBytes []byte) (throttle string, steering string, err error) {
	x, err := extractFeatures(imgBytes)
	if err != nil {
		return "", "", err
	}

	// Ensure the feature vector has the exact size expected by the net.
	if len(x) != b.inputSize {
		x = padOrTrim(x, b.inputSize)
	}

	// --- forward pass --- //

	// Hidden layer (ReLU).
	h := make([]float64, b.hiddenSize)
	for j := 0; j < b.hiddenSize; j++ {
		sum := b.b1[j]
		for i, xi := range x {
			sum += xi * b.w1[i][j]
		}
		if sum < 0 {
			sum = 0
		}
		h[j] = sum
	}

	// Throttle head.
	logitsT := make([]float64, 3)
	for k := 0; k < 3; k++ {
		sum := b.b2T[k]
		for j, hj := range h {
			sum += hj * b.w2T[j][k]
		}
		logitsT[k] = sum
	}

	// Steering head.
	logitsS := make([]float64, 3)
	for k := 0; k < 3; k++ {
		sum := b.b2S[k]
		for j, hj := range h {
			sum += hj * b.w2S[j][k]
		}
		logitsS[k] = sum
	}

	probsT := softmax(logitsT)
	probsS := softmax(logitsS)

	idxT := argmax(probsT)
	idxS := argmax(probsS)

	throttleMap := []string{"accelerate", "brake", ""}
	steeringMap := []string{"left", "right", ""}

	return throttleMap[idxT], steeringMap[idxS], nil
}

/* -------------------------------------------------------------------------- */
/*                              Helper functions                              */
/* -------------------------------------------------------------------------- */

func argmax(v []float64) int {
	idx := 0
	max := v[0]
	for i, x := range v {
		if x > max {
			max = x
			idx = i
		}
	}
	return idx
}

func softmax(v []float64) []float64 {
	max := v[0]
	for _, x := range v {
		if x > max {
			max = x
		}
	}

	expSum := 0.0
	out := make([]float64, len(v))
	for i, x := range v {
		e := math.Exp(x - max)
		out[i] = e
		expSum += e
	}

	for i := range out {
		out[i] /= expSum
	}

	return out
}

// extractFeatures decodes the JPEG, flattens it into grayscale values in the
// range [0,1].
func extractFeatures(imgBytes []byte) ([]float64, error) {
	img, err := jpeg.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	vec := make([]float64, 0, w*h)
	_ = h // silence unused variable warning
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			vec = append(vec, gray/65535.0) // normalise to [0,1]
		}
	}

	return vec, nil
}

// padOrTrim adjusts the slice to the exact size expected by the network. If
// it's too long we truncate, if it's too short we zero-pad.
func padOrTrim(v []float64, size int) []float64 {
	if len(v) == size {
		return v
	}

	if len(v) > size {
		return v[:size]
	}

	out := make([]float64, size)
	copy(out, v)
	return out
}
