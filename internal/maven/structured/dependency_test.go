package structured

import (
	"os"
	"strings"
	"testing"
)

func TestDependencyTreeParser(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/maven_output/sample_dependency_tree.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

	parser := &DependencyTreeParser{}

	// Find dependency tree at start of file
	startIdx := 0
	ok, markerLen := parser.StartMarker(lines, startIdx)
	if !ok || markerLen == 0 {
		t.Fatal("DependencyTreeParser StartMarker failed")
	}

	allParsers := []Parser{parser}
	found, consumed, ok := parser.ExtractLines(lines, startIdx, allParsers)
	if !ok {
		t.Fatal("ExtractLines failed")
	}

	meta := parser.ParseMetaData(found)

	// Verify root
	root, ok := meta["root"].(map[string]any)
	if !ok {
		t.Fatal("Expected root in meta")
	}
	if root["groupId"] != "com.example" || root["artifactId"] != "module-a" {
		t.Errorf("Expected com.example:module-a, got %v:%v", root["groupId"], root["artifactId"])
	}

	// Verify dependencies
	deps, ok := meta["dependencies"].([]map[string]any)
	if !ok {
		t.Fatal("Expected dependencies in meta")
	}
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// First dep: junit
	if deps[0]["groupId"] != "junit" || deps[0]["artifactId"] != "junit" {
		t.Errorf("First dep should be junit:junit, got %v:%v", deps[0]["groupId"], deps[0]["artifactId"])
	}

	// Second dep: guava
	if deps[1]["groupId"] != "com.google.guava" || deps[1]["artifactId"] != "guava" {
		t.Errorf("Second dep should be com.google.guava:guava, got %v:%v", deps[1]["groupId"], deps[1]["artifactId"])
	}

	_ = consumed
}

func TestDependencyTreeWithFilter(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/maven_output/sample_dependency_tree.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

	parser := &DependencyTreeParser{}

	// Find and parse tree
	found, _, ok := parser.ExtractLines(lines, 0, []Parser{parser})
	if !ok {
		t.Fatal("ExtractLines failed")
	}

	meta := parser.ParseMetaData(found)

	// Test filter for junit
	filtered := filterDependencyTree(meta, "junit")
	if len(filtered) != 2 { // root + 1 dep
		t.Errorf("Filter 'junit' should return 2 lines, got %d", len(filtered))
	}

	// Test filter for guava
	filtered = filterDependencyTree(meta, "guava")
	if len(filtered) != 2 {
		t.Errorf("Filter 'guava' should return 2 lines, got %d", len(filtered))
	}

	// Test filter for non-existent
	filtered = filterDependencyTree(meta, "nonexistent")
	if len(filtered) != 1 { // only root
		t.Errorf("Filter 'nonexistent' should return 1 line, got %d", len(filtered))
	}

	// Test full group:artifact filter
	filtered = filterDependencyTree(meta, "junit:junit")
	if len(filtered) != 2 {
		t.Errorf("Filter 'junit:junit' should return 2 lines, got %d", len(filtered))
	}
}

func filterDependencyTree(meta map[string]any, filter string) []string {
	var lines []string

	if root, ok := meta["root"].(map[string]any); ok {
		groupID, _ := root["groupId"].(string)
		artifactID, _ := root["artifactId"].(string)
		version, _ := root["version"].(string)
		lines = append(lines, groupID+":"+artifactID+":"+version)
	}

	if deps, ok := meta["dependencies"].([]map[string]any); ok {
		for _, dep := range deps {
			groupID, _ := dep["groupId"].(string)
			artifactID, _ := dep["artifactId"].(string)
			version, _ := dep["version"].(string)
			scope, _ := dep["scope"].(string)

			matchStr := groupID + ":" + artifactID
			fullMatch := groupID + ":" + artifactID + ":" + version
			if filter != groupID && filter != artifactID && filter != matchStr && filter != fullMatch {
				continue
			}

			line := "+- " + groupID + ":" + artifactID + ":" + version
			if scope != "" {
				line += ":" + scope
			}
			lines = append(lines, line)
		}
	}

	return lines
}
