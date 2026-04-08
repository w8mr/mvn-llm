package structured

type Node struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Lines    []string       `json:"lines,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []Node         `json:"children,omitempty"`
	Parent   *Node          `json:"-"`
}

type StructuredOutput struct {
	Root Node `json:"root"`
}

// AcceptanceMap defines what child node types each parent node type accepts
var AcceptanceMap = map[string][]string{
	"root":           {"initialization", "module", "summary", "unparsable"},
	"module":         {"build-block", "unparsable"},
	"build-block":    {},
	"summary":        {},
	"initialization": {},
}

// CanInsert checks if a child node type can be inserted into a parent node type
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
