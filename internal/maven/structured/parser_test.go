package structured

import (
	"encoding/json"
	"os"
	"testing"
)

// ...existing tests...

func Test_NoDuplicateCleanPluginBlocks(t *testing.T) {
	lines := []string{
		"[INFO] Scanning for projects...",
		"[INFO] ------------------------------------------------------------------------",
		"[INFO] Reactor Build Order:",
		"[INFO]",
		"[INFO] module-a                                                           [jar]",
		"[INFO]",
		"[INFO] ------------------------< com.example:module-a >------------------------",
		"[INFO] Building module-a 1.0-SNAPSHOT                                     [1/1]",
		"[INFO]   from module-a/pom.xml",
		"[INFO] --------------------------------[ jar ]---------------------------------",
		"[INFO]",
		"[INFO] --- clean:3.2.0:clean (default-clean) @ module-a ---",
		"[INFO] Deleting /some/path/target",
		"[INFO]",
	}
	r := NewDefaultRegistry()
	parsed := r.ParseOutput(lines)

	// Find the module phase for module-a
	var modulePhase *PhaseOutput
	for _, phase := range parsed.Phases {
		if phase.Name == "module-a" {
			modulePhase = &phase
			break
		}
	}
	if modulePhase == nil {
		t.Fatalf("Could not find module-a phase")
	}

	// Count build-blocks for clean plugin
	count := 0
	for _, block := range modulePhase.Blocks {
		if block.Type == "build-block" && block.Meta != nil && block.Meta["plugin"] == "clean" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected exactly one build-block for the clean plugin in module-a, got %d", count)
		for _, block := range modulePhase.Blocks {
			t.Logf("Block: type=%s meta=%v lines=%v", block.Type, block.Meta, block.Lines)
		}
	}
}

func Test_MultiModuleClean_NoDuplicates(t *testing.T) {
	lines := []string{
		"[INFO] Scanning for projects...",
		"[INFO] ------------------------------------------------------------------------",
		"[INFO] Reactor Build Order:",
		"[INFO] ",
		"[INFO] sample-multi                                                       [pom]",
		"[INFO] module-a                                                           [jar]",
		"[INFO] module-b                                                           [jar]",
		"[INFO] ",
		"[INFO] ----------------------< com.example:sample-multi >----------------------",
		"[INFO] Building sample-multi 1.0-SNAPSHOT                                 [1/3]",
		"[INFO]   from pom.xml",
		"[INFO] --------------------------------[ pom ]---------------------------------",
		"[INFO] ",
		"[INFO] --- clean:3.2.0:clean (default-clean) @ sample-multi ---",
		"[INFO] ",
		"[INFO] ------------------------< com.example:module-a >------------------------",
		"[INFO] Building module-a 1.0-SNAPSHOT                                     [2/3]",
		"[INFO]   from module-a/pom.xml",
		"[INFO] --------------------------------[ jar ]---------------------------------",
		"[INFO] ",
		"[INFO] --- clean:3.2.0:clean (default-clean) @ module-a ---",
		"[INFO] ",
		"[INFO] ------------------------< com.example:module-b >------------------------",
		"[INFO] Building module-b 1.0-SNAPSHOT                                     [3/3]",
		"[INFO]   from module-b/pom.xml",
		"[INFO] --------------------------------[ jar ]---------------------------------",
		"[INFO] ",
		"[INFO] --- clean:3.2.0:clean (default-clean) @ module-b ---",
		"[INFO] ------------------------------------------------------------------------",
		"[INFO] Reactor Summary for sample-multi 1.0-SNAPSHOT:",
		"[INFO] ",
		"[INFO] sample-multi ....................................... SUCCESS [  0.064 s]",
		"[INFO] module-a ........................................... SUCCESS [  0.001 s]",
		"[INFO] module-b ........................................... SUCCESS [  0.001 s]",
		"[INFO] ------------------------------------------------------------------------",
		"[INFO] BUILD SUCCESS",
		"[INFO] ------------------------------------------------------------------------",
		"[INFO] Total time:  0.127 s",
		"[INFO] Finished at: 2026-04-07T16:13:23+02:00",
		"[INFO] ------------------------------------------------------------------------",
	}
	r := NewDefaultRegistry()
	parsed := r.ParseOutput(lines)

	modules := []string{"sample-multi", "module-a", "module-b"}
	for _, module := range modules {
		var foundModule *PhaseOutput
		for _, phase := range parsed.Phases {
			if phase.Name == module {
				foundModule = &phase
				break
			}
		}
		if foundModule == nil {
			t.Errorf("Could not find module phase %s", module)
			continue
		}
		count := 0
		for _, block := range foundModule.Blocks {
			if block.Type == "build-block" && block.Meta != nil && block.Meta["plugin"] == "clean" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("Expected exactly one build-block for the clean plugin in module %s, got %d", module, count)
			for _, block := range foundModule.Blocks {
				t.Logf("Module %s: Block: type=%s meta=%v lines=%v", module, block.Type, block.Meta, block.Lines)
			}
		}
	}
}

// Integration test: read the REAL Maven CLI log, parse, check JSON output for duplication
func Test_RealLog_NoDuplicateCleanInJSON(t *testing.T) {
	logPath := "../testdata/actual-multimodule-clean.log" // Path relative to structured/
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()
	var lines []string
	buf := make([]byte, 1024*64)
	n, err := f.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read file: %v", err)
	}
	for _, l := range splitLines(string(buf[:n])) {
		lines = append(lines, l)
	}

	r := NewDefaultRegistry()
	parsed := r.ParseOutput(lines)
	j, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("Could not marshal JSON: %v", err)
	}
	t.Logf("Structured JSON output:\n%s", string(j))

	modules := []string{"sample-multi", "module-a", "module-b"}
	for _, module := range modules {
		var foundModule *PhaseOutput
		for _, phase := range parsed.Phases {
			if phase.Name == module {
				foundModule = &phase
				break
			}
		}
		if foundModule == nil {
			t.Errorf("Could not find module phase %s", module)
			continue
		}
		count := 0
		for _, block := range foundModule.Blocks {
			if block.Type == "build-block" && block.Meta != nil && block.Meta["plugin"] == "clean" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("Expected exactly one clean build-block for %s in JSON output, got %d", module, count)
			for _, block := range foundModule.Blocks {
				if block.Meta["plugin"] == "clean" {
					t.Logf("Module %s: Block: meta=%v lines=%v", module, block.Meta, block.Lines)
				}
			}
		}
	}
}

func splitLines(s string) []string {
	res := []string{}
	l := 0
	for i := range s {
		if s[i] == '\n' {
			res = append(res, s[l:i])
			l = i + 1
		}
	}
	if l < len(s) {
		res = append(res, s[l:])
	}
	return res
}
