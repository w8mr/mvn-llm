package structured

import (
	"os"
	"strings"
	"testing"
)

func parseTestFile(filename string) *StructuredOutput {
	data, err := os.ReadFile("../testdata/" + filename)
	if err != nil {
		panic("unable to read test data: " + filename)
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

func TestParse_SampleDependencyTree(t *testing.T) {
	parsed := parseTestFile("sample_dependency_tree.txt")

	root := &parsed.Root
	if root.Name != "maven-build" {
		t.Errorf("Expected root name 'maven-build', got '%s'", root.Name)
	}
}
