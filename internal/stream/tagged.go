package stream

import (
	"bufio"
	"io"
	"strings"

	"github.com/pivaldi/stream-stepper/internal/processor"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

// TaggedHandler handles tagged pipe input mode ([OUT] and [ERR] prefixes)
type TaggedHandler struct {
	display ui.Display
	reader  io.Reader
}

// NewTaggedHandler creates a handler for tagged pipe mode
func NewTaggedHandler(display ui.Display, reader io.Reader) *TaggedHandler {
	return &TaggedHandler{
		display: display,
		reader:  reader,
	}
}

// Start begins reading from the tagged pipe
func (h *TaggedHandler) Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error {
	// Cast to access SetTitle method
	if tvd, ok := h.display.(*ui.TViewDisplay); ok {
		tvd.SetTitle(" Pipe: Tagged Stdin ")
	}

	scanner := bufio.NewScanner(h.reader)
	var isErr bool

	for scanner.Scan() {
		line := scanner.Text()

		if after, ok := strings.CutPrefix(line, "[ERR] "); ok {
			line = after
			isErr = true
		} else {
			line = strings.TrimPrefix(line, "[OUT] ")
			isErr = false
		}

		result := proc.ProcessLine(line, isErr)
		h.display.WriteLog(result.FormattedText)
	}

	onComplete(0, nil)

	return nil
}
