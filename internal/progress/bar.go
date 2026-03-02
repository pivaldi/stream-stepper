package progress

import (
	"fmt"
	"strings"
)

type Status int

const (
	defaultColor            = "#33E5FF"
	fractionalSteps         = 8 // Number of sub-character progress steps for smooth rendering
	StatusProcessing Status = iota
	StatusError
	StatusOK
)

func NewStatus(progressing, hasError bool) Status {
	if hasError {
		return StatusError
	}

	if progressing {
		return StatusProcessing
	}

	return StatusOK
}

func (s Status) color() string {
	switch s {
	case StatusOK:
		return "green"
	case StatusError:
		return "red"
	default:
		return defaultColor
	}
}

// BuildProgressBar creates a Unicode progress bar string
// Ported from existing buildProgressBar and currentProgressBar functions
func BuildProgressBar(currentSteps, totalSteps int32, status Status, width int) string {
	fractions := []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}
	nfractions := len(fractions)

	color := status.color()

	// Calculate progress ratio
	progress := float64(currentSteps) / float64(totalSteps)
	if currentSteps > totalSteps {
		progress = 1.0
	}

	filledLen := int(progress * float64(width))

	if filledLen >= width {
		return fmt.Sprintf("[%s]%s", color, strings.Repeat("█", width))
	}

	fractionIdx := int(((progress * float64(width)) - float64(filledLen)) * fractionalSteps)
	if fractionIdx < 0 {
		fractionIdx = 0
	} else if fractionIdx >= nfractions {
		fractionIdx = nfractions - 1
	}

	fractionStr := fractions[fractionIdx]
	emptyLen := width - filledLen
	if fractionIdx > 0 {
		emptyLen--
	}
	if emptyLen < 0 {
		emptyLen = 0
	}

	format := "[%s]%s[#3388FF]%s[#444444]%s"

	return fmt.Sprintf(format, color, strings.Repeat("█", filledLen), fractionStr, strings.Repeat("─", emptyLen))
}
