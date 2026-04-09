package structured

// Parser is the interface implemented by all phase parsers.
// Each parser attempts to match and parse a specific section of Maven output.
type Parser interface {
	Parse(lines []string, startIdx int) (*Node, int, bool)
	NodeType() string
}

// BaseParser provides the common pattern of:
// 1. Find boundaries (ExtractLines)
// 2. Extract metadata (ParseMetaData)
// 3. Combine into Node (Parse)
type BaseParser struct{}

// ExtractLines finds one segment starting at startIdx.
// Returns: found lines, lines consumed, success
func (p *BaseParser) ExtractLines(lines []string, startIdx int) ([]string, int, bool) {
	return nil, 0, false
}

// ParseMetaData extracts metadata from the found lines.
// To be overridden by each parser.
func (p *BaseParser) ParseMetaData(lines []string) map[string]any {
	return nil
}
