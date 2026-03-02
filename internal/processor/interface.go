package processor

// LineProcessor processes input lines and detects progress triggers
type LineProcessor interface {
	ProcessLine(line string, isStderr bool) ProcessedLine
}

// ProcessedLine represents a processed line with metadata
type ProcessedLine struct {
	FormattedText  string // The text to display (with color codes, indentation)
	IsProgressStep bool   // Whether this line triggered a progress step
	StatusMessage  string // Extracted status message (if IsProgressStep is true)
}
