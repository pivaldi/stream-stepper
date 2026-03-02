package stream

import (
	"github.com/pivaldi/stream-stepper/internal/processor"
)

// Handler abstracts different input modes (exec, pipe, fifo, tagged)
type Handler interface {
	Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error
}
