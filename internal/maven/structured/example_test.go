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

	r := NewOutputParser()
	parsed := r.ParseOutput(lines)

	if len(parsed.Root.Children) == 0 || parsed.Root.Children[0].Type != "initialization" {
		t.Errorf("Expected first node to be 'initialization', got: %+v", parsed.Root.Children)
	}
	if len(parsed.Root.Children[0].Lines) == 0 {
		t.Errorf("Expected lines in initialization node, got: %d", len(parsed.Root.Children[0].Lines))
	}
	if !strings.Contains(parsed.Root.Children[0].Lines[0], "Scanning for projects") {
		t.Errorf("Expected initialization node to contain 'Scanning for projects', got: %v", parsed.Root.Children[0].Lines)
	}
}
