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
	if len(parsed.Root.Children) < 2 {
		t.Fatalf("Expected at least 2 phases (modules and summary), got: %d", len(parsed.Root.Children))
	}

	var summaryNode *Node
	for i := range parsed.Root.Children {
		if parsed.Root.Children[i].Type == "summary" {
			summaryNode = &parsed.Root.Children[i]
			break
		}
	}
	if summaryNode == nil {
		t.Errorf("Expected to find a 'summary' node")
	} else {
		if len(summaryNode.Lines) == 0 {
			t.Errorf("Expected at least one line in summary node, got 0")
		}
		found := false
		for _, l := range summaryNode.Lines {
			if strings.Contains(l, "BUILD SUCCESS") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected summary node to contain 'BUILD SUCCESS'")
		}
	}
}
