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

	// Summary line pattern for module results (in reactor summary)
	// Matches: [INFO] module-name ........................ SUCCESS [ 9.788 s]
	//         [INFO] module-name ........................ SKIPPED
	// Capture groups: [1]=name, [2]=status, [3]=time (optional)
	ModuleResultRegex = regexp.MustCompile(`^\[INFO\] (.+?)[. ]+(SUCCESS|FAILURE|SKIPPED)(?:\[ (.+) \])?$`)

	// Build status patterns
	BuildSuccessRegex = regexp.MustCompile(`^\[INFO\] BUILD SUCCESS$`)
	BuildFailureRegex = regexp.MustCompile(`^\[INFO\] BUILD FAILURE$`)

	// Time patterns
	TotalTimeRegex  = regexp.MustCompile(`^\[INFO\] Total time: (.+)$`)
	FinishedAtRegex = regexp.MustCompile(`^\[INFO\] Finished at: (.+)$`)

	// Initialization pattern
	ReactorSummaryForRegex = regexp.MustCompile(`^\[INFO\] Reactor Summary( for)?(:)?$`)
)

// Module header detection patterns (prefix-based, no extraction needed)
func isModuleBuildingLine(line string) bool { return strings.HasPrefix(line, "[INFO] Building") }
func isModulePomFileLine(line string) bool  { return strings.HasPrefix(line, "[INFO]   from") }

// Module header detection (full line matching)
func isModuleHeader(line string) bool {
	return ModuleHeaderRegex.MatchString(line)
}

// Module separator detection (full line matching)
func isModuleSeparator(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.HasSuffix(line, "---------------------------------")
}

// Plugin/build block detection (prefix-based, fast)
func isPluginHeader(line string) bool {
	return PluginHeaderRegex.MatchString(line)
}

// Alias for clarity
var isBuildBlockHeader = isPluginHeader

// Long separator detection (full 70-dash line) - used for build block termination
func isLongSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "[INFO] ------------------------------------------------------------------------"
}

// Separator detection (any dash line) - for general use
func isSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "[INFO] -------------------------------------------------------"
}

// Alias for compatibility
var isBuildSeparator = isSeparator

// Empty line detection
func isEmptyInfoLine(line string) bool {
	return strings.HasPrefix(line, "[INFO] ") && strings.TrimSpace(line[7:]) == ""
}

// Reactor header detection
func isReactorHeader(line string) bool {
	return line == "[INFO] Reactor Build Order:"
}

// Reactor Summary for header detection
func isReactorSummaryFor(line string) bool {
	return ReactorSummaryForRegex.MatchString(line)
}

// Initialization separator detection (module header OR plugin header)
func isInitializationSeparator(line string) bool {
	return isModuleHeader(line) || isPluginHeader(line)
}

// Build status detection
func isBuildSuccess(line string) bool { return BuildSuccessRegex.MatchString(line) }
func isBuildFailure(line string) bool { return BuildFailureRegex.MatchString(line) }

// Module result line detection (in reactor summary)
func isModuleResultLine(line string) bool {
	return ModuleResultRegex.MatchString(line)
}

// Time info detection
func isTotalTime(line string) bool  { return TotalTimeRegex.MatchString(line) }
func isFinishedAt(line string) bool { return FinishedAtRegex.MatchString(line) }

// Empty line detection (any line with only whitespace after [INFO] prefix)
func isEmptyLine(line string) bool { return strings.TrimSpace(line) == "" }
