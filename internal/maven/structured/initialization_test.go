package structured

import (
	"strings"
	"testing"
)

func TestParse_InitializationWithPreludeAndInfoLines(t *testing.T) {
	input := `Some random prelude line
Another prelude line
[INFO] Some info line before scanning
[INFO] Another info line
[INFO] Scanning for projects...
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---------------------< com.example:my-app >----------------------
[INFO] Building my-app 1.0.0                                          [1/1]
[INFO]   from pom.xml
[INFO] --------------------------------[ jar ]---------------------------------`

	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
		if len(child.Lines) > 0 {
			t.Logf("  First line: %q", child.Lines[0])
		}
	}

	// Expected structure:
	// - unparsable (prelude: 2 lines)
	// - initialization (starts at "[INFO] Some info line...", contains Scanning marker)
	// - module (my-app)

	if len(parsed.Root.Children) < 2 {
		t.Fatalf("Expected at least 2 children (unparsable prelude + initialization), got %d", len(parsed.Root.Children))
	}

	// First child should be unparsable prelude
	if parsed.Root.Children[0].Type != "unparsable" {
		t.Errorf("Expected first child to be unparsable, got %q", parsed.Root.Children[0].Type)
	}

	// Second child should be initialization
	var initNode *Node
	for i := range parsed.Root.Children {
		if parsed.Root.Children[i].Type == "initialization" {
			initNode = &parsed.Root.Children[i]
			break
		}
	}

	if initNode == nil {
		t.Fatalf("Expected to find initialization node")
	}

	// Initialization should start with "[INFO] Some info line before scanning"
	if len(initNode.Lines) == 0 {
		t.Fatalf("Initialization node has no lines")
	}

	firstLine := initNode.Lines[0]
	if firstLine != "[INFO] Some info line before scanning" {
		t.Errorf("Expected initialization to start with '[INFO] Some info line before scanning', got %q", firstLine)
	}

	// Initialization should contain "Scanning for projects..."
	foundScanning := false
	for _, line := range initNode.Lines {
		if line == "[INFO] Scanning for projects..." {
			foundScanning = true
			break
		}
	}
	if !foundScanning {
		t.Error("Expected initialization block to contain '[INFO] Scanning for projects...'")
	}

	// Initialization should contain "Reactor Build Order:"
	foundReactor := false
	for _, line := range initNode.Lines {
		if strings.Contains(line, "Reactor Build Order:") {
			foundReactor = true
			break
		}
	}
	if !foundReactor {
		t.Error("Expected initialization block to contain 'Reactor Build Order:'")
	}
}

func TestParse_InitializationStartsWithApacheMaven(t *testing.T) {
	input := `Apache Maven 3.8.1 (05c21c65bdfed0f71a2f2ada8b84da59348c4c5d)
Maven home: /usr/local/maven
Java version: 11.0.12, vendor: Oracle Corporation
[INFO] Scanning for projects...
[INFO] 
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---------------------< com.example:my-app >----------------------`

	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
	}

	// Should have initialization starting with "Apache Maven"
	var initNode *Node
	for i := range parsed.Root.Children {
		if parsed.Root.Children[i].Type == "initialization" {
			initNode = &parsed.Root.Children[i]
			break
		}
	}

	if initNode == nil {
		t.Fatalf("Expected to find initialization node")
	}

	if len(initNode.Lines) == 0 {
		t.Fatalf("Initialization node has no lines")
	}

	if !strings.HasPrefix(initNode.Lines[0], "Apache Maven") {
		t.Errorf("Expected initialization to start with 'Apache Maven', got %q", initNode.Lines[0])
	}
}
