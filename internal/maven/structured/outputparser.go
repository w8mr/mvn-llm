package structured

import (
	"reflect"

	"github.com/agentic-ai/mvn-llm/internal/errors"
)

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

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

// getValidTypesUpChain returns all valid child types for the current node and all its ancestors.
// This allows parsers to match at any valid level in the hierarchy.
func getValidTypesUpChain(node *Node) []string {
	var result []string
	seen := make(map[string]bool)
	for current := node; current != nil; current = current.Parent {
		for _, t := range AcceptanceMap[current.Type] {
			if !seen[t] {
				seen[t] = true
				result = append(result, t)
			}
		}
	}
	return result
}

// bubbleUpAndInsert attempts to insert a node by first checking the current level,
// then bubbling up the parent chain to find a valid insertion point.
// Returns true if insertion succeeded, false if no valid parent was found.
// After insertion, if the node accepts children, the insertion point moves into it.
func (p *OutputParser) bubbleUpAndInsert(root *Node, node Node) bool {
	// Try current level, then bubble up through parent chain
	for p.currentInsertionNode != nil {
		if CanInsert(p.currentInsertionNode.Type, node.Type) {
			node.Parent = p.currentInsertionNode
			p.currentInsertionNode.Children = append(p.currentInsertionNode.Children, node)
			// If inserted node accepts children, move into it; otherwise stay at parent
			if len(AcceptanceMap[node.Type]) > 0 {
				p.currentInsertionNode = &p.currentInsertionNode.Children[len(p.currentInsertionNode.Children)-1]
			}
			return true
		}
		// Move to parent (loop will exit when we reach root with no parent)
		p.currentInsertionNode = p.currentInsertionNode.Parent
	}

	// No valid parent found - return false to let caller handle
	return false
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

		// Get valid child types for current level AND all ancestor levels
		validTypes := getValidTypesUpChain(p.currentInsertionNode)

		// Only try parsers that can produce valid node types for current or ancestor levels
		for _, parser := range p.Parsers {
			if contains(validTypes, parser.NodeType()) {
				node, consumed, ok := parser.Parse(lines, idx)
				if ok {
					if !p.bubbleUpAndInsert(&root, *node) {
						errors.FatalWithMavenLog(lines, "Parser could not find valid insertion point for node type %q", node.Type)
					}
					idx += consumed
					matched = true
					break
				}
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
			// New unparsable node - bubble up if needed
			unparsable := Node{
				Name:  "unparsable",
				Type:  "unparsable",
				Lines: []string{lines[idx]},
			}
			if !p.bubbleUpAndInsert(&root, unparsable) {
				errors.FatalWithMavenLog(lines, "Parser could not find valid insertion point for node type %q", unparsable.Type)
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
		if !LinesMatch(lines, collected) {
			errors.FatalWithMavenLog(lines, "Parsing may have lost lines. Original: %d, Parsed: %d", len(lines), len(collected))
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

// LinesMatch compares two slices of strings for exact equality.
// Used to verify parsing preserves all original lines.
func LinesMatch(original, parsed []string) bool {
	return reflect.DeepEqual(original, parsed)
}
