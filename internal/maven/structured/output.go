package structured

import "strings"

// Node represents a parsed section of Maven log output.
// It forms a tree structure where each node may have children.
// Parent is a back-reference for traversal (not serialized).
type Node struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Lines    []string       `json:"lines,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []Node         `json:"children,omitempty"`
	Parent   *Node          `json:"-"`
}

// StructuredOutput is the top-level result of parsing Maven output.
type StructuredOutput struct {
	Root Node `json:"root"`
}

// AcceptanceMap defines which child node types each parent node type can contain.
// Empty list means the node type cannot have children.
var AcceptanceMap = map[string][]string{
	"root":           {"initialization", "module", "summary", "unparsable"},
	"module":         {"build-block", "unparsable"},
	"build-block":    {},
	"summary":        {},
	"initialization": {},
}

// CanInsert checks whether a child node can be inserted as a child of the given parent type.
// It consults AcceptanceMap to determine validity.
func CanInsert(parentType, childType string) bool {
	allowed, ok := AcceptanceMap[parentType]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == childType {
			return true
		}
	}
	return false
}

// TextSummary generates a text summary from the structured output.
// Uses the pre-enriched status and summary from module meta.
func TextSummary(out *StructuredOutput) string {
	var moduleErrs, moduleWarnSucc []string
	overallStatus := "SUCCESS"

	for _, child := range out.Root.Children {
		if child.Type != "module" {
			continue
		}

		moduleName := child.Name
		meta := child.Meta
		status, _ := meta["status"].(string)
		summary, _ := meta["summary"].(string)

		if status == "FAILED" {
			if overallStatus != "FAILURE" {
				overallStatus = "FAILURE"
			}
			moduleErrs = append(moduleErrs, moduleName+":\n"+summary)
		} else if status == "SUCCESS-WITH-WARNINGS" {
			if overallStatus == "SUCCESS" {
				overallStatus = "SUCCESS-WITH-WARNINGS"
			}
			moduleWarnSucc = append(moduleWarnSucc, moduleName+":\n"+summary)
		} else {
			moduleWarnSucc = append(moduleWarnSucc, moduleName+":\n"+summary)
		}
	}

	var lines []string
	if overallStatus == "FAILURE" {
		for i, m := range moduleErrs {
			if i > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, m)
		}
	} else if len(moduleErrs) > 0 && len(moduleWarnSucc) > 0 {
		lines = append(lines, moduleErrs...)
		lines = append(lines, "")
		lines = append(lines, moduleWarnSucc...)
	} else if len(moduleErrs) > 0 {
		lines = append(lines, moduleErrs...)
	} else {
		lines = append(lines, moduleWarnSucc...)
	}

	if overallStatus == "FAILURE" {
		return "Failure:\n" + strings.Join(lines, "\n")
	}
	return "Successful:\n" + strings.Join(lines, "\n")
}
