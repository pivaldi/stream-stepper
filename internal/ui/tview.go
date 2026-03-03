package ui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TViewDisplay implements Display using tview library
type TViewDisplay struct {
	app        *tview.Application
	mainView   *tview.TextView
	statusView *tview.TextView
	layout     *tview.Flex
	tuiWriter  io.Writer
	writeMu    sync.Mutex
}

// NewTViewDisplay creates a new tview-based display
func NewTViewDisplay() *TViewDisplay {
	return &TViewDisplay{
		app: tview.NewApplication(),
	}
}

// Initialize sets up the TUI layout
func (d *TViewDisplay) Initialize() error {
	d.mainView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	d.mainView.SetBorder(true).SetTitle(" Logs ").SetBorderColor(tview.Styles.PrimaryTextColor)
	d.mainView.SetChangedFunc(func() {
		d.mainView.ScrollToEnd()
		d.app.Draw()
	})

	// Set a global input capture function
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ, tcell.KeyCtrlC:
			d.Stop()

			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' || event.Rune() == 'Q' {
				d.Stop()
				os.Exit(1)
			}
		}

		// Return the original event to allow normal processing
		return event
	})

	d.statusView = tview.NewTextView().SetDynamicColors(true)

	d.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.mainView, 0, 1, false).
		AddItem(d.statusView, 1, 0, false)

	d.tuiWriter = tview.ANSIWriter(d.mainView)

	return nil
}

// WriteLog appends text to the main scrollable view
func (d *TViewDisplay) WriteLog(text string) {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	fmt.Fprintln(d.tuiWriter, text)
}

// UpdateStatus updates the status bar with progress information
func (d *TViewDisplay) UpdateStatus(spinner, progressBar, percentage, elapsed, eta, message string) {
	d.app.QueueUpdateDraw(func() {
		symbFmt := " %s "
		barFmt := " %s "
		pctFmt := "[#dcdccc]%s%%[white]"
		sepFmt := " [#555555]│[white] "
		timeFmt := "[#dcdccc]%s/%s[white]"
		msgFmt := "%s"
		ctrlFmt := " [#888888]Press Ctrl+C (exit 0) or q/Q (exit 1)[white]"
		format := symbFmt + barFmt + pctFmt + sepFmt + timeFmt + sepFmt + msgFmt + ctrlFmt

		d.statusView.SetText(fmt.Sprintf(format, spinner, progressBar, percentage, elapsed, eta, message))
	})
}

// Run starts the tview event loop (blocking)
func (d *TViewDisplay) Run() error {
	if err := d.app.SetRoot(d.layout, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("error running display: %w", err)
	}

	return nil
}

func (d *TViewDisplay) Escape(test string) string {
	return tview.Escape(test)
}

// Stop gracefully stops the application
func (d *TViewDisplay) Stop() {
	logContent := d.mainView.GetText(true)
	d.app.Stop()

	fmt.Println(logContent)
}

// SetTitle sets the main view title
func (d *TViewDisplay) SetTitle(title string) {
	d.mainView.SetTitle(title)
}

// QueueUpdate queues a function to run on the UI thread
func (d *TViewDisplay) QueueUpdate(fn func()) {
	d.app.QueueUpdateDraw(fn)
}
