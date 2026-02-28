package main

import (
	"fmt"
	"strings"
	"sync/atomic"
)

var pbWidth int

func buildProgressBar(color string, progress float64, width int) string {
	fractions := []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}
	nfractions := len(fractions)
	filledLen := int(progress * float64(width))

	if hasError {
		color = red
	}
	if filledLen >= width {
		return fmt.Sprintf("[%s]%s", color, strings.Repeat("█", width))
	}

	fractionIdx := int(((progress * float64(width)) - float64(filledLen)) * 8)
	if fractionIdx < 0 {
		fractionIdx = 0
	} else if fractionIdx > nfractions {
		fractionIdx = nfractions
	}

	fractionStr := fractions[fractionIdx]
	emptyLen := width - filledLen
	if fractionIdx > 0 {
		emptyLen--
	}
	if emptyLen < 0 {
		emptyLen = 0
	}

	format := "[%s]%s[#3388FF]%s[#444444]%s"

	return fmt.Sprintf(format, color, strings.Repeat("█", filledLen), fractionStr, strings.Repeat("─", emptyLen))
}

func currentProgressBar(color string) (currentPos float64, bar string) {
	currentPos = float64(min(atomic.LoadInt32(&currentSteps), totalSteps))
	bar = buildProgressBar(color, float64(currentSteps)/float64(totalSteps), pbWidth)

	return
}
