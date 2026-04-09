package structured

import (
	"testing"
)

// testStructuredOutput creates a structured output for testing TextSummary.
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

// runTextSummaryTest runs TextSummary with the given modules and checks expected output.
func runTextSummaryTest(t *testing.T, name string, modules []map[string]any, expectedSummary string) {
	out := testStructuredOutput(modules)
	result := TextSummary(out)
	if result != expectedSummary {
		t.Errorf("Test %q:\nExpected:\n%s\nGot:\n%s", name, expectedSummary, result)
	}
}

func TestTextSummary_AllSuccess(t *testing.T) {
	modules := []map[string]any{
		{
			"name": "my-app",
			"builds": []map[string]any{
				{"plugin": "clean", "status": "SUCCESS", "summary": "Successful"},
				{"plugin": "compile", "status": "SUCCESS", "summary": "Successful: Compiled 10 files"},
				{"plugin": "jar", "status": "SUCCESS", "summary": "Successful: Building jar"},
			},
		},
	}
	runTextSummaryTest(t, "all success", modules, "Successful:\nmy-app: Building jar")
}

func TestTextSummary_WithWarnings(t *testing.T) {
	modules := []map[string]any{
		{
			"name": "my-app",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "SUCCESS", "summary": "Successful"},
				{"plugin": "compile", "status": "SUCCESS-WITH-WARNINGS", "summary": "Successful: Some warning"},
			},
		},
	}
	runTextSummaryTest(t, "with warnings", modules, "Successful:\nmy-app: Some warning")
}

func TestTextSummary_WithErrors(t *testing.T) {
	modules := []map[string]any{
		{
			"name": "my-app",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "FAILED", "summary": "Failure: Compilation failed"},
			},
		},
	}
	runTextSummaryTest(t, "with errors", modules, "Failure:\nmy-app: Compilation failed")
}

func TestTextSummary_MultipleModulesErrorsFirst(t *testing.T) {
	modules := []map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "SUCCESS", "summary": "Successful: OK"},
			},
		},
		{
			"name": "module-b",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "FAILED", "summary": "Failure: Error in B"},
			},
		},
	}
	// Errors shown first, then warnings/success with empty line between
	runTextSummaryTest(t, "multiple modules errors first", modules, "Failure:\nmodule-b: Error in B\n\nmodule-a: OK")
}

func TestTextSummary_WarningsBeforeErrors(t *testing.T) {
	modules := []map[string]any{
		{
			"name": "module-a",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "SUCCESS-WITH-WARNINGS", "summary": "Successful: Warning"},
			},
		},
		{
			"name": "module-b",
			"builds": []map[string]any{
				{"plugin": "compile", "status": "SUCCESS", "summary": "Successful"},
			},
		},
	}
	// When no errors, shows warnings first (highest status) then success
	runTextSummaryTest(t, "warnings before errors", modules, "Successful:\nmodule-a: Warning\n  module-b: Successful")
}
