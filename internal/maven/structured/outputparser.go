package structured

type OutputParser struct {
	Parsers []Parser
}

func NewOutputParser() *OutputParser {
	return &OutputParser{
		Parsers: []Parser{
			&InitializationPhaseParser{},
			&ModulePhaseParser{
				SubParsers: []Parser{
					&BuildPhaseParser{},
				},
			},
			&SummaryPhaseParser{},
		},
	}
}

func (p *OutputParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx != 0 {
		return nil, 0, false
	}

	root := Node{
		Name:     "maven-build",
		Type:     "root",
		Children: []Node{},
	}

	idx := 0
	for idx < len(lines) {
		matched := false

		for _, parser := range p.Parsers {
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

	return &root, len(lines), true
}

func (p *OutputParser) ParseOutput(lines []string) *StructuredOutput {
	root, _, _ := p.Parse(lines, 0)
	return &StructuredOutput{Root: *root}
}
