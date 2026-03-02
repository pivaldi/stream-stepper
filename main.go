package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pivaldi/stream-stepper/internal/processor"
	defaultprocessor "github.com/pivaldi/stream-stepper/internal/processor/default"
	stbashprocessor "github.com/pivaldi/stream-stepper/internal/processor/stbash"
	"github.com/pivaldi/stream-stepper/internal/progress"
	"github.com/pivaldi/stream-stepper/internal/stream"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

const (
	defaultTriggerFlag = "==>"
	defaultPBWidth     = 40
	tickerIntervalMS   = 100
	percentMultiplier  = 100
)

type paramsT struct {
	stepsPtr     *int
	flagPtr      *string
	taggedPtr    *bool
	fifoPtr      *string
	pbWidthPtr   *int
	processorPtr *string
}

func main() {
	params := parseFlags()
	tracker, display := initializeComponents(*params.stepsPtr)
	var proc processor.LineProcessor
	switch *params.processorPtr {
	case "stbash":
		proc = stbashprocessor.New()
	default:
		proc = defaultprocessor.New(*params.flagPtr)
	}
	handler := selectHandler(display, tracker, *params.taggedPtr, *params.fifoPtr)
	done := make(chan struct{})
	onComplete := createCompletionCallback(tracker, display, *params.pbWidthPtr, done)

	go startTicker(display, tracker, *params.pbWidthPtr, done)
	go startHandler(handler, proc, onComplete)

	if err := display.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() paramsT {
	p := paramsT{}
	p.stepsPtr = flag.Int("steps", 0, "Required total steps for 100%")
	p.flagPtr = flag.String("flag", defaultTriggerFlag, "Trigger string for progress")
	p.taggedPtr = flag.Bool("tagged", false, "Read stdin expecting [OUT] and [ERR] prefixes")
	p.fifoPtr = flag.String("err-fifo", "", "Path to a named pipe (FIFO) to read stderr from")
	p.pbWidthPtr = flag.Int("pb-width", defaultPBWidth, "Optional progress-bar width")
	p.processorPtr = flag.String("processor", "", `Parsing processor. Only stbash supported for now:
see https://github.com/pivaldi/bash-stepper`)

	flag.Parse()

	if *p.stepsPtr <= 0 {
		fmt.Println("Error: --steps is required and must be > 0")
		os.Exit(1)
	}

	return p
}

func initializeComponents(steps int) (*progress.Tracker, ui.Display) {
	tracker := progress.NewTracker(int32(steps))
	display := ui.NewTViewDisplay()
	if err := display.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize display: %v\n", err)
		os.Exit(1)
	}

	return tracker, display
}

func selectHandler(display ui.Display, tracker *progress.Tracker, tagged bool, fifoPath string) stream.Handler {
	switch {
	case flag.NArg() > 0:
		return stream.NewExecHandler(display, tracker, flag.Arg(0))
	case tagged:
		return stream.NewTaggedHandler(display, tracker, os.Stdin)
	case fifoPath != "":
		return stream.NewFIFOHandler(display, tracker, os.Stdin, fifoPath)
	default:
		return stream.NewPipeHandler(display, tracker, os.Stdin)
	}
}

func createCompletionCallback(
	tracker *progress.Tracker,
	display ui.Display,
	pbWidth int,
	done chan struct{},
) func(int, error) {
	return func(_ int, err error) {
		close(done) // Prevents the ticker updates the status later.

		tracker.Finish()
		elapsed := tracker.GetElapsed()

		finishDisplay(display, tracker, pbWidth, elapsed, err)
	}
}

func startTicker(display ui.Display, tracker *progress.Tracker, pbWidth int, done chan struct{}) {
	ticker := time.NewTicker(tickerIntervalMS * time.Millisecond)
	defer ticker.Stop()
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := 0

	for {
		select {
		case <-ticker.C:
			updateStatus(display, tracker, pbWidth, frames[idx])
			idx = (idx + 1) % len(frames)
		case <-done:

			return
		}
	}
}

func startHandler(handler stream.Handler, proc processor.LineProcessor, onComplete func(int, error)) {
	_ = handler.Start(proc, onComplete)
}

func updateStatus(display ui.Display, tracker *progress.Tracker, pbWidth int, spinner string) {
	currentSteps := tracker.GetCurrentSteps()
	totalSteps := tracker.GetTotalSteps()
	statusMsg := tracker.GetStatusMessage()
	elapsed := tracker.GetElapsed()
	pbStatus := progress.NewStatus(true, tracker.HasError())

	progressBar := progress.BuildProgressBar(currentSteps, totalSteps, pbStatus, pbWidth)
	pct := int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	eta := progress.CalculateETA(elapsed, currentSteps, totalSteps)

	display.UpdateStatus(
		spinner,
		progressBar,
		strconv.Itoa(pct),
		progress.FormatTime(elapsed),
		progress.FormatTime(eta),
		statusMsg,
	)
}

func finishDisplay(display ui.Display, tracker *progress.Tracker, pbWidth int, elapsed time.Duration, err error) {
	hasError := tracker.HasError()
	totalSteps := tracker.GetTotalSteps()
	currentSteps := tracker.GetCurrentSteps()

	symbol := "✓"
	color := "green"
	doneMsg := fmt.Sprintf("[%s] Done.[white]", color)
	completionMsg := "\n[green]--- Process completed[white] ---"

	if hasError || err != nil {
		symbol = "✗"
		color = "red"
		doneMsg = fmt.Sprintf("[%s] Done with errors.[white]", color)
		completionMsg = "\n[red]--- Process completed with errors[white] ---"
	}

	symbol = fmt.Sprintf("[%s]%s[white]", color, symbol)

	// Build final progress bar
	progressBar := ""
	pct := percentMultiplier
	pbStatus := progress.NewStatus(false, hasError || err != nil)

	if err == nil {
		progressBar = progress.BuildProgressBar(totalSteps, totalSteps, pbStatus, pbWidth)
	} else {
		doneMsg = fmt.Sprintf("[%s]Process aborted.[white]", color)
		completionMsg = fmt.Sprintf("[%s]Process aborted: %v[white]", color, err)
		progressBar = progress.BuildProgressBar(currentSteps, totalSteps, pbStatus, pbWidth)
		pct = int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	}

	display.UpdateStatus(
		symbol,
		progressBar,
		strconv.Itoa(pct),
		progress.FormatTime(elapsed),
		"00:00",
		doneMsg,
	)

	display.WriteLog(completionMsg)
}
