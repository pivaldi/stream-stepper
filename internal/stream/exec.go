package stream

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/pivaldi/stream-stepper/internal/processor"
	"github.com/pivaldi/stream-stepper/internal/ui"
)

// ExecHandler handles exec input mode (runs command as child process)
type ExecHandler struct {
	display ui.Display
	cmdStr  string
}

// NewExecHandler creates a handler for exec mode
func NewExecHandler(display ui.Display, cmdStr string) *ExecHandler {
	return &ExecHandler{
		display: display,
		cmdStr:  cmdStr,
	}
}

// Start begins executing the command
func (h *ExecHandler) Start(proc processor.LineProcessor, onComplete func(exitCode int, err error)) error {
	h.setTitle()

	//#nosec G204 -- command execution is the intended purpose of this CLI tool
	cmd := exec.CommandContext(context.Background(), "sh", "-c", h.cmdStr)

	stdout, stderr, err := h.setupPipes(cmd)
	if err != nil {
		onComplete(1, err)

		return err
	}

	wg := h.startReaders(proc, stdout, stderr)

	if err := cmd.Start(); err != nil {
		onComplete(1, err)

		return fmt.Errorf("%s failed with error: %w", h.cmdStr, err)
	}

	wg.Wait()
	exitCode, err := h.waitForCommand(cmd)
	onComplete(exitCode, err)

	return err
}

func (h *ExecHandler) setTitle() {
	if tvd, ok := h.display.(*ui.TViewDisplay); ok {
		tvd.SetTitle(fmt.Sprintf(" Exec: %s ", h.cmdStr))
	}
}

func (h *ExecHandler) setupPipes(cmd *exec.Cmd) (stdout, stderr *bufio.Scanner, err error) {
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stderr pipe error: %w", err)
	}

	stdout, stderr, err = bufio.NewScanner(stdoutReader), bufio.NewScanner(stderrReader), nil

	return
}

func (h *ExecHandler) startReaders(
	proc processor.LineProcessor,
	stdout, stderr *bufio.Scanner,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(2)

	go h.readStream(&wg, proc, stdout, false)
	go h.readStream(&wg, proc, stderr, true)

	return &wg
}

func (h *ExecHandler) readStream(
	wg *sync.WaitGroup,
	proc processor.LineProcessor,
	scanner *bufio.Scanner,
	isStderr bool,
) {
	defer wg.Done()
	for scanner.Scan() {
		result := proc.ProcessLine(scanner.Text(), isStderr)
		_ = h.display.WriteLog(result.FormattedText)
	}
}

func (h *ExecHandler) waitForCommand(cmd *exec.Cmd) (int, error) {
	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), err
		}

		return 1, fmt.Errorf("cmd '%s' error: %w", cmd.String(), err)
	}

	return 0, nil
}
