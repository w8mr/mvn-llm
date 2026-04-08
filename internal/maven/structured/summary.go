package structured

import (
	"regexp"
	"strings"
)

// SummaryPhaseParser parses the Maven build summary section at the end of the log.
// It extracts overall build status, total time, finish time, and per-module results.
type SummaryPhaseParser struct{}

// Parse attempts to parse a summary block starting at startIdx.
// Returns the parsed Node (with overall status, module results), number of lines consumed, and whether parsing succeeded.
func (p *SummaryPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	if lines[startIdx] != "[INFO] ------------------------------------------------------------------------" {
		return nil, 0, false
	}
	start := -1
	end := len(lines)
	foundSummaryStart := false
	for i := startIdx; i < len(lines); i++ {
		if !foundSummaryStart && lines[i] == "[INFO] ------------------------------------------------------------------------" {
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "[INFO] Reactor Summary for") {
				start = i
				foundSummaryStart = true
				i++
			}
			continue
		}
		if !foundSummaryStart && startIdx > 0 && i == startIdx {
			continue
		}
		if foundSummaryStart && start != -1 {
			var hasBuildStatus, hasTotalTime, hasFinishedAt, hasFinalSep bool
			for j := i; j < len(lines); j++ {
				if !hasBuildStatus && (lines[j] == "[INFO] BUILD SUCCESS" || lines[j] == "[INFO] BUILD FAILURE") {
					hasBuildStatus = true
				}
				if hasBuildStatus && !hasTotalTime && strings.HasPrefix(lines[j], "[INFO] Total time:") {
					hasTotalTime = true
				}
				if hasTotalTime && !hasFinishedAt && strings.HasPrefix(lines[j], "[INFO] Finished at:") {
					hasFinishedAt = true
				}
				if hasFinishedAt && strings.HasPrefix(lines[j], "[INFO] --------") {
					end = j + 1
					i = j
					hasFinalSep = true
					break
				}
			}
			if hasBuildStatus && hasTotalTime && hasFinishedAt && hasFinalSep {
				break
			}
		}
	}
	if start != -1 && end > start {
		extendedEnd := end
		if extendedEnd < len(lines) && strings.TrimSpace(lines[extendedEnd]) == "" {
			extendedEnd++
		}

		meta := map[string]any{}
		moduleList := []map[string]any{}
		var overallStatus, totalTime, finishedAt string

		summaryStart := -1
		for i := start; i < extendedEnd; i++ {
			if strings.HasPrefix(lines[i], "[INFO] Reactor Summary for") {
				summaryStart = i
				break
			}
		}

		mLinesStart := -1
		mLinesEnd := -1
		for i := summaryStart + 2; i < extendedEnd; i++ {
			if mLinesStart == -1 && strings.TrimSpace(lines[i]) != "" {
				mLinesStart = i
			}
			if strings.HasPrefix(lines[i], "[INFO] ------------------------------------------------------------------------") {
				mLinesEnd = i
				break
			}
		}
		if mLinesStart == -1 {
			mLinesStart = summaryStart + 2
		}
		if mLinesEnd == -1 {
			for i := mLinesStart; i < extendedEnd; i++ {
				if strings.HasPrefix(lines[i], "[INFO] BUILD SUCCESS") || strings.HasPrefix(lines[i], "[INFO] BUILD FAILURE") {
					mLinesEnd = i
					break
				}
			}
		}

		if mLinesStart != -1 && mLinesEnd != -1 && mLinesEnd > mLinesStart {
			modRegex := regexp.MustCompile(`^(.*?) +[. ]+ *(SUCCESS|FAILURE|SKIPPED) *\[ *([^\]]+?) *\]$`)
			for _, modLine := range lines[mLinesStart:mLinesEnd] {
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

		for i := extendedEnd - 1; i >= start; i-- {
			if overallStatus == "" && (strings.Contains(lines[i], "BUILD SUCCESS") || strings.Contains(lines[i], "BUILD FAILURE")) {
				overallStatus = strings.TrimSpace(strings.TrimPrefix(lines[i], "[INFO] "))
			}
			if totalTime == "" && strings.HasPrefix(lines[i], "[INFO] Total time:") {
				totalTime = strings.TrimSpace(strings.TrimPrefix(lines[i], "[INFO] Total time:"))
			}
			if finishedAt == "" && strings.HasPrefix(lines[i], "[INFO] Finished at:") {
				finishedAt = strings.TrimSpace(strings.TrimPrefix(lines[i], "[INFO] Finished at:"))
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

		node := &Node{
			Name:  "summary",
			Type:  "summary",
			Lines: lines[start:extendedEnd],
			Meta:  meta,
		}
		return node, extendedEnd - start, true
	}
	return nil, 0, false
}
