package integration

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMvnLlmDeps(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--no-clean", "--project-root=.", "deps")
	outBytes, err := cmd.CombinedOutput()
	_ = string(outBytes)
	if err == nil {
		t.Error("Expected error exit from deps in non-maven dir")
	}
}

func TestMvnLlmInstallSuccess(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "install")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err != nil {
		t.Errorf("Expected success exit, got: %v\nOutput:\n%s", err, out)
	}
	if !strings.Contains(out, "SUCCESS") {
		t.Errorf("Expected SUCCESS in output, got:\n%s", out)
	}
}

func TestMvnLlmTestSuccess(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "test")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err != nil {
		t.Errorf("Expected success exit, got: %v\nOutput:\n%s", err, out)
	}
	if !strings.Contains(out, "SUCCESS") {
		t.Errorf("Expected SUCCESS in output, got:\n%s", out)
	}
}

func TestMvnLlmInstallFailureCompile(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-buildfail", "install")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err == nil {
		t.Error("Expected error exit from build failure")
	}
	if !strings.Contains(out, "COMPILE_ERROR") {
		t.Errorf("Expected COMPILE_ERROR in output, got:\n%s", out)
	}
	if !strings.Contains(out, "module-a") {
		t.Errorf("Expected module-a in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Calculator.java:") {
		t.Errorf("Expected Calculator.java: in output, got:\n%s", out)
	}
}

func TestMvnLlmTestFailure(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-testfail", "test")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err == nil {
		t.Error("Expected error exit from test failure")
	}
	if !strings.Contains(out, "FAILURE") && !strings.Contains(out, "Failures:") {
		t.Errorf("Expected test failure details in output, got:\n%s", out)
	}
}

func TestMvnLlmSurefireReportsGenerated(t *testing.T) {
	moduleAPath := filepath.Join("..", "testdata", "sample", "module-a", "target", "surefire-reports")
	moduleBPath := filepath.Join("..", "testdata", "sample", "module-b", "target", "surefire-reports")

	cmdA := exec.Command("ls", moduleAPath)
	outA, errA := cmdA.CombinedOutput()
	if errA != nil {
		t.Fatalf("Failed to list surefire reports for module-a: %v\n%s", errA, outA)
	}
	if !strings.Contains(string(outA), "TEST-") {
		t.Errorf("Expected TEST-*.xml files in module-a surefire-reports, got:\n%s", outA)
	}

	cmdB := exec.Command("ls", moduleBPath)
	outB, errB := cmdB.CombinedOutput()
	if errB != nil {
		t.Fatalf("Failed to list surefire reports for module-b: %v\n%s", errB, outB)
	}
	if !strings.Contains(string(outB), "TEST-") {
		t.Errorf("Expected TEST-*.xml files in module-b surefire-reports, got:\n%s", outB)
	}
}

func TestMvnLlmSurefireFailureDetails(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-testfail", "test")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)

	if err == nil {
		t.Error("Expected error exit from test failure")
	}

	// Should have either "Tests run:" or "TEST_FAILURE"
	if !strings.Contains(out, "Tests run:") && !strings.Contains(out, "TEST_FAILURE") {
		t.Errorf("Expected test failure info in output, got:\n%s", out)
	}

	// Should have module info
	if !strings.Contains(out, "module-a") {
		t.Errorf("Expected module-a in output, got:\n%s", out)
	}
}

func TestMvnLlmInvalidPOM(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-invalid-pom", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	if !strings.Contains(out, "INVALID_POM") && !strings.Contains(out, "parsing") {
		t.Errorf("Expected POM parsing error in output, got: %s", out)
	}
}

func TestMvnLlmInvalidDependency(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-invalid-dep", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	if !strings.Contains(out, "INVALID_DEPENDENCY") {
		t.Errorf("Expected INVALID_DEPENDENCY in output, got: %s", out)
	}
}

func TestMvnLlmCircularDeps(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-circular-dep", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	if !strings.Contains(out, "CIRCULAR_DEPENDENCY") {
		t.Errorf("Expected CIRCULAR_DEPENDENCY in output, got: %s", out)
	}
}

func TestMvnLlmReactorIssue(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-reactor-issue", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	if !strings.Contains(out, "REACTOR_ERROR") {
		t.Errorf("Expected REACTOR_ERROR in output, got: %s", out)
	}
}

func TestMvnLlmTestCompileError(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-testcompfail", "install")
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err == nil {
		t.Error("Expected error exit from test compile failure")
	}
	if !strings.Contains(out, "TEST_COMPILE_ERROR") {
		t.Errorf("Expected TEST_COMPILE_ERROR in output, got: %s", out)
	}
	if !strings.Contains(out, "BrokenTest.java:") {
		t.Errorf("Expected BrokenTest.java: in output, got: %s", out)
	}
}

func TestMvnLlmPackageError(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-packagefail", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	if !strings.Contains(out, "SUCCESS") && !strings.Contains(out, "FAILURE") {
		t.Errorf("Expected status in output, got: %s", out)
	}
}

func TestMvnLlmJsonOutputSuccess(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	if !strings.Contains(out, `"status": "SUCCESS"`) {
		t.Errorf("Expected status=SUCCESS in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"testSummary":`) {
		t.Errorf("Expected testSummary in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonOutputCompileError(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-buildfail", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	if !strings.Contains(out, `"status": "COMPILE_ERROR"`) {
		t.Errorf("Expected status=COMPILE_ERROR in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"failedModule": "module-a"`) {
		t.Errorf("Expected failedModule in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"failureLocation": "module-a/src/main/java/com/example/Calculator.java:6:5"`) {
		t.Errorf("Expected failureLocation in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonOutputTestFailure(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-testfail", "-o", "json", "test")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Status should be TEST_FAILURE
	if !strings.Contains(out, `"status": "TEST_FAILURE"`) {
		t.Errorf("Expected status=TEST_FAILURE in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonOutputInvalidPom(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-invalid-pom", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Status should be the failure type
	if !strings.Contains(out, `"status": "INVALID_POM"`) {
		t.Errorf("Expected status=INVALID_POM in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonResumeCommand(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-buildfail", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	if !strings.Contains(out, "resumeCommand") {
		t.Errorf("Expected resumeCommand in JSON, got: %s", out)
	}
	if !strings.Contains(out, "-rf module-a") {
		t.Errorf("Expected resume command with module-a, got: %s", out)
	}
}

func TestMvnLlmJsonAgentSummarySuccess(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	if !strings.Contains(out, "agentSummary") {
		t.Errorf("Expected agentSummary in JSON, got: %s", out)
	}
	if !strings.Contains(out, "SUCCESS") {
		t.Errorf("Expected SUCCESS in agentSummary, got: %s", out)
	}
}

func TestMvnLlmJsonAgentSummaryFailure(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-buildfail", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	if !strings.Contains(out, "agentSummary") {
		t.Errorf("Expected agentSummary in JSON, got: %s", out)
	}
	if !strings.Contains(out, "COMPILE_ERROR") {
		t.Errorf("Expected COMPILE_ERROR in agentSummary, got: %s", out)
	}
	if !strings.Contains(out, "module-a") {
		t.Errorf("Expected module-a in agentSummary, got: %s", out)
	}
	if !strings.Contains(out, "Calculator.java:6:5") {
		t.Errorf("Expected Calculator.java:6:5 in agentSummary, got: %s", out)
	}
}

func TestMvnLlmJsonSuccessNoResumeCommand(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Successful builds should NOT have resumeCommand
	if strings.Contains(out, `"resumeCommand"`) {
		t.Errorf("Expected no resumeCommand for successful build, got: %s", out)
	}
	// Should have status = SUCCESS
	if !strings.Contains(out, `"status": "SUCCESS"`) {
		t.Errorf("Expected status=SUCCESS in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonInvalidPomModuleAndResume(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-invalid-pom", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Pre-build failures should NOT have failedModule
	if strings.Contains(out, `"failedModule"`) {
		t.Errorf("Expected no failedModule for pre-build failure (INVALID_POM), got: %s", out)
	}
	// Pre-build failures should NOT have resumeCommand
	if strings.Contains(out, `"resumeCommand"`) {
		t.Errorf("Expected no resumeCommand for pre-build failure (INVALID_POM), got: %s", out)
	}
	// Status should contain the failure type
	if !strings.Contains(out, `"status": "INVALID_POM"`) {
		t.Errorf("Expected status=INVALID_POM in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonInvalidDepModuleAndResume(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-invalid-dep", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Pre-build failures should NOT have failedModule
	if strings.Contains(out, `"failedModule"`) {
		t.Errorf("Expected no failedModule for pre-build failure (INVALID_DEPENDENCY), got: %s", out)
	}
	// Pre-build failures should NOT have resumeCommand
	if strings.Contains(out, `"resumeCommand"`) {
		t.Errorf("Expected no resumeCommand for pre-build failure (INVALID_DEPENDENCY), got: %s", out)
	}
	// Status should contain the failure type
	if !strings.Contains(out, `"status": "INVALID_DEPENDENCY"`) {
		t.Errorf("Expected status=INVALID_DEPENDENCY in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonCircularDepModuleAndResume(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-circular-dep", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Pre-build failures should NOT have failedModule
	if strings.Contains(out, `"failedModule"`) {
		t.Errorf("Expected no failedModule for pre-build failure (CIRCULAR_DEPENDENCY), got: %s", out)
	}
	// Pre-build failures should NOT have resumeCommand
	if strings.Contains(out, `"resumeCommand"`) {
		t.Errorf("Expected no resumeCommand for pre-build failure (CIRCULAR_DEPENDENCY), got: %s", out)
	}
	// Status should contain the failure type
	if !strings.Contains(out, `"status": "CIRCULAR_DEPENDENCY"`) {
		t.Errorf("Expected status=CIRCULAR_DEPENDENCY in JSON, got: %s", out)
	}
}

func TestMvnLlmJsonReactorErrorModuleAndResume(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample-reactor-issue", "-o", "json", "install")
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)

	// Pre-build failures should NOT have failedModule
	if strings.Contains(out, `"failedModule"`) {
		t.Errorf("Expected no failedModule for pre-build failure (REACTOR_ERROR), got: %s", out)
	}
	// Pre-build failures should NOT have resumeCommand
	if strings.Contains(out, `"resumeCommand"`) {
		t.Errorf("Expected no resumeCommand for pre-build failure (REACTOR_ERROR), got: %s", out)
	}
	// Status should contain the failure type
	if !strings.Contains(out, `"status": "REACTOR_ERROR"`) {
		t.Errorf("Expected status=REACTOR_ERROR in JSON, got: %s", out)
	}
}
