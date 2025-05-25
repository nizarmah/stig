package brain

import (
	"encoding/gob"
	"fmt"
	"os"
)

// brainSerializable is an exported representation of Brain for gob encoding.
// All fields are exported so that gob can access them.
// This struct should mirror the Brain fields exactly.

type brainSerializable struct {
	InputSize  int
	HiddenSize int

	W1  [][]float64
	B1  []float64
	W2T [][]float64
	B2T []float64
	W2S [][]float64
	B2S []float64
}

func toSerializable(b *Brain) *brainSerializable {
	return &brainSerializable{
		InputSize:  b.inputSize,
		HiddenSize: b.hiddenSize,
		W1:         b.w1,
		B1:         b.b1,
		W2T:        b.w2T,
		B2T:        b.b2T,
		W2S:        b.w2S,
		B2S:        b.b2S,
	}
}

func fromSerializable(s *brainSerializable) *Brain {
	return &Brain{
		inputSize:  s.InputSize,
		hiddenSize: s.HiddenSize,
		w1:         s.W1,
		b1:         s.B1,
		w2T:        s.W2T,
		b2T:        s.B2T,
		w2S:        s.W2S,
		b2S:        s.B2S,
	}
}

// Save writes the brain weights to disk at the given path.
func (b *Brain) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open brain file: %w", err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(toSerializable(b)); err != nil {
		return fmt.Errorf("encode brain: %w", err)
	}
	return nil
}

// Load reads a brain from disk. If the file does not exist or decoding fails
// an error is returned.
func Load(path string) (*Brain, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open brain file: %w", err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	var s brainSerializable
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("decode brain: %w", err)
	}

	return fromSerializable(&s), nil
}
