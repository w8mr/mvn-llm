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

	// Check for module separator line with any reasonable number of dashes
	// e.g., "------------------------< ... >------------------------" or "---------------------< ... >----------------------"
	if end < len(lines) {
		sepLine := lines[end]
		if strings.HasPrefix(sepLine, "[INFO] ") && strings.HasSuffix(sepLine, "---------------------------------") {
			// Check that it has matching dash count on both sides (at least 20 dashes each side)
			inner := sepLine[len("[INFO] "):]
			dashCount := strings.Count(inner, "-")
			if dashCount >= 40 { // At least 20 dashes on each side
				if end+1 < len(lines) && strings.HasPrefix(lines[end+1], "[INFO] ") && strings.TrimSpace(lines[end+1][7:]) == "" {
					end += 2
				}
			}
		}
	}

	// After the standard header lines, keep reading until we hit a build block line
	for end < len(lines) {
		line := lines[end]
		// Stop at build block header (e.g., "[INFO] --- maven-clean-plugin:3.2.0:clean (default-clean) @ module-name ---")
		if isPluginHeader(line) {
			break
		}
		// Include any other line in the header
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
	if strings.HasPrefix(l2, "Building ") {
		// Format: "Building <name> <version> [<moduleIndex>/<totalModules>]"
		rest := strings.TrimPrefix(l2, "Building ")

		// Find version by looking for patterns like -SNAPSHOT, -RELEASE, etc. or by finding [
		versionEnd := strings.Index(rest, " [")
		if versionEnd == -1 {
			versionEnd = len(rest)
		}

		// Look for version pattern - last word before [ is typically version
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

		// Extract module index from [x/y] if present
		idxBracket1 := strings.Index(rest, "[")
		idxBracket2 := strings.Index(rest, "/")
		idxBracket3 := strings.Index(rest, "]")
		if idxBracket1 != -1 && idxBracket2 != -1 && idxBracket3 != -1 {
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
		if strings.HasPrefix(l3, "[INFO]   from") {
			meta["pomFile"] = strings.TrimSpace(strings.TrimPrefix(l3, "[INFO]   from"))
		}
	}

	for i := 3; i < len(headerLines); i++ {
		l := headerLines[i]
		if strings.HasPrefix(l, "[INFO] ") && strings.HasSuffix(l, "---------------------------------") {
			inner := l[len("[INFO] "):]
			dashCount := strings.Count(inner, "-")
			if dashCount >= 40 {
				// Look for [packaging] pattern between two dash sequences
				// Format: "[INFO] ----...----[ packaging ]----...----"
				// The [ comes after the dashes, then packaging, then ]
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
