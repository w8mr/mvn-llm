package maven

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type SurefireTestCase struct {
	Name    string
	Class   string
	Time    float64
	Error   string
	Failure string
}

type SurefireTestSuite struct {
	Name      string
	Tests     int
	Failures  int
	Errors    int
	Skipped   int
	Time      float64
	TestCases []SurefireTestCase
}

type SurefireReport struct {
	Suites []SurefireTestSuite
}

// MavenOutput represents the structured output from a Maven build
type MavenOutput struct {
	Status          string
	FailedModule    string
	FailureLocation string
	ErrorMessage    string
	Errors          []string
	TestSummary     string
	TestSuites      []SurefireTestSuite
	ResumeModule    string
}

// ParseMavenOutput parses Maven output and extracts structured information
func ParseMavenOutput(output string, projectRoot string) MavenOutput {
	result := MavenOutput{
		Errors: []string{},
	}

	lines := strings.Split(output, "\n")
	var statusRE = regexp.MustCompile(`\[INFO\] (BUILD SUCCESS|BUILD FAILURE)`)
	var moduleRE = regexp.MustCompile(`\[INFO\] ([^\s]+)\s+\.+\s+(SUCCESS|FAILURE|SKIPPED)`)
	var compilationRE = regexp.MustCompile(`\[ERROR\] (.*):(\d+):?(\d+)? (.+)`)
	var fatalRE = regexp.MustCompile(`\[FATAL\]|Non-parseable POM`)
	var testSummaryRE = regexp.MustCompile(`Tests run:`)

	// Local variable to track raw Maven status during parsing
	rawStatus := ""

	// Local variable to track specific failure type (COMPILE_ERROR, TEST_FAILURE, etc.)
	detectedFailure := ""

	// Local variable to track test assertion message for test failures
	testAssertionMsg := ""

	// Local variable to track test class name for test failures
	testClassName := ""

	// Local variable to track test method name for test failures
	testMethodName := ""

	// Local variable to track test line number for test failures
	testLineNum := ""

	// Store project root for making paths relative
	projRoot := projectRoot

	for _, line := range lines {
		// Build status - only set if not already a specific failure type
		if m := statusRE.FindStringSubmatch(line); m != nil {
			if rawStatus == "" || rawStatus == "BUILD FAILURE" {
				rawStatus = m[1]
			}
		}

		// Also detect BUILD FAILURE from non-INFO lines - but don't overwrite specific failure types
		if strings.Contains(line, "BUILD FAILURE") {
			if rawStatus == "" {
				rawStatus = "BUILD FAILURE"
			}
		}

		// If we have errors but no status yet, set status to BUILD FAILURE
		if rawStatus == "" && len(result.Errors) > 0 {
			rawStatus = "BUILD FAILURE"
		}

		// Check for test summary
		if testSummaryRE.MatchString(line) && strings.Contains(line, "Failures:") {
			result.TestSummary = extractTestSummary(line)
		}

		// Failed module from reactor summary
		if m := moduleRE.FindStringSubmatch(line); m != nil {
			if m[2] == "FAILURE" {
				result.FailedModule = m[1]
			}
		}

		// Handle both [ERROR] and [FATAL] lines
		errLine := line
		if strings.HasPrefix(line, "[FATAL]") {
			errLine = strings.TrimPrefix(line, "[FATAL] ")
		} else if strings.HasPrefix(line, "[ERROR]") {
			errLine = strings.TrimPrefix(line, "[ERROR] ")
		} else {
			continue
		}
		result.Errors = append(result.Errors, errLine)

		// Extract project/module name from error messages
		// Pattern: "on project sample-invalid-dep" or "on project module-a"
		if result.FailedModule == "" {
			projRE := regexp.MustCompile(`on project ([^:]+)`)
			if m := projRE.FindStringSubmatch(errLine); m != nil {
				result.FailedModule = m[1]
			}
		}

		// For POM parsing errors, extract project name from pom.xml path
		// e.g., "/path/to/sample-invalid-pom/pom.xml" -> "sample-invalid-pom"
		if result.FailedModule == "" && strings.Contains(errLine, "Non-parseable POM") {
			pomPathRE := regexp.MustCompile(`/([^/]+)/pom\.xml`)
			if m := pomPathRE.FindStringSubmatch(errLine); m != nil {
				result.FailedModule = m[1]
			}
		}

		// For cyclic dependency errors, extract module names from the cycle
		// e.g., "com.example:module-a:1.0-SNAPSHOT --> com.example:module-b:1.0-SNAPSHOT"
		if result.FailedModule == "" && (strings.Contains(errLine, "cyclic") || strings.Contains(errLine, "Cycle")) {
			cycleRE := regexp.MustCompile(`com\.example:([^:]+):`)
			matches := cycleRE.FindAllStringSubmatch(errLine, -1)
			if len(matches) >= 2 {
				result.FailedModule = matches[0][1] + " <-> " + matches[1][1]
			}
		}

		// For reactor/missing module errors, extract project name from path
		// e.g., "Child module /path/to/sample-reactor-issue/module-nonexistent ... does not exist"
		if result.FailedModule == "" && strings.Contains(errLine, "does not exist") {
			missingModRE := regexp.MustCompile(`/([^/]+)/[^/]+ does not exist`)
			if m := missingModRE.FindStringSubmatch(errLine); m != nil {
				result.FailedModule = m[1]
			}
		}

		// Detect failure type from error lines
		if fatalRE.MatchString(line) || strings.Contains(errLine, "Non-parseable POM") {
			detectedFailure = "INVALID_POM"
			result.ErrorMessage = strings.TrimSpace(errLine)
		} else if strings.Contains(errLine, "Could not find artifact") || strings.Contains(errLine, "Could not resolve dependencies") {
			detectedFailure = "INVALID_DEPENDENCY"
			result.ErrorMessage = strings.TrimSpace(errLine)
		} else if strings.Contains(errLine, "Circular dependency") || strings.Contains(errLine, "Cycle") {
			detectedFailure = "CIRCULAR_DEPENDENCY"
			result.ErrorMessage = strings.TrimSpace(errLine)
		} else if strings.Contains(errLine, "does not exist") {
			detectedFailure = "REACTOR_ERROR"
			result.ErrorMessage = strings.TrimSpace(errLine)
		} else if strings.Contains(errLine, "COMPILATION ERROR") {
			detectedFailure = "COMPILE_ERROR"
			result.ErrorMessage = strings.TrimSpace(errLine)
		} else if strings.Contains(errLine, "Failures:") && strings.Contains(errLine, "Tests run:") {
			detectedFailure = "TEST_FAILURE"
			result.ErrorMessage = extractTestSummary(errLine)
		} else if strings.Contains(errLine, "test-compile") || strings.Contains(errLine, "Compilation failure") {
			if strings.Contains(errLine, "test") {
				detectedFailure = "TEST_COMPILE_ERROR"
				result.ErrorMessage = strings.TrimSpace(errLine)
			}
		}

		// Extract compile error message from lines with file:[line,col] pattern
		// Pattern: "/path/to/file.ext:[line,col] <error message>"
		// More generic - handles .java, .kt, .scala, .groovy, etc.
		compileErrRE := regexp.MustCompile(`(/[^/]+/src/[^:]+\.[a-zA-Z0-9]+):\[(\d+),(\d+)\]\s+(.+)$`)
		if m := compileErrRE.FindStringSubmatch(errLine); m != nil {
			result.ErrorMessage = m[4]
		} else if strings.Contains(errLine, "Failed to execute goal") && strings.Contains(errLine, "package") {
			detectedFailure = "PACKAGE_ERROR"
			result.ErrorMessage = errLine
		}

		// Extract test failure location and assertion from error lines BEFORE setting error message
		// Pattern: "  ClassName.methodName:lineNumber assertion message"
		// Example: "  CalculatorTest.testFail:9 This test always fails expected:<0> but was:<1>"
		// Only set method name and line number here (class name comes from FAILURE pattern)
		if detectedFailure == "TEST_FAILURE" {
			testLocRE := regexp.MustCompile(`^\s*([A-Z][a-zA-Z0-9_]*)\.([a-zA-Z0-9_]+):(\d+)\s+(.+)$`)
			if m := testLocRE.FindStringSubmatch(errLine); m != nil {
				// Only set className if not already captured (to preserve full package name from FAILURE pattern)
				if testClassName == "" {
					testClassName = m[1]
				}
				testMethodName = m[2]
				testLineNum = m[3]
				testAssertionMsg = m[4]
			}
		}

		// Extract test class from "FAILURE! -- in com.example.ClassName" pattern
		if detectedFailure == "TEST_FAILURE" && testClassName == "" {
			failClassRE := regexp.MustCompile(`FAILURE!\s*--\s*in\s+(.+)$`)
			if m := failClassRE.FindStringSubmatch(errLine); m != nil {
				testClassName = m[1]
			}
		}

		// Check for test failures in output
		if strings.Contains(line, "Tests run:") && strings.Contains(line, "Failures:") {
			result.TestSummary = extractTestSummary(line)
		}

		// Extract file location from any error line
		if loc := compilationRE.FindStringSubmatch(errLine); loc != nil {
			result.FailureLocation = loc[1] + ":" + loc[2]
		}

		// Extract file:line:column from any error line containing .ext:[line,col] pattern
		// More generic - handles .java, .kt, .scala, .groovy, etc.
		if result.FailureLocation == "" {
			// Match full absolute path with [line,col] suffix
			// Pattern: /.../src/.../file.ext:[line,col]
			extFileRE := regexp.MustCompile(`(/[^/]+/src/[^:]+\.[a-zA-Z0-9]+):\[(\d+),(\d+)\]`)
			if m := extFileRE.FindStringSubmatch(errLine); m != nil {
				fullPath := m[1]
				// Make path relative to project root
				relPath := strings.TrimPrefix(fullPath, projRoot+"/")
				// Trim leading slash if present
				relPath = strings.TrimPrefix(relPath, "/")
				result.FailureLocation = relPath + ":" + m[2] + ":" + m[3]
			}
		}

		// Check for test failures in output
		if strings.Contains(line, "Tests run:") && strings.Contains(line, "Failures:") {
			result.TestSummary = extractTestSummary(line)
		}
	}

	// Parse surefire reports if test output exists
	if result.TestSummary != "" {
		report := ParseAllSurefireReports(projectRoot)
		result.TestSuites = report.Suites
	}

	// Set resume module - only for build failures that can be resumed (not pre-build failures)
	// Pre-build failures (INVALID_POM, INVALID_DEPENDENCY, CIRCULAR_DEPENDENCY, REACTOR_ERROR) can't be resumed
	if result.FailedModule != "" && !isPreBuildFailure(result.Status) {
		result.ResumeModule = result.FailedModule
	}

	// Set final Status based on what we detected
	// If we detected a specific failure type, use that; otherwise use rawStatus
	if detectedFailure != "" {
		result.Status = detectedFailure
	} else if rawStatus != "" {
		result.Status = rawStatus
	}

	// For test failures, prepend the assertion message if we captured one
	if testAssertionMsg != "" && result.ErrorMessage != "" {
		result.ErrorMessage = testAssertionMsg + " | " + result.ErrorMessage
	}

	// For test failures, construct failure location from class+method+line
	if detectedFailure == "TEST_FAILURE" && testClassName != "" {
		if testMethodName != "" && testLineNum != "" {
			result.FailureLocation = testClassName + "." + testMethodName + ":" + testLineNum
		} else if testMethodName != "" {
			result.FailureLocation = testClassName + "." + testMethodName
		} else {
			result.FailureLocation = testClassName
		}
	}

	return result
}

// isPreBuildFailure returns true if the failure type is a pre-build failure that can't be resumed
func isPreBuildFailure(failureType string) bool {
	switch failureType {
	case "INVALID_POM", "INVALID_DEPENDENCY", "CIRCULAR_DEPENDENCY", "REACTOR_ERROR":
		return true
	}
	return false
}

func extractTestSummary(line string) string {
	re := regexp.MustCompile(`Tests run: (\d+), Failures: (\d+), Errors: (\d+), Skipped: (\d+)`)
	matches := re.FindStringSubmatch(line)
	if matches != nil {
		return fmt.Sprintf("Tests run: %s, Failures: %s, Errors: %s", matches[1], matches[2], matches[3])
	}
	re2 := regexp.MustCompile(`Tests run: (\d+), Failures: (\d+), Errors: (\d+), Skipped: (\d+),`)
	matches2 := re2.FindStringSubmatch(line)
	if matches2 != nil {
		return fmt.Sprintf("Tests run: %s, Failures: %s, Errors: %s", matches2[1], matches2[2], matches2[3])
	}
	return ""
}

// ToJSON converts MavenOutput to JSON string
func (m *MavenOutput) ToJSON() string {
	agentSummary := m.GetAgentSummary()

	// Build JSON with only non-empty fields
	parts := []string{}

	// Status: map internal status to output format
	var statusOutput string
	if m.Status == "BUILD SUCCESS" || m.Status == "SUCCESS" {
		statusOutput = "SUCCESS"
	} else if m.Status != "" && m.Status != "BUILD FAILURE" {
		// Status contains a failure type (COMPILE_ERROR, INVALID_POM, etc.)
		statusOutput = m.Status
	} else if m.Status == "BUILD FAILURE" || (m.Status == "" && len(m.Errors) > 0) {
		// Generic failure
		statusOutput = "BUILD FAILURE"
	} else {
		statusOutput = "BUILD FAILURE"
	}
	parts = append(parts, fmt.Sprintf(`"status": "%s"`, statusOutput))

	// Only include failedModule for non-pre-build failures
	if m.FailedModule != "" && !isPreBuildFailure(m.Status) {
		parts = append(parts, fmt.Sprintf(`"failedModule": "%s"`, m.FailedModule))
	}

	if m.FailureLocation != "" {
		parts = append(parts, fmt.Sprintf(`"failureLocation": "%s"`, m.FailureLocation))
	}

	if m.ErrorMessage != "" {
		parts = append(parts, fmt.Sprintf(`"errorMessage": "%s"`, m.ErrorMessage))
	}

	if len(m.Errors) > 0 {
		errStr := ""
		for i, err := range m.Errors {
			if i > 0 {
				errStr += ","
			}
			errStr += fmt.Sprintf(`"%s"`, err)
		}
		parts = append(parts, fmt.Sprintf(`"errors": [%s]`, errStr))
	}

	if m.TestSummary != "" {
		parts = append(parts, fmt.Sprintf(`"testSummary": "%s"`, m.TestSummary))
	}

	if agentSummary != "" {
		parts = append(parts, fmt.Sprintf(`"agentSummary": "%s"`, agentSummary))
	}

	// Only include resumeCommand for non-pre-build failures and when there's a resume module
	if m.ResumeModule != "" && !isPreBuildFailure(m.Status) {
		parts = append(parts, fmt.Sprintf(`"resumeCommand": "mvn-llm install -rf %s"`, m.ResumeModule))
	}

	return "{\n  " + strings.Join(parts, ",\n  ") + "\n}"
}

// GetAgentSummary returns a clean one-line summary suitable for agents
func (m *MavenOutput) GetAgentSummary() string {
	// Success case
	if m.Status == "BUILD SUCCESS" || m.Status == "SUCCESS" {
		if m.TestSummary != "" {
			return fmt.Sprintf("SUCCESS - %s", m.TestSummary)
		}
		return "SUCCESS"
	}

	// Build failure - create concise summary
	var summary string

	// Use failure type if available (but not raw "BUILD FAILURE")
	if m.Status != "" && m.Status != "BUILD FAILURE" && m.Status != "BUILD SUCCESS" {
		summary = m.Status
	} else if len(m.Errors) > 0 {
		// Extract key info from first error - truncate if too long
		firstErr := m.Errors[0]
		// Get first line only, truncate if too long
		firstLine := strings.Split(firstErr, "\n")[0]
		if len(firstLine) > 80 {
			firstLine = firstLine[:80] + "..."
		}
		summary = firstLine
	} else {
		summary = "BUILD_FAILURE"
	}

	// Add module if known - but not for pre-build failures
	if m.FailedModule != "" && !isPreBuildFailure(m.Status) {
		summary += " (module: " + m.FailedModule + ")"
	}

	// Add location if available
	if m.FailureLocation != "" {
		summary += " at " + m.FailureLocation
	}

	// Add error message for test failures, compile errors, and pre-build failures
	if m.ErrorMessage != "" && (m.Status == "TEST_FAILURE" || m.Status == "TEST_COMPILE_ERROR" || m.Status == "COMPILE_ERROR" || m.Status == "INVALID_POM" || m.Status == "INVALID_DEPENDENCY" || m.Status == "REACTOR_ERROR") {
		errMsg := strings.TrimSpace(m.ErrorMessage)
		summary += " | " + errMsg
	}

	// Add test failure details from surefire
	if len(m.TestSuites) > 0 {
		for _, suite := range m.TestSuites {
			if suite.Failures > 0 {
				for _, tc := range suite.TestCases {
					if tc.Failure != "" {
						summary += fmt.Sprintf(" | test %s failed: %s", tc.Name, tc.Failure)
						break
					}
				}
				break
			}
		}
	}

	return summary
}

// MinimalParseResult holds summary and errors for MVP impl
type MinimalParseResult struct {
	Status string
	Errors []string
}

// MinimalParse parses output and extracts build status + [ERROR] lines
func MinimalParse(output string) MinimalParseResult {
	lines := strings.Split(output, "\n")
	errLines := []string{}
	buildStatus := "UNKNOWN"
	var statusRE = regexp.MustCompile(`\[INFO\] (BUILD SUCCESS|BUILD FAILURE)`)
	for _, line := range lines {
		if strings.HasPrefix(line, "[ERROR]") {
			errLines = append(errLines, strings.TrimPrefix(line, "[ERROR] "))
		}
		if m := statusRE.FindStringSubmatch(line); m != nil {
			buildStatus = m[1]
		}
	}
	return MinimalParseResult{Status: buildStatus, Errors: errLines}
}

// ParseSurefireXML reads a surefire TEST-*.xml file and extracts test results
func ParseSurefireXML(path string) (SurefireTestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SurefireTestSuite{}, err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	ts := SurefireTestSuite{}

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		if elem, ok := token.(xml.StartElement); ok {
			if elem.Name.Local == "testsuite" {
				// Extract attributes from testsuite
				for _, attr := range elem.Attr {
					switch attr.Name.Local {
					case "name":
						ts.Name = attr.Value
					case "tests":
						ts.Tests, _ = strconv.Atoi(attr.Value)
					case "failures":
						ts.Failures, _ = strconv.Atoi(attr.Value)
					case "errors":
						ts.Errors, _ = strconv.Atoi(attr.Value)
					case "skipped":
						ts.Skipped, _ = strconv.Atoi(attr.Value)
					case "time":
						ts.Time, _ = strconv.ParseFloat(attr.Value, 64)
					}
				}
			} else if elem.Name.Local == "testcase" {
				tc := SurefireTestCase{}
				for _, attr := range elem.Attr {
					switch attr.Name.Local {
					case "name":
						tc.Name = attr.Value
					case "classname":
						tc.Class = attr.Value
					case "time":
						tc.Time, _ = strconv.ParseFloat(attr.Value, 64)
					case "message":
						tc.Failure = attr.Value
					}
				}
				ts.TestCases = append(ts.TestCases, tc)
			}
		}
	}

	return ts, nil
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// FindSurefireReports searches for surefire XML reports in the given directory and subdirectories
func FindSurefireReports(projectRoot string) []string {
	var reports []string
	matches, err := filepath.Glob(filepath.Join(projectRoot, "**", "target", "surefire-reports", "TEST-*.xml"))
	if err == nil && len(matches) > 0 {
		reports = append(reports, matches...)
	}
	return reports
}

// ParseAllSurefireReports parses all surefire XML reports in a project
func ParseAllSurefireReports(projectRoot string) SurefireReport {
	report := SurefireReport{}
	reports := FindSurefireReports(projectRoot)
	for _, r := range reports {
		suite, err := ParseSurefireXML(r)
		if err == nil {
			report.Suites = append(report.Suites, suite)
		}
	}
	return report
}

// GetTestSummary returns a summary string from surefire reports
func (r SurefireReport) GetTestSummary() string {
	totalTests := 0
	totalFailures := 0
	totalErrors := 0
	totalSkipped := 0

	for _, suite := range r.Suites {
		totalTests += suite.Tests
		totalFailures += suite.Failures
		totalErrors += suite.Errors
		totalSkipped += suite.Skipped
	}

	if totalTests == 0 {
		return "No tests run"
	}

	parts := []string{}
	parts = append(parts, "Tests run:")
	parts = append(parts, formatInt(totalTests))
	if totalFailures > 0 {
		parts = append(parts, "Failures:")
		parts = append(parts, formatInt(totalFailures))
	}
	if totalErrors > 0 {
		parts = append(parts, "Errors:")
		parts = append(parts, formatInt(totalErrors))
	}
	if totalSkipped > 0 {
		parts = append(parts, "Skipped:")
		parts = append(parts, formatInt(totalSkipped))
	}
	return strings.Join(parts, " ")
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
