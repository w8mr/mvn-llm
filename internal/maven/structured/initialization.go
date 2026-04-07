package structured

import "strings"

type InitializationPhaseParser struct{}

func (p *InitializationPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx > 0 {
		return nil, 0, false
	}

	// Must start with [INFO] to be valid initialization line
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "[INFO]") {
		return nil, 0, false
	}

	start := startIdx
	end := startIdx
	seenReactorHeader := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if line == "[INFO] Reactor Build Order:" {
			seenReactorHeader = true
		}
		if seenReactorHeader {
			// Stop at next phase boundary
			if len(line) > 10 && line[:10] == "[INFO] ---" {
				break
			}
			if len(line) > 30 && (strings.HasPrefix(line, "[INFO] ----------------------< ") || strings.HasPrefix(line, "[INFO] ------------------------< ")) {
				break
			}
			if line == "[INFO] ------------------------------------------------------------------------" {
				break
			}
			// If line is not a valid [INFO] reactor line, stop here
			// This handles unparsable lines after the reactor section
			if !strings.HasPrefix(line, "[INFO]") {
				break
			}
		}
		end = i + 1
	}
	if end > start {
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
					continue
				}
				s := line[len("[INFO] "):]
				if len(s) == 0 {
					continue
				}
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
					if name != "" && typ != "" {
						modules = append(modules, map[string]string{"module": name, "packaging": typ})
					}
				} else if lastBracket == -1 && lastClose == -1 && !strings.HasPrefix(s, "-") {
					inReactorOrder = false
				}
			}
		}
		meta := map[string]any{}
		if len(modules) > 0 {
			meta["modules"] = modules
		}
		node := &Node{
			Name:  "initialization",
			Type:  "initialization",
			Lines: lines[start:end],
			Meta:  meta,
		}
		return node, end - start, true
	}
	return nil, 0, false
}
