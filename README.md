# StreamStepper

A Go-based CLI tool that parses shell command output to render a scrollable terminal UI (TUI) and dynamic progress bar.

StreamStepper intercepts `stdout` and `stderr` from your scripts. It listens for a specific trigger string to increment a progress bar and extract status messages, while keeping the raw logs available in a scrollable view.

## Installation

Ensure you have Go installed, then compile the binary:

```bash
git clone https://github.com/yourusername/stream-stepper.git
cd stream-stepper
go build -o stream-stepper main.go

# Optional: move to your binary path
sudo mv stream-stepper /usr/local/bin/

```

## How It Works

StreamStepper scans text line-by-line. If a line contains the trigger flag (default: `==>`), two things happen:

1. The progress bar increments by 1 step.
2. Any text following the flag on that same line is displayed as the current status message in the bottom panel.

All other output is printed normally in the main scrolling window.

## Usage

### Command-Line Flags

* `--steps` (Required): Total number of steps required to reach 100%.
* `--flag` (Optional): The string trigger to look for. Default is `==>`.
* `--tagged` (Optional): Boolean flag to read from `stdin` expecting `[OUT]` and `[ERR]` prefixes.
* `--err-fifo` (Optional): Path to a named pipe to read `stderr` from.
---

### Operating Modes

StreamStepper supports three ways to route your data:

#### 1. Exec Mode (Default)

Pass the command as an argument. StreamStepper executes it as a child process and handles `stdout` and `stderr` automatically.

```bash
stream-stepper --steps=5 "./deploy.sh"

```

#### 2. Standard Pipe Mode

Pipe data directly into StreamStepper.
**Note:** as pipe redirect `stderr` to `stdout`, StreamStepper can not capture errors with this usage.

```bash
./deploy.sh 2>&1 | stream-stepper --steps=5

```

#### 3. Tagged Pipe Mode (Separated stdout/stderr)

If you want to pipe data but keep `stderr` visually distinct (rendered in red), you can use Bash to tag the streams before they enter the UI.
In this case as `sed` blocks buffering, the parsing is not realtime.

```bash
{ ./deploy.sh 2>&1 1>&3 | sed 's/^/[ERR] /' >&2; } 3>&1 1>&2 | sed 's/^/[OUT] /' | stream-stepper --steps=5 --tagged

```

To fix unbuffered `sed` using `GNU sed`:

```bash
{ stdbuf -oL -eL ./deploy.sh 2>&1 1>&3 | sed -u 's/^/[ERR] /' >&2; } 3>&1 1>&2 | sed -u 's/^/[OUT] /' | stream-stepper --steps=5 --tagged

```

#### 4. FIFO Mode (Named Pipes)

Route `stderr` through a named pipe, and `stdout` through the standard pipe.

```bash
mkfifo err_pipe
./deploy.sh 2> err_pipe | stream-stepper --steps=5 --err-fifo="err_pipe"
rm err_pipe
```

## Example Script

If you run `stream-stepper --steps=3 "./script.sh"`, your `script.sh` should look something like this:

```bash
#!/bin/bash

echo "Starting process..."

echo "==> Downloading files..."
sleep 2

echo "==> Compiling assets..."
sleep 2

echo "==> Cleaning up..."
sleep 1
```

## Dependencies

* [tview](https://github.com/rivo/tview) - Terminal UI library
* [tcell](https://github.com/gdamore/tcell) - Terminal handling
