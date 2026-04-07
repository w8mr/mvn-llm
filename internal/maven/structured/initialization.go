package structured

import "strings"

// InitializationPhaseParser parses initialization-related output blocks.
type InitializationPhaseParser struct{}

func (p *InitializationPhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx > 0 {
		return nil, 0, false // Only matches from the top
	}
	start := startIdx
	end := startIdx
	seenReactorHeader := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		// Mark when we see the Reactor Build Order
		if line == "[INFO] Reactor Build Order:" {
			seenReactorHeader = true
		}
		// After we've seen Reactor Build Order, only stop at a real phase boundary
		if seenReactorHeader {
			if len(line) > 10 && line[:10] == "[INFO] ---" {
				break // plugin header
			}
			if len(line) > 30 && line[:30] == "[INFO] ----------------------< " {
				break // artifact separator (next module)
			}
			if line == "[INFO] ------------------------------------------------------------------------" {
				break // summary separator
			}
		}
		end = i + 1
	}
	if end > start {
		block := BlockOutput{
			Type:  "initialization",
			Lines: lines[start:end],
		}

		// Extract modules and packaging from Reactor Build Order if present
		var modules []map[string]string
		inReactorOrder := false
		for i := start; i < end; i++ {
			line := lines[i]
			if line == "[INFO] Reactor Build Order:" {
				inReactorOrder = true
				continue
			}
			if inReactorOrder {
				if line == "[INFO]" || line == "[INFO] " || len(line) < 12 {
					continue // skip blank info lines
				}
				// Match against e.g. '[INFO] module-a       [jar]'
				s := line[len("[INFO] "):]
				if len(s) == 0 {
					continue
				}
				// Find last '[' and last ']', everything before is name, in brackets is type
				lastBracket := -1
				lastClose := -1
				for j := len(s) - 1; j >= 0; j-- {
					if s[j] == '[' && lastBracket == -1 {
						lastBracket = j
					}
					if s[j] == ']' && lastClose == -1 {
						lastClose = j
					}
				}
				if lastBracket != -1 && lastClose != -1 && lastBracket < lastClose {
					name := s[:lastBracket]
					typ := s[lastBracket+1 : lastClose]
					name = strings.TrimSpace(name)
					typ = strings.TrimSpace(typ)
					// Add to list
					if name != "" && typ != "" {
						modules = append(modules, map[string]string{"module": name, "packaging": typ})
					}
				} else if lastBracket == -1 && lastClose == -1 && !strings.HasPrefix(s, "-") {
					// End of reactor order list
					inReactorOrder = false
				}
			}
		}
		if len(modules) > 0 {
			block.Meta = map[string]any{"modules": modules}
		}
		return block, end - start, true
	}
	return nil, 0, false
}
