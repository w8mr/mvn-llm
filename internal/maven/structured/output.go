package structured

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
// Uses the pre-enriched summary node's meta.
func TextSummary(out *StructuredOutput) string {
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
