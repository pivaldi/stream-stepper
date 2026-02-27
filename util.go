package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rivo/tview"
)

// Helper to read a stream
func readStream(wg *sync.WaitGroup, r io.Reader, isErr bool) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		processLine(scanner.Text(), isErr)
	}
}

// The line processor
func processLine(line string, isErr bool) {
	var color string
	var indent string

	_, after, ok := strings.Cut(line, triggerFlag)
	if ok {
		color = flagColor
		atomic.AddInt32(&currentSteps, 1)
		restOfLine := strings.TrimSpace(after)
		if restOfLine != "" {
			msgMu.Lock()
			statusMsg = colorizeLine(color, restOfLine)
			msgMu.Unlock()
		}
	} else {
		indent = indentStr
	}

	writeMu.Lock()
	defer writeMu.Unlock()
	if isErr {
		hasError = true
		color = red
	}

	//nolint:gosec // G705: TUI application, no HTML/XSS risk here
	fmt.Fprintln(tuiWriter, colorizeLine(color, indent+line))
}

func colorizeLine(color, line string) string {
	if color != "" {
		return fmt.Sprintf("[%s]%s[white]", color, line)
	}

	return line
}

// Setup the TUI
func setupTUI() {
	mainView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	mainView.SetBorder(true).SetTitle(" Logs ").SetBorderColor(tview.Styles.PrimaryTextColor)
	mainView.SetChangedFunc(func() { mainView.ScrollToEnd(); app.Draw() })

	statusView = tview.NewTextView().SetDynamicColors(true)

	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainView, 0, 1, false).
		AddItem(statusView, 1, 0, false)

	tuiWriter = tview.ANSIWriter(mainView)
}

func finish(err error) {
	finishedMu.Lock()
	defer finishedMu.Unlock()

	if finished {
		return
	}

	symbol := "✓"
	color := "green"
	doneMsg := colorizeLine(color, " Done.")
	withErrorMsg := ""
	if hasError || err != nil {
		symbol = "✗"
		color = red
		doneMsg = colorizeLine(color, " Done with errors.")
		withErrorMsg = colorizeLine(color, " with errors.")
	}

	app.QueueUpdateDraw(func() {
		format := " [%s]%s [white]%s [green]100%% [#555555]│%s [#888888]Press Ctrl+C.[white]"
		pg := ""

		if err == nil {
			pg = buildProgressBar(progressBarColor, 1.0, pbWidth)
		} else {
			var curr float64
			curr, pg = updateProgressBar(red)
			pct := int((curr / float64(totalSteps)) * 100)
			format = " [%s]%s [white]%s [green]" + strconv.Itoa(pct) + "%% "
			format += "[#555555]│[red] Process aborted.%s [#888888]Press Ctrl+C.[white]"
		}

		statusView.SetText(fmt.Sprintf(format, color, symbol, pg, doneMsg))
	})

	writeMu.Lock()
	txt := fmt.Sprintf("\n[green]--- Process Complete[white]%s[green] ---[white]", withErrorMsg)
	if err != nil {
		txt = colorizeLine(color, fmt.Sprintf("Process aborted: %v", err))
	}

	fmt.Fprintln(tuiWriter, txt)
	writeMu.Unlock()
	finished = true
}
