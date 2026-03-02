package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pivaldi/stream-stepper/internal/processor"
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

func main() {
	stepsPtr, flagPtr, taggedPtr, fifoPtr, pbWidthPtr := parseFlags()
	tracker, display := initializeComponents(*stepsPtr)
	proc := processor.New(*flagPtr, tracker)
	handler := selectHandler(display, *taggedPtr, *fifoPtr)
	done := make(chan struct{})
	onComplete := createCompletionCallback(tracker, display, *pbWidthPtr, done)

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		defer display.Stop()
		os.Exit(0)
	}()

	startTicker(display, tracker, *pbWidthPtr, done)
	startHandler(handler, proc, onComplete)

	if err := display.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() (stepsPtr *int, flagPtr *string, taggedPtr *bool, fifoPtr *string, pbWidthPtr *int) {
	stepsPtr = flag.Int("steps", 0, "Required total steps for 100%")
	flagPtr = flag.String("flag", defaultTriggerFlag, "Trigger string for progress")
	taggedPtr = flag.Bool("tagged", false, "Read stdin expecting [OUT] and [ERR] prefixes")
	fifoPtr = flag.String("err-fifo", "", "Path to a named pipe (FIFO) to read stderr from")
	pbWidthPtr = flag.Int("pb-width", defaultPBWidth, "Optional progress-bar width")
	flag.Parse()

	if *stepsPtr <= 0 {
		fmt.Println("Error: --steps is required and must be > 0")
		os.Exit(1)
	}

	return stepsPtr, flagPtr, taggedPtr, fifoPtr, pbWidthPtr
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

func selectHandler(display ui.Display, tagged bool, fifoPath string) stream.Handler {
	switch {
	case flag.NArg() > 0:
		return stream.NewExecHandler(display, flag.Arg(0))
	case tagged:
		return stream.NewTaggedHandler(display, os.Stdin)
	case fifoPath != "":
		return stream.NewFIFOHandler(display, os.Stdin, fifoPath)
	default:
		return stream.NewPipeHandler(display, os.Stdin)
	}
}

func createCompletionCallback(
	tracker *progress.Tracker,
	display ui.Display,
	pbWidth int,
	done chan struct{},
) func(int, error) {
	return func(_ int, err error) {
		tracker.Finish()
		elapsed := tracker.GetElapsed()

		// Queue the final display update
		if tvd, ok := display.(*ui.TViewDisplay); ok {
			tvd.QueueUpdate(func() {
				finishDisplay(display, tracker, pbWidth, elapsed, err)
			})
		} else {
			log.Fatal("unexpectd error: can not cast display to ui.TViewDisplay")
		}

		close(done)
	}
}

func startTicker(display ui.Display, tracker *progress.Tracker, pbWidth int, done chan struct{}) {
	go func() {
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
				updateStatus(display, tracker, pbWidth, "✓")

				return
			}
		}
	}()
}

func startHandler(handler stream.Handler, proc processor.LineProcessor, onComplete func(int, error)) {
	go func() {
		_ = handler.Start(proc, onComplete)
	}()
}

func updateStatus(display ui.Display, tracker *progress.Tracker, pbWidth int, spinner string) {
	currentSteps := tracker.GetCurrentSteps()
	totalSteps := tracker.GetTotalSteps()
	hasError := tracker.HasError()
	statusMsg := tracker.GetStatusMessage()
	elapsed := tracker.GetElapsed()

	progressBar := progress.BuildProgressBar(currentSteps, totalSteps, hasError, pbWidth)
	pct := int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	eta := progress.CalculateETA(elapsed, currentSteps, totalSteps)

	_ = display.UpdateStatus(
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
	if err == nil {
		progressBar = progress.BuildProgressBar(totalSteps, totalSteps, hasError, pbWidth)
	} else {
		doneMsg = fmt.Sprintf("[%s]Process aborted.[white]", color)
		completionMsg = fmt.Sprintf("[%s]Process aborted: %v[white]", color, err)
		progressBar = progress.BuildProgressBar(currentSteps, totalSteps, true, pbWidth)
		pct = int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	}

	// Cast to TViewDisplay to use QueueUpdate
	if tvd, ok := display.(*ui.TViewDisplay); ok {
		tvd.QueueUpdate(func() {
			_ = display.UpdateStatus(
				symbol,
				progressBar,
				strconv.Itoa(pct),
				progress.FormatTime(elapsed),
				"00:00",
				doneMsg,
			)
		})
	}

	_ = display.WriteLog(completionMsg)
	fmt.Fprintf(os.Stderr, "[FINISH] Completion message written: %s\n", completionMsg)
}
