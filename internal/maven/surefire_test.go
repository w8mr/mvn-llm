package maven

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseSurefireXML(t *testing.T) {
	projectRoot := filepath.Join("..", "..", "testdata", "sample")
	t.Logf("Project root: %s", projectRoot)

	exec.Command("mvn", "clean", "test", "-f", filepath.Join(projectRoot, "pom.xml")).Run()

	reports := FindSurefireReports(projectRoot)
	t.Logf("Found reports: %v", reports)

	report := ParseAllSurefireReports(projectRoot)
	t.Logf("Parsed suites: %d", len(report.Suites))
	for _, s := range report.Suites {
		t.Logf("Suite: %s, Tests: %d, Failures: %d", s.Name, s.Tests, s.Failures)
	}

	if len(report.Suites) == 0 {
		t.Fatal("Expected at least one surefire report suite")
	}

	foundModuleA := false
	foundModuleB := false

	for _, suite := range report.Suites {
		if suite.Name == "com.example.CalculatorTest" {
			foundModuleA = true
			if suite.Tests != 2 {
				t.Errorf("Expected 2 tests in CalculatorTest, got %d", suite.Tests)
			}
			if suite.Failures != 0 {
				t.Errorf("Expected 0 failures in CalculatorTest, got %d", suite.Failures)
			}
		}
		if suite.Name == "com.example.GreeterTest" {
			foundModuleB = true
			if suite.Tests != 1 {
				t.Errorf("Expected 1 test in GreeterTest, got %d", suite.Tests)
			}
			if suite.Failures != 0 {
				t.Errorf("Expected 0 failures in GreeterTest, got %d", suite.Failures)
			}
		}
	}

	if !foundModuleA {
		t.Error("Expected to find CalculatorTest suite")
	}
	if !foundModuleB {
		t.Error("Expected to find GreeterTest suite")
	}

	summary := report.GetTestSummary()
	if summary == "No tests run" {
		t.Error("Expected test summary with results")
	}
}

func TestSurefireReportSummary(t *testing.T) {
	projectRoot := filepath.Join("..", "..", "testdata", "sample")

	exec.Command("mvn", "clean", "test", "-f", filepath.Join(projectRoot, "pom.xml")).Run()

	report := ParseAllSurefireReports(projectRoot)
	summary := report.GetTestSummary()

	if summary == "No tests run" {
		t.Fatalf("Expected test results in summary, got: %s", summary)
	}

	if !contains(summary, "Tests run:") {
		t.Errorf("Expected summary to contain 'Tests run:', got: %s", summary)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
