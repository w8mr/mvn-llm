package structured

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentic-ai/mvn-llm/internal/testutil"
)

// helper for finding a child node by type
func findChildByType(children []Node, typ string) *Node {
	for _, c := range children {
		if c.Type == typ {
			return &c
		}
	}
	return nil
}

func TestParse_UnparsablePhaseBlockBeforeModuleHeader(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "unparsable_phase_before_module.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
		if len(child.Lines) > 0 && len(child.Lines) < 5 {
			for _, l := range child.Lines {
				t.Logf("  - %s", l)
			}
		}
	}

	// We expect the phase block (the "clean" block) to be grouped as an unparsable node before the module node
	if len(parsed.Root.Children) < 2 {
		t.Fatalf("Expected at least 2 children (unparsable, module...), got %d", len(parsed.Root.Children))
	}

	unparsableNode := findChildByType(parsed.Root.Children, "unparsable")
	if unparsableNode == nil {
		t.Error("Expected an unparsable node at root")
	} else {
		if len(unparsableNode.Lines) == 0 {
			t.Error("Unparsable node should have at least one line from the phase block")
		}
		if !strings.Contains(unparsableNode.Lines[0], "--- clean:3.2.0:clean (default-clean) @ module-a ---") {
			t.Errorf("Unparsable node should contain the phase block line, got: %v", unparsableNode.Lines)
		}
	}

	moduleNode := findChildByType(parsed.Root.Children, "module")
	if moduleNode == nil {
		t.Error("Expected module node")
	}
}

func TestParse_ModuleHeaderWithAlternateDashes(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "module_header_alternate_dashes.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
	}

	if len(parsed.Root.Children) < 2 {
		t.Fatalf("Expected at least 2 children (initialization, module...), got %d", len(parsed.Root.Children))
	}

	moduleNode := findChildByType(parsed.Root.Children, "module")
	if moduleNode == nil {
		t.Error("Expected module node")
	} else {
		if moduleNode.Name != "Baker Types" {
			t.Errorf("Expected module name 'Baker Types', got %q", moduleNode.Name)
		}
		if packaging, ok := moduleNode.Meta["packaging"].(string); !ok || packaging != "jar" {
			t.Errorf("Expected packaging 'jar', got %v", moduleNode.Meta["packaging"])
		}
		if groupId, ok := moduleNode.Meta["groupId"].(string); !ok || groupId != "com.ing.baker" {
			t.Errorf("Expected groupId 'com.ing.baker', got %v", moduleNode.Meta["groupId"])
		}
		if artifactId, ok := moduleNode.Meta["artifactId"].(string); !ok || artifactId != "baker-types" {
			t.Errorf("Expected artifactId 'baker-types', got %v", moduleNode.Meta["artifactId"])
		}
		if version, ok := moduleNode.Meta["version"].(string); !ok || version != "5.1.0-SNAPSHOT" {
			t.Errorf("Expected version '5.1.0-SNAPSHOT', got %v", moduleNode.Meta["version"])
		}
		if moduleIndex, ok := moduleNode.Meta["moduleIndex"].(int); !ok || moduleIndex != 2 {
			t.Errorf("Expected moduleIndex 2, got %v", moduleNode.Meta["moduleIndex"])
		}
		if moduleCount, ok := moduleNode.Meta["moduleCount"].(int); !ok || moduleCount != 28 {
			t.Errorf("Expected moduleCount 28, got %v", moduleNode.Meta["moduleCount"])
		}
	}

	buildBlockNode := findChildByType(moduleNode.Children, "build-block")
	if buildBlockNode == nil {
		t.Error("Expected build-block node inside module")
	}
}

func TestParse_TwoModules(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "two_modules.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
	}

	// Should have: initialization, module-a, module-b
	if len(parsed.Root.Children) != 3 {
		t.Fatalf("Expected 3 children (init, module-a, module-b), got %d", len(parsed.Root.Children))
	}

	moduleA := findChildByType(parsed.Root.Children, "module")
	if moduleA == nil {
		t.Error("Expected module node for module-a")
	} else {
		if moduleA.Name != "module-a" {
			t.Errorf("Expected module name 'module-a', got %q", moduleA.Name)
		}
		if len(moduleA.Children) == 0 {
			t.Error("Expected children inside module-a")
		}
	}

	// Find module-b (second module)
	moduleCount := 0
	var moduleB *Node
	for _, child := range parsed.Root.Children {
		if child.Type == "module" {
			moduleCount++
			if moduleCount == 2 {
				moduleB = &child
			}
		}
	}
	if moduleB == nil {
		t.Error("Expected module node for module-b")
	} else {
		if moduleB.Name != "module-b" {
			t.Errorf("Expected module name 'module-b', got %q", moduleB.Name)
		}
		if len(moduleB.Children) == 0 {
			t.Error("Expected children inside module-b")
		}
	}
}

func TestParse_ModuleWithWarnings(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "module_with_warnings.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
	}

	moduleNode := findChildByType(parsed.Root.Children, "module")
	if moduleNode == nil {
		t.Error("Expected module node")
	} else {
		t.Logf("Module lines: %v", moduleNode.Lines)
		if !strings.Contains(moduleNode.Lines[0], "module-a") {
			t.Errorf("Expected module header to contain module-a, got: %s", moduleNode.Lines[0])
		}
		// Check that warnings are included in module header
		hasWarnings := false
		for _, line := range moduleNode.Lines {
			if strings.Contains(line, "[WARNING]") {
				hasWarnings = true
				break
			}
		}
		if !hasWarnings {
			t.Error("Expected module header to include warnings")
		}
	}
}

func TestParse_SummaryWithReactor(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "summary_with_reactor.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q", i, child.Type, child.Name)
	}

	summaryNode := findChildByType(parsed.Root.Children, "summary")
	if summaryNode == nil {
		t.Error("Expected summary node")
	} else {
		t.Logf("Summary node found with %d lines", len(summaryNode.Lines))
		if meta, ok := summaryNode.Meta["modules"].([]map[string]any); ok {
			t.Logf("Found %d modules in summary", len(meta))
			for i, m := range meta {
				t.Logf("  Module %d: %+v", i, m)
			}
		}
	}
}

func TestParse_SampleInstall(t *testing.T) {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "maven_output", "sample_install.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	moduleA := findChildByName(parsed.Root.Children, "module-a")
	if moduleA == nil {
		t.Fatal("Expected module-a node")
	}

	surefireBlock := findChildByName(moduleA.Children, "surefire")
	if surefireBlock == nil {
		t.Fatal("Expected surefire build-block inside module-a")
	}

	hasTestHeader := false
	hasTestsRun := false
	for _, line := range surefireBlock.Lines {
		if strings.Contains(line, "T E S T S") {
			hasTestHeader = true
		}
		if strings.Contains(line, "Tests run:") {
			hasTestsRun = true
		}
	}
	if !hasTestHeader {
		t.Error("Expected surefire block to contain test header")
	}
	if !hasTestsRun {
		t.Error("Expected surefire block to contain test results")
	}
}

func findChildByName(children []Node, name string) *Node {
	for _, c := range children {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

func TestBuildPhaseParser_ExtractSummary(t *testing.T) {
	p := &BuildPhaseParser{}

	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name: "last info line",
			lines: []string{
				"[INFO] --- maven-compiler-plugin:3.15.0:compile (default-compile) @ my-app ---",
				"[INFO] Compiling 1 source file",
				"[INFO] ",
			},
			expected: "Successful: Compiling 1 source file",
		},
		{
			name: "last warning line",
			lines: []string{
				"[INFO] --- maven-compiler-plugin:3.15.0:compile (default-compile) @ my-app ---",
				"[WARNING] Some warning",
				"[INFO] Compiling 1 source file",
			},
			expected: "Successful: Some warning",
		},
		{
			name: "last error line",
			lines: []string{
				"[INFO] --- maven-compiler-plugin:3.15.0:compile (default-compile) @ my-app ---",
				"[INFO] Something failed",
				"[ERROR] Compilation failure",
			},
			expected: "Failure: Compilation failure",
		},
		{
			name:     "empty content",
			lines:    []string{"[INFO] --- maven-compiler-plugin:3.15.0:compile @ my-app ---"},
			expected: "Successful",
		},
		{
			name: "only header with empty info",
			lines: []string{
				"[INFO] --- maven-compiler-plugin:3.15.0:compile (default-compile) @ my-app ---",
				"[INFO] ",
			},
			expected: "Successful",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := p.ExtractSummary(tc.lines)
			if result != tc.expected {
				t.Errorf("ExtractSummary(%v) = %q, want %q", tc.name, result, tc.expected)
			}
		})
	}
}
