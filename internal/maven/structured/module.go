package structured

import (
	"strconv"
	"strings"
)

// ModulePhaseParser parses Maven module phases (from module header until next phase).
type ModulePhaseParser struct{}

func (p *ModulePhaseParser) NodeType() string { return "module" }

// ExtractLines finds one module block starting at startIdx.
func (p *ModulePhaseParser) ExtractLines(lines []string, startIdx int) ([]string, int, bool) {
	if len(lines) < startIdx+1 {
		return nil, 0, false
	}

	// Check for standard header OR simple "Building <name>" format OR multi-line simple format
	if !isModuleHeader(lines[startIdx]) && !isSimpleModuleHeaderMultiLine(lines, startIdx) && !isSimpleBuildingLine(lines[startIdx]) {
		return nil, 0, false
	}

	// Track what type we started with
	startedWithStandardHeader := isModuleHeader(lines[startIdx])
	startedWithMultiLineSimple := isSimpleModuleHeaderMultiLine(lines, startIdx)
	startedWithSimpleHeader := !startedWithStandardHeader && !startedWithMultiLineSimple && isSimpleBuildingLine(lines[startIdx])

	// Start after header (skip the multi-line separator + "Building" + separator if present)
	end := startIdx + 1
	if startedWithMultiLineSimple {
		end = startIdx + 3 // Skip first separator, "Building" line, and second separator
	}

	for ; end < len(lines); end++ {
		line := lines[end]
		isSimpleFormat := startedWithSimpleHeader || startedWithMultiLineSimple

		// END CONDITIONS: Stop when we detect the START of another module header OR build block

		// Check for multi-line module header starting at current position
		if isSimpleModuleHeaderMultiLine(lines, end) {
			break
		}

		// Check for standard module header
		if isModuleHeader(line) {
			break
		}

		// For standard format: stop at plugin headers (build blocks are parsed separately)
		// For simple format: plugin headers are part of module content
		if !isSimpleFormat && isPluginHeader(line) {
			break
		}

		// Other end markers (reactor summary, etc.)
		if isReactorHeader(line) || isReactorSummaryFor(line) {
			break
		}

		// Only break on long separator if NOT using simple format
		if !isSimpleFormat && isLongSeparator(line) {
			break
		}
	}

	if end > startIdx {
		return lines[startIdx:end], end - startIdx, true
	}
	return nil, 0, false
}

// isSimpleModuleHeader checks if line is a simple "Building <name> <version>" format
// and checks that previous line is NOT a standard module header (to avoid splitting modules)
func isSimpleModuleHeader(lines []string, idx int) bool {
	if idx >= len(lines) {
		return false
	}
	// Check for multi-line format: separator + "Building ..." + separator
	if isSimpleModuleHeaderMultiLine(lines, idx) {
		return true
	}
	line := lines[idx]
	// Don't use simple format if previous line was a standard header
	if idx > 0 && isModuleHeader(lines[idx-1]) {
		return false
	}
	return ModuleHeaderSimpleRegex.MatchString(line)
}

// ParseMetaData extracts metadata from the found module block using extractModuleMetadata.
func (p *ModulePhaseParser) ParseMetaData(found []string) map[string]any {
	return extractModuleMetadata(found)
}

// Parse combines ExtractLines and ParseMetaData.
func (p *ModulePhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	name := ""
	artifactId := ""
	groupId := ""

	// Check if this was a multi-line simple header starting at startIdx
	isMultiLineSimpleAtStart := isSimpleModuleHeaderMultiLine(lines, startIdx)

	// Try to get groupId:artifactId from standard header first line
	for _, l := range found {
		if isModuleHeader(l) {
			if substrs := ModuleHeaderRegex.FindStringSubmatch(l); len(substrs) == 2 {
				ga := substrs[1]
				if idx := strings.Index(ga, ":"); idx != -1 {
					groupId = ga[:idx]
					artifactId = ga[idx+1:]
					meta["groupId"] = groupId
					meta["artifactId"] = artifactId
				}
				break
			}
		}
	}

	// For multi-line simple format, extract name from the "Building" line in found[1]
	if isMultiLineSimpleAtStart && len(found) > 1 {
		if strings.HasPrefix(found[1], "[INFO] Building ") {
			content := found[1][len("[INFO] Building "):]
			words := strings.Fields(content)
			if len(words) >= 2 {
				// Last word is version, everything before is the name
				name = strings.Join(words[:len(words)-1], " ")
			} else if len(words) == 1 {
				name = words[0]
			}
		}
	}

	// Fallback: try meta["name"] if name still empty
	if name == "" {
		if n, ok := meta["name"].(string); ok && n != "" {
			name = n
		}
	}

	// If name not set from multi-line, try meta["name"] (e.g., "Baker Types" from the "Building ..." line)
	if name == "" {
		if n, ok := meta["name"].(string); ok && n != "" {
			name = n
		} else if artifactId != "" {
			name = artifactId
		} else if groupId != "" {
			name = groupId + ":" + artifactId
		}
	}
	node := &Node{
		Name:  name,
		Type:  "module",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

// extractModuleMetadata parses headerLines to extract module metadata defensively.
func extractModuleMetadata(headerLines []string) map[string]any {
	meta := make(map[string]any)
	if len(headerLines) < 2 {
		return meta
	}
	l2 := headerLines[1]
	var name, version string

	// Check if the line starts with "Building "
	// The actual line format is "[INFO] Building ..." so we need to check after [INFO]
	l2Content := l2
	if len(l2) > len("[INFO] ") && strings.HasPrefix(l2, "[INFO] ") {
		l2Content = l2[len("[INFO] "):]
	}

	if strings.HasPrefix(l2Content, "Building ") {
		rest := strings.TrimPrefix(l2Content, "Building ")
		versionEnd := strings.Index(rest, " [")
		if versionEnd == -1 {
			versionEnd = len(rest)
		}
		words := strings.Fields(rest[:versionEnd])
		if len(words) >= 2 {
			version = words[len(words)-1]
			name = strings.Join(words[:len(words)-1], " ")
		} else if len(words) == 1 {
			name = words[0]
		}
		if name != "" {
			meta["name"] = name
		}
		if version != "" {
			meta["version"] = version
		}
		idxBracket1 := strings.Index(rest, "[")
		idxBracket2 := strings.Index(rest, "/")
		idxBracket3 := strings.Index(rest, "]")
		if idxBracket1 != -1 && idxBracket2 != -1 && idxBracket3 != -1 && idxBracket1 < idxBracket2 && idxBracket2 < idxBracket3 {
			x := rest[idxBracket1+1 : idxBracket2]
			y := rest[idxBracket2+1 : idxBracket3]
			if xi, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
				meta["moduleIndex"] = xi
			}
			if yi, err := strconv.Atoi(strings.TrimSpace(y)); err == nil {
				meta["moduleCount"] = yi
			}
		}
	}
	if len(headerLines) > 2 {
		l3 := headerLines[2]
		if len(l3) > len("[INFO]   from") && strings.HasPrefix(l3, "[INFO]   from") {
			meta["pomFile"] = strings.TrimSpace(strings.TrimPrefix(l3, "[INFO]   from"))
		}
	}
	for i := 3; i < len(headerLines); i++ {
		l := headerLines[i]
		if strings.HasPrefix(l, "[INFO] ") && strings.HasSuffix(l, "---------------------------------") {
			inner := l[len("[INFO] "):]
			dashCount := strings.Count(inner, "-")
			if dashCount >= 40 {
				bracketStart := strings.Index(inner, "[")
				if bracketStart != -1 {
					bracketEnd := strings.Index(inner[bracketStart:], "]")
					if bracketEnd != -1 {
						packaging := strings.TrimSpace(inner[bracketStart+1 : bracketStart+bracketEnd])
						if packaging != "" {
							meta["packaging"] = packaging
						}
					}
				}
				break
			}
		}
	}
	return meta
}
