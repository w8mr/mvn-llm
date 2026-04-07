package structured

// BuildPhaseParser parses the main build phase output blocks.
type BuildPhaseParser struct{}

func (p *BuildPhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	start := startIdx
	// Find the first build marker
	for start < len(lines) && !(len(lines[start]) > 10 && lines[start][:10] == "[INFO] ---") {
		start++
	}
	if start >= len(lines) {
		return nil, 0, false
	}
	end := start
	for end < len(lines) {
		if lines[end] == "[INFO] ------------------------------------------------------------------------" {
			break // Stop BEFORE the summary separator line
		}
		end++
	}
	if end > start {
		block := BlockOutput{
			Type:  "build-block",
			Lines: lines[start:end],
		}
		return block, end - start, true
	}
	return nil, 0, false
}
