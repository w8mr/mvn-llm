package structured

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	testResultsRegex = regexp.MustCompile(`\[INFO\]\s+T E S T S`)
	testSummaryRegex = regexp.MustCompile(`\[ERROR\]\s+Tests run:\s+(\d+),\s+Failures:\s+(\d+),\s+Errors:\s+(\d+),\s+Skipped:\s+(\d+)`)
	testFailureRegex = regexp.MustCompile(`\[ERROR\]\s+(.+?)\s+--\s+Time elapsed`)
	testMethodRegex  = regexp.MustCompile(`at\s+(.+?)\.(.+?)\(([^:]+):(\d+)\)`)
	testRunningRegex = regexp.MustCompile(`\[INFO\]\s+Running\s+(.+)`)
)

type FailsafePhaseParser struct {
	BuildPhaseParser
}

func (p *FailsafePhaseParser) NodeType() string {
	return "build-block"
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

func (p *FailsafePhaseParser) ParseMetaData(found []string) map[string]any {
	meta := p.BuildPhaseParser.ParseMetaData(found)

	var testResults map[string]any
	var failures []map[string]any

	for _, line := range found {
		if testRunningRegex.MatchString(line) {
			matches := testRunningRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				meta["testClass"] = matches[1]
			}
		}

		if matches := testSummaryRegex.FindStringSubmatch(line); len(matches) >= 5 {
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
