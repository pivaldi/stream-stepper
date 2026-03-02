package processor

import (
	"fmt"
	"strings"

	"github.com/pivaldi/stream-stepper/internal/progress"
)

const (
	indentStr = "\t"
	flagColor = "blue"
	redColor  = "red"
)

// Processor implements LineProcessor interface
type Processor struct {
	triggerFlag string
	tracker     *progress.Tracker
}

// New creates a new line processor
func New(triggerFlag string, tracker *progress.Tracker) *Processor {
	return &Processor{
		triggerFlag: triggerFlag,
		tracker:     tracker,
	}
}

// ProcessLine examines a line for triggers, applies formatting, and returns result
func (p *Processor) ProcessLine(line string, isStderr bool) ProcessedLine {
	result := ProcessedLine{
		FormattedText:  line,
		IsProgressStep: false,
		StatusMessage:  "",
	}

	var color string
	var indent string

	// Check for trigger flag
	_, after, hasTrigger := strings.Cut(line, p.triggerFlag)
	if hasTrigger {
		color = flagColor
		result.IsProgressStep = true

		// Extract status message after trigger
		restOfLine := strings.TrimSpace(after)
		if restOfLine != "" {
			result.StatusMessage = colorizeLine(color, restOfLine)
		}

		// Increment progress in tracker
		p.tracker.IncrementStep(result.StatusMessage)
	} else {
		// Check for "** " prefix (yellow highlight)
		if strings.HasPrefix(line, "** ") {
			color = "yellow"
		} else {
			// Normal lines get indented
			indent = indentStr
		}
	}

	// Stderr lines are red
	if isStderr {
		p.tracker.SetError()
		color = redColor
	}

	result.FormattedText = colorizeLine(color, indent+line)

	return result
}

func colorizeLine(color, line string) string {
	if color != "" {
		return fmt.Sprintf("[%s]%s[white]", color, line)
	}

	return line
}
