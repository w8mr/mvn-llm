package structured

import (
	"strconv"
	"strings"
)

// ModulePhaseParser parses the per-module header and hands off the remaining lines to subparsers (e.g., for build/plugin blocks)
type ModulePhaseParser struct{}

func (p *ModulePhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	// Look for module section separator (less strict, supports all module header lines)
	if strings.HasPrefix(lines[startIdx], "[INFO] ") && strings.Contains(lines[startIdx], "<") && strings.Contains(lines[startIdx], ">") {
		idx := startIdx
		// Loosen header: require only the first 3 canonical lines
		if idx+2 < len(lines) &&
			strings.HasPrefix(lines[idx+1], "[INFO] Building") &&
			strings.HasPrefix(lines[idx+2], "[INFO]   from") {
			end := idx + 3
			// Optionally include the artifact separator and its blank
			if end+1 < len(lines) &&
				strings.HasPrefix(lines[end], "[INFO] --------------------------------[") &&
				strings.HasSuffix(lines[end], "]---------------------------------") &&
				strings.HasPrefix(lines[end+1], "[INFO] ") && strings.TrimSpace(lines[end+1][7:]) == "" {
				end += 2
			}
			// Claim any additional trailing blank [INFO] lines, so phase loop cannot land on these before plugin header
			for end < len(lines) && strings.HasPrefix(lines[end], "[INFO] ") && strings.TrimSpace(lines[end][7:]) == "" {
				end++
			}
			// Parse metadata
			meta := make(map[string]any)
			// Line 1: <groupId:artifactId>
			l1 := lines[idx]
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
			// Line 2: Building name version [x/y]
			l2 := lines[idx+1][len("[INFO] "):]
			// e.g., 'Building module-a 1.0-SNAPSHOT [2/3]'
			var name, version string
			if sp := strings.SplitN(l2, " ", 3); len(sp) == 3 && strings.HasPrefix(sp[0], "Building") {
				toks := strings.SplitN(l2, " ", 4)
				if len(toks) >= 3 {
					name = toks[1]
					version = toks[2]
					meta["name"] = name
					meta["version"] = version
					if len(toks) == 4 {
						idxBracket1 := strings.Index(toks[3], "[")
						idxBracket2 := strings.Index(toks[3], "/")
						idxBracket3 := strings.Index(toks[3], "]")
						if idxBracket1 != -1 && idxBracket2 != -1 && idxBracket3 != -1 {
							// parse [2/3]
							x := toks[3][idxBracket1+1 : idxBracket2]
							y := toks[3][idxBracket2+1 : idxBracket3]
							// best effort Atoi
							if xi, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
								meta["moduleIndex"] = xi
							}
							if yi, err := strconv.Atoi(strings.TrimSpace(y)); err == nil {
								meta["moduleCount"] = yi
							}
						}
					}
				}
			}

			// Line 3: from ...
			l3 := lines[idx+2]
			if strings.HasPrefix(l3, "[INFO]   from") {
				meta["pomFile"] = strings.TrimSpace(strings.TrimPrefix(l3, "[INFO]   from"))
			}
			// Find artifact separator with packaging between lines idx+3 and end
			for i := idx + 3; i < end; i++ {
				l := lines[i]
				if strings.HasPrefix(l, "[INFO] --------------------------------[") && strings.HasSuffix(l, "]---------------------------------") {
					// Find the bracket pair containing the packaging, immediately after '[INFO] --------------------------------['
					sepPrefix := "[INFO] --------------------------------["
					start := strings.Index(l, sepPrefix)
					if start != -1 {
						left := start + len(sepPrefix)
						right := strings.Index(l[left:], "]")
						if right != -1 {
							meta["packaging"] = strings.TrimSpace(l[left : left+right])
						}
					}
					break
				}
			}
			block := BlockOutput{
				Type:  "module-header",
				Lines: lines[idx:end],
				Meta:  meta,
			}
			return block, end - idx, true
		}
	}
	return nil, 0, false
}
