package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
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
		_, _, ok := strings.Cut(line, "** ")
		if ok {
			color = "yellow"
		} else {
			indent = indentStr
		}
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

	symbol = colorizeLine(color, symbol)

	app.QueueUpdateDraw(func() {
		pg := ""
		pct := 100

		if err == nil {
			pg = buildProgressBar(color, 1.0, pbWidth)
		} else {
			doneMsg = colorizeLine(color, "Process aborted.")
			var curr float64
			curr, pg = currentProgressBar(color)
			pct = int((curr / float64(totalSteps)) * 100)
		}

		setStatusView(symbol, pg, pct, doneMsg)
	})

	writeMu.Lock()
	txt := fmt.Sprintf("\n[green]--- Process completed[white]%s[green] ---[white]", withErrorMsg)
	if err != nil {
		txt = colorizeLine(color, fmt.Sprintf("Process aborted: %v", err))
	}

	fmt.Fprintln(tuiWriter, txt)
	writeMu.Unlock()
	finished = true
}
