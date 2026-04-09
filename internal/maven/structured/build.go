package structured

import "strings"

// BuildPhaseParser parses Maven plugin execution blocks (e.g., [INFO] --- maven-compiler-plugin:3.11.0:compile @ my-app ---).
// Each block represents a single plugin invocation with its output.
type BuildPhaseParser struct {
	BaseParser
}

// NodeType returns the node type this parser produces.
func (p *BuildPhaseParser) NodeType() string {
	return "build-block"
}

// ExtractLines finds one build block starting at startIdx.
func (p *BuildPhaseParser) ExtractLines(lines []string, startIdx int) ([]string, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	if !isPluginHeader(lines[startIdx]) {
		return nil, 0, false
	}
	start := startIdx
	end := start + 1
	for end < len(lines) {
		if isPluginHeader(lines[end]) || isModuleHeader(lines[end]) || isLongSeparator(lines[end]) {
			break
		}
		end++
	}
	for end < len(lines) && isEmptyInfoLine(lines[end]) {
		end++
	}
	return lines[start:end], end - startIdx, true
}

// ExtractSummary extracts a summary line from the found lines.
// Priority: last ERROR line → last WARNING line → last INFO line.
// Skips the first line (plugin header).
// Format: "Successful: <line>" or "Failure: <line>", or just "Success"/"Failure"
func (p *BuildPhaseParser) ExtractSummary(found []string) string {
	var lastError, lastWarning, lastInfo string

	for i, l := range found {
		if i == 0 {
			continue // Skip header
		}
		if len(l) == 0 {
			continue
		}
		if strings.HasPrefix(l, "[ERROR] ") && len(l) > 8 {
			lastError = strings.TrimSpace(l[8:])
		} else if strings.HasPrefix(l, "[WARNING] ") && len(l) > 10 {
			lastWarning = strings.TrimSpace(l[10:])
		} else if strings.HasPrefix(l, "[INFO] ") && len(l) > 6 {
			info := strings.TrimSpace(l[6:])
			if info != "" {
				lastInfo = info
			}
		}
	}

	if lastError != "" {
		return "Failure: " + lastError
	}
	if lastWarning != "" {
		return "Successful: " + lastWarning
	}
	if lastInfo != "" {
		return "Successful: " + lastInfo
	}
	return "Successful"
}

// ParseMetaData extracts metadata from the found lines.
func (p *BuildPhaseParser) ParseMetaData(found []string) map[string]any {
	if len(found) == 0 {
		return nil
	}

	status := "SUCCESS"
	for _, l := range found {
		if len(l) > 0 && l[0] == '[' {
			if len(l) > 8 && l[:8] == "[ERROR] " {
				status = "FAILED"
				break
			} else if len(l) > 10 && l[:10] == "[WARNING] " && status != "FAILED" {
				status = "SUCCESS-WITH-WARNINGS"
			}
		}
	}

	header := found[0]
	plugin := ""
	version := ""
	goal := ""
	executionId := ""
	artifactId := ""
	m := PluginHeaderRegex.FindStringSubmatch(header)
	if len(m) >= 7 {
		plugin = m[1]
		version = m[2]
		goal = m[3]
		if len(m[5]) > 0 {
			executionId = m[5]
		}
		artifactId = m[6]
	}

	meta := map[string]any{
		"status":  status,
		"summary": p.ExtractSummary(found),
	}
	if plugin != "" {
		meta["plugin"] = plugin
		meta["version"] = version
		meta["goal"] = goal
		meta["artifactId"] = artifactId
		if executionId != "" {
			meta["executionId"] = executionId
		}
	}

	return meta
}

// Parse combines ExtractLines and ParseMetaData for backward compatibility.
func (p *BuildPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}
