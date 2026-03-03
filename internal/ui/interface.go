package ui

import "github.com/pivaldi/stream-stepper/internal/progress"

// Display abstracts the TUI layer
type Display interface {
	Initialize() error
	WriteLog(text string)
	UpdateStatus(spinner, progressBar, percentage, elapsed, eta, message string)
	Run() error
	Stop()
	Escape(text string) string
}

type TUI struct {
	Display Display
	Tracker *progress.Tracker
}

func New(display Display, tracker *progress.Tracker) TUI {
	return TUI{
		Display: display,
		Tracker: tracker,
	}
}
