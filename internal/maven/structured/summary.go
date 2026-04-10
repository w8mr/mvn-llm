package structured

import (
	"regexp"
	"strings"
)

// SummaryPhaseParser parses the Maven build summary section at the end of the log.
// It extracts overall build status, total time, finish time, and per-module results.
type SummaryPhaseParser struct {
	BaseParser
}

// StartMarker: detects summary section start (long separator followed by Reactor Summary)
func (p *SummaryPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	// Summary starts with a long separator line followed by "Reactor Summary"
	if !isLongSeparator(lines[idx]) {
		return false, 0
	}
	// Must be followed by "Reactor Summary" to be the start of summary block
	if idx+1 < len(lines) {
		nextLine := lines[idx+1]
		if strings.HasPrefix(nextLine, "[INFO] Reactor Summary") {
			return true, 1
		}
	}
	return false, 0
}

// NodeType returns the node type this parser produces.
func (p *SummaryPhaseParser) NodeType() string {
	return "summary"
}

// ExtractLines finds the summary section starting at startIdx.
func (p *SummaryPhaseParser) ExtractLines(lines []string, startIdx int, allParsers []Parser) ([]string, int, bool) {
	ok, markerLen := p.StartMarker(lines, startIdx)
	if !ok {
		return nil, 0, false
	}

	startOfContent := startIdx + markerLen
	_, consumed := ParseUntilNextBlock(lines, startOfContent, allParsers, "summary")

	// Summary typically goes to end of file - consume all remaining lines
	totalConsumed := markerLen + consumed
	return lines[startIdx : startIdx+totalConsumed], totalConsumed, true
}

// ParseMetaData extracts metadata from the found summary lines.
func (p *SummaryPhaseParser) ParseMetaData(found []string) map[string]any {
	if len(found) == 0 {
		return nil
	}

	meta := map[string]any{}
	moduleList := []map[string]any{}
	var overallStatus, totalTime, finishedAt string

	// Find "Reactor Summary for" line
	summaryStart := -1
	for i, line := range found {
		if strings.HasPrefix(line, "[INFO] Reactor Summary for") {
			summaryStart = i
			break
		}
	}
	if summaryStart == -1 {
		summaryStart = 0
	}

	// Find module results table boundaries
	mLinesStart := -1
	mLinesEnd := -1
	for i := summaryStart + 2; i < len(found); i++ {
		if mLinesStart == -1 && strings.TrimSpace(found[i]) != "" {
			mLinesStart = i
		}
		if strings.HasPrefix(found[i], "[INFO] ------------------------------------------------------------------------") {
			mLinesEnd = i
			break
		}
	}
	if mLinesStart == -1 {
		mLinesStart = summaryStart + 2
	}
	if mLinesEnd == -1 {
		for i := mLinesStart; i < len(found); i++ {
			if strings.HasPrefix(found[i], "[INFO] BUILD SUCCESS") || strings.HasPrefix(found[i], "[INFO] BUILD FAILURE") {
				mLinesEnd = i
				break
			}
		}
	}

	// Extract module results
	if mLinesStart != -1 && mLinesEnd != -1 && mLinesEnd > mLinesStart {
		modRegex := regexp.MustCompile(`^(.*?) +[. ]+ *(SUCCESS|FAILURE|SKIPPED) *\[ *([^\]]+?) *\]$`)
		for _, modLine := range found[mLinesStart:mLinesEnd] {
			if !strings.HasPrefix(modLine, "[INFO] ") {
				continue
			}
			ml := strings.TrimPrefix(modLine, "[INFO] ")
			if matches := modRegex.FindStringSubmatch(ml); matches != nil {
				name := strings.TrimSpace(matches[1])
				status := matches[2]
				time := strings.TrimSpace(matches[3])
				moduleList = append(moduleList, map[string]any{
					"name":   name,
					"status": status,
					"time":   time,
				})
			} else {
				fields := strings.Fields(ml)
				name, status, time := "", "", ""
				if len(fields) >= 3 && strings.HasSuffix(fields[len(fields)-2], "SUCCESS") {
					status = "SUCCESS"
					name = strings.Join(fields[:len(fields)-2], " ")
					time = strings.Trim(fields[len(fields)-1], "[]")
				} else if len(fields) >= 3 && strings.HasSuffix(fields[len(fields)-2], "FAILURE") {
					status = "FAILURE"
					name = strings.Join(fields[:len(fields)-2], " ")
					time = strings.Trim(fields[len(fields)-1], "[]")
				} else if len(fields) >= 3 && strings.HasSuffix(fields[len(fields)-2], "SKIPPED") {
					status = "SKIPPED"
					name = strings.Join(fields[:len(fields)-2], " ")
					time = strings.Trim(fields[len(fields)-1], "[]")
				}
				if name != "" && status != "" {
					moduleList = append(moduleList, map[string]any{
						"name":   name,
						"status": status,
						"time":   time,
					})
				}
			}
		}
	}

	// Search backwards for overallStatus, totalTime, finishedAt
	for i := len(found) - 1; i >= 0; i-- {
		if overallStatus == "" && (strings.Contains(found[i], "BUILD SUCCESS") || strings.Contains(found[i], "BUILD FAILURE")) {
			overallStatus = strings.TrimSpace(strings.TrimPrefix(found[i], "[INFO] "))
		}
		if totalTime == "" && strings.HasPrefix(found[i], "[INFO] Total time:") {
			totalTime = strings.TrimSpace(strings.TrimPrefix(found[i], "[INFO] Total time:"))
		}
		if finishedAt == "" && strings.HasPrefix(found[i], "[INFO] Finished at:") {
			finishedAt = strings.TrimSpace(strings.TrimPrefix(found[i], "[INFO] Finished at:"))
		}
	}

	if len(moduleList) > 0 {
		meta["modules"] = moduleList
	}
	if overallStatus != "" {
		meta["overallStatus"] = overallStatus
	}
	if totalTime != "" {
		meta["totalTime"] = totalTime
	}
	if finishedAt != "" {
		meta["finishedAt"] = finishedAt
	}

	return meta
}

// Parse combines ExtractLines and ParseMetaData for backward compatibility.
func (p *SummaryPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  "summary",
		Type:  "summary",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}
