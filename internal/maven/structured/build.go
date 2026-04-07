package structured

import (
	"regexp"
)

type BuildPhaseParser struct {
	SubParsers []Parser
}

var pluginHeaderRegex = regexp.MustCompile(`^\[INFO\] --- [\w\-\.]+:\d+[\w\.]*:[\w\-]+( \([^)]+\))? @ [^ ]+ ---$`)
var moduleArtifactSeparatorRegex = regexp.MustCompile(`^\[INFO\] [-]+< [^>]+ >[-]+$`)

func (p *BuildPhaseParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx >= len(lines) {
		return nil, 0, false
	}
	if !pluginHeaderRegex.MatchString(lines[startIdx]) {
		return nil, 0, false
	}
	start := startIdx
	end := start + 1
	for end < len(lines) {
		if pluginHeaderRegex.MatchString(lines[end]) {
			break
		}
		if moduleArtifactSeparatorRegex.MatchString(lines[end]) {
			break
		}
		if lines[end] == "[INFO] ------------------------------------------------------------------------" {
			break
		}
		end++
	}
	for end < len(lines) && lines[end] == "[INFO] " {
		end++
	}
	status := "SUCCESS"
	for _, l := range lines[start:end] {
		if len(l) > 0 {
			if l[0] == '[' {
				if len(l) > 8 && l[:8] == "[ERROR] " {
					status = "FAILED"
					break
				} else if len(l) > 10 && l[:10] == "[WARNING] " && status != "FAILED" {
					status = "SUCCESS-WITH-WARNINGS"
				}
			}
		}
	}

	header := lines[start]
	plugin := ""
	version := ""
	goal := ""
	executionId := ""
	artifactId := ""
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

	meta := map[string]any{"status": status}
	if plugin != "" {
		meta["plugin"] = plugin
		meta["version"] = version
		meta["goal"] = goal
		meta["artifactId"] = artifactId
		if executionId != "" {
			meta["executionId"] = executionId
		}
	}

	node := &Node{
		Name:  plugin,
		Type:  "build-block",
		Lines: lines[start:end],
		Meta:  meta,
	}

	return node, end - startIdx, true
}
