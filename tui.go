package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rivo/tview"
)

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

func setStatusView(spinerSymob, progressBar string, pct int, msg string, elapsed, eta time.Duration) {
	symbFmt := " %s "
	barFmt := " %s "
	pctFmt := "[#dcdccc]%s%%[white]"
	sepFmt := " [#555555]│[white] "
	timeFmt := "[#dcdccc]%s/%s[white]" // "elapsed/ETA"
	msgFmt := "%s"
	ctrlFmt := " [#888888]Press Ctrl+C.[white]"
	format := symbFmt + barFmt + pctFmt + sepFmt + timeFmt + sepFmt + msgFmt + ctrlFmt
	pctStr := strconv.Itoa(pct)

	etaStr := formatTime(eta)

	statusView.SetText(fmt.Sprintf(format, spinerSymob, progressBar, pctStr, formatTime(elapsed), etaStr, msg))
}

func finish(err error, elapsed time.Duration) {
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

		setStatusView(symbol, pg, pct, doneMsg, elapsed, 0)
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
