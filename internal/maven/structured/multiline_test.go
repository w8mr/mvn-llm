package structured

import (
	"strings"
	"testing"
)

func TestParse_MultiLineModuleHeader(t *testing.T) {
	input := `[INFO] Scanning for projects...
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] module-a                                                         [jar]
[INFO] module-b                                                         [jar]
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building HawtJNI 1.6
[INFO] ------------------------------------------------------------------------ 
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ HawtJNI ---
[INFO] 
[INFO] --- compiler:3.8.1:compile (default-compile) @ HawtJNI ---
[INFO] Compiling 1 source file
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building HawtJNI-native 1.8
[INFO] ------------------------------------------------------------------------ 
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ HawtJNI-native ---
[INFO] 
[INFO] --- compiler:3.8.1:compile (default-compile) @ HawtJNI-native ---
[INFO] Compiling 1 source file
`
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
		if child.Type == "module" {
			t.Logf("  Meta: %v", child.Meta)
			for j, gc := range child.Children {
				t.Logf("    Child %d: Type=%q, Name=%q", j, gc.Type, gc.Name)
			}
		}
	}

	// Should have: initialization, module HawtJNI, module HawtJNI-native
	if len(parsed.Root.Children) != 3 {
		t.Fatalf("Expected 3 children (init, HawtJNI, HawtJNI-native), got %d", len(parsed.Root.Children))
	}

	// Check first module name
	firstModule := findModuleByName(parsed.Root.Children, "HawtJNI")
	if firstModule == nil {
		t.Error("Expected to find module with name 'HawtJNI'")
	}

	// Check second module name
	secondModule := findModuleByName(parsed.Root.Children, "HawtJNI-native")
	if secondModule == nil {
		t.Error("Expected to find module with name 'HawtJNI-native'")
	}

	// Check that modules were parsed correctly
	if firstModule != nil {
		t.Logf("First module has %d children and %d lines", len(firstModule.Children), len(firstModule.Lines))
		// For simple format, build blocks are included in the module lines but not parsed as children
		// This is acceptable - they can be parsed later if needed
		if len(firstModule.Lines) < 5 {
			t.Errorf("Expected module 'HawtJNI' to have content lines, got %d lines", len(firstModule.Lines))
		}
	}

	if secondModule != nil {
		t.Logf("Second module has %d children and %d lines", len(secondModule.Children), len(secondModule.Lines))
		if len(secondModule.Lines) < 5 {
			t.Errorf("Expected module 'HawtJNI-native' to have content lines, got %d lines", len(secondModule.Lines))
		}
	}
}

func findModuleByName(children []Node, name string) *Node {
	for i := range children {
		if children[i].Type == "module" && children[i].Name == name {
			return &children[i]
		}
	}
	return nil
}

func TestParse_MultiLineModuleHeader_MultiWordNames(t *testing.T) {
	input := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] HawtJNI
[INFO] HawtJNI Runtime
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building HawtJNI 1.6
[INFO] ------------------------------------------------------------------------
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building HawtJNI Runtime 1.6
[INFO] ------------------------------------------------------------------------
[INFO] 
[INFO] --- maven-resources-plugin:2.6:resources (default-resources) @ hawtjni-runtime ---
[INFO] Compiling 14 source files
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] HawtJNI ........................................... SUCCESS [0.001s]
[INFO] HawtJNI Runtime ................................... FAILURE [2.813s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE
[INFO] ------------------------------------------------------------------------`
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")

	parsed := NewOutputParser().ParseOutput(lines, nil, nil)

	t.Logf("Root children count: %d", len(parsed.Root.Children))
	for i, child := range parsed.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q, Lines=%d", i, child.Type, child.Name, len(child.Lines))
		if child.Type == "module" {
			t.Logf("  Meta: %v", child.Meta)
		}
	}

	// Should have: initialization, module HawtJNI, module HawtJNI Runtime, summary
	if len(parsed.Root.Children) < 3 {
		t.Fatalf("Expected at least 3 children (init, HawtJNI, HawtJNI Runtime), got %d", len(parsed.Root.Children))
	}

	// Check first module name
	firstModule := findModuleByName(parsed.Root.Children, "HawtJNI")
	if firstModule == nil {
		t.Error("Expected to find module with name 'HawtJNI'")
	} else {
		if name, ok := firstModule.Meta["name"].(string); !ok || name != "HawtJNI" {
			t.Errorf("Expected module meta name 'HawtJNI', got %q", name)
		}
	}

	// Check second module name (multi-word)
	secondModule := findModuleByName(parsed.Root.Children, "HawtJNI Runtime")
	if secondModule == nil {
		t.Error("Expected to find module with name 'HawtJNI Runtime'")
	} else {
		if name, ok := secondModule.Meta["name"].(string); !ok || name != "HawtJNI Runtime" {
			t.Errorf("Expected module meta name 'HawtJNI Runtime', got %q", name)
		}
		if version, ok := secondModule.Meta["version"].(string); !ok || version != "1.6" {
			t.Errorf("Expected module version '1.6', got %q", version)
		}
	}
}
