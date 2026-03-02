package progress

import (
	"sync"
	"time"
)

// Tracker encapsulates all progress tracking state with thread-safe access
type Tracker struct {
	totalSteps    int32
	currentSteps  int32
	startTime     time.Time
	endTime       time.Time
	hasError      bool
	statusMessage string
	mu            sync.RWMutex
}

// NewTracker creates a new progress tracker
func NewTracker(totalSteps int32) *Tracker {
	return &Tracker{
		totalSteps:    totalSteps,
		currentSteps:  0,
		startTime:     time.Now(),
		statusMessage: "Waiting for data...",
	}
}

// IncrementStep increments the progress counter and updates status message
func (t *Tracker) IncrementStep(statusMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentSteps++
	if statusMsg != "" {
		t.statusMessage = statusMsg
	}
}

// GetCurrentSteps returns the current step count
func (t *Tracker) GetCurrentSteps() int32 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.currentSteps
}

// GetTotalSteps returns the total step count
func (t *Tracker) GetTotalSteps() int32 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.totalSteps
}

// GetStatusMessage returns the current status message
func (t *Tracker) GetStatusMessage() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.statusMessage
}

// SetError marks that an error has occurred
func (t *Tracker) SetError() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hasError = true
}

// HasError returns whether an error has occurred
func (t *Tracker) HasError() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.hasError
}

// GetStartTime returns when tracking started
func (t *Tracker) GetStartTime() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.startTime
}

// Finish marks the tracking as complete
func (t *Tracker) Finish() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.endTime = time.Now()
}

// GetElapsed returns elapsed time (uses endTime if finished, otherwise calculates from now)
func (t *Tracker) GetElapsed() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.endTime.IsZero() {
		return t.endTime.Sub(t.startTime)
	}

	return time.Since(t.startTime)
}
