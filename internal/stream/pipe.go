package stream

import (
	"bufio"
	"io"
	"sync"

	"github.com/pivaldi/stream-stepper/internal/processor"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

// PipeHandler handles standard pipe input mode
type PipeHandler struct {
	tui    ui.TUI
	reader io.Reader
}

// NewPipeHandler creates a handler for standard pipe mode
func NewPipeHandler(tui ui.TUI, reader io.Reader) *PipeHandler {
	return &PipeHandler{
		tui:    tui,
		reader: reader,
	}
}

// Start begins reading from the pipe
func (h *PipeHandler) Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error {
	// Cast to access SetTitle method
	if tvd, ok := h.tui.Display.(*ui.TViewDisplay); ok {
		tvd.SetTitle(" Pipe: Standard Stdin ")
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		scanner := bufio.NewScanner(h.reader)
		for scanner.Scan() {
			proc.ProcessLine(scanner.Text(), false).Trigger(h.tui.Display, h.tui.Tracker)
		}
	})

	wg.Wait()
	onComplete(0, nil)

	return nil
}
