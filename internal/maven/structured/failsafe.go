package structured

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	testResultsRegex       = regexp.MustCompile(`\[INFO\]\s+T E S T S`)
	testSummaryFullRegex   = regexp.MustCompile(`\[ERROR\]\s+Tests run:\s+(\d+),\s+Failures:\s+(\d+),\s+Errors:\s+(\d+),\s+Skipped:\s+(\d+)(?:,\s+Time elapsed:.*)?`)
	testFailureRegex       = regexp.MustCompile(`\[ERROR\]\s+(.+?)\s+--\s+Time elapsed`)
	testFailureDetailRegex = regexp.MustCompile(`\[ERROR\]\s+(\w+)\.([^\s:]+).*:\s*(.*)$`)
	testRunningRegex       = regexp.MustCompile(`\[INFO\]\s+Running\s+(.+)`)
	testProviderRegex      = regexp.MustCompile(`\[INFO\]\s+Using\s+(.+?)\s+provider\s+(.+)`)
)

type SurefirePhaseParser struct {
	BaseParser
}

func (p *SurefirePhaseParser) NodeType() string {
	return "surefire-block"
}

func (p *SurefirePhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "surefire" || strings.HasPrefix(plugin, "maven-surefire") {
			return true, 1
		}
	}
	return false, 0
}

func (p *SurefirePhaseParser) ExtractLines(lines []string, startIdx int, allParsers []Parser) ([]string, int, bool) {
	ok, markerLen := p.StartMarker(lines, startIdx)
	if !ok {
		return nil, 0, false
	}

	startOfContent := startIdx + markerLen
	_, consumed := ParseUntilNextBlock(lines, startOfContent, allParsers, "surefire-block")

	totalConsumed := markerLen + consumed
	return lines[startIdx : startIdx+totalConsumed], totalConsumed, true
}

func (p *SurefirePhaseParser) ParseMetaData(found []string) map[string]any {
	meta := make(map[string]any)

	header := found[0]
	plugin := ""
	version := ""
	goal := ""
	executionId := ""
	artifactId := ""
	m := PluginHeaderRegex.FindStringSubmatch(header)
	if len(m) >= 7 {
		plugin = m[1]
		version = m[2]
		goal = m[3]
		if len(m[5]) > 0 {
			executionId = m[5]
		}
		artifactId = m[6]
	}

	status := "SUCCESS"
	for _, l := range found {
		if len(l) > 0 && l[0] == '[' {
			if len(l) > 8 && l[:8] == "[ERROR] " {
				status = "FAILED"
			}
		}
	}

	meta["plugin"] = plugin
	meta["version"] = version
	meta["goal"] = goal
	meta["executionId"] = executionId
	meta["artifactId"] = artifactId
	meta["status"] = status
	meta["summary"] = extractTestSummary(found)

	var testResults map[string]any
	var failures []map[string]any

	for _, line := range found {
		if matches := testProviderRegex.FindStringSubmatch(line); len(matches) >= 3 {
			meta["provider"] = strings.TrimSpace(matches[2])
		}

		if testRunningRegex.MatchString(line) {
			matches := testRunningRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				meta["testClass"] = matches[1]
			}
		}

		if matches := testSummaryFullRegex.FindStringSubmatch(line); len(matches) >= 5 {
			runs, _ := strconv.Atoi(matches[1])
			failuresCount, _ := strconv.Atoi(matches[2])
			errorsCount, _ := strconv.Atoi(matches[3])
			skippedCount, _ := strconv.Atoi(matches[4])

			testResults = map[string]any{
				"runs":     runs,
				"failures": failuresCount,
				"errors":   errorsCount,
				"skipped":  skippedCount,
			}
		}

		if matches := testFailureRegex.FindStringSubmatch(line); len(matches) >= 2 {
			failures = append(failures, map[string]any{
				"test": matches[1],
			})
		}

		if matches := testFailureDetailRegex.FindStringSubmatch(line); len(matches) >= 4 {
			failures = append(failures, map[string]any{
				"class":  matches[1],
				"method": matches[2],
				"error":  matches[3],
			})
		}
	}

	if testResults != nil {
		meta["testResults"] = testResults
	}
	if len(failures) > 0 {
		meta["testFailures"] = failures
	}

	return meta
}

func (p *SurefirePhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
	// Verify this is actually a surefire plugin block
	if !isPluginHeader(lines[startIdx]) {
		return nil, 0, false
	}
	plugin := extractPluginName(lines[startIdx])
	if plugin != "surefire" && !strings.HasPrefix(plugin, "maven-surefire") {
		return nil, 0, false
	}

	found, consumed, ok := p.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		return nil, 0, false
	}
	meta := p.ParseMetaData(found)

	// Trace: check what's in meta
	if meta["testResults"] != nil {
		// meta has testResults! This should work
	}

	node := &Node{
		Name:  meta["plugin"].(string),
		Type:  "build-block",
		Lines: found,
		Meta:  meta,
	}
	return node, consumed, true
}

func extractTestSummary(found []string) string {
	var lastError, lastWarning, lastInfo string

	for i, l := range found {
		if i == 0 {
			continue
		}
		if len(l) == 0 {
			continue
		}
		if strings.HasPrefix(l, "[ERROR] ") && len(l) > 8 {
			lastError = strings.TrimSpace(l[8:])
		} else if strings.HasPrefix(l, "[WARNING] ") && len(l) > 10 {
			lastWarning = strings.TrimSpace(l[10:])
		} else if strings.HasPrefix(l, "[INFO] ") && len(l) > 6 {
			info := strings.TrimSpace(l[6:])
			if info != "" {
				lastInfo = info
			}
		}
	}

	if lastError != "" {
		return lastError
	}
	if lastWarning != "" {
		return lastWarning
	}
	if lastInfo != "" {
		return lastInfo
	}
	return "Successful"
}

type FailsafePhaseParser struct {
	BaseParser
}

func (p *FailsafePhaseParser) NodeType() string {
	return "failsafe-block"
}

func (p *FailsafePhaseParser) StartMarker(lines []string, idx int) (bool, int) {
	if idx >= len(lines) {
		return false, 0
	}
	if isPluginHeader(lines[idx]) {
		plugin := extractPluginName(lines[idx])
		if plugin == "failsafe" || strings.HasPrefix(plugin, "maven-failsafe") {
			return true, 1
		}
	}
	return false, 0
}

func (p *FailsafePhaseParser) ExtractLines(lines []string, startIdx int, allParsers []Parser) ([]string, int, bool) {
	ok, markerLen := p.StartMarker(lines, startIdx)
	if !ok {
		return nil, 0, false
	}

	startOfContent := startIdx + markerLen
	_, consumed := ParseUntilNextBlock(lines, startOfContent, allParsers, "failsafe-block")

	totalConsumed := markerLen + consumed
	return lines[startIdx : startIdx+totalConsumed], totalConsumed, true
}

func (p *FailsafePhaseParser) ParseMetaData(found []string) map[string]any {
	meta := make(map[string]any)

	header := found[0]
	plugin := ""
	version := ""
	goal := ""
	executionId := ""
	artifactId := ""
	m := PluginHeaderRegex.FindStringSubmatch(header)
	if len(m) >= 7 {
		plugin = m[1]
		version = m[2]
		goal = m[3]
		if len(m[5]) > 0 {
			executionId = m[5]
		}
		artifactId = m[6]
	}

	status := "SUCCESS"
	for _, l := range found {
		if len(l) > 0 && l[0] == '[' {
			if len(l) > 8 && l[:8] == "[ERROR] " {
				status = "FAILED"
			}
		}
	}

	meta["plugin"] = plugin
	meta["version"] = version
	meta["goal"] = goal
	meta["executionId"] = executionId
	meta["artifactId"] = artifactId
	meta["status"] = status
	meta["summary"] = extractTestSummary(found)

	var testResults map[string]any
	var failures []map[string]any

	for _, line := range found {
		if testRunningRegex.MatchString(line) {
			matches := testRunningRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				meta["testClass"] = matches[1]
			}
		}

		if matches := testSummaryFullRegex.FindStringSubmatch(line); len(matches) >= 5 {
			runs, _ := strconv.Atoi(matches[1])
			failuresCount, _ := strconv.Atoi(matches[2])
			errorsCount, _ := strconv.Atoi(matches[3])
			skippedCount, _ := strconv.Atoi(matches[4])

			testResults = map[string]any{
				"runs":     runs,
				"failures": failuresCount,
				"errors":   errorsCount,
				"skipped":  skippedCount,
			}
		}

		if matches := testFailureRegex.FindStringSubmatch(line); len(matches) >= 2 {
			failures = append(failures, map[string]any{
				"test": matches[1],
			})
		}

		if matches := testFailureDetailRegex.FindStringSubmatch(line); len(matches) >= 4 {
			failures = append(failures, map[string]any{
				"class":  matches[1],
				"method": matches[2],
				"error":  matches[3],
			})
		}
	}

	if testResults != nil {
		meta["testResults"] = testResults
	}
	if len(failures) > 0 {
		meta["testFailures"] = failures
	}

	return meta
}

func (p *FailsafePhaseParser) Parse(lines []string, startIdx int, allParsers []Parser) (*Node, int, bool) {
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
