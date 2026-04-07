package structured

type StructuredOutput struct {
	Phases []PhaseOutput `json:"phases"`
}

type PhaseOutput struct {
	Name   string        `json:"name"`
	Blocks []BlockOutput `json:"blocks"`
}

type BlockOutput struct {
	Type  string         `json:"type"`
	Lines []string       `json:"lines"`
	Meta  map[string]any `json:"meta,omitempty"`
}
