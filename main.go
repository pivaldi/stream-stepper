package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rivo/tview"
)

const (
	progressBarColor   = "#33E5FF"
	defaultTriggerFlag = "==>"
	defaultPBWidth     = 50
)

var (
	writeMu      sync.Mutex // Protects tview writer from concurrent goroutines
	msgMu        sync.Mutex
	finishedMu   sync.Mutex
	finished     bool
	mainView     *tview.TextView
	statusView   *tview.TextView
	layout       *tview.Flex
	tuiWriter    io.Writer
	app                = tview.NewApplication()
	done               = make(chan struct{})
	statusMsg          = "Waiting for data..."
	currentSteps int32 = 0
	indentStr          = "\t"
	flagColor          = "blue"
	hasError     bool
	red          = "red"
	triggerFlag  string
	totalSteps   int32
)

func main() {
	stepsPtr := flag.Int("steps", 0, "Required total steps for 100%")
	flagPtr := flag.String("flag", defaultTriggerFlag, "Trigger string for progress")
	taggedPtr := flag.Bool("tagged", false, "Read stdin expecting [OUT] and [ERR] prefixes")
	fifoPtr := flag.String("err-fifo", "", "Path to a named pipe (FIFO) to read stderr from")
	flag.IntVar(&pbWidth, "pb-width", defaultPBWidth, "Optional progress-bar width")
	flag.Parse()

	if *stepsPtr <= 0 {
		fmt.Println("Error: --steps is required and must be > 0")
		os.Exit(1)
	}

	totalSteps = int32(*stepsPtr)
	triggerFlag = *flagPtr

	setupTUI()

	// --- 4. Ticker (Spinner & Progress) ---
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := 0

		for {
			select {
			case <-ticker.C:
				msgMu.Lock()
				msg := statusMsg
				msgMu.Unlock()

				app.QueueUpdateDraw(func() {
					curr, bar := currentProgressBar(progressBarColor)
					pct := int((curr / float64(totalSteps)) * 100)
					idx = (idx + 1) % len(frames)
					statusView.SetText(fmt.Sprintf(" [#00E5FF]%s [white]%s %3d%% [#555555]│[white] %s", frames[idx], bar, pct, msg))
				})

			case <-done:
				finish(nil)
				return
			}
		}
	}()

	// Data Routing Modes ---
	go func() {
		defer close(done)
		var wg sync.WaitGroup

		// MODE 1: EXEC MODE
		if flag.NArg() > 0 {
			cmdStr := flag.Arg(0)
			handleExec(&wg, cmdStr)
			// MODE 2: TAGGED PIPE MODE
		} else if *taggedPtr {
			handleTagged()

			// MODE 3: FIFO PIPE MODE
		} else if *fifoPtr != "" {
			handleFIFO(&wg, *fifoPtr)
			// FALLBACK: Standard Pipe Mode
		} else {
			mainView.SetTitle(" Pipe: Standard Stdin ")
			wg.Add(1)
			go readStream(&wg, os.Stdin, false)
			wg.Wait()
		}
	}()

	// Start UI
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
