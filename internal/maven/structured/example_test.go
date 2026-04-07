package structured

import (
	"os"
	"strings"
	"testing"
)

func TestRegistry_ParseInitialization(t *testing.T) {
	data, err := os.ReadFile("../testdata/sample_install.txt")
	if err != nil {
		t.Fatalf("unable to read sample_install.txt: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	r := NewDefaultRegistry()
	parsed := r.ParseOutput(lines)

	if len(parsed.Phases) == 0 || parsed.Phases[0].Name != "initialization" {
		t.Errorf("Expected first phase to be 'initialization', got: %+v", parsed.Phases)
	}
	if len(parsed.Phases[0].Blocks) != 1 {
		t.Errorf("Expected one block in initialization phase, got: %d", len(parsed.Phases[0].Blocks))
	}
	if !strings.Contains(parsed.Phases[0].Blocks[0].Lines[0], "Scanning for projects") {
		t.Errorf("Expected initialization block to contain 'Scanning for projects', got: %v", parsed.Phases[0].Blocks[0].Lines)
	}
}
