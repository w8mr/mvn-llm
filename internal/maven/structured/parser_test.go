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

	parsed := NewOutputParser().ParseOutput(lines)

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
