package structured

import "strings"

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
	if startIdx > 0 {
		return nil, 0, false
	}
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "[INFO]") {
		return nil, 0, false
	}

	start := startIdx
	end := startIdx
	inReactorSection := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if isReactorHeader(line) {
			inReactorSection = true
		}
		if inReactorSection {
			if isInitializationSeparator(line) || isBuildSeparator(line) {
				break
			}
			if !strings.HasPrefix(line, "[INFO]") {
				break
			}
			end = i + 1
		}
	}
	if end > start {
		return lines[start:end], end - startIdx, true
	}
	return nil, 0, false
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
