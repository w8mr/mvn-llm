package structured

import (
	"fmt"
	"os"
	"reflect"
)

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

func (p *OutputParser) canAnyParserMatch(lines []string, idx int) bool {
	for _, parser := range p.Parsers {
		if p.canParserMatch(parser, lines, idx) {
			return true
		}
	}
	return false
}

func (p *OutputParser) canParserMatch(parser Parser, lines []string, idx int) bool {
	if _, _, ok := parser.Parse(lines, idx); ok {
		return true
	}
	if mp, ok := parser.(*ModulePhaseParser); ok {
		for _, sub := range mp.SubParsers {
			if p.canParserMatch(sub, lines, idx) {
				return true
			}
		}
	}
	if bp, ok := parser.(*BuildPhaseParser); ok {
		for _, sub := range bp.SubParsers {
			if p.canParserMatch(sub, lines, idx) {
				return true
			}
		}
	}
	return false
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
			// Before adding to unparsable, check if any parser CAN match at current position
			// If a parser can match, skip creating unparsable - it will match on next iteration
			canMatch := p.canAnyParserMatch(lines, idx)

			if !canMatch {
				// Check if previous node was also unparsable - combine if adjacent
				if len(root.Children) > 0 {
					lastIdx := len(root.Children) - 1
					if root.Children[lastIdx].Type == "unparsable" {
						root.Children[lastIdx].Lines = append(root.Children[lastIdx].Lines, lines[idx])
						idx++
						continue
					}
				}
				// New unparsable node
				root.Children = append(root.Children, Node{
					Name:  "unparsable",
					Type:  "unparsable",
					Lines: []string{lines[idx]},
				})
				idx++
			} else {
				// A parser can match here - skip this line, let parser handle it
				idx++
			}
		}
	}

	return &root, len(lines), true
}

func (p *OutputParser) ParseOutput(lines []string) *StructuredOutput {
	return p.ParseOutputStrict(lines, false)
}

func (p *OutputParser) ParseOutputStrict(lines []string, strict bool) *StructuredOutput {
	root, _, ok := p.Parse(lines, 0)

	if strict && ok {
		collected := collectAllLines(root)
		missing := findMissingLines(lines, collected)
		extra := findExtraLines(lines, collected)

		if len(missing) > 0 || len(extra) > 0 {
			fmt.Fprintf(os.Stderr, "ERROR: Parsing may have lost lines.\n")
			fmt.Fprintf(os.Stderr, "  Original lines: %d\n", len(lines))
			fmt.Fprintf(os.Stderr, "  Parsed lines: %d\n", len(collected))

			if len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "  Missing: %d lines\n", len(missing))
				for i, line := range missing {
					if i >= 3 {
						break
					}
					fmt.Fprintf(os.Stderr, "    %d: %s\n", i+1, line)
				}
			}

			if len(extra) > 0 {
				fmt.Fprintf(os.Stderr, "  Extra (unparsed): %d lines\n", len(extra))
				for i, line := range extra {
					if i >= 3 {
						break
					}
					fmt.Fprintf(os.Stderr, "    %d: %s\n", i+1, line)
				}
			}

			fmt.Fprintf(os.Stderr, "\nPlease report this issue at: https://github.com/anomalyco/maven-tool/issues\n")
			fmt.Fprintf(os.Stderr, "To disable strict mode: mvn-llm --no-strict ...\n")
			os.Exit(1)
		}
	}

	return &StructuredOutput{Root: *root}
}

func collectAllLines(node *Node) []string {
	var lines []string
	lines = append(lines, node.Lines...)
	for _, child := range node.Children {
		lines = append(lines, collectAllLines(&child)...)
	}
	return lines
}

func findMissingLines(original, parsed []string) []string {
	parsedSet := make(map[string]bool)
	for _, line := range parsed {
		parsedSet[line] = true
	}

	var missing []string
	for _, line := range original {
		if !parsedSet[line] {
			missing = append(missing, line)
		}
	}
	return missing
}

func findExtraLines(original, parsed []string) []string {
	originalSet := make(map[string]bool)
	for _, line := range original {
		originalSet[line] = true
	}

	var extra []string
	for _, line := range parsed {
		if !originalSet[line] {
			extra = append(extra, line)
		}
	}
	return extra
}

func LinesMatch(original, parsed []string) bool {
	return reflect.DeepEqual(original, parsed)
}
