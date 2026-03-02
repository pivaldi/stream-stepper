package processor

import (
	"github.com/pivaldi/stream-stepper/internal/progress"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

// LineProcessor processes input lines and detects progress triggers
type LineProcessor interface {
	ProcessLine(line string, isStderr bool) ProcessedLine
}

// ProcessedLine represents a processed line with metadata
type ProcessedLine struct {
	FormattedText  string // The text to display (with color codes, indentation)
	IsProgressStep bool   // Whether this line triggered a progress step
	StatusMessage  string // Extracted status message (if IsProgressStep is true)
	IsError        bool
}

func (pl ProcessedLine) Trigger(display ui.Display, tracker *progress.Tracker) {
	display.WriteLog(pl.FormattedText)
	if pl.IsError {
		tracker.SetError()
	}

	if pl.IsProgressStep {
		tracker.IncrementStep(pl.StatusMessage)
	}
}
