package structured

import (
	"regexp"
	"strings"
)

// Common regex patterns for Maven output parsing

var (
	// Plugin header pattern: "[INFO] --- plugin:version:goal (execution) @ artifact ---"
	PluginHeaderRegex = regexp.MustCompile(`^\[INFO\] --- [\w\-\.]+:\d+[\w\.]*:[\w\-]+( \([^)]+\))? @ [^ ]+ ---$`)

	// Module artifact separator: "[INFO] ----...< artifact >----...----"
	ModuleArtifactSeparatorRegex = regexp.MustCompile(`^\[INFO\] [-]+< [^>]+ >[-]+$`)

	// Summary line pattern for module results
	ModuleResultRegex = regexp.MustCompile(`^(.*?) +[. ]+ *(SUCCESS|FAILURE|SKIPPED) *\[ *([^\]]+?) *\]$`)
)

// isPluginHeader checks if a line is a Maven plugin execution header.
// Example: "[INFO] --- maven-clean-plugin:3.2.0:clean (default-clean) @ module-name ---"
func isPluginHeader(line string) bool {
	return strings.HasPrefix(line, "[INFO] --- ")
}

// isModuleHeader checks if a line is a Maven module header.
// Example: "[INFO] ---------------------< com.example:module-name >----------------------"
func isModuleHeader(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.Contains(line, "<") && strings.Contains(line, ">")
}

// isSummary checks if a line indicates the start of a summary section.
func isSummary(line string) bool {
	return line == "[INFO] ------------------------------------------------------------------------" ||
		strings.HasPrefix(line, "[INFO] Reactor Summary for") ||
		strings.HasPrefix(line, "[INFO] BUILD ")
}

// isEmptyInfoLine checks if a line is an empty [INFO] line.
func isEmptyInfoLine(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.TrimSpace(line[7:]) == ""
}

// isBuildSeparator checks if a line is a build separator line (line of dashes).
// Used to detect end of various sections.
func isBuildSeparator(line string) bool {
	// Standard Maven separator
	if line == "[INFO] ------------------------------------------------------------------------" {
		return true
	}
	// Shorter separator in summary
	if strings.HasPrefix(line, "[INFO] --------") {
		return true
	}
	return false
}

// isModuleSeparator checks if a line is a module packaging separator.
// Example: "[INFO] --------------------------------[ jar ]---------------------------------"
func isModuleSeparator(line string) bool {
	if !strings.HasPrefix(line, "[INFO] ") || !strings.HasSuffix(line, "---------------------------------") {
		return false
	}
	inner := line[len("[INFO] "):]
	dashCount := strings.Count(inner, "-")
	return dashCount >= 40 // At least 20 dashes on each side
}

// isReactorHeader checks if a line is the Reactor Build Order header.
func isReactorHeader(line string) bool {
	return line == "[INFO] Reactor Build Order:"
}

// isInitializationSeparator checks if a line marks the end of initialization (module header).
// This includes both the standard module header pattern and the alternate dash format.
func isInitializationSeparator(line string) bool {
	// Standard pattern: "[INFO] --- plugin:goal @ artifact ---"
	if len(line) > 10 && line[:10] == "[INFO] ---" {
		return true
	}
	// Alternate module header: "[INFO] ---< groupId:artifactId >---"
	if len(line) > 30 && (strings.HasPrefix(line, "[INFO] ----------------------< ") || strings.HasPrefix(line, "[INFO] ------------------------< ")) {
		return true
	}
	// The standard separator line at end of module header
	if isModuleSeparator(line) {
		return true
	}
	return false
}
