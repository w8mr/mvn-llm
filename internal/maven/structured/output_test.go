package structured

import (
	"testing"
)

// Test helper to create a structured output for testing
func testStructuredOutput(modules []map[string]any) *StructuredOutput {
	root := Node{Type: "root", Name: "maven-build"}
	for _, m := range modules {
		moduleName := m["name"].(string)
		builds := m["builds"].([]map[string]any)

		moduleNode := Node{Type: "module", Name: moduleName}
		for _, b := range builds {
			buildNode := Node{
				Type: "build-block",
				Name: b["plugin"].(string),
				Meta: map[string]any{
					"status":  b["status"],
					"summary": b["summary"],
				},
			}
			moduleNode.Children = append(moduleNode.Children, buildNode)
		}
		root.Children = append(root.Children, moduleNode)
	}
	return &StructuredOutput{Root: root}
}

func TestTextSummary_AllSuccess(t *testing.T) {
	out := testStructuredOutput([]map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "clean", "status": "SUCCESS", "summary": "Successful"},
				{"plugin": "compile", "status": "SUCCESS", "summary": "Successful: Compiled 10 files"},
			},
		},
	})
	result := TextSummary(out)
	expected := "Successful:\nmodule-a: Compiled 10 files"
	if result != expected {
		t.Errorf("AllSuccess:\n%s\nwant:\n%s", result, expected)
	}
}

func TestTextSummary_WithWarnings(t *testing.T) {
	out := testStructuredOutput([]map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "clean", "status": "SUCCESS", "summary": "Successful"},
				{"plugin": "compile", "status": "SUCCESS-WITH-WARNINGS", "summary": "Successful: Some warning"},
			},
		},
	})
	result := TextSummary(out)
	expected := "Successful:\nmodule-a: Some warning"
	if result != expected {
		t.Errorf("WithWarnings:\n%s\nwant:\n%s", result, expected)
	}
}

func TestTextSummary_WithErrors(t *testing.T) {
	out := testStructuredOutput([]map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "FAILED", "summary": "Failure: Compilation failed"},
			},
		},
	})
	result := TextSummary(out)
	expected := "Failure:\nmodule-a: Compilation failed"
	if result != expected {
		t.Errorf("WithErrors:\n%s\nwant:\n%s", result, expected)
	}
}

func TestTextSummary_MixedModules(t *testing.T) {
	out := testStructuredOutput([]map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "FAILED", "summary": "Failure: Error in A"},
			},
		},
		{
			"name": "module-b",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "SUCCESS-WITH-WARNINGS", "summary": "Successful: Warning in B"},
				{"plugin": "test", "status": "SUCCESS", "summary": "Successful: Tests passed"},
			},
		},
	})
	result := TextSummary(out)
	expected := "Failure:\nmodule-a: Error in A\n\nmodule-b: Warning in B"
	if result != expected {
		t.Errorf("Got: %q", result)
	}
}
