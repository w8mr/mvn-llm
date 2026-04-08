package intent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentic-ai/mvn-llm/internal/maven"
)

import "github.com/agentic-ai/mvn-llm/internal/testutil"

func readTestFile(t *testing.T, filename string) string {
	t.Helper()
	repoRoot := testutil.FindRepoRoot()
	path := filepath.Join(repoRoot, "testdata", "maven_output", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filename, err)
	}
	return string(data)
}

// Simulated CLI runner for phase
func fakeRunMaven(ctx context.Context, projectRoot string, args []string, opts maven.MavenOpts) (string, error) {
	goal := args[0]
	switch goal {
	case "dependency:tree":
		return readTestFile(&testing.T{}, "sample_dependency_tree.txt"), nil
	case "test":
		return readTestFile(&testing.T{}, "sample_testfail_test.txt"), nil
	case "install":
		return readTestFile(&testing.T{}, "sample_install.txt"), nil
	default:
		return "[INFO] Unknown command executed", nil
	}
}

func TestHandleMavenPhaseIntent(t *testing.T) {
	// Override RunMaven for test isolation
	saved := maven.RunMaven
	defer func() { maven.RunMaven = saved }()
	maven.RunMaven = fakeRunMaven

	tests := []struct {
		phase  string
		expect string
	}{
		{"dependency:tree", "DepsOutput"},
		{"unknown-goal", "string"},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			result, err := HandleMavenPhaseIntent(context.Background(), "/fake/path", tc.phase, maven.MavenOpts{})
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}
			// Check type string of the result
			resultType := ""
			switch result.(type) {
			case maven.DepsOutput:
				resultType = "DepsOutput"
			case string:
				resultType = "string"
			default:
				resultType = "unknown"
			}
			// Only check expected type for deps and unknown-goal (text lines)
			if tc.phase == "dependency:tree" || tc.phase == "unknown-goal" {
				if resultType != tc.expect {
					t.Errorf("For phase %s, expected type %s, got %s", tc.phase, tc.expect, resultType)
				}
			}
			// Optional: check a piece of the raw string result for unknown-goal
			if tc.phase == "unknown-goal" {
				if !strings.Contains(result.(string), "Unknown command") {
					t.Errorf("Expected 'Unknown command' for unknown-goal, got: %s", result)
				}
			}
		})
	}
}
