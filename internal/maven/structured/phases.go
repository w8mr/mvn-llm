package structured

import (
	"regexp"
	"strings"
)

var (
	jarCreatedRegex      = regexp.MustCompile(`\[INFO\]\s+Building jar:\s+(.+)`)
	jarManifestRegex     = regexp.MustCompile(`\[INFO\]\s+Manifest:\s+(.+)`)
	installDeployedRegex = regexp.MustCompile(`\[INFO\]\s+Installing\s+(.+)\s+to\s+(.+)`)
)

type JarPhaseParser struct {
	BuildPhaseParser
}

func (p *JarPhaseParser) NodeType() string {
	return "build-block"
}

func (p *JarPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "jar" || strings.HasPrefix(plugin, "maven-jar") {
			return true, 1
		}
	}
	return false, 0
}

func (p *JarPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := jarCreatedRegex.FindStringSubmatch(line); len(matches) >= 2 {
			meta["jarFile"] = matches[1]
		}
		if matches := jarManifestRegex.FindStringSubmatch(line); len(matches) >= 2 {
			meta["manifest"] = matches[1]
		}
	}

	return meta
}

func (p *JarPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

type InstallPhaseParser struct {
	BuildPhaseParser
}

func (p *InstallPhaseParser) NodeType() string {
	return "build-block"
}

func (p *InstallPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "install" || strings.HasPrefix(plugin, "maven-install") {
			return true, 1
		}
	}
	return false, 0
}

func (p *InstallPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := installDeployedRegex.FindStringSubmatch(line); len(matches) >= 3 {
			meta["artifact"] = matches[1]
			meta["path"] = matches[2]
		}
	}

	return meta
}

func (p *InstallPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

type DeployPhaseParser struct {
	BuildPhaseParser
}

func (p *DeployPhaseParser) NodeType() string {
	return "build-block"
}

func (p *DeployPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "deploy" || strings.HasPrefix(plugin, "maven-deploy") {
			return true, 1
		}
	}
	return false, 0
}

func (p *DeployPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := installDeployedRegex.FindStringSubmatch(line); len(matches) >= 3 {
			meta["artifact"] = matches[1]
			meta["path"] = matches[2]
		}
	}

	return meta
}

func (p *DeployPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

var (
	resourcesCopiedRegex = regexp.MustCompile(`\[INFO\]\s+Using\s+(.+?)\s+encoding\s+to\s+copy\s+(.+)`)
	resourcesTargetRegex = regexp.MustCompile(`\[INFO\]\s+Copying\s+(.+)\s+to\s+(.+)`)
	resourcesSkipRegex   = regexp.MustCompile(`\[INFO\]\s+skip\s+(.+)\s+resourceDirectory`)
)

type ResourcesPhaseParser struct {
	BuildPhaseParser
}

func (p *ResourcesPhaseParser) NodeType() string {
	return "build-block"
}

func (p *ResourcesPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "resources" || strings.HasPrefix(plugin, "maven-resource") {
			return true, 1
		}
		if plugin == "testResources" || strings.HasPrefix(plugin, "maven-test-resources") {
			return true, 1
		}
	}
	return false, 0
}

func (p *ResourcesPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := resourcesSkipRegex.FindStringSubmatch(line); len(matches) >= 2 {
			meta["skipped"] = matches[1]
		}
		if matches := resourcesCopiedRegex.FindStringSubmatch(line); len(matches) >= 3 {
			meta["encoding"] = matches[1]
			meta["resourceType"] = matches[2]
		}
		if matches := resourcesTargetRegex.FindStringSubmatch(line); len(matches) >= 3 {
			meta["source"] = matches[1]
			meta["target"] = matches[2]
		}
	}

	return meta
}

func (p *ResourcesPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

var (
	sourceGeneratedRegex = regexp.MustCompile(`\[INFO\]\s+Building\s+jar:\s+(.+)`)
)

type SourcePhaseParser struct {
	BuildPhaseParser
}

func (p *SourcePhaseParser) NodeType() string {
	return "build-block"
}

func (p *SourcePhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "source" || strings.HasPrefix(plugin, "maven-source") {
			return true, 1
		}
	}
	return false, 0
}

func (p *SourcePhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := sourceGeneratedRegex.FindStringSubmatch(line); len(matches) >= 2 {
			meta["sourceJar"] = matches[1]
		}
	}

	return meta
}

func (p *SourcePhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

var (
	surefireProviderRegex = regexp.MustCompile(`\[INFO\]\s+Using\s+(.+?)\s+provider`)
	testsSummaryRegex     = regexp.MustCompile(`\[ERROR\]\s+Tests run:\s+(\d+),\s+Failures:\s+(\d+),\s+Errors:\s+(\d+),\s+Skipped:\s+(\d+)`)
)

type CleanPhaseParser struct {
	BuildPhaseParser
}

func (p *CleanPhaseParser) NodeType() string {
	return "build-block"
}

func (p *CleanPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "clean" || strings.HasPrefix(plugin, "maven-clean") {
			return true, 1
		}
	}
	return false, 0
}

func (p *CleanPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		content := strings.TrimSpace(line[6:])
		if strings.HasPrefix(content, "Deleting") {
			meta["deleted"] = content
		}
	}

	return meta
}

func (p *CleanPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

var (
	warOverlayRegex = regexp.MustCompile(`\[INFO\]\s+--- war:.*Overlay`)
	warWebXmlRegex  = regexp.MustCompile(`\[INFO\]\s+webxml\.webXml\s+->\s+(.+)`)
)

type WarPhaseParser struct {
	BuildPhaseParser
}

func (p *WarPhaseParser) NodeType() string {
	return "build-block"
}

func (p *WarPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "war" || strings.HasPrefix(plugin, "maven-war") {
			return true, 1
		}
	}
	return false, 0
}

func (p *WarPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	for _, line := range found {
		if matches := warOverlayRegex.FindStringSubmatch(line); len(matches) >= 1 {
			meta["overlay"] = "enabled"
		}
		if matches := warWebXmlRegex.FindStringSubmatch(line); len(matches) >= 2 {
			meta["webXml"] = matches[1]
		}
	}

	return meta
}

func (p *WarPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

var (
	earIncludedRegex = regexp.MustCompile(`\[INFO\]\s+Including:\s+(.+)`)
	earExcludedRegex = regexp.MustCompile(`\[INFO\]\s+Excluding:\s+(.+)`)
)

type EarPhaseParser struct {
	BuildPhaseParser
}

func (p *EarPhaseParser) NodeType() string {
	return "build-block"
}

func (p *EarPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "ear" || strings.HasPrefix(plugin, "maven-ear") {
			return true, 1
		}
	}
	return false, 0
}

func (p *EarPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	var inclusions []string
	var exclusions []string

	for _, line := range found {
		if matches := earIncludedRegex.FindStringSubmatch(line); len(matches) >= 2 {
			inclusions = append(inclusions, matches[1])
		}
		if matches := earExcludedRegex.FindStringSubmatch(line); len(matches) >= 2 {
			exclusions = append(exclusions, matches[1])
		}
	}

	if len(inclusions) > 0 {
		meta["includes"] = inclusions
	}
	if len(exclusions) > 0 {
		meta["excludes"] = exclusions
	}

	return meta
}

func (p *EarPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)
	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}
