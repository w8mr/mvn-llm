package structured

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	compilationErrorRegex   = regexp.MustCompile(`\[ERROR\]\s+(.+?):\[(\d+),(\d+)\]\s+(.*)`)
	compilationWarningRegex = regexp.MustCompile(`\[WARNING\]\s+(.+?):\[(\d+),(\d+)\]\s+(.*)`)
	compilingFileRegex      = regexp.MustCompile(`Compiling\s+(\d+)\s+source`)
	compiledFilesRegex      = regexp.MustCompile(`Compiling\s+(\d+)\s+files?`)
	compilerArgsRegex       = regexp.MustCompile(`with\s+javac\s+\[(.*?)\]`)
	incrementalRegex        = regexp.MustCompile(`Recompiling the module|because of changed`)
	sourceVersionRegex      = regexp.MustCompile(`-source\s+(\d+)`)
	targetVersionRegex      = regexp.MustCompile(`-target\s+(\d+)`)
	releaseVersionRegex     = regexp.MustCompile(`--release\s+(\d+)`)
)

// CompilerPhaseParser parses Maven compiler plugin output with enhanced error/warning extraction.
type CompilerPhaseParser struct {
	BuildPhaseParser
}

// NodeType returns "build-block" (same as parent, but with enhanced metadata)
func (p *CompilerPhaseParser) NodeType() string {
	return "build-block"
}

// StartMarker detects if lines[idx] is a compiler plugin block
func (p *CompilerPhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "compiler" || strings.HasPrefix(plugin, "maven-compiler") {
			return true, 1
		}
	}
	return false, 0
}

func extractPluginName(header string) string {
	m := PluginHeaderRegex.FindStringSubmatch(header)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// ParseMetaData extracts enhanced metadata for compiler plugin
func (p *CompilerPhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	var errors []map[string]any
	var warnings []map[string]any
	var sourceFiles int
	var compilerArgs []string
	var sourceVersion, targetVersion, releaseVersion string
	var incremental bool

	for _, line := range found {
		// Parse compilation errors: [ERROR] /path/File.java:[line,col] message
		if matches := compilationErrorRegex.FindStringSubmatch(line); len(matches) >= 5 {
			file := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])
			message := matches[4]
			errors = append(errors, map[string]any{
				"file":    file,
				"line":    lineNum,
				"column":  colNum,
				"message": message,
			})
		}

		// Parse compilation warnings
		if matches := compilationWarningRegex.FindStringSubmatch(line); len(matches) >= 5 {
			file := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])
			message := matches[4]
			warnings = append(warnings, map[string]any{
				"file":    file,
				"line":    lineNum,
				"column":  colNum,
				"message": message,
			})
		}

		// Check for compiling X source files
		if matches := compiledFilesRegex.FindStringSubmatch(line); len(matches) >= 2 {
			n, _ := strconv.Atoi(matches[1])
			sourceFiles = n
		}

		// Check for compiler args
		if matches := compilerArgsRegex.FindStringSubmatch(line); len(matches) >= 2 {
			compilerArgs = strings.Split(matches[1], " ")
		}

		// Check for incremental build
		if incrementalRegex.MatchString(line) {
			incremental = true
		}

		// Extract version flags
		if matches := sourceVersionRegex.FindStringSubmatch(line); len(matches) >= 2 {
			sourceVersion = matches[1]
		}
		if matches := targetVersionRegex.FindStringSubmatch(line); len(matches) >= 2 {
			targetVersion = matches[1]
		}
		if matches := releaseVersionRegex.FindStringSubmatch(line); len(matches) >= 2 {
			releaseVersion = matches[1]
		}
	}

	// Add compiler-specific fields to meta
	if len(errors) > 0 {
		meta["compilationErrors"] = errors
	}
	if len(warnings) > 0 {
		meta["compilationWarnings"] = warnings
	}
	if sourceFiles > 0 {
		meta["sourceFiles"] = sourceFiles
	}
	if len(compilerArgs) > 0 {
		meta["compilerArgs"] = compilerArgs
	}
	if sourceVersion != "" {
		meta["sourceVersion"] = sourceVersion
	}
	if targetVersion != "" {
		meta["targetVersion"] = targetVersion
	}
	if releaseVersion != "" {
		meta["releaseVersion"] = releaseVersion
	}
	if incremental {
		meta["incremental"] = true
	}

	return meta
}

// Parse combines ExtractLines and ParseMetaData
func (p *CompilerPhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
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
