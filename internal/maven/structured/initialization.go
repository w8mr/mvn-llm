package structured

// InitializationPhaseParser parses initialization-related output blocks.
type InitializationPhaseParser struct{}

func (p *InitializationPhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx > 0 {
		return nil, 0, false // Only matches from the top
	}
	start := startIdx
	end := startIdx
	inReactor := false
	for i, line := range lines {
		if line == "[INFO] Reactor Build Order:" {
			inReactor = true
		}
		if inReactor {
			if line == "" || (i > 0 && lines[i-1] != "" && line == "") {
				end = i + 1 // include blank after modules
				break
			}
		} else if i > 0 && line == "" && lines[i-1] == "[INFO] Reactor Build Order:" {
			end = i + 1
			break
		} else if i > 0 && line == "[INFO]" && lines[i-1] == "[INFO] Reactor Build Order:" {
			end = i + 1
			break
		}
		end = i + 1
		// Stop if we hit the first module build marker
		if len(line) > 10 && line[:10] == "[INFO] ---" {
			break
		}
	}
	if end > start {
		block := BlockOutput{
			Type:  "initialization",
			Lines: lines[start:end],
		}
		return block, end - start, true
	}
	return nil, 0, false
}
