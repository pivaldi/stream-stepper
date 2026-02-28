package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func handleExec(wg *sync.WaitGroup, cmdStr string) {
	// mainView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	mainView.SetTitle(fmt.Sprintf(" Exec: %s ", cmdStr))
	cmd := exec.CommandContext(context.Background(), "sh", "-c", cmdStr)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	wg.Add(2)
	go readStream(wg, stdout, false)
	go readStream(wg, stderr, true)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		hasError = true
		finish(err, elapsedTime)
	}
}

func handleTagged() {
	mainView.SetTitle(" Pipe: Tagged Stdin ")
	scanner := bufio.NewScanner(os.Stdin)
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

		processLine(line, isErr)
	}
}

func handleFIFO(wg *sync.WaitGroup, fifoPath string) {
	mainView.SetTitle(" Pipe: Stdin + FIFO ")
	wg.Add(2)

	// Read Standard Out from normal stdin
	go readStream(wg, os.Stdin, false)

	// Read Standard Err from the Named Pipe
	go func() {
		defer wg.Done()
		// This blocks until the bash script writes to it!
		file, err := os.Open(fifoPath)
		if err != nil {
			processLine(fmt.Sprintf("Failed to open FIFO: %v", err), true)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			processLine(scanner.Text(), true)
		}
	}()

	wg.Wait()
}
