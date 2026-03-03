package stream

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pivaldi/stream-stepper/internal/processor"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

const (
	streamCount = 2 // Number of input streams to read (stdin and FIFO)
)

// FIFOHandler handles FIFO pipe input mode (stderr from named pipe, stdout from stdin)
type FIFOHandler struct {
	tui      ui.TUI
	stdin    io.Reader
	fifoPath string
}

// NewFIFOHandler creates a handler for FIFO mode
func NewFIFOHandler(tui ui.TUI, stdin io.Reader, fifoPath string) *FIFOHandler {
	return &FIFOHandler{
		tui:      tui,
		stdin:    stdin,
		fifoPath: fifoPath,
	}
}

// Start begins reading from stdin and FIFO
func (h *FIFOHandler) Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error {
	// Cast to access SetTitle method
	if tvd, ok := h.tui.Display.(*ui.TViewDisplay); ok {
		tvd.SetTitle(" Pipe: Stdin + FIFO ")
	}

	var wg sync.WaitGroup
	wg.Add(streamCount)

	// Read stdout from stdin
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(h.stdin)
		for scanner.Scan() {
			proc.ProcessLine(scanner.Text(), false).Trigger(h.tui.Display, h.tui.Tracker)
		}
	}()

	// Read stderr from FIFO
	go func() {
		defer wg.Done()
		file, err := os.Open(h.fifoPath)
		if err != nil {
			proc.ProcessLine(fmt.Sprintf("Failed to open FIFO: %v", err), true).Trigger(h.tui.Display, h.tui.Tracker)

			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			result := proc.ProcessLine(scanner.Text(), true)
			h.tui.Display.WriteLog(result.FormattedText)
		}
	}()

	wg.Wait()
	onComplete(0, nil)

	return nil
}
