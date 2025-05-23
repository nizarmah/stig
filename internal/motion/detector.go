// Package motion provides motion detection from screenshots.
package motion

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"sync"
	"time"
)

// Detector tracks motion between consecutive screenshots.
type Detector struct {
	mu sync.Mutex

	// Previous frame data
	prevFrame []byte
	prevTime  time.Time

	// Motion tracking
	lastMotionTime time.Time
	totalDistance  float64
	frameCount     int

	// Configuration
	threshold float64 // Minimum difference to consider as motion
	debug     bool
}

// NewDetector creates a new motion detector.
func NewDetector(threshold float64, debug bool) *Detector {
	return &Detector{
		threshold:      threshold,
		debug:          debug,
		lastMotionTime: time.Now(),
	}
}

// ProcessFrame analyzes a new frame and returns the motion score.
func (d *Detector) ProcessFrame(frameData []byte) (float64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	currentTime := time.Now()

	// If this is the first frame, just store it
	if d.prevFrame == nil {
		d.prevFrame = frameData
		d.prevTime = currentTime
		d.frameCount++
		return 0, nil
	}

	// Calculate motion between frames
	motion, err := d.calculateMotion(d.prevFrame, frameData)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate motion: %w", err)
	}

	// Update tracking
	d.frameCount++
	if motion > d.threshold {
		d.lastMotionTime = currentTime
		d.totalDistance += motion
	}

	// Calculate time since last frame
	timeDelta := currentTime.Sub(d.prevTime).Seconds()

	if d.debug {
		log.Printf("Motion: %.4f (threshold: %.4f), Time delta: %.3fs, Total distance: %.2f, Frames: %d",
			motion, d.threshold, timeDelta, d.totalDistance, d.frameCount)
	}

	// Store current frame for next comparison
	d.prevFrame = frameData
	d.prevTime = currentTime

	return motion, nil
}

// calculateMotion computes the motion score between two frames.
func (d *Detector) calculateMotion(prev, curr []byte) (float64, error) {
	// Decode previous frame
	prevImg, err := jpeg.Decode(bytes.NewReader(prev))
	if err != nil {
		return 0, fmt.Errorf("failed to decode previous frame: %w", err)
	}

	// Decode current frame
	currImg, err := jpeg.Decode(bytes.NewReader(curr))
	if err != nil {
		return 0, fmt.Errorf("failed to decode current frame: %w", err)
	}

	// Calculate difference focusing on the center region (where the road is)
	motion := d.calculateCenterWeightedDifference(prevImg, currImg)

	return motion, nil
}

// calculateCenterWeightedDifference calculates motion with more weight on the center.
func (d *Detector) calculateCenterWeightedDifference(prev, curr image.Image) float64 {
	bounds := prev.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Focus on the center 60% horizontally and bottom 70% vertically (where the road is)
	xStart := int(float64(width) * 0.2)
	xEnd := int(float64(width) * 0.8)
	yStart := int(float64(height) * 0.3)
	yEnd := height

	totalDiff := 0.0
	pixelCount := 0

	// Sample every 4th pixel for efficiency
	for y := yStart; y < yEnd; y += 4 {
		for x := xStart; x < xEnd; x += 4 {
			// Get pixel values
			pr1, pg1, pb1, _ := prev.At(x, y).RGBA()
			pr2, pg2, pb2, _ := curr.At(x, y).RGBA()

			// Calculate RGB difference
			dr := float64(pr1) - float64(pr2)
			dg := float64(pg1) - float64(pg2)
			db := float64(pb1) - float64(pb2)

			// Euclidean distance in RGB space
			diff := math.Sqrt(dr*dr+dg*dg+db*db) / 65535.0 // Normalize to [0,1]

			// Apply center weighting (stronger weight towards center and bottom)
			centerX := float64(width) / 2
			centerWeight := 1.0 - math.Abs(float64(x)-centerX)/centerX*0.5
			bottomWeight := float64(y-yStart) / float64(yEnd-yStart)
			weight := centerWeight * (0.5 + 0.5*bottomWeight)

			totalDiff += diff * weight
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0
	}

	// Average motion score
	return totalDiff / float64(pixelCount)
}

// GetTotalDistance returns the accumulated motion distance.
func (d *Detector) GetTotalDistance() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.totalDistance
}

// GetTimeSinceLastMotion returns the duration since motion was last detected.
func (d *Detector) GetTimeSinceLastMotion() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return time.Since(d.lastMotionTime)
}

// Reset clears the motion detector state.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.prevFrame = nil
	d.prevTime = time.Time{}
	d.lastMotionTime = time.Now()
	d.totalDistance = 0
	d.frameCount = 0
}

// GetStats returns current motion statistics.
func (d *Detector) GetStats() (totalDistance float64, frameCount int, timeSinceMotion time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.totalDistance, d.frameCount, time.Since(d.lastMotionTime)
}
