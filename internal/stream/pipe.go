package stream

import (
	"bufio"
	"io"
	"sync"

	"github.com/pivaldi/stream-stepper/internal/processor"
	"github.com/pivaldi/stream-stepper/internal/progress"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

// PipeHandler handles standard pipe input mode
type PipeHandler struct {
	display ui.Display
	tracker *progress.Tracker
	reader  io.Reader
}

// NewPipeHandler creates a handler for standard pipe mode
func NewPipeHandler(display ui.Display, tracker *progress.Tracker, reader io.Reader) *PipeHandler {
	return &PipeHandler{
		display: display,
		tracker: tracker,
		reader:  reader,
	}
}

// Start begins reading from the pipe
func (h *PipeHandler) Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error {
	// Cast to access SetTitle method
	if tvd, ok := h.display.(*ui.TViewDisplay); ok {
		tvd.SetTitle(" Pipe: Standard Stdin ")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(h.reader)
		for scanner.Scan() {
			proc.ProcessLine(scanner.Text(), false).Trigger(h.display, h.tracker)
		}
	}()

	wg.Wait()
	onComplete(0, nil)

	return nil
}
