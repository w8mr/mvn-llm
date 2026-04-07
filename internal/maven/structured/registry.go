package structured

// Central dispatcher for structured Maven output parsers.
type Registry struct {
	Phases []Parser
}

func NewRegistry() *Registry {
	return &Registry{}
}

// NewDefaultRegistry initializes the registry with all default phase parsers in order.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.RegisterPhase(&InitializationPhaseParser{})
	r.RegisterPhase(&BuildPhaseParser{})
	r.RegisterPhase(&SummaryPhaseParser{})
	return r
}

func (r *Registry) RegisterPhase(p Parser) {
	r.Phases = append(r.Phases, p)
}

// ParseOutput runs all registered phase parsers in order.
func (r *Registry) ParseOutput(lines []string) *StructuredOutput {
	output := &StructuredOutput{}
	idx := 0
	for idx < len(lines) {
		matched := false
		for _, phase := range r.Phases {
			block, consumed, ok := phase.Parse(lines, idx)
			if ok {
				typeName := "phase"
				if blockOut, yes := block.(BlockOutput); yes {
					typeName = blockOut.Type
					found := false
					for i := range output.Phases {
						if output.Phases[i].Name == typeName {
							output.Phases[i].Blocks = append(output.Phases[i].Blocks, blockOut)
							found = true
							break
						}
					}
					if !found {
						output.Phases = append(output.Phases, PhaseOutput{Name: typeName, Blocks: []BlockOutput{blockOut}})
					}
				}
				idx += consumed
				matched = true
				break
			}
		}
		if !matched {
			// No parser claimed this line: add as unparsable
			block := BlockOutput{
				Type:  "unparsable",
				Lines: []string{lines[idx]},
			}
			found := false
			for i := range output.Phases {
				if output.Phases[i].Name == "unparsable" {
					output.Phases[i].Blocks = append(output.Phases[i].Blocks, block)
					found = true
					break
				}
			}
			if !found {
				output.Phases = append(output.Phases, PhaseOutput{Name: "unparsable", Blocks: []BlockOutput{block}})
			}
			idx++ // move ahead one line
		}
	}
	return output
}
