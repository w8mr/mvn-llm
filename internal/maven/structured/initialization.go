package structured

import (
	"strings"
)

// InitializationPhaseParser parses the Maven initialization section at the start of the log.
// This includes the Reactor Build Order section listing all modules to be built.
type InitializationPhaseParser struct {
	BaseParser
}

// StartMarker: only matches strong initialization markers.
// This is used by ParseUntilNextBlock to detect boundaries - must be strict
// to avoid stopping inside module content at generic [INFO] lines.
func (p *InitializationPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if len(lines) == 0 || idx >= len(lines) {
		return false, 0
	}
	line := lines[idx]

	// Accept Apache Maven header
	if strings.HasPrefix(line, "Apache Maven") {
		return true, 1
	}

	// Accept strong initialization markers only
	if line == "[INFO] Scanning for projects..." {
		return true, 1
	}
	if isReactorHeader(line) || isReactorHeaderMultiLine(lines, idx) {
		return true, 1
	}

	return false, 0
}

// NodeType returns the node type this parser produces.
func (p *InitializationPhaseParser) NodeType() string {
	return "initialization"
}

// ExtractLines finds one initialization block starting at startIdx.
// Can start at a strong marker OR a generic Maven log line (soft marker).
// Must contain key initialization markers to be valid.
func (p *InitializationPhaseParser) ExtractLines(lines []string, startIdx int, allParsers []Parser) ([]string, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}

	// Check for strong markers first
	strongMarker, _ := p.StartMarker(lines, startIdx)

	// Also accept soft markers (generic Maven log lines)
	line := lines[startIdx]
	softMarker := isMavenLogLine(line)

	// Must start with either strong or soft marker
	if !strongMarker && !softMarker {
		return nil, 0, false
	}

	// Track key initialization markers for validation
	hasScanningForProjects := false
	hasReactorBuildOrder := false

	// Scan through content starting from startIdx to find required markers and detect end
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]

		// Check for key initialization markers
		if line == "[INFO] Scanning for projects..." {
			hasScanningForProjects = true
		}
		if isReactorHeader(line) || isReactorHeaderMultiLine(lines, i) {
			hasReactorBuildOrder = true
		}

		// Check if any other parser's block is starting (end of initialization)
		if i > startIdx {
			for _, parser := range allParsers {
				// Don't check against ourselves
				if parser.NodeType() == "initialization" {
					continue
				}
				if ok, _ := parser.StartMarker(lines, i); ok {
					// Found start of another block - end initialization here if valid
					if hasScanningForProjects || hasReactorBuildOrder {
						return lines[startIdx:i], i - startIdx, true
					}
					return nil, 0, false
				}
			}
		}
	}

	// End of file - valid if we have at least one key marker
	if hasScanningForProjects || hasReactorBuildOrder {
		return lines[startIdx:], len(lines) - startIdx, true
	}

	return nil, 0, false
}

// isMavenLogLine checks if a line is Maven output (valid log levels like [INFO], [DEBUG], etc.)
func isMavenLogLine(line string) bool {
	return strings.HasPrefix(line, "[INFO]") ||
		strings.HasPrefix(line, "[DEBUG]") ||
		strings.HasPrefix(line, "[WARNING]") ||
		strings.HasPrefix(line, "[ERROR]")
}

// ParseMetaData extracts metadata from the found lines.
func (p *InitializationPhaseParser) ParseMetaData(found []string) map[string]any {
	var modules []map[string]string
	inReactorOrder := false
	for _, line := range found {
		if isReactorHeader(line) {
			inReactorOrder = true
			continue
		}
		if inReactorOrder {
			if line == "[INFO]" || line == "[INFO] " || len(line) < 12 {
				continue
			}
			s := line[len("[INFO] "):]
			if len(s) == 0 {
				continue
			}
			lastBracket := -1
			lastClose := -1
			for j := len(s) - 1; j >= 0; j-- {
				if s[j] == '[' && lastBracket == -1 {
					lastBracket = j
				}
				if s[j] == ']' && lastClose == -1 {
					lastClose = j
				}
			}
			if lastBracket != -1 && lastClose != -1 && lastBracket < lastClose {
				name := s[:lastBracket]
				typ := s[lastBracket+1 : lastClose]
				name = strings.TrimSpace(name)
				typ = strings.TrimSpace(typ)
				if name != "" && typ != "" {
					modules = append(modules, map[string]string{"module": name, "packaging": typ})
				}
			} else if lastBracket == -1 && lastClose == -1 && !strings.HasPrefix(s, "-") {
				inReactorOrder = false
			}
		}
	}
	meta := map[string]any{}
	if len(modules) > 0 {
		meta["modules"] = modules
	}
	return meta
}

// Parse combines ExtractLines and ParseMetaData for backward compatibility.
func (p *InitializationPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  "initialization",
		Type:  "initialization",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}
