package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type BlockOutput struct {
	Type  string         `json:"type"`
	Lines []string       `json:"lines"`
	Meta  map[string]any `json:"meta"`
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

// Integration test for go run CLI output file deduplication
func TestMvnLlmClean_NoDuplicateCleanBlock(t *testing.T) {
	outputFile := "/tmp/test.log"
	cmd := exec.Command("go", "run", "../cmd/mvn-llm", "--output-file", outputFile, "--project-root=../testdata/sample", "clean")
	cmd.Env = append(os.Environ(), "GOFLAGS=") // ensures go run works consistently in test
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mvn-llm failed: %v\nOutput:\n%s", err, string(out))
	}
	f, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("could not open output file: %v", err)
	}
	defer f.Close()
	var out StructuredOutput
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		t.Fatalf("could not json-decode: %v", err)
	}
	modules := []string{"sample-multi", "module-a", "module-b"}
	for _, module := range modules {
		var found *PhaseOutput
		for i := range out.Phases {
			if out.Phases[i].Name == module {
				found = &out.Phases[i]
				break
			}
		}
		if found == nil {
			t.Errorf("Module %s not found in output", module)
			continue
		}
		count := 0
		for _, block := range found.Blocks {
			if block.Type == "build-block" && block.Meta != nil && block.Meta["plugin"] == "clean" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("Expected 1 clean build-block for %s, got %d", module, count)
			for _, block := range found.Blocks {
				if block.Meta["plugin"] == "clean" {
					t.Logf("Module %s: clean block meta=%v lines=%v", module, block.Meta, block.Lines)
				}
			}
		}
	}
}
