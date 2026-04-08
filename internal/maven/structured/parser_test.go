package structured

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

import "github.com/agentic-ai/mvn-llm/internal/testutil"

func parseTestFile(filename string) *StructuredOutput {
	repoRoot := testutil.FindRepoRoot()
	filePath := filepath.Join(repoRoot, "testdata", "maven_output", filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic("unable to read test data: " + filePath)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	return NewOutputParser().ParseOutput(lines)
}

func findChildByType(children []Node, nodeType string) *Node {
	for i := range children {
		if children[i].Type == nodeType {
			return &children[i]
		}
	}
	return nil
}

func findChildrenByType(children []Node, nodeType string) []Node {
	var result []Node
	for i := range children {
		if children[i].Type == nodeType {
			result = append(result, children[i])
		}
	}
	return result
}

func findChildByName(children []Node, name string) *Node {
	for i := range children {
		if children[i].Name == name {
			return &children[i]
		}
	}
	return nil
}

func findNodeByType(root *Node, nodeType string) *Node {
	if root.Type == nodeType {
		return root
	}
	for i := range root.Children {
		if found := findNodeByType(&root.Children[i], nodeType); found != nil {
			return found
		}
	}
	return nil
}

func findBuildBlocksWithStatus(parsed *StructuredOutput, status string) []Node {
	var result []Node
	for _, child := range parsed.Root.Children {
		if child.Type == "module" {
			for _, block := range child.Children {
				if block.Type == "build-block" {
					if metaStatus, ok := block.Meta["status"].(string); ok && metaStatus == status {
						result = append(result, block)
					}
				}
			}
		}
	}
	return result
}

func findBuildBlockByName(parsed *StructuredOutput, name string) *Node {
	for _, child := range parsed.Root.Children {
		if child.Type == "module" {
			for _, block := range child.Children {
				if block.Type == "build-block" && block.Name == name {
					return &block
				}
			}
		}
	}
	return nil
}

func containsError(parsed *StructuredOutput) bool {
	var check func(nodes []Node) bool
	check = func(nodes []Node) bool {
		for _, n := range nodes {
			for _, line := range n.Lines {
				if strings.Contains(line, "[ERROR]") {
					return true
				}
			}
			if len(n.Children) > 0 && check(n.Children) {
				return true
			}
		}
		return false
	}
	return check([]Node{parsed.Root})
}

func getMetaString(node *Node, key string) string {
	if node == nil {
		return ""
	}
	if v, ok := node.Meta[key].(string); ok {
		return v
	}
	return ""
}

func getMetaInt(node *Node, key string) int {
	if node == nil {
		return 0
	}
	if v, ok := node.Meta[key].(int); ok {
		return v
	}
	return 0
}

func getMetaModules(node *Node) any {
	if node == nil {
		return nil
	}
	return node.Meta["modules"]
}

func TestParse_SampleInstall_Success(t *testing.T) {
	parsed := parseTestFile("sample_install.txt")

	root := &parsed.Root
	if root.Name != "maven-build" {
		t.Errorf("Expected root name 'maven-build', got '%s'", root.Name)
	}
	if root.Type != "root" {
		t.Errorf("Expected root type 'root', got '%s'", root.Type)
	}

	children := root.Children
	if len(children) < 4 {
		t.Errorf("Expected at least 4 children (init + 3 modules + summary), got %d", len(children))
	}

	initNode := findChildByType(children, "initialization")
	if initNode == nil {
		t.Errorf("Expected initialization node")
	} else {
		modules := getMetaModules(initNode)
		if modules == nil {
			t.Errorf("Expected modules in initialization meta")
		}
	}

	moduleNodes := findChildrenByType(children, "module")
	if len(moduleNodes) != 3 {
		t.Errorf("Expected 3 module nodes, got %d", len(moduleNodes))
	}

	for _, mod := range moduleNodes {
		if getMetaString(&mod, "name") == "" {
			t.Errorf("Module missing name in meta")
		}
		if getMetaString(&mod, "groupId") == "" {
			t.Errorf("Module missing groupId in meta")
		}

		blocks := findChildrenByType(mod.Children, "build-block")
		if len(blocks) == 0 {
			t.Errorf("Module %s has no build blocks", mod.Name)
		}

		for _, block := range blocks {
			if getMetaString(&block, "status") == "" {
				t.Errorf("Build block missing status in meta")
			}
		}
	}

	summaryNode := findChildByType(children, "summary")
	if summaryNode == nil {
		t.Errorf("Expected summary node")
	} else {
		if getMetaString(summaryNode, "overallStatus") != "BUILD SUCCESS" {
			t.Errorf("Expected overallStatus 'BUILD SUCCESS', got '%s'", getMetaString(summaryNode, "overallStatus"))
		}

		modules := getMetaModules(summaryNode)
		if modules == nil {
			t.Errorf("Expected modules in summary meta")
		}
	}
}

func TestParse_SampleBuildFail(t *testing.T) {
	parsed := parseTestFile("sample_buildfail_install.txt")

	children := parsed.Root.Children

	summaryNode := findChildByType(children, "summary")
	if summaryNode == nil {
		t.Errorf("Expected summary node")
	} else {
		if getMetaString(summaryNode, "overallStatus") != "BUILD FAILURE" {
			t.Errorf("Expected overallStatus 'BUILD FAILURE', got '%s'", getMetaString(summaryNode, "overallStatus"))
		}
	}

	failedBlocks := findBuildBlocksWithStatus(parsed, "FAILED")
	if len(failedBlocks) == 0 {
		t.Errorf("Expected at least one build block with status FAILED")
	}
}

func TestParse_SampleTestFail(t *testing.T) {
	parsed := parseTestFile("sample_testfail_test.txt")

	children := parsed.Root.Children

	summaryNode := findChildByType(children, "summary")
	if summaryNode == nil {
		t.Errorf("Expected summary node")
	} else {
		if getMetaString(summaryNode, "overallStatus") != "BUILD FAILURE" {
			t.Errorf("Expected overallStatus 'BUILD FAILURE', got '%s'", getMetaString(summaryNode, "overallStatus"))
		}
	}

	surefire := findBuildBlockByName(parsed, "surefire")
	if surefire == nil {
		t.Errorf("Expected surefire build block")
	} else {
		if surefire.Meta == nil {
			t.Errorf("surefire block missing meta")
		}
	}
}

func TestParse_SampleTestCompFail(t *testing.T) {
	parsed := parseTestFile("sample_testcompfail_install.txt")

	children := parsed.Root.Children

	if len(children) == 0 {
		t.Errorf("Expected parsed children")
		return
	}

	summaryNode := findChildByType(children, "summary")
	if summaryNode == nil {
		t.Logf("No summary node (may be single module format)")
	}
}

func TestParse_SampleInvalidDep(t *testing.T) {
	parsed := parseTestFile("sample_invalid_dep_install.txt")

	if !containsError(parsed) {
		t.Errorf("Expected error in output for invalid dependency")
	}

	children := parsed.Root.Children

	if len(children) == 0 {
		t.Errorf("Expected parsed children")
	}
}

func TestParse_SampleReactorIssue(t *testing.T) {
	parsed := parseTestFile("sample_reactor_issue_install.txt")

	if !containsError(parsed) {
		t.Errorf("Expected error in output for reactor issue")
	}

	root := &parsed.Root

	children := root.Children
	if len(children) == 0 {
		t.Logf("No children (expected for early failure)")
	}
}

func TestParse_SampleCircularDep(t *testing.T) {
	parsed := parseTestFile("sample_circular_dep_install.txt")

	if !containsError(parsed) {
		t.Errorf("Expected error in output for circular dependency")
	}

	children := parsed.Root.Children

	summaryNode := findChildByType(children, "summary")
	if summaryNode != nil {
		if getMetaString(summaryNode, "overallStatus") != "BUILD FAILURE" {
			t.Errorf("Expected overallStatus 'BUILD FAILURE', got '%s'", getMetaString(summaryNode, "overallStatus"))
		}
	}
}

func TestParse_UnparsableLinesCombined(t *testing.T) {
	// Read test data file with unparsable lines
	data, err := os.ReadFile("testdata/unparsable_lines.txt")
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	// Parse the lines
	p := NewOutputParser()
	result, consumed, ok := p.Parse(lines, 0)

	t.Logf("Parse result: ok=%v consumed=%d", ok, consumed)
	if result != nil {
		t.Logf("Root children: %d", len(result.Children))
		for i, c := range result.Children {
			t.Logf("  Child %d: type=%s name=%s lines=%d", i, c.Type, c.Name, len(c.Lines))
		}
	}

	if !ok {
		t.Error("Parse should succeed")
		return
	}

	// Should consume all lines
	if consumed != len(lines) {
		t.Errorf("Expected to consume %d lines, got %d", len(lines), consumed)
	}

	// Verify combining: consecutive unparsable lines should be ONE node
	unparsableNodeCount := 0
	var totalUnparsableLines int
	for _, child := range result.Children {
		if child.Type == "unparsable" {
			unparsableNodeCount++
			totalUnparsableLines += len(child.Lines)
			t.Logf("Unparsable node with %d lines: %v", len(child.Lines), child.Lines)
		}
	}

	// Key assertion: consecutive unparsable lines should be combined into ONE node
	if unparsableNodeCount != 1 {
		t.Errorf("Expected 1 unparsable node (consecutive lines combined), got %d", unparsableNodeCount)
	}

	// All lines should be in that single node
	if totalUnparsableLines != 3 {
		t.Errorf("Expected 3 lines in unparsable node, got %d", totalUnparsableLines)
	}
}

func TestParse_UnparsableBetweenInitializationAndModule(t *testing.T) {
	data, err := os.ReadFile("testdata/unparsable_in_middle.txt")
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines)

	t.Logf("Root children: %d", len(parsed.Root.Children))
	for i, c := range parsed.Root.Children {
		t.Logf("  Child %d: type=%s name=%s lines=%d children=%d", i, c.Type, c.Name, len(c.Lines), len(c.Children))
		for j, child := range c.Children {
			t.Logf("    Grandchild %d: type=%s name=%s", j, child.Type, child.Name)
		}
	}

	// Should have: initialization, module, summary = 3 direct children
	if len(parsed.Root.Children) != 3 {
		t.Errorf("Expected 3 children (init + module + summary), got %d", len(parsed.Root.Children))
	}

	// Verify unparsable node exists inside module (not at root) and has 2 lines
	moduleNode := findChildByType(parsed.Root.Children, "module")
	if moduleNode == nil {
		t.Error("Expected module node")
	} else {
		unparsableInModule := findChildByType(moduleNode.Children, "unparsable")
		if unparsableInModule == nil {
			t.Error("Expected unparsable node inside module")
		} else if len(unparsableInModule.Lines) != 2 {
			t.Errorf("Expected unparsable with 2 lines, got %d", len(unparsableInModule.Lines))
		}

		// Verify build-block is also inside module
		buildBlockInModule := findChildByType(moduleNode.Children, "build-block")
		if buildBlockInModule == nil {
			t.Error("Expected build-block inside module")
		}
	}

	initNode := findChildByType(parsed.Root.Children, "initialization")
	if initNode == nil {
		t.Error("Expected initialization node")
	}

	summaryNode := findChildByType(parsed.Root.Children, "summary")
	if summaryNode == nil {
		t.Error("Expected summary node")
	}
}

func TestParse_UnparsableBetweenModuleHeaderLines(t *testing.T) {
	data, err := os.ReadFile("testdata/unparsable_between_module_header.txt")
	if err != nil {
		t.Fatalf("unable to read test data: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines)

	// Should have: initialization, unparsable, module, summary = 4 children
	if len(parsed.Root.Children) != 4 {
		t.Errorf("Expected 4 children (init + unparsable + module + summary), got %d", len(parsed.Root.Children))
	}

	initNode := findChildByType(parsed.Root.Children, "initialization")
	if initNode == nil {
		t.Error("Expected initialization node")
	}

	unparsableNode := findChildByType(parsed.Root.Children, "unparsable")
	if unparsableNode == nil {
		t.Error("Expected unparsable node")
	} else if len(unparsableNode.Lines) != 3 {
		t.Errorf("Expected unparsable with 3 lines (module header + 2 random), got %d", len(unparsableNode.Lines))
	}

	moduleNode := findChildByType(parsed.Root.Children, "module")
	if moduleNode == nil {
		t.Error("Expected module node")
	}

	summaryNode := findChildByType(parsed.Root.Children, "summary")
	if summaryNode == nil {
		t.Error("Expected summary node")
	}
}
