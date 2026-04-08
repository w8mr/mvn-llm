package structured

import (
	"strconv"
	"strings"
)

// ModulePhaseParser parses Maven module headers (e.g., [INFO] --- module-a <artifact:version> ---).
// It extracts module metadata and returns a node representing the module container.
// Build blocks within the module are handled by OutputParser's insertion-point tracking.
type ModulePhaseParser struct{}

// NodeType returns the node type this parser produces.
func (p *ModulePhaseParser) NodeType() string {
	return "module"
}

// Parse attempts to parse a module header starting at startIdx.
// Returns the parsed Node (with metadata), number of lines consumed, and whether parsing succeeded.
func (p *ModulePhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	line := lines[startIdx]
	isMH := isModuleHeader(line)
	if !isMH {
		return nil, 0, false
	}

	headerEnd := findHeaderEnd(lines, startIdx)
	if headerEnd <= startIdx {
		return nil, 0, false
	}

	meta := extractModuleMetadata(lines[startIdx:headerEnd])
	moduleName := "module"
	if n, ok := meta["name"].(string); ok {
		moduleName = n
	}

	node := &Node{
		Name:     moduleName,
		Type:     "module",
		Lines:    lines[startIdx:headerEnd],
		Meta:     meta,
		Children: []Node{},
	}

	// Build-phase parsing is now handled by OutputParser's insertion-point tracking
	// This returns only the module header, OutputParser handles children insertion

	return node, headerEnd - startIdx, true
}

func isModuleHeader(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.Contains(line, "<") && strings.Contains(line, ">")
}

func isSummary(line string) bool {
	return line == "[INFO] ------------------------------------------------------------------------" ||
		strings.HasPrefix(line, "[INFO] Reactor Summary for") ||
		strings.HasPrefix(line, "[INFO] BUILD ")
}

func isPluginHeader(line string) bool {
	return strings.HasPrefix(line, "[INFO] --- ")
}

func findHeaderEnd(lines []string, startIdx int) int {
	if startIdx+2 >= len(lines) {
		return startIdx
	}
	if !strings.HasPrefix(lines[startIdx+1], "[INFO] Building") {
		return startIdx
	}
	if !strings.HasPrefix(lines[startIdx+2], "[INFO]   from") {
		return startIdx
	}

	end := startIdx + 3

	if end < len(lines) &&
		strings.HasPrefix(lines[end], "[INFO] --------------------------------[") &&
		strings.HasSuffix(lines[end], "]---------------------------------") {
		if end+1 < len(lines) && strings.HasPrefix(lines[end+1], "[INFO] ") && strings.TrimSpace(lines[end+1][7:]) == "" {
			end += 2
		}
	}

	for end < len(lines) && strings.HasPrefix(lines[end], "[INFO] ") && strings.TrimSpace(lines[end][7:]) == "" {
		end++
	}

	return end
}

func extractModuleMetadata(headerLines []string) map[string]any {
	meta := make(map[string]any)

	if len(headerLines) < 3 {
		return meta
	}

	l1 := headerLines[0]
	if i1 := strings.Index(l1, "<"); i1 != -1 {
		if i2 := strings.Index(l1, ">"); i2 != -1 && i2 > i1 {
			ga := strings.TrimSpace(l1[i1+1 : i2])
			parts := strings.SplitN(ga, ":", 2)
			if len(parts) == 2 {
				meta["groupId"] = parts[0]
				meta["artifactId"] = parts[1]
			}
		}
	}

	l2 := headerLines[1][len("[INFO] "):]
	var name, version string
	if sp := strings.SplitN(l2, " ", 3); len(sp) == 3 && strings.HasPrefix(sp[0], "Building") {
		toks := strings.SplitN(l2, " ", 4)
		if len(toks) >= 3 {
			name = toks[1]
			version = toks[2]
			meta["name"] = name
			meta["version"] = version
			if len(toks) == 4 {
				idxBracket1 := strings.Index(toks[3], "[")
				idxBracket2 := strings.Index(toks[3], "/")
				idxBracket3 := strings.Index(toks[3], "]")
				if idxBracket1 != -1 && idxBracket2 != -1 && idxBracket3 != -1 {
					x := toks[3][idxBracket1+1 : idxBracket2]
					y := toks[3][idxBracket2+1 : idxBracket3]
					if xi, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
						meta["moduleIndex"] = xi
					}
					if yi, err := strconv.Atoi(strings.TrimSpace(y)); err == nil {
						meta["moduleCount"] = yi
					}
				}
			}
		}
	}

	if len(headerLines) > 2 {
		l3 := headerLines[2]
		if strings.HasPrefix(l3, "[INFO]   from") {
			meta["pomFile"] = strings.TrimSpace(strings.TrimPrefix(l3, "[INFO]   from"))
		}
	}

	for i := 3; i < len(headerLines); i++ {
		l := headerLines[i]
		if strings.HasPrefix(l, "[INFO] --------------------------------[") && strings.HasSuffix(l, "]---------------------------------") {
			sepPrefix := "[INFO] --------------------------------["
			start := strings.Index(l, sepPrefix)
			if start != -1 {
				left := start + len(sepPrefix)
				right := strings.Index(l[left:], "]")
				if right != -1 {
					meta["packaging"] = strings.TrimSpace(l[left : left+right])
				}
			}
			break
		}
	}

	return meta
}
