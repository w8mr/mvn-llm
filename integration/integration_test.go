package integration

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

type BlockOutput struct {
	Type  string   `json:"type"`
	Lines []string `json:"lines"`
}

type PhaseOutput struct {
	Name   string        `json:"name"`
	Blocks []BlockOutput `json:"blocks"`
}

type StructuredOutput struct {
	Phases []PhaseOutput `json:"phases"`
}

func TestMvnLlmStructuredJsonInstallSuccess(t *testing.T) {
	cmd := exec.Command("../mvn-llm", "--project-root=../testdata/sample", "install")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected success exit, got: %v\nOutput:\n%s", err, string(outBytes))
	}
	var out StructuredOutput
	if err := json.Unmarshal(outBytes, &out); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nRaw Out:\n%s", err, string(outBytes))
	}
	if len(out.Phases) < 3 {
		t.Errorf("Expected at least 3 phases, got %d", len(out.Phases))
	}
	// Check that summary contains BUILD SUCCESS
	found := false
	for _, p := range out.Phases {
		if p.Name == "summary" {
			for _, b := range p.Blocks {
				for _, line := range b.Lines {
					if strings.Contains(line, "BUILD SUCCESS") {
						found = true
						break
					}
				}
			}
		}
	}
	if !found {
		t.Errorf("Did not find BUILD SUCCESS in summary phase")
	}
}
