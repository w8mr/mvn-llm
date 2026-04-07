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

// Pre-compiled regexes for better performance
var (
	statusRE       = regexp.MustCompile(`\[INFO\] (BUILD SUCCESS|BUILD FAILURE)`)
	moduleRE       = regexp.MustCompile(`\[INFO\] ([^\s]+)\s+\.+\s+(SUCCESS|FAILURE|SKIPPED)`)
	testSummaryRE  = regexp.MustCompile(`Tests run: (\d+), Failures: (\d+), Errors: (\d+), Skipped: (\d+)`)
	projectRE      = regexp.MustCompile(`on project ([^:]+)`)
	pomPathRE      = regexp.MustCompile(`/([^/]+)/pom\.xml`)
	cycleRE        = regexp.MustCompile(`com\.example:([^:]+):`)
	missingModRE   = regexp.MustCompile(`/([^/]+)/[^/]+ does not exist`)
	compileErrRE   = regexp.MustCompile(`(/[^/]+/src/[^:]+\.[a-zA-Z0-9]+):\[(\d+),(\d+)\]\s+(.+)$`)
	fileLocationRE = regexp.MustCompile(`(/[^/]+/src/[^:]+\.[a-zA-Z0-9]+):\[(\d+),(\d+)\]`)
	testLocRE      = regexp.MustCompile(`^\s*([A-Z][a-zA-Z0-9_]*)\.([a-zA-Z0-9_]+):(\d+)\s+(.+)$`)
	failClassRE    = regexp.MustCompile(`FAILURE!\s*--\s*in\s+(.+)$`)
)

// FailureType represents the type of build failure
type FailureType string

const (
	FailureNone        FailureType = ""
	FailureInvalidPOM  FailureType = "INVALID_POM"
	FailureInvalidDep  FailureType = "INVALID_DEPENDENCY"
	FailureCircularDep FailureType = "CIRCULAR_DEPENDENCY"
	FailureReactor     FailureType = "REACTOR_ERROR"
	FailureCompile     FailureType = "COMPILE_ERROR"
	FailureTestCompile FailureType = "TEST_COMPILE_ERROR"
	FailureTest        FailureType = "TEST_FAILURE"
	FailurePackage     FailureType = "PACKAGE_ERROR"
	FailureBuild       FailureType = "BUILD FAILURE"
	StatusSuccess      FailureType = "BUILD SUCCESS"
)

// parseContext holds state during parsing
type parseContext struct {
	rawStatus      string
	failureType    FailureType
	testClassName  string
	testMethodName string
	testLineNum    string
	testAssertMsg  string
	projectRoot    string
}

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

// --- NEW AGENT ERROR SUMMARY HELPER ---
func (m *MavenOutput) getAgentErrorSummaryLines(maxLines int) []string {
	type errKey struct {
		file string
		line string
		col  string
	}
	fileErrors := make(map[errKey]string)   // for compile errors
	otherErrors := make(map[string]string)  // catch-all for non-file-specific errors
	testFailures := make(map[string]string) // for test method failures
	for _, err := range m.Errors {
		if match := compileErrRE.FindStringSubmatch(err); match != nil {
			key := errKey{file: match[1], line: match[2], col: match[3]}
			msg := match[4]
			relPath := match[1]
			if strings.HasPrefix(relPath, m.FailureLocation) {
				// already relative
			} else if i := strings.Index(relPath, "/src/"); i >= 0 {
				relPath = relPath[i+1:]
			}
			fileErrors[key] = relPath + ":" + match[2] + ":" + match[3] + ": " + msg
			continue
		}
		if match := testLocRE.FindStringSubmatch(err); match != nil {
			key := match[1] + "." + match[2] + ":" + match[3]
			testFailures[key] = key + ": " + match[4]
			continue
		}
		msg := err
		if len(msg) > 120 {
			msg = msg[:120] + "..."
		}
		otherErrors[msg] = msg
	}
	lines := []string{}
	for _, v := range fileErrors {
		lines = append(lines, v)
	}
	for _, v := range testFailures {
		lines = append(lines, v)
	}
	if len(lines) == 0 {
		for _, v := range otherErrors {
			lines = append(lines, v)
		}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines,
			fmt.Sprintf("...and %d more errors (total %d)", len(fileErrors)+len(testFailures)+len(otherErrors)-maxLines, len(fileErrors)+len(testFailures)+len(otherErrors)),
		)
	}
	return lines
}

// === BEGIN RESTORED CANONICAL LOGIC ===
// ([31mFull canonical functions follow[0m)

// ParseMavenOutput parses Maven output and extracts structured information
func ParseMavenOutput(output string, projectRoot string) MavenOutput {
	result := MavenOutput{
		Errors: []string{},
	}
	ctx := &parseContext{
		projectRoot: projectRoot,
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		ctx.detectBuildStatus(line)
		result.extractTestSummary(line)
		result.extractFailedModule(line)

		errLine, isError := extractErrorLine(line)
		if !isError {
			continue
		}

		result.Errors = append(result.Errors, errLine)
		result.extractModuleFromError(errLine, ctx)
		ctx.detectFailureType(errLine, line, &result)
		ctx.extractTestFailureInfo(errLine)
		result.extractFailureLocation(errLine, ctx)
	}

	result.finalize(ctx, projectRoot)
	return result
}

// detectBuildStatus detects the build status from a line
func (ctx *parseContext) detectBuildStatus(line string) {
	if m := statusRE.FindStringSubmatch(line); m != nil {
		if ctx.rawStatus == "" || ctx.rawStatus == "BUILD FAILURE" {
			ctx.rawStatus = m[1]
		}
	}

	if strings.Contains(line, "BUILD FAILURE") && ctx.rawStatus == "" {
		ctx.rawStatus = "BUILD FAILURE"
	}
}

// extractErrorLine extracts the error content from [ERROR] or [FATAL] lines
func extractErrorLine(line string) (string, bool) {
	if strings.HasPrefix(line, "[FATAL]") {
		return strings.TrimPrefix(line, "[FATAL] "), true
	}
	if strings.HasPrefix(line, "[ERROR]") {
		return strings.TrimPrefix(line, "[ERROR] "), true
	}
	return "", false
}

// extractTestSummary extracts test summary from a line
func (m *MavenOutput) extractTestSummary(line string) {
	if strings.Contains(line, "Tests run:") && strings.Contains(line, "Failures:") {
		if summary := formatTestSummary(line); summary != "" {
			m.TestSummary = summary
		}
	}
}

// extractFailedModule extracts the failed module from reactor summary
func (m *MavenOutput) extractFailedModule(line string) {
	if match := moduleRE.FindStringSubmatch(line); match != nil {
		if match[2] == "FAILURE" {
			m.FailedModule = match[1]
		}
	}
}

// extractModuleFromError extracts module name from various error patterns
func (m *MavenOutput) extractModuleFromError(errLine string, ctx *parseContext) {
	if m.FailedModule != "" {
		return
	}

	// Pattern: "on project sample-invalid-dep" or "on project module-a"
	if match := projectRE.FindStringSubmatch(errLine); match != nil {
		m.FailedModule = match[1]
		return
	}

	// POM parsing errors: extract from pom.xml path
	if strings.Contains(errLine, "Non-parseable POM") {
		if match := pomPathRE.FindStringSubmatch(errLine); match != nil {
			m.FailedModule = match[1]
			return
		}
	}

	// Cyclic dependency: extract module names from cycle
	if strings.Contains(errLine, "cyclic") || strings.Contains(errLine, "Cycle") {
		if matches := cycleRE.FindAllStringSubmatch(errLine, -1); len(matches) >= 2 {
			m.FailedModule = matches[0][1] + " <-> " + matches[1][1]
			return
		}
	}

	// Reactor/missing module: extract from path
	if strings.Contains(errLine, "does not exist") {
		if match := missingModRE.FindStringSubmatch(errLine); match != nil {
			m.FailedModule = match[1]
		}
	}
}

// detectFailureType identifies the type of failure from error lines
func (ctx *parseContext) detectFailureType(errLine, originalLine string, result *MavenOutput) {
	switch {
	case strings.Contains(originalLine, "[FATAL]") || strings.Contains(errLine, "Non-parseable POM"):
		ctx.failureType = FailureInvalidPOM
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "Could not find artifact") || strings.Contains(errLine, "Could not resolve dependencies"):
		ctx.failureType = FailureInvalidDep
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "Circular dependency") || strings.Contains(errLine, "Cycle"):
		ctx.failureType = FailureCircularDep
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "does not exist"):
		ctx.failureType = FailureReactor
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "COMPILATION ERROR"):
		ctx.failureType = FailureCompile
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "Failures:") && strings.Contains(errLine, "Tests run:"):
		ctx.failureType = FailureTest
		result.ErrorMessage = formatTestSummary(errLine)

	case (strings.Contains(errLine, "test-compile") || strings.Contains(errLine, "Compilation failure")) && strings.Contains(errLine, "test"):
		ctx.failureType = FailureTestCompile
		result.ErrorMessage = strings.TrimSpace(errLine)

	case strings.Contains(errLine, "Failed to execute goal") && strings.Contains(errLine, "package"):
		ctx.failureType = FailurePackage
		result.ErrorMessage = errLine
	}

	// Extract compile error message from file:[line,col] pattern
	if match := compileErrRE.FindStringSubmatch(errLine); match != nil {
		result.ErrorMessage = match[4]
	}
}

// extractTestFailureInfo extracts test class, method, and assertion info
func (ctx *parseContext) extractTestFailureInfo(errLine string) {
	if ctx.failureType != FailureTest {
		return
	}

	// Extract test location: ClassName.methodName:lineNumber assertion
	if match := testLocRE.FindStringSubmatch(errLine); match != nil {
		if ctx.testClassName == "" {
			ctx.testClassName = match[1]
		}
		ctx.testMethodName = match[2]
		ctx.testLineNum = match[3]
		ctx.testAssertMsg = match[4]
	}

	// Extract test class from "FAILURE! -- in com.example.ClassName"
	if ctx.testClassName == "" {
		if match := failClassRE.FindStringSubmatch(errLine); match != nil {
			ctx.testClassName = match[1]
		}
	}
}

// extractFailureLocation extracts file:line:column from error
func (m *MavenOutput) extractFailureLocation(errLine string, ctx *parseContext) {
	if m.FailureLocation != "" {
		return
	}

	if match := fileLocationRE.FindStringSubmatch(errLine); match != nil {
		fullPath := match[1]
		relPath := strings.TrimPrefix(fullPath, ctx.projectRoot+"/")
		relPath = strings.TrimPrefix(relPath, "/")
		m.FailureLocation = relPath + ":" + match[2] + ":" + match[3]
	}
}

// finalize completes the parsing by setting final status and processing test results
func (m *MavenOutput) finalize(ctx *parseContext, projectRoot string) {
	// Set status based on detected failure type or raw status
	if ctx.failureType != FailureNone {
		m.Status = string(ctx.failureType)
	} else if ctx.rawStatus != "" {
		m.Status = ctx.rawStatus
	}

	// Handle test failure specifics
	if ctx.failureType == FailureTest {
		m.finalizeTestFailure(ctx)
	}

	// Parse surefire reports if tests were run
	if m.TestSummary != "" {
		report := ParseAllSurefireReports(projectRoot)
		m.TestSuites = report.Suites
	}

	// Set resume module for resumable failures
	if m.FailedModule != "" && !isPreBuildFailure(m.Status) {
		m.ResumeModule = m.FailedModule
	}
}

// finalizeTestFailure handles test-specific finalization
func (m *MavenOutput) finalizeTestFailure(ctx *parseContext) {
	// Prepend assertion message if available
	if ctx.testAssertMsg != "" && m.ErrorMessage != "" {
		m.ErrorMessage = ctx.testAssertMsg + " | " + m.ErrorMessage
	}

	// Construct failure location from test info
	if ctx.testClassName != "" {
		switch {
		case ctx.testMethodName != "" && ctx.testLineNum != "":
			m.FailureLocation = ctx.testClassName + "." + ctx.testMethodName + ":" + ctx.testLineNum
		case ctx.testMethodName != "":
			m.FailureLocation = ctx.testClassName + "." + ctx.testMethodName
		default:
			m.FailureLocation = ctx.testClassName
		}
	}
}

// isPreBuildFailure returns true if the failure type is a pre-build failure that can't be resumed
func isPreBuildFailure(failureType string) bool {
	switch FailureType(failureType) {
	case FailureInvalidPOM, FailureInvalidDep, FailureCircularDep, FailureReactor:
		return true
	}
	return false
}

// formatTestSummary extracts and formats test summary from a line
func formatTestSummary(line string) string {
	matches := testSummaryRE.FindStringSubmatch(line)
	if matches != nil {
		return fmt.Sprintf("Tests run: %s, Failures: %s, Errors: %s", matches[1], matches[2], matches[3])
	}
	return ""
}

// extractTestSummary is kept for backward compatibility
func extractTestSummary(line string) string {
	return formatTestSummary(line)
}

// ToJSON converts MavenOutput to JSON string
func (m *MavenOutput) ToJSON() string {
	agentSummary := m.GetAgentSummary()
	parts := []string{}

	// Status: map internal status to output format
	statusOutput := m.getStatusOutput()
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
		errStrs := make([]string, len(m.Errors))
		for i, err := range m.Errors {
			errStrs[i] = fmt.Sprintf(`"%s"`, err)
		}
		parts = append(parts, fmt.Sprintf(`"errors": [%s]`, strings.Join(errStrs, ",")))
	}

	if m.TestSummary != "" {
		parts = append(parts, fmt.Sprintf(`"testSummary": "%s"`, m.TestSummary))
	}

	if agentSummary != "" {
		parts = append(parts, fmt.Sprintf(`"agentSummary": "%s"`, agentSummary))
	}

	// Agent error summary lines for LLMs and downstream agents
	agentErrorLines := m.getAgentErrorSummaryLines(6)
	if len(agentErrorLines) > 0 {
		lineReprs := make([]string, len(agentErrorLines))
		for i, l := range agentErrorLines {
			lineReprs[i] = fmt.Sprintf(`"%s"`, l)
		}
		parts = append(parts, fmt.Sprintf(`"agentErrorSummaryLines": [%s]`, strings.Join(lineReprs, ", ")))
	}

	// Only include resumeCommand for non-pre-build failures
	if m.ResumeModule != "" && !isPreBuildFailure(m.Status) {
		parts = append(parts, fmt.Sprintf(`"resumeCommand": "mvn-llm install -rf %s"`, m.ResumeModule))
	}

	return "{\n  " + strings.Join(parts, ",\n  ") + "\n}"
}

// getStatusOutput returns the status string for output
func (m *MavenOutput) getStatusOutput() string {
	switch {
	case m.Status == "BUILD SUCCESS" || m.Status == "SUCCESS":
		return "SUCCESS"
	case m.Status != "" && m.Status != "BUILD FAILURE":
		return m.Status
	default:
		return "BUILD FAILURE"
	}
}

// GetAgentSummary returns a clean one-line summary suitable for agents
func (m *MavenOutput) GetAgentSummary() string {
	if m.Status == "BUILD SUCCESS" || m.Status == "SUCCESS" {
		return m.getSuccessSummary()
	}
	return m.getFailureSummary()
}

func (m *MavenOutput) getSuccessSummary() string {
	if m.TestSummary != "" {
		return fmt.Sprintf("SUCCESS - %s", m.TestSummary)
	}
	return "SUCCESS"
}

func (m *MavenOutput) getFailureSummary() string {
	summary := m.getBaseSummary()

	// Add module if known (but not for pre-build failures)
	if m.FailedModule != "" && !isPreBuildFailure(m.Status) {
		summary += " (module: " + m.FailedModule + ")"
	}

	// Add location if available
	if m.FailureLocation != "" {
		summary += " at " + m.FailureLocation
	}

	// Add error message for specific failure types
	if m.shouldIncludeErrorMessage() {
		summary += " | " + strings.TrimSpace(m.ErrorMessage)
	}

	// Add test failure details from surefire
	summary += m.getSurefireFailureDetails()

	return summary
}

func (m *MavenOutput) getBaseSummary() string {
	if m.Status != "" && m.Status != "BUILD FAILURE" && m.Status != "BUILD SUCCESS" {
		return m.Status
	}
	if len(m.Errors) > 0 {
		firstLine := strings.Split(m.Errors[0], "\n")[0]
		if len(firstLine) > 80 {
			firstLine = firstLine[:80] + "..."
		}
		return firstLine
	}
	return "BUILD_FAILURE"
}

func (m *MavenOutput) shouldIncludeErrorMessage() bool {
	if m.ErrorMessage == "" {
		return false
	}
	switch FailureType(m.Status) {
	case FailureTest, FailureTestCompile, FailureCompile, FailureInvalidPOM, FailureInvalidDep, FailureReactor:
		return true
	}
	return false
}

func (m *MavenOutput) getSurefireFailureDetails() string {
	for _, suite := range m.TestSuites {
		if suite.Failures > 0 {
			for _, tc := range suite.TestCases {
				if tc.Failure != "" {
					return fmt.Sprintf(" | test %s failed: %s", tc.Name, tc.Failure)
				}
			}
		}
	}
	return ""
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
			switch elem.Name.Local {
			case "testsuite":
				parseTestSuiteAttrs(&ts, elem.Attr)
			case "testcase":
				tc := parseTestCaseAttrs(elem.Attr)
				ts.TestCases = append(ts.TestCases, tc)
			}
		}
	}

	return ts, nil
}

func parseTestSuiteAttrs(ts *SurefireTestSuite, attrs []xml.Attr) {
	for _, attr := range attrs {
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
}

func parseTestCaseAttrs(attrs []xml.Attr) SurefireTestCase {
	tc := SurefireTestCase{}
	for _, attr := range attrs {
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
	return tc
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

	parts := []string{fmt.Sprintf("Tests run: %d", totalTests)}
	if totalFailures > 0 {
		parts = append(parts, fmt.Sprintf("Failures: %d", totalFailures))
	}
	if totalErrors > 0 {
		parts = append(parts, fmt.Sprintf("Errors: %d", totalErrors))
	}
	if totalSkipped > 0 {
		parts = append(parts, fmt.Sprintf("Skipped: %d", totalSkipped))
	}
	return strings.Join(parts, " ")
}
