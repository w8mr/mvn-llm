package structured

import (
	"regexp"
	"strings"
)

var (
	treeHeaderRegex     = regexp.MustCompile(`\[INFO\] --- ([\w\-\.]+):(\d+[\w\.]*):([\w\-]+)( \(([^)]+)\))? @ ([^ ]+) ---$`)
	treeRootRegex       = regexp.MustCompile(`\[INFO\]\s+([\w\.]+):([\w\-]+):([\w\-]+):([\S]+)`)
	treeDependencyRegex = regexp.MustCompile(`\[INFO\]\s+[+|\\](\-?)\s+([\w\.]+):([\w\-]+):([\w\-]+):([\S]+)(:(\w+))?`)
	treeWhitespaceRegex = regexp.MustCompile(`\[INFO\]\s+(\s+)`)
)

type DependencyTreeParser struct {
	BaseParser
}

func (p *DependencyTreeParser) NodeType() string {
	return "dependency-tree"
}

func (p *DependencyTreeParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "dependency" || strings.HasPrefix(plugin, "maven-dependency") {
			return true, 1
		}
	}
	return false, 0
}

func (p *DependencyTreeParser) ExtractLines(lines []string, startIdx int, allParsers []Parser) ([]string, int, bool) {
	ok, markerLen := p.StartMarker(lines, startIdx)
	if !ok {
		return nil, 0, false
	}

	startOfContent := startIdx + markerLen
	end := startOfContent
	for end < len(lines) {
		line := lines[end]
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "[INFO] ---") {
			break
		}
		if strings.HasPrefix(line, "[INFO] ") && !strings.Contains(line, ":") {
			break
		}
		end++
	}
	return lines[startIdx:end], end - startIdx, true
}

func (p *DependencyTreeParser) ParseMetaData(found []string) map[string]any {
	meta := make(map[string]any)

	var rootGroup, rootArtifact, rootVersion string
	var dependencies []map[string]any

	for _, line := range found {
		if matches := treeRootRegex.FindStringSubmatch(line); len(matches) >= 5 {
			rootGroup = matches[1]
			rootArtifact = matches[2]
			rootVersion = matches[4]
			meta["root"] = map[string]any{
				"groupId":    rootGroup,
				"artifactId": rootArtifact,
				"version":    rootVersion,
			}
			continue
		}

		if matches := treeDependencyRegex.FindStringSubmatch(line); len(matches) >= 7 {
			scope := ""
			if len(matches) > 7 && matches[7] != "" {
				scope = matches[7]
			}
			dep := map[string]any{
				"groupId":    matches[2],
				"artifactId": matches[3],
				"version":    matches[5],
			}
			if scope != "" {
				dep["scope"] = scope
			}
			dependencies = append(dependencies, dep)
		}
	}

	if len(dependencies) > 0 {
		meta["dependencies"] = dependencies
	}

	return meta
}

func (p *DependencyTreeParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	// Verify this is a dependency tree block
	if !isPluginHeader(lines[startIdx]) {
		return nil, 0, false
	}
	plugin := extractPluginName(lines[startIdx])
	if plugin != "dependency" && !strings.HasPrefix(plugin, "maven-dependency") {
		return nil, 0, false
	}

	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)

	name := "dependency-tree"
	if root, ok := meta["root"].(map[string]any); ok {
		if art, ok := root["artifactId"].(string); ok {
			name = art + "-tree"
		}
	}

	node := &Node{
		Name:  name,
		Type:  "dependency-tree",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}
