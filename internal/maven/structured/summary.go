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

// NodeType returns the node type this parser produces.
func (p *SummaryPhaseParser) NodeType() string {
	return "summary"
}

// ExtractLines finds the summary section starting at startIdx.
// Returns lines from startIdx to end of slice.
func (p *SummaryPhaseParser) ExtractLines(lines []string, startIdx int) ([]string, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	if lines[startIdx] != "[INFO] ------------------------------------------------------------------------" {
		return nil, 0, false
	}

	// Scan forward to find "Reactor Summary"
	foundSummaryStart := false
	for i := startIdx; i < startIdx+3 && i < len(lines); i++ {
		if isReactorSummaryFor(lines[i]) {
			foundSummaryStart = true
			break
		}
	}
	if !foundSummaryStart {
		return nil, 0, false
	}

	// Return everything to end of file
	end := len(lines)
	if end > startIdx && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[startIdx:end], end - startIdx, true
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
func (p *SummaryPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx)
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
