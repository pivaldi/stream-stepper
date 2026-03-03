package defaultprocessor

import (
	"fmt"
	"strings"

	"github.com/pivaldi/stream-stepper/internal/processor"
)

const (
	indentStr = "\t"
	flagColor = "blue"
	redColor  = "red"
)

// Processor implements LineProcessor interface
type Processor struct {
	escape      func(string) string
	triggerFlag string
}

func (p *Processor) Escape(text string) string {
	return p.escape(text)
}

// New creates a new line processor
func New(escape func(string) string, triggerFlag string) *Processor {
	return &Processor{
		escape:      escape,
		triggerFlag: triggerFlag,
	}
}

// ProcessLine examines a line for triggers, applies formatting, and returns result
func (p *Processor) ProcessLine(line string, isStderr bool) processor.ProcessedLine {
	result := processor.ProcessedLine{}

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
			result.StatusMessage = colorizeLine(color, p.Escape(restOfLine))
		}

		result.FormattedText = result.StatusMessage

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
		result.IsError = true
		color = redColor
	}

	result.FormattedText = colorizeLine(color, p.Escape(indent+line))

	return result
}

func colorizeLine(color, line string) string {
	if color != "" {
		return fmt.Sprintf("[%s]%s[white]", color, line)
	}

	return line
}
