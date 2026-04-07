package structured

// Parser is the interface for all structured Maven output parsers.
type Parser interface {
	// Parse attempts to parse a section of lines. Returns (output, linesConsumed, matched)
	Parse(lines []string, startIdx int) (any, int, bool)
}
