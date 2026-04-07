package structured

import (
	"regexp"
)

// BuildPhaseParser parses the main build phase output blocks.
type BuildPhaseParser struct{}

var pluginHeaderRegex = regexp.MustCompile(`^\[INFO\] --- [\w\-\.]+:\d+[\w\.]*:[\w\-]+( \([^)]+\))? @ [^ ]+ ---$`)
var moduleArtifactSeparatorRegex = regexp.MustCompile(`^\[INFO\] [-]+< [^>]+ >[-]+$`)

func (p *BuildPhaseParser) Parse(lines []string, startIdx int) (any, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	// Check if current line is a plugin header - do NOT scan forward
	if !pluginHeaderRegex.MatchString(lines[startIdx]) {
		return nil, 0, false
	}
	start := startIdx
	end := start + 1
	for end < len(lines) {
		if pluginHeaderRegex.MatchString(lines[end]) {
			break // next plugin phase
		}
		if moduleArtifactSeparatorRegex.MatchString(lines[end]) {
			break // next module starts; do not include this line
		}
		if lines[end] == "[INFO] ------------------------------------------------------------------------" {
			break // summary separator
		}
		end++
	}
	// Claim any immediate trailing blank [INFO] lines, so registry cannot re-match this plugin header.
	for end < len(lines) && lines[end] == "[INFO] " {
		end++
	}
	// Determine status by scanning for ERROR, WARNING, INFO lines
	status := "SUCCESS"
	for _, l := range lines[start:end] {
		if len(l) > 0 {
			if l[0] == '[' {
				if len(l) > 8 && l[:8] == "[ERROR] " {
					status = "FAILED"
					break // error always wins
				} else if len(l) > 10 && l[:10] == "[WARNING] " && status != "FAILED" {
					status = "SUCCESS-WITH-WARNINGS"
				}
			}
		}
	}

	// parse build/plugin header meta fields from the first line
	header := lines[start]
	plugin := ""
	version := ""
	goal := ""
	executionId := ""
	artifactId := ""
	// Header sample: [INFO] --- compiler:3.15.0:compile (default-compile) @ module-a ---
	// Regex: ^\[INFO\] --- ([\w\-\.]+):(\d+[\w\.]*):([\w\-]+)( \(([^)]+)\))? @ ([^ ]+) ---$
	pluginHeaderParseRe := regexp.MustCompile(`^\[INFO\] --- ([\w\-\.]+):(\d+[\w\.]*):([\w\-]+)( \(([^)]+)\))? @ ([^ ]+) ---`)
	m := pluginHeaderParseRe.FindStringSubmatch(header)
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
		"status": status,
	}
	if plugin != "" {
		meta["plugin"] = plugin
	}
	if version != "" {
		meta["version"] = version
	}
	if goal != "" {
		meta["goal"] = goal
	}
	if executionId != "" {
		meta["executionId"] = executionId
	}
	if artifactId != "" {
		meta["artifactId"] = artifactId
	}

	block := BlockOutput{
		Type:  "build-block",
		Lines: lines[start:end],
		Meta:  meta,
	}
	return block, end - startIdx, true
}
