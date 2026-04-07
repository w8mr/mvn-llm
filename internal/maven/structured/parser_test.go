package structured

import (
	"os"
	"strings"
	"testing"
)

func TestRegistry_ParseBuildAndSummaryPhases(t *testing.T) {
	data, err := os.ReadFile("../testdata/sample_install.txt")
	if err != nil {
		t.Fatalf("unable to read sample_install.txt: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	r := NewDefaultRegistry()
	parsed := r.ParseOutput(lines)
	if len(parsed.Phases) < 3 {
		t.Fatalf("Expected at least 3 phases (init, build, summary), got: %d", len(parsed.Phases))
	}

	if parsed.Phases[1].Name != "build-block" {
		t.Errorf("Expected second phase to be 'build-block', got: %s", parsed.Phases[1].Name)
	}
	if len(parsed.Phases[1].Blocks) == 0 {
		t.Errorf("Expected at least one block in build phase, got 0")
	}
	found := false
	for _, l := range parsed.Phases[1].Blocks[0].Lines {
		if strings.Contains(l, "Building") || strings.Contains(l, "Compiling") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected build block to contain compilation log lines")
	}

	if parsed.Phases[2].Name != "summary" {
		t.Errorf("Expected third phase to be 'summary', got: %s", parsed.Phases[2].Name)
	}
	if len(parsed.Phases[2].Blocks) == 0 {
		t.Errorf("Expected at least one block in summary phase, got 0")
	}
	found = false
	for _, l := range parsed.Phases[2].Blocks[0].Lines {
		if strings.Contains(l, "BUILD SUCCESS") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected summary block to contain 'BUILD SUCCESS'")
	}
}
