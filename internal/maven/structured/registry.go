package structured

import "strings"

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
	r.RegisterPhase(&ModulePhaseParser{})
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
	var currentModule *PhaseOutput = nil
	for idx < len(lines) {
		matched := false
		for _, phase := range r.Phases {
			block, consumed, ok := phase.Parse(lines, idx)
			if ok {
				if blockOut, yes := block.(BlockOutput); yes {
					switch blockOut.Type {
					case "module-header":
						// Extract module name from the module-header block ("Building <modulename>")
						if currentModule != nil {
							output.Phases = append(output.Phases, *currentModule)
						}
						moduleName := "module"
						for _, line := range blockOut.Lines {
							if strings.HasPrefix(line, "[INFO] Building ") {
								parts := strings.SplitN(line[len("[INFO] Building "):], " ", 2)
								if len(parts) > 0 && parts[0] != "" {
									moduleName = parts[0]
								}
								break
							}
						}
						currentModule = &PhaseOutput{Name: moduleName, Blocks: []BlockOutput{blockOut}}
					case "build-block":
						if currentModule != nil {
							currentModule.Blocks = append(currentModule.Blocks, blockOut)
						} // else: orphan block, ignore (do not emit at top-level)

					case "summary", "initialization":
						if currentModule != nil {
							output.Phases = append(output.Phases, *currentModule)
							currentModule = nil
						}
						output.Phases = append(output.Phases, PhaseOutput{Name: blockOut.Type, Blocks: []BlockOutput{blockOut}})
					default:
						output.Phases = append(output.Phases, PhaseOutput{Name: blockOut.Type, Blocks: []BlockOutput{blockOut}})
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
			output.Phases = append(output.Phases, PhaseOutput{Name: "unparsable", Blocks: []BlockOutput{block}})
			idx++
		}
	}
	// If still inside a module at end, flush it
	if currentModule != nil {
		output.Phases = append(output.Phases, *currentModule)
	}
	return output
}
