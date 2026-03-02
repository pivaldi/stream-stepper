package ui

// Display abstracts the TUI layer
type Display interface {
	Initialize() error
	WriteLog(text string)
	UpdateStatus(spinner, progressBar, percentage, elapsed, eta, message string)
	Run() error
	Stop()
}
