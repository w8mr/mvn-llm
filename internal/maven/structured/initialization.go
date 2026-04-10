package structured

import (
	"strings"
)

// InitializationPhaseParser parses the Maven initialization section at the start of the log.
// This includes the Reactor Build Order section listing all modules to be built.
type InitializationPhaseParser struct {
	BaseParser
}

// NodeType returns the node type this parser produces.
func (p *InitializationPhaseParser) NodeType() string {
	return "initialization"
}

// ExtractLines finds one initialization block starting at startIdx.
func (p *InitializationPhaseParser) ExtractLines(lines []string, startIdx int) ([]string, int, bool) {
	if len(lines) == 0 || startIdx >= len(lines) {
		return nil, 0, false
	}

	// Must start exactly at startIdx with Maven output (valid log level OR Apache Maven header)
	if !isMavenLogLine(lines[startIdx]) && !strings.HasPrefix(lines[startIdx], "Apache Maven") {
		return nil, 0, false
	}

	// Track key initialization markers
	hasScanningForProjects := false
	hasReactorBuildOrder := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]

		// Check for key initialization markers (single line or multi-line)
		if line == "[INFO] Scanning for projects..." {
			hasScanningForProjects = true
		}
		// Check for single-line reactor header OR multi-line format
		if isReactorHeader(line) || isReactorHeaderMultiLine(lines, i) {
			hasReactorBuildOrder = true
		}

		// END CONDITIONS:
		// Stop when we detect the START of a module header (standard or multi-line)
		// Check for multi-line module header starting at current position
		if isSimpleModuleHeaderMultiLine(lines, i) {
			// Multi-line module header starts here - stop before it
			if hasScanningForProjects || hasReactorBuildOrder {
				return lines[startIdx:i], i - startIdx, true
			}
			return nil, 0, false
		}

		// Check for standard module header or plugin header
		if isInitializationSeparator(line) {
			// Valid initialization block requires at least one key marker
			if hasScanningForProjects || hasReactorBuildOrder {
				return lines[startIdx:i], i - startIdx, true
			}
			// Not a valid initialization block without key markers
			return nil, 0, false
		}

		// For simple "Building <name>" lines (single-line format): only end block if we have valid initialization markers
		// This prevents early termination for files like HawtJNI that use simple format
		if isSimpleBuildingLine(line) && (hasScanningForProjects || hasReactorBuildOrder) {
			return lines[startIdx:i], i - startIdx, true
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
func (p *InitializationPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx)
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
