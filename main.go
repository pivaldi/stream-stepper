package main

import (
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
	"github.com/spf13/cobra"
)

const (
	defaultTriggerFlag = "==>"
	defaultPBWidth     = 40
	tickerIntervalMS   = 100
	percentMultiplier  = 100
)

var (
	steps         int
	triggerFlag   string
	tagged        bool
	wait          int
	errFifo       string
	pbWidth       int
	processorType string

	rootCmd = &cobra.Command{
		Use:   "stream-stepper [flags] [command]",
		Short: "A CLI tool that parses shell command output and renders it in a TUI with a dynamic progress bar",
		Long: `StreamStepper intercepts stdout/stderr, detects trigger strings (default: "==>"),
increments a progress bar, extracts status messages, and colorizes output.

Examples:
  # Execute a command and monitor progress
  stream-stepper --steps 7 ./examples/deploy.sh

  # Read from stdin with tagged output
  ./your-script.sh | stream-stepper --steps 5 --tagged

  # Use a FIFO for stderr
  stream-stepper --steps 10 --err-fifo /tmp/stderr.fifo

  # Use stbash processor for bash-stepper output
  stream-stepper --steps 10 --processor stbash ./script.sh`,
		Args: cobra.MaximumNArgs(1),
		Run:  runStreamStepper,
	}
)

func initFlags() {
	rootCmd.Flags().IntVarP(&steps, "steps", "s", 0, "Total steps for 100% progress (required)")
	rootCmd.Flags().IntVarP(&wait, "wait", "", -1, "Wait time in seconds before exiting")
	rootCmd.Flags().StringVarP(&triggerFlag, "flag", "f", defaultTriggerFlag, "Trigger string for progress detection")
	rootCmd.Flags().BoolVarP(&tagged, "tagged", "t", false, "Read stdin expecting [OUT] and [ERR] prefixes")
	rootCmd.Flags().StringVar(&errFifo, "err-fifo", "", "Path to a named pipe (FIFO) to read stderr from")
	rootCmd.Flags().IntVarP(&pbWidth, "pb-width", "w", defaultPBWidth, "Progress bar width in characters")
	rootCmd.Flags().StringVarP(&processorType, "processor", "p", "",
		"Parsing processor (only 'stbash' supported, see https://github.com/pivaldi/bash-stepper)")

	// Mark steps as required
	if err := rootCmd.MarkFlagRequired("steps"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	initFlags()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStreamStepper(_ *cobra.Command, args []string) {
	if steps <= 0 {
		fmt.Fprintf(os.Stderr, "Error: --steps must be greater than 0\n")
		os.Exit(1)
	}

	tui := ui.New(initializeComponents(steps))

	var proc processor.LineProcessor
	switch processorType {
	case "stbash":
		proc = stbashprocessor.New(tui.Display.Escape)
	default:
		proc = defaultprocessor.New(tui.Display.Escape, triggerFlag)
	}

	handler := selectHandler(tui, tagged, errFifo, args)
	done := make(chan struct{})
	onComplete := createCompletionCallback(tui, pbWidth, wait, done)

	go startTicker(tui, pbWidth, done)
	go startHandler(handler, proc, onComplete)

	if err := tui.Display.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}

func initializeComponents(steps int) (ui.Display, *progress.Tracker) {
	tracker := progress.NewTracker(int32(steps))
	display := ui.NewTViewDisplay()
	if err := display.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize display: %v\n", err)
		os.Exit(1)
	}

	return display, tracker
}

func selectHandler(tui ui.TUI, tagged bool, fifoPath string, args []string) stream.Handler {
	switch {
	case len(args) > 0:
		return stream.NewExecHandler(tui, args[0])
	case tagged:
		return stream.NewTaggedHandler(tui, os.Stdin)
	case fifoPath != "":
		return stream.NewFIFOHandler(tui, os.Stdin, fifoPath)
	default:
		return stream.NewPipeHandler(tui, os.Stdin)
	}
}

func createCompletionCallback(
	tui ui.TUI,
	pbWidth int,
	wait int,
	done chan struct{},
) func(int, error) {
	return func(_ int, err error) {
		close(done) // Prevents the ticker updates the status later.

		tui.Tracker.Finish()
		elapsed := tui.Tracker.GetElapsed()

		finishDisplay(tui, pbWidth, wait, elapsed, err)
	}
}

func startTicker(tui ui.TUI, pbWidth int, done chan struct{}) {
	ticker := time.NewTicker(tickerIntervalMS * time.Millisecond)
	defer ticker.Stop()
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := 0

	for {
		select {
		case <-ticker.C:
			updateStatus(tui, pbWidth, frames[idx])
			idx = (idx + 1) % len(frames)
		case <-done:

			return
		}
	}
}

func startHandler(handler stream.Handler, proc processor.LineProcessor, onComplete func(int, error)) {
	_ = handler.Start(proc, onComplete)
}

func updateStatus(tui ui.TUI, pbWidth int, spinner string) {
	currentSteps := tui.Tracker.GetCurrentSteps()
	totalSteps := tui.Tracker.GetTotalSteps()
	statusMsg := tui.Tracker.GetStatusMessage()
	elapsed := tui.Tracker.GetElapsed()
	pbStatus := progress.NewStatus(true, tui.Tracker.HasError())

	progressBar := progress.BuildProgressBar(currentSteps, totalSteps, pbStatus, pbWidth)
	pct := int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	eta := progress.CalculateETA(elapsed, currentSteps, totalSteps)

	tui.Display.UpdateStatus(
		spinner,
		progressBar,
		strconv.Itoa(pct),
		progress.FormatTime(elapsed),
		progress.FormatTime(eta),
		statusMsg,
	)
}

func finishDisplay(tui ui.TUI, pbWidth, wait int, elapsed time.Duration, err error) {
	hasError := err != nil || tui.Tracker.HasError()
	totalSteps := tui.Tracker.GetTotalSteps()
	currentSteps := tui.Tracker.GetCurrentSteps()

	symbol := "✓"
	color := "green"
	doneMsg := fmt.Sprintf("[%s] Done.[white]", color)
	completionMsg := "\n[green]--- Process completed[white] ---"

	if hasError {
		symbol = "✗"
		color = "red"
		doneMsg = fmt.Sprintf("[%s] Done with errors.[white]", color)
		completionMsg = "\n[red]--- Process completed with errors[white] ---"
	}

	symbol = fmt.Sprintf("[%s]%s[white]", color, symbol)

	// Build final progress bar
	progressBar := ""
	pct := percentMultiplier
	pbStatus := progress.NewStatus(false, hasError)

	if err == nil {
		progressBar = progress.BuildProgressBar(totalSteps, totalSteps, pbStatus, pbWidth)
	} else {
		doneMsg = fmt.Sprintf("[%s]Process aborted.[white]", color)
		completionMsg = fmt.Sprintf("[%s]Process aborted: %v[white]", color, err)
		progressBar = progress.BuildProgressBar(currentSteps, totalSteps, pbStatus, pbWidth)
		pct = int((float64(currentSteps) / float64(totalSteps)) * percentMultiplier)
	}

	tui.Display.UpdateStatus(
		symbol,
		progressBar,
		strconv.Itoa(pct),
		progress.FormatTime(elapsed),
		"00:00",
		doneMsg,
	)

	tui.Display.WriteLog(completionMsg)
	if wait >= 0 {
		time.Sleep(time.Duration(wait) * time.Second)
		tui.Display.Stop()
		if hasError {
			os.Exit(1)
		}
	}
}
