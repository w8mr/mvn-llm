package structured

import (
	"fmt"
	"os"
	"reflect"
)

// OutputParser coordinates multiple phase parsers to build a hierarchical tree structure.
// It maintains insertion-point tracking to ensure nodes are placed in valid parent positions
// according to the AcceptanceMap rules.
type OutputParser struct {
	Parsers              []Parser
	currentInsertionNode *Node
}

// NewOutputParser creates a new OutputParser with all available phase parsers.
// The parser list is flat (all at root level); hierarchy is maintained via insertion-point tracking.
func NewOutputParser() *OutputParser {
	return &OutputParser{
		Parsers: []Parser{
			&InitializationPhaseParser{},
			&ModulePhaseParser{},
			&BuildPhaseParser{},
			&SummaryPhaseParser{},
		},
		currentInsertionNode: nil,
	}
}

// canInsert checks whether the current insertion node can accept a child of the given type.
// Returns false if currentInsertionNode is nil or the type is not accepted.
func (p *OutputParser) canInsert(nodeType string) bool {
	if p.currentInsertionNode == nil {
		return false
	}
	return CanInsert(p.currentInsertionNode.Type, nodeType)
}

// insertNode appends a node as a child of the current insertion point.
// If the node accepts children (has entries in AcceptanceMap), the insertion point moves into it.
// Exception: unparsable nodes never change the insertion point to prevent breaking the hierarchy.
func (p *OutputParser) insertNode(node Node) {
	if p.currentInsertionNode != nil {
		p.currentInsertionNode.Children = append(p.currentInsertionNode.Children, node)
		// If inserted node accepts children AND is not unparsable, move into it
		// Unparsable nodes should NOT change the insertion point
		if node.Type != "unparsable" && len(AcceptanceMap[node.Type]) > 0 {
			p.currentInsertionNode = &p.currentInsertionNode.Children[len(p.currentInsertionNode.Children)-1]
		}
	}
}

// bubbleUpAndInsert attempts to insert a node by first checking the current level,
// then bubbling up the parent chain to find a valid insertion point.
// If no valid parent is found in the chain, the node is inserted at root level.
// After insertion, if the node accepts children, the insertion point moves into it.
func (p *OutputParser) bubbleUpAndInsert(root *Node, node Node) {
	// Try current level first
	if p.currentInsertionNode != nil && CanInsert(p.currentInsertionNode.Type, node.Type) {
		node.Parent = p.currentInsertionNode
		p.currentInsertionNode.Children = append(p.currentInsertionNode.Children, node)
		// If inserted node accepts children, move into it; otherwise stay at parent
		if len(AcceptanceMap[node.Type]) > 0 {
			p.currentInsertionNode = &p.currentInsertionNode.Children[len(p.currentInsertionNode.Children)-1]
		}
		return
	}

	// Bubble up via parent chain
	for p.currentInsertionNode != nil && p.currentInsertionNode.Parent != nil {
		p.currentInsertionNode = p.currentInsertionNode.Parent
		if CanInsert(p.currentInsertionNode.Type, node.Type) {
			node.Parent = p.currentInsertionNode
			p.currentInsertionNode.Children = append(p.currentInsertionNode.Children, node)
			// If inserted node accepts children, move into it; otherwise stay at parent
			if len(AcceptanceMap[node.Type]) > 0 {
				p.currentInsertionNode = &p.currentInsertionNode.Children[len(p.currentInsertionNode.Children)-1]
			}
			return
		}
	}

	// Insert at root
	node.Parent = nil
	root.Children = append(root.Children, node)
	p.currentInsertionNode = &root.Children[len(root.Children)-1]
}

// Parse parses Maven log output lines into a hierarchical Node tree.
// It iterates through lines, trying each parser in order. When a parser matches,
// it attempts to insert at the current level. If insertion fails (parent doesn't accept
// the node type), it bubbles up to find a valid parent. Unparsable lines are combined
// into single nodes when consecutive.
func (p *OutputParser) Parse(lines []string, startIdx int) (*Node, int, bool) {
	if startIdx != 0 {
		return nil, 0, false
	}

	root := Node{
		Name:     "maven-build",
		Type:     "root",
		Children: []Node{},
		Parent:   nil,
	}
	p.currentInsertionNode = &root

	idx := 0
	for idx < len(lines) {
		matched := false

		// Try all parsers
		for _, parser := range p.Parsers {
			node, consumed, ok := parser.Parse(lines, idx)
			if ok {
				// Try to insert at current level
				if p.canInsert(node.Type) {
					p.insertNode(*node)
				} else {
					// Cannot insert at current - bubble up to find a valid parent
					p.bubbleUpAndInsert(&root, *node)
				}
				idx += consumed
				matched = true
				break
			}
		}

		if !matched {
			// Check if previous node was also unparsable - combine if adjacent
			if p.currentInsertionNode != nil && len(p.currentInsertionNode.Children) > 0 {
				lastIdx := len(p.currentInsertionNode.Children) - 1
				if p.currentInsertionNode.Children[lastIdx].Type == "unparsable" {
					p.currentInsertionNode.Children[lastIdx].Lines = append(
						p.currentInsertionNode.Children[lastIdx].Lines, lines[idx])
					idx++
					continue
				}
			}
			// New unparsable node - try to insert, bubble if needed
			unparsable := Node{
				Name:  "unparsable",
				Type:  "unparsable",
				Lines: []string{lines[idx]},
			}
			if p.canInsert("unparsable") {
				p.insertNode(unparsable)
			} else {
				p.bubbleUpAndInsert(&root, unparsable)
			}
			idx++
		}
	}

	return &root, len(lines), true
}

// ParseOutput parses Maven log lines into a StructuredOutput with a hierarchical Node tree.
// This is the main entry point for parsing Maven output.
func (p *OutputParser) ParseOutput(lines []string) *StructuredOutput {
	return p.ParseOutputStrict(lines, false)
}

// ParseOutputStrict parses Maven log lines and optionally verifies no lines were lost.
// If strict=true, it compares original and parsed lines; on mismatch, it prints an error
// and exits. Use strict mode for debugging parsing issues.
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

// collectAllLines recursively collects all lines from a node and its children.
func collectAllLines(node *Node) []string {
	var lines []string
	lines = append(lines, node.Lines...)
	for _, child := range node.Children {
		lines = append(lines, collectAllLines(&child)...)
	}
	return lines
}

// findMissingLines returns lines from the original input that don't appear in the parsed output.
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

// findExtraLines returns lines in the parsed output that don't appear in the original input.
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

// LinesMatch compares two slices of strings for exact equality.
// Used to verify parsing preserves all original lines.
func LinesMatch(original, parsed []string) bool {
	return reflect.DeepEqual(original, parsed)
}
