package progress

import (
	"fmt"
	"time"
)

const (
	secondsPerMinute = 60 // Number of seconds in a minute
)

// CalculateETA estimates time remaining based on linear extrapolation
// Ported from existing calculateETA function
func CalculateETA(elapsed time.Duration, currentSteps, totalSteps int32) time.Duration {
	if currentSteps == 0 {
		return 0
	}

	remainingSteps := totalSteps - currentSteps
	if remainingSteps <= 0 {
		return 0
	}

	avgTimePerStep := float64(elapsed) / float64(currentSteps)
	eta := time.Duration(avgTimePerStep * float64(remainingSteps))

	return eta
}

// FormatTime formats a duration as MM:SS
// Ported from existing formatTime function
func FormatTime(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / secondsPerMinute
	seconds := totalSeconds % secondsPerMinute

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
