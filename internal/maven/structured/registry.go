package structured

type Registry struct {
	Parsers []Parser
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) RegisterPhase(p Parser) {
	r.Parsers = append(r.Parsers, p)
}

func (r *Registry) ParseOutput(lines []string) *StructuredOutput {
	root := Node{
		Name:     "maven-build",
		Type:     "root",
		Children: []Node{},
	}

	idx := 0
	for idx < len(lines) {
		matched := false

		for _, parser := range r.Parsers {
			node, consumed, ok := parser.Parse(lines, idx)
			if ok {
				root.Children = append(root.Children, *node)
				idx += consumed
				matched = true
				break
			}
		}

		if !matched {
			root.Children = append(root.Children, Node{
				Name:  "unparsable",
				Type:  "unparsable",
				Lines: []string{lines[idx]},
			})
			idx++
		}
	}

	return &StructuredOutput{Root: root}
}

func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.RegisterPhase(&InitializationPhaseParser{})
	r.RegisterPhase(&ModulePhaseParser{
		SubParsers: []Parser{
			&BuildPhaseParser{
				SubParsers: nil,
			},
		},
	})
	r.RegisterPhase(&SummaryPhaseParser{})
	return r
}
