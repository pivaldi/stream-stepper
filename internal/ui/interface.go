package ui

// Display abstracts the TUI layer
type Display interface {
	Initialize() error
	WriteLog(text string) error
	UpdateStatus(spinner, progressBar, percentage, elapsed, eta, message string) error
	Run() error
	Stop()
}
