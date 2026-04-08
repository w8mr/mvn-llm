package structured

// Parser is the interface implemented by all phase parsers.
// Each parser attempts to match and parse a specific section of Maven output.
type Parser interface {
	Parse(lines []string, startIdx int) (*Node, int, bool)
}
