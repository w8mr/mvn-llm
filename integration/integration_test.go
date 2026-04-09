package integration

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

type Node struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Lines    []string       `json:"lines,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []Node         `json:"children,omitempty"`
}

type StructuredOutput struct {
	Root Node `json:"root"`
}

func TestMvnLlmStructuredJsonInstallSuccess(t *testing.T) {
	cmd := exec.Command("go", "run", "../cmd/mvn-llm", "-goal=install", "-project-root=../testdata/sample")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected success exit, got: %v\nOutput:\n%s", err, string(outBytes))
	}
	var out StructuredOutput
	if err := json.Unmarshal(outBytes, &out); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nRaw Out:\n%s", err, string(outBytes))
	}
	if len(out.Root.Children) < 2 {
		t.Errorf("Expected at least 2 children, got %d", len(out.Root.Children))
	}

	// Check for unparsable nodes - should be none
	for _, child := range out.Root.Children {
		if child.Type == "unparsable" {
			t.Errorf("Found unparsable node with lines: %v", child.Lines)
		}
	}

	found := false
	for _, child := range out.Root.Children {
		if child.Type == "summary" {
			for _, line := range child.Lines {
				if strings.Contains(line, "BUILD SUCCESS") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Errorf("Did not find BUILD SUCCESS in summary node")
	}
}
