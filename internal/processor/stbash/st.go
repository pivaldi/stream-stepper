package stbashprocessor

import (
	"fmt"
	"strings"

	"github.com/pivaldi/stream-stepper/internal/processor"
)

type grammar struct {
	flag           string
	formatText     string
	formatStatus   string
	isError        bool
	isProgressStep bool
}

var grammars = map[string]grammar{
	"h1": {
		flag:       "st.h1> ",
		formatText: "[blue::b]━━━ %s ━━━[::-]",
	},
	"h2": {
		flag:       "st.h2> ",
		formatText: "[teal::b] ━━ %s ━━[::-]",
	},
	"h3": {
		flag:       "st.h3> ",
		formatText: "[#5FAFAF::b] ━ %s ━[::-]",
	},
	"doing": {
		flag:           "st.doing> ",
		formatText:     "[cyan] ⟳ %s[white]",
		formatStatus:   "%s",
		isProgressStep: true,
	},
	"done": {
		flag:       "st.done> ",
		formatText: "[green] ✓ %s[white]",
	},
	"nothingtd": {
		flag:       "st.nothingtd> ",
		formatText: "[green] ⏭ %s[white]",
	},
	"skipped": {
		flag:       "st.skipped> ",
		formatText: "[#33E5FF] ⏭ %s[white]",
	},
	"warn": {
		flag:       "st.warn> ",
		formatText: "[yellow::b] ⚠ %s[::-]",
	},
	"fail": {
		flag:         "st.fail> ",
		formatText:   "[red] ✗ %s[white]",
		formatStatus: "FAILED: %s",
	},
	"do": {
		flag:       "st.do> ",
		formatText: "[#AFAFAF] $ %s[white]",
	},
	"success": {
		flag:       "st.success> ",
		formatText: "[green::b] ✓ %s[::-]",
	},
	"abort": {
		flag:         "st.abort> ",
		formatText:   "[red::b] ✗ %s[::-]",
		formatStatus: "ABORTED: %s",
	},
}

type Processor struct {
	escape func(string) string
}

func New(escape func(string) string) *Processor {
	return &Processor{
		escape: escape,
	}
}

func (p *Processor) Escape(text string) string {
	return p.escape(text)
}

// ProcessLine parses the st.* grammar and triggers UI/Tracker updates
func (p *Processor) ProcessLine(line string, isStderr bool) processor.ProcessedLine {
	result := processor.ProcessedLine{
		FormattedText: line,
	}

	// If it's a standard error line without a grammar prefix, print it in red
	if isStderr && !strings.HasPrefix(line, "st.") {
		result.FormattedText = fmt.Sprintf("[red::b]%s[::-]", p.Escape(line))
		result.IsError = true

		return result
	}

	pass := false
	for _, g := range grammars {
		after, ok := strings.CutPrefix(line, g.flag)
		if !ok {
			continue
		}

		msg := strings.TrimSpace(after)
		result.FormattedText = fmt.Sprintf(g.formatText, p.Escape(msg))
		if g.formatStatus != "" {
			result.StatusMessage = fmt.Sprintf(g.formatStatus, p.Escape(msg))
		}

		result.IsError = g.isError
		result.IsProgressStep = g.isProgressStep
		pass = true

		break
	}

	if !pass {
		result.FormattedText = "\t" + p.Escape(line)
	}

	return result
}
