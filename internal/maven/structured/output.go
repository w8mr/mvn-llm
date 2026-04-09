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
// Per module: if WARNING join all warnings, else take last build phase with that status.
func TextSummary(out *StructuredOutput) string {
	var moduleErrs, moduleWarnSucc []string
	overallStatus := "SUCCESS"

	for _, child := range out.Root.Children {
		if child.Type != "module" {
			continue
		}

		moduleName := child.Name
		status, summary := moduleSummary(child.Children, moduleName)

		if status == "FAILED" {
			if overallStatus != "FAILURE" {
				overallStatus = "FAILURE"
			}
			moduleErrs = append(moduleErrs, moduleName+": "+summary)
		} else if status == "SUCCESS-WITH-WARNINGS" {
			if overallStatus == "SUCCESS" {
				overallStatus = "SUCCESS-WITH-WARNINGS"
			}
			moduleWarnSucc = append(moduleWarnSucc, moduleName+": "+summary)
		} else {
			moduleWarnSucc = append(moduleWarnSucc, moduleName+": "+summary)
		}
	}

	var lines []string
	if len(moduleErrs) > 0 && len(moduleWarnSucc) > 0 {
		lines = append(lines, moduleNamePrefix(moduleErrs)...)
		lines = append(lines, "")
		lines = append(lines, moduleNamePrefix(moduleWarnSucc)...)
	} else if len(moduleErrs) > 0 {
		lines = append(lines, moduleNamePrefix(moduleErrs)...)
	} else {
		lines = append(lines, moduleNamePrefix(moduleWarnSucc)...)
	}

	if overallStatus == "FAILURE" {
		return "Failure:\n" + strings.Join(lines, "\n")
	}
	return "Successful:\n" + strings.Join(lines, "\n")
}

// moduleSummary returns status and summary for a module.
// Priority: last with highest status: FAILED > SUCCESS-WITH-WARNINGS > SUCCESS
func moduleSummary(children []Node, moduleName string) (status, summary string) {
	var lastErr, lastWarn, lastSucc string
	for _, child := range children {
		if child.Type != "build-block" {
			continue
		}
		meta := child.Meta
		st, _ := meta["status"].(string)
		sm, _ := meta["summary"].(string)

		// Clean prefix
		if strings.HasPrefix(sm, "Successful: ") {
			sm = sm[12:]
		} else if strings.HasPrefix(sm, "Failure: ") {
			sm = sm[9:]
		}

		if st == "FAILED" {
			lastErr = sm
		} else if st == "SUCCESS-WITH-WARNINGS" {
			lastWarn = sm
		} else {
			lastSucc = sm
		}
	}

	// Priority: FAILED > SUCCESS-WITH-WARNINGS > SUCCESS
	if lastErr != "" {
		return "FAILED", lastErr
	}
	if lastWarn != "" {
		return "SUCCESS-WITH-WARNINGS", lastWarn
	}
	return "SUCCESS", lastSucc
}

func moduleNamePrefix(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, line)
		} else {
			result = append(result, "  "+line)
		}
	}
	return result
}
