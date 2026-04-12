package structured

import "strings"

type Node struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Lines     []string       `json:"lines,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	Children  []Node         `json:"children,omitempty"`
	Parent    *Node          `json:"-"`
	StartLine int            `json:"startLine,omitempty"`
	EndLine   int            `json:"endLine,omitempty"`
}

// StructuredOutput is the top-level result of parsing Maven output.
type StructuredOutput struct {
	Root Node `json:"root"`
}

// AcceptanceMap defines which child node types each parent node type can contain.
// Empty list means the node type cannot have children.
var AcceptanceMap = map[string][]string{
	"root":            {"initialization", "module", "summary", "unparsable", "dependency-tree"},
	"module":          {"build-block", "surefire-block", "failsafe-block", "compiler-block", "jar-block", "install-block", "deploy-block", "resources-block", "source-block", "clean-block", "war-block", "ear-block", "unparsable"},
	"build-block":     {},
	"summary":         {},
	"initialization":  {},
	"surefire-block":  {},
	"failsafe-block":  {},
	"compiler-block":  {},
	"jar-block":       {},
	"install-block":   {},
	"deploy-block":    {},
	"resources-block": {},
	"source-block":    {},
	"clean-block":     {},
	"war-block":       {},
	"ear-block":       {},
	"dependency-tree": {},
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
// Uses the pre-enriched summary node's meta.
func TextSummary(out *StructuredOutput) string {
	return TextSummaryWithConfig(out, nil)
}

// TextSummaryWithConfig generates a text summary with additional config options.
func TextSummaryWithConfig(out *StructuredOutput, config ParseConfig) string {
	// Check for dependency tree first
	for _, child := range out.Root.Children {
		if child.Type == "dependency-tree" {
			depAncestor, _ := config["depAncestor"].(string)
			return formatDependencyTree(&child, depAncestor)
		}
	}

	// Fall back to summary-based output
	for _, child := range out.Root.Children {
		if child.Type != "summary" {
			continue
		}
		meta := child.Meta
		overallStatus, _ := meta["overallStatus"].(string)
		summary, _ := meta["summary"].(string)

		if overallStatus == "BUILD FAILURE" {
			return "Failure:\n" + summary
		}
		return "Successful:\n" + summary
	}
	return "Successful:\n"
}

func formatDependencyTree(node *Node, filter string) string {
	var lines []string
	meta := node.Meta

	if root, ok := meta["root"].(map[string]any); ok {
		groupID, _ := root["groupId"].(string)
		artifactID, _ := root["artifactId"].(string)
		version, _ := root["version"].(string)
		lines = append(lines, groupID+":"+artifactID+":"+version)
	}

	hasFilter := filter != ""

	if deps, ok := meta["dependencies"].([]map[string]any); ok {
		for _, dep := range deps {
			groupID, _ := dep["groupId"].(string)
			artifactID, _ := dep["artifactId"].(string)
			version, _ := dep["version"].(string)
			scope, _ := dep["scope"].(string)

			// Check if this dependency matches the filter
			if hasFilter {
				matchStr := groupID + ":" + artifactID
				fullMatch := groupID + ":" + artifactID + ":" + version
				if filter != groupID && filter != artifactID && filter != matchStr && filter != fullMatch {
					continue
				}
			}

			line := "+- " + groupID + ":" + artifactID + ":" + version
			if scope != "" {
				line += ":" + scope
			}
			lines = append(lines, line)
		}
	}

	// If filter was specified but no matches found, indicate that
	if hasFilter && len(lines) == 1 {
		return "No matches found for: " + filter
	}

	return strings.Join(lines, "\n")
}
