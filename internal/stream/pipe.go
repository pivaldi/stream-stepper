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
	display ui.Display
	reader  io.Reader
}

// NewPipeHandler creates a handler for standard pipe mode
func NewPipeHandler(display ui.Display, reader io.Reader) *PipeHandler {
	return &PipeHandler{
		display: display,
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
			result := proc.ProcessLine(scanner.Text(), false)
			h.display.WriteLog(result.FormattedText)
		}
	}()

	wg.Wait()
	onComplete(0, nil)

	return nil
}
