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
	block := parsed.Phases[2].Blocks[0]
	found = false
	for _, l := range block.Lines {
		if strings.Contains(l, "BUILD SUCCESS") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected summary block to contain 'BUILD SUCCESS'")
	}

	// Validate summary meta fields
	if block.Meta == nil {
		t.Errorf("Expected summary block to have Meta field, got nil")
	} else {
		if mods, ok := block.Meta["modules"].([]interface{}); !ok || len(mods) != 3 {
			t.Errorf("Expected 3 modules in meta, got %v", block.Meta["modules"])
		}
		if status, ok := block.Meta["overallStatus"].(string); !ok || status != "BUILD SUCCESS" {
			t.Errorf("Expected overallStatus 'BUILD SUCCESS', got %v", block.Meta["overallStatus"])
		}
		if tt, ok := block.Meta["totalTime"].(string); !ok || tt == "" {
			t.Errorf("Expected non-empty totalTime, got %v", block.Meta["totalTime"])
		}
		if fa, ok := block.Meta["finishedAt"].(string); !ok || fa == "" {
			t.Errorf("Expected non-empty finishedAt, got %v", block.Meta["finishedAt"])
		}
	}
}
