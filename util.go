package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

func formatTime(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func calculateETA(elapsed time.Duration, currentSteps, totalSteps int32) time.Duration {
	if currentSteps == 0 {
		return 0
	}

	remainingSteps := totalSteps - currentSteps
	if remainingSteps <= 0 {
		return 0
	}

	avgTimePerStep := float64(elapsed) / float64(currentSteps)
	eta := time.Duration(avgTimePerStep * float64(remainingSteps))

	return eta
}
