package stbashprocessor

import (
	"fmt"
	"strings"

	"github.com/pivaldi/stream-stepper/internal/processor"
)

type Processor struct {
}

func New() *Processor {
	return new(Processor)
}

// ProcessLine parses the st.* grammar and triggers UI/Tracker updates
func (p *Processor) ProcessLine(line string, isStderr bool) processor.ProcessedLine {
	result := processor.ProcessedLine{
		FormattedText: line,
	}

	// If it's a standard error line without a grammar prefix, print it in red
	if isStderr && !strings.HasPrefix(line, "st.") {
		result.FormattedText = fmt.Sprintf("[red::b]%s[::-]", line)
		result.IsError = true

		return result
	}

	// Route based on the grammar prefix
	switch {
	case strings.HasPrefix(line, "st.h1> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.h1> "))
		result.FormattedText = fmt.Sprintf("\n[blue::b]━━━ %s ━━━[::-]", msg)

	case strings.HasPrefix(line, "st.h2> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.h2> "))
		result.FormattedText = fmt.Sprintf("\n[teal::b] ━━ %s ━━[::-]", msg)

	case strings.HasPrefix(line, "st.h3> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.h3> "))
		result.FormattedText = fmt.Sprintf("[#5FAFAF::b] ━ %s ━[::-]", msg)

	case strings.HasPrefix(line, "st.doing> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.doing> "))
		result.IsProgressStep = true
		result.StatusMessage = msg
		result.FormattedText = fmt.Sprintf("[cyan] ⟳ %s[white]", msg)

	case strings.HasPrefix(line, "st.done> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.done> "))
		result.FormattedText = fmt.Sprintf("[green] ✓ %s[white]", msg)

	case strings.HasPrefix(line, "st.nothingtd> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.nothingtd>"))
		result.FormattedText = fmt.Sprintf("[green] ⏭ %s[white]", msg)

	case strings.HasPrefix(line, "st.skiped> "): // Matches the typo in your bash script
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.skiped> "))
		result.FormattedText = fmt.Sprintf("[#33E5FF] ⏭ %s[white]", msg)

	case strings.HasPrefix(line, "st.warn> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.warn> "))
		result.FormattedText = fmt.Sprintf("[yellow::b] ⚠ %s[::-]", msg)

	case strings.HasPrefix(line, "st.fail> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.fail> "))
		result.StatusMessage = "FAILED: " + msg
		result.FormattedText = fmt.Sprintf("[red::b] ✗ %s[::-]", msg)

	case strings.HasPrefix(line, "st.do> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.do> "))
		result.FormattedText = fmt.Sprintf("[#AFAFAF] $ %s[white]", msg)

	case strings.HasPrefix(line, "st.success> "):
		msg := strings.TrimSpace(strings.TrimPrefix(line, "st.success> "))
		result.FormattedText = fmt.Sprintf("[green::b] ✓ %s[::-]", strings.ToUpper(msg))

	default:
		// Standard log lines pass through normally
		result.FormattedText = "\t" + line
	}

	return result
}
