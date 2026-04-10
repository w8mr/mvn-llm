package structured

// Parser is the interface implemented by all phase parsers.
// Each parser must implement three methods to support the unified parsing flow:
//
// 1. StartMarker: Detects if lines[idx:] match this parser's block header.
//   - Returns (matched, markerLen) where markerLen is the number of lines in the marker
//   - Markers can be single-line or multi-line (e.g., separator + content + separator)
//   - Should be fast and specific to avoid false positives
//
// 2. Parse: Extracts and parses a complete block starting at startIdx.
//   - Uses StartMarker to validate the start position
//   - Uses ParseUntilNextBlock to find block boundaries
//   - Returns a Node with parsed metadata and lines
//
// 3. NodeType: Returns the node type string (e.g., "module", "build-block", "summary")
//
// Design rationale:
//   - Separating StartMarker from Parse allows the main OutputParser to efficiently
//     scan for block boundaries without invoking full parsing logic
//   - The unified flow ensures consistent boundary detection across all parsers
//   - Multi-line marker support enables proper handling of complex Maven output formats
type Parser interface {
	// Detects if lines[idx:] matches this parser's block. Returns (matched, markerLen)
	StartMarker(lines []string, idx int) (bool, int)
	Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool)
	NodeType() string
}

// BaseParser provides the common pattern of:
// 1. Find boundaries (ExtractLines)
// 2. Extract metadata (ParseMetaData)
// 3. Combine into Node (Parse)
type BaseParser struct{}

// Default StartMarker implementation always returns not-matched. To be overridden.
func (p *BaseParser) StartMarker(lines []string, idx int) (bool, int) {
	return false, 0
}

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
