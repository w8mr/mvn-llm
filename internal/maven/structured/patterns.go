package structured

import (
	"regexp"
	"strings"
)

// Common regex patterns for Maven output parsing
// Each regex is designed to both MATCH and EXTRACT data

var (
	// Module header pattern: captures groupId:artifactId
	// Matches: [INFO] ---------------------< com.example:module-name >----------------------
	// Capture groups: [1] = groupId:artifactId
	ModuleHeaderRegex = regexp.MustCompile(`^\[INFO\] [-]+< ([^>]+) >[-]+$`)

	// Module packaging separator pattern: captures packaging type
	// Matches: [INFO] --------------------------------[ jar ]---------------------------------
	// Capture groups: [1] = packaging (jar, pom, war, etc.)
	ModuleSeparatorRegex = regexp.MustCompile(`\[INFO\] ([-]+)\[([^\]]+)\]([-]+)$`)

	// Plugin header pattern: captures plugin details
	// Matches: [INFO] --- plugin:version:goal (execution) @ artifact ---
	// Capture groups: [1]=plugin, [2]=version, [3]=goal, [5]=executionId, [6]=artifactId
	PluginHeaderRegex = regexp.MustCompile(`^\[INFO\] --- ([\w\-\.]+):(\d+[\w\.]*):([\w\-]+)( \(([^)]+)\))? @ ([^ ]+) ---$`)

	// Summary line pattern for module results
	// Matches: module-name ............................... SUCCESS [  1.234 s]
	// Capture groups: [1]=name, [2]=status, [3]=time
	ModuleResultRegex = regexp.MustCompile(`^(.*?) +[. ]+ *(SUCCESS|FAILURE|SKIPPED) *\[ *([^\]]+?) *\]$`)
)

// isModuleHeader checks if a line is a Maven module header.
// Example: [INFO] ---------------------< com.example:module-name >----------------------
func isModuleHeader(line string) bool {
	return ModuleHeaderRegex.MatchString(line)
}

// isModuleSeparator checks if a line is a module packaging separator.
// Example: [INFO] --------------------------------[ jar ]---------------------------------
func isModuleSeparator(line string) bool {
	return ModuleSeparatorRegex.MatchString(line)
}

// isPluginHeader checks if a line is a Maven plugin execution header.
// Example: [INFO] --- maven-clean-plugin:3.2.0:clean (default-clean) @ module-name ---
func isPluginHeader(line string) bool {
	return strings.HasPrefix(line, "[INFO] --- ")
}

// isBuildBlockHeader is an alias for isPluginHeader (they are the same pattern)
func isBuildBlockHeader(line string) bool {
	return isPluginHeader(line)
}

// isSeparator checks if a line is a build separator (all dashes).
// Example: [INFO] ------------------------------------------------------------------------
func isSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "[INFO] ------------------------------------------------------------------------" ||
		strings.HasPrefix(trimmed, "[INFO] ----")
}

// isSummary checks if a line indicates the start of a summary section.
func isSummary(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "[INFO] ------------------------------------------------------------------------" ||
		strings.HasPrefix(trimmed, "[INFO] Reactor Summary for") ||
		strings.HasPrefix(trimmed, "[INFO] BUILD ")
}

// isEmptyInfoLine checks if a line is an empty [INFO] line.
func isEmptyInfoLine(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.TrimSpace(line[7:]) == ""
}

// isBuildSeparator is an alias for isSeparator (kept for backward compatibility)
var isBuildSeparator = isSeparator

// isReactorHeader checks if a line is the Reactor Build Order header.
func isReactorHeader(line string) bool {
	return line == "[INFO] Reactor Build Order:"
}

// isInitializationSeparator checks if a line marks the end of initialization phase.
// Returns true for module headers or build block headers.
func isInitializationSeparator(line string) bool {
	return isModuleHeader(line) || isPluginHeader(line)
}
