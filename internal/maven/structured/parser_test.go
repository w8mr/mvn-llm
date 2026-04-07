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
	if len(parsed.Phases) < 2 {
		t.Fatalf("Expected at least 2 phases (modules and summary), got: %d", len(parsed.Phases))
	}

	// With the hierarchical structure, phases are now organized by module
	// Find the summary phase
	var summaryPhase *PhaseOutput
	for i := range parsed.Phases {
		if parsed.Phases[i].Name == "summary" {
			summaryPhase = &parsed.Phases[i]
			break
		}
	}
	if summaryPhase == nil {
		t.Errorf("Expected to find a 'summary' phase")
	} else {
		if len(summaryPhase.Blocks) == 0 {
			t.Errorf("Expected at least one block in summary phase, got 0")
		}
		found := false
		for _, b := range summaryPhase.Blocks {
			for _, l := range b.Lines {
				if strings.Contains(l, "BUILD SUCCESS") {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("Expected summary block to contain 'BUILD SUCCESS'")
		}
	}
}
