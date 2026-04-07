package structured

type Node struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Lines    []string       `json:"lines,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []Node         `json:"children,omitempty"`
}

type StructuredOutput struct {
	Root Node `json:"root"`
}
