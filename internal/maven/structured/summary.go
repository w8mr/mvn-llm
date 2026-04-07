package structured

import "strings"

// SummaryPhaseParser parses the build summary output block.
type SummaryPhaseParser struct{}

func (p *SummaryPhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	start := -1
	end := len(lines)
	foundSummaryStart := false
	for i := startIdx; i < len(lines); i++ {
		// Look for the first summary separator and 'Reactor Summary for'
		if !foundSummaryStart && lines[i] == "[INFO] ------------------------------------------------------------------------" {
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "[INFO] Reactor Summary for") {
				start = i
				foundSummaryStart = true
				i++ // skip one line ahead to continue scan
			}
		}
		if foundSummaryStart && start != -1 {
			// Look for ending sequence: BUILD result, Total time, Finished at, final separator
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
					i = j // move i forward
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
		// Extend summary to include a blank or whitespace-only line immediately after the summary if present
		extendedEnd := end
		if extendedEnd < len(lines) && strings.TrimSpace(lines[extendedEnd]) == "" {
			extendedEnd++
		}
		block := BlockOutput{
			Type:  "summary",
			Lines: lines[start:extendedEnd],
		}
		return block, extendedEnd - start, true
	}
	return nil, 0, false
}
