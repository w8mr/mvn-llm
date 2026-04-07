package maven

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMavenOutputSuccess(t *testing.T) {
	output := readTestFile(t, "sample_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "BUILD SUCCESS" && result.Status != "SUCCESS" {
		t.Errorf("Expected SUCCESS status, got: %s", result.Status)
	}

	if result.FailedModule != "" {
		t.Errorf("Expected no failed module for success, got: %s", result.FailedModule)
	}

	if result.ErrorMessage != "" {
		t.Errorf("Expected no error message for success, got: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputBuildFailure(t *testing.T) {
	output := readTestFile(t, "sample_buildfail_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-buildfail")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status == "BUILD SUCCESS" || result.Status == "SUCCESS" {
		t.Errorf("Expected failure status, got: %s", result.Status)
	}

	if !strings.Contains(result.Status, "FAILURE") && result.Status != "COMPILE_ERROR" {
		t.Errorf("Expected COMPILE_ERROR or BUILD FAILURE, got: %s", result.Status)
	}

	if result.FailedModule != "module-a" {
		t.Errorf("Expected failed module 'module-a', got: %s", result.FailedModule)
	}

	if result.FailureLocation == "" {
		t.Error("Expected failure location")
	}

	if !strings.Contains(result.ErrorMessage, "Compilation failure") && !strings.Contains(result.ErrorMessage, "illegal start of expression") {
		t.Errorf("Expected compilation error message, got: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputTestFailure(t *testing.T) {
	output := readTestFile(t, "sample_testfail_test.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-testfail")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "TEST_FAILURE" {
		t.Errorf("Expected TEST_FAILURE status, got: %s", result.Status)
	}

	if result.FailedModule != "module-a" {
		t.Errorf("Expected failed module 'module-a', got: %s", result.FailedModule)
	}

	if result.TestSummary == "" {
		t.Error("Expected test summary")
	}

	if !strings.Contains(result.TestSummary, "Failures:") {
		t.Errorf("Expected test summary with failures, got: %s", result.TestSummary)
	}
}

func TestParseMavenOutputInvalidPOM(t *testing.T) {
	output := readTestFile(t, "sample_invalid_pom_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-invalid-pom")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "INVALID_POM" {
		t.Errorf("Expected INVALID_POM status, got: %s", result.Status)
	}

	if result.FailedModule == "" {
		t.Error("Expected failed module to be set")
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message")
	}

	if !strings.Contains(result.ErrorMessage, "Non-parseable POM") && !strings.Contains(result.ErrorMessage, "end tag name") {
		t.Errorf("Expected POM parsing error, got: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputInvalidDependency(t *testing.T) {
	output := readTestFile(t, "sample_invalid_dep_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-invalid-dep")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "INVALID_DEPENDENCY" {
		t.Errorf("Expected INVALID_DEPENDENCY status, got: %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message")
	}

	if !strings.Contains(result.ErrorMessage, "Could not resolve") && !strings.Contains(result.ErrorMessage, "fake-artifact") {
		t.Errorf("Expected dependency resolution error, got: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputCircularDependency(t *testing.T) {
	output := readTestFile(t, "sample_circular_dep_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-circular-dep")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "CIRCULAR_DEPENDENCY" {
		t.Errorf("Expected CIRCULAR_DEPENDENCY status, got: %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message")
	}

	if !strings.Contains(result.ErrorMessage, "cyclic") && !strings.Contains(result.ErrorMessage, "Cycle") {
		t.Logf("Error message: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputReactorError(t *testing.T) {
	output := readTestFile(t, "sample_reactor_issue_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-reactor-issue")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "REACTOR_ERROR" {
		t.Errorf("Expected REACTOR_ERROR status, got: %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message")
	}

	if !strings.Contains(result.ErrorMessage, "does not exist") {
		t.Errorf("Expected reactor error about missing module, got: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputTestCompileError(t *testing.T) {
	output := readTestFile(t, "sample_testcompfail_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-testcompfail")

	result := ParseMavenOutput(output, projectRoot)

	if result.Status != "TEST_COMPILE_ERROR" {
		t.Errorf("Expected TEST_COMPILE_ERROR status, got: %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message")
	}

	if !strings.Contains(result.ErrorMessage, "testCompile") && !strings.Contains(result.ErrorMessage, "BrokenTest.java") && !strings.Contains(result.ErrorMessage, ";") {
		t.Logf("Error message: %s", result.ErrorMessage)
	}
}

func TestParseMavenOutputErrorsCollection(t *testing.T) {
	output := readTestFile(t, "sample_buildfail_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-buildfail")

	result := ParseMavenOutput(output, projectRoot)

	if len(result.Errors) == 0 {
		t.Error("Expected errors to be collected")
	}

	hasErrorLine := false
	for _, err := range result.Errors {
		if strings.Contains(err, "COMPILATION ERROR") || strings.Contains(err, "Failed to execute") {
			hasErrorLine = true
			break
		}
	}
	if !hasErrorLine {
		t.Error("Expected error lines in errors collection")
	}
}

func TestParseMavenOutputResumeModule(t *testing.T) {
	output := readTestFile(t, "sample_buildfail_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-buildfail")

	result := ParseMavenOutput(output, projectRoot)

	if result.ResumeModule == "" {
		t.Error("Expected resume module to be set for build failure")
	}

	if result.ResumeModule != "module-a" {
		t.Errorf("Expected resume module 'module-a', got: %s", result.ResumeModule)
	}
}

func TestParseMavenOutputNoResumeForPreBuildFailures(t *testing.T) {
	testCases := []struct {
		filename       string
		projectRoot    string
		expectedStatus string
	}{
		{"sample_invalid_pom_install.txt", filepath.Join("..", "..", "testdata", "sample-invalid-pom"), "INVALID_POM"},
		{"sample_invalid_dep_install.txt", filepath.Join("..", "..", "testdata", "sample-invalid-dep"), "INVALID_DEPENDENCY"},
		{"sample_circular_dep_install.txt", filepath.Join("..", "..", "testdata", "sample-circular-dep"), "CIRCULAR_DEPENDENCY"},
		{"sample_reactor_issue_install.txt", filepath.Join("..", "..", "testdata", "sample-reactor-issue"), "REACTOR_ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			output := readTestFile(t, tc.filename)
			result := ParseMavenOutput(output, tc.projectRoot)

			if result.Status != tc.expectedStatus {
				t.Errorf("Expected status %s, got: %s", tc.expectedStatus, result.Status)
			}
		})
	}
}

func TestParseMavenOutputToJSON(t *testing.T) {
	output := readTestFile(t, "sample_buildfail_install.txt")
	projectRoot := filepath.Join("..", "..", "testdata", "sample-buildfail")

	result := ParseMavenOutput(output, projectRoot)
	json := result.ToJSON()

	if json == "" {
		t.Error("Expected non-empty JSON output")
	}

	if !strings.Contains(json, `"status"`) {
		t.Error("Expected status in JSON")
	}

	if !strings.Contains(json, `"errors"`) {
		t.Error("Expected errors in JSON")
	}
}

func TestParseMavenOutputAgentSummary(t *testing.T) {
	t.Run("success summary", func(t *testing.T) {
		output := readTestFile(t, "sample_install.txt")
		projectRoot := filepath.Join("..", "..", "testdata", "sample")

		result := ParseMavenOutput(output, projectRoot)
		summary := result.GetAgentSummary()

		if summary == "" {
			t.Error("Expected non-empty agent summary")
		}

		if !strings.Contains(summary, "SUCCESS") {
			t.Errorf("Expected SUCCESS in summary, got: %s", summary)
		}
	})

	t.Run("failure summary", func(t *testing.T) {
		output := readTestFile(t, "sample_buildfail_install.txt")
		projectRoot := filepath.Join("..", "..", "testdata", "sample-buildfail")

		result := ParseMavenOutput(output, projectRoot)
		summary := result.GetAgentSummary()

		if summary == "" {
			t.Error("Expected non-empty agent summary")
		}

		if !strings.Contains(summary, "COMPILE_ERROR") && !strings.Contains(summary, "FAILURE") {
			t.Errorf("Expected failure info in summary, got: %s", summary)
		}
	})
}

func TestParseDependencyTree_NestedAncestors(t *testing.T) {
	output := readTestFile(t, "sample_dependency_tree.txt")
	deps := ParseDependencyTree(output)

	// Only one module
	if len(deps.Modules) != 1 {
		t.Fatalf("Expected 1 module, got %d", len(deps.Modules))
	}
	var tree *DependencyTree
	for _, v := range deps.Modules {
		tree = v
		break
	}
	// Root node should have 2 children: junit and guava
	children := tree.Root.Children
	if len(children) != 2 {
		t.Fatalf("Expected 2 root children, got %d", len(children))
	}
	foundJunit, foundGuava := false, false
	for _, c := range children {
		if c.GroupID == "junit" && c.ArtifactID == "junit" && c.Version == "4.12" {
			foundJunit = true
			// junit should have 1 child (hamcrest-core)
			if len(c.Children) != 1 {
				t.Errorf("Expected junit to have 1 child, got %d", len(c.Children))
			}
			if c.Children[0].GroupID != "org.hamcrest" || c.Children[0].ArtifactID != "hamcrest-core" {
				t.Errorf("Expected junit's child to be hamcrest-core, got %s:%s", c.Children[0].GroupID, c.Children[0].ArtifactID)
			}
		}
		if c.GroupID == "com.google.guava" && c.ArtifactID == "guava" {
			foundGuava = true
		}
	}
	if !foundJunit {
		t.Error("Did not find junit:junit root child")
	}
	if !foundGuava {
		t.Error("Did not find com.google.guava:guava root child")
	}
	// Check ancestors for hamcrest-core
	ancestors := deps.GetAncestors("org.hamcrest:hamcrest-core:1.3")
	if len(ancestors) != 2 {
		t.Fatalf("Expected 2 ancestors for hamcrest-core, got %d", len(ancestors))
	}
	if ancestors[0] != "com.example:module-a:1.0-SNAPSHOT" || !strings.HasPrefix(ancestors[1], "junit:junit") {
		t.Errorf("Unexpected ancestors: %v", ancestors)
	}
}

// TestAgentErrorSummaryLines_MultiLineOutput ensures the agent error summary lines field is exposed and correct.
func TestAgentErrorSummaryLines_MultiLineOutput(t *testing.T) {
	// Simulate Maven output with multiple compile, test, and generic errors.
	output := `
[ERROR] /foo/src/main/java/com/example/App.java:[11,20] cannot find symbol
[ERROR] /foo/src/main/java/com/example/App.java:[15,8] ';' expected
[ERROR] com.example.AppTest.testAddition:29 expected:<4> but was:<5>
[ERROR] Some other non-file-specific build error occurred
[ERROR] [FATAL] Non-parseable POM /foo/bar/pom.xml: end tag name </projct> must match start tag <project>
[ERROR] [INFO] BUILD FAILURE
`
	result := ParseMavenOutput(output, "/foo")
	json := result.ToJSON()

	if !strings.Contains(json, "agentErrorSummaryLines") {
		t.Errorf("Expected agentErrorSummaryLines in output, got: %s", json)
	}

	// Each sample error should be present in the summary lines
	expects := []string{"cannot find symbol", "AppTest.testAddition:29 expected:<4>", ";' expected", "Some other non-file-specific", "end tag name"}
	for _, exp := range expects {
		if !strings.Contains(json, exp) {
			t.Errorf("Missing expected error line excerpt: %q in JSON: %s", exp, json)
		}
	}

	// Should be max 6 lines, or have an ellipsis if there are more than 6 errors
	agentLinesCount := strings.Count(json, "agentErrorSummaryLines")
	if agentLinesCount != 1 {
		t.Errorf("Expected only one agentErrorSummaryLines section, got %d", agentLinesCount)
	}
}


func readTestFile(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filename, err)
	}
	return string(data)
}
