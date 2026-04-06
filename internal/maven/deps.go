package maven

import (
	"regexp"
	"strings"
)

// Dependency represents a single dependency node
type Dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
	Children   []Dependency
	Optional   bool
}

// DependencyTree represents the parsed dependency tree for a module
type DependencyTree struct {
	ModuleName string
	Artifact   string
	Root       Dependency
	ByArtifact map[string][]Dependency
	ByGroup    map[string][]Dependency
}

// DepsOutput represents the structured output from dependency analysis
type DepsOutput struct {
	Modules   map[string]*DependencyTree
	FullTree  string
	Ancestors map[string][]string
}

// ParseDependencyTree parses the output of 'mvn dependency:tree'
func ParseDependencyTree(output string) DepsOutput {
	result := DepsOutput{
		Modules:   make(map[string]*DependencyTree),
		Ancestors: make(map[string][]string),
	}

	lines := strings.Split(output, "\n")
	var currentModule string
	var currentTree *DependencyTree
	re := regexp.MustCompile(`-+< (.+:[^>]+) >-+`)
	re2 := regexp.MustCompile(`@ ([\w\-.]+) +---$`)

	var stack []*Dependency // stack to maintain hierarchy; root above all

	rootSet := false
	for _, line := range lines {
		line = strings.TrimPrefix(line, "[INFO] ")
		line = strings.TrimPrefix(line, "[INFO]")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if m := re.FindStringSubmatch(line); len(m) >= 2 {
			currentModule = m[1]
			currentTree = &DependencyTree{
				ModuleName: currentModule,
				ByArtifact: make(map[string][]Dependency),
				ByGroup:    make(map[string][]Dependency),
			}
			currentTree.Root = Dependency{}
			rootSet = false
			result.Modules[currentModule] = currentTree
			stack = stack[:0]
			stack = append(stack, &currentTree.Root)
			continue
		}

		// SUPPORT: also detect mvn-dependency-plugin header ("@ module-a ---")
		if m := re2.FindStringSubmatch(line); len(m) >= 2 {
			currentModule = m[1]
			currentTree = &DependencyTree{
				ModuleName: currentModule,
				ByArtifact: make(map[string][]Dependency),
				ByGroup:    make(map[string][]Dependency),
			}
			currentTree.Root = Dependency{}
			rootSet = false
			result.Modules[currentModule] = currentTree
			stack = stack[:0]
			stack = append(stack, &currentTree.Root)
			continue
		}

		if currentTree == nil {
			continue
		}

		if strings.HasPrefix(line, "+- ") || strings.HasPrefix(line, "\\- ") || strings.HasPrefix(line, "| ") {
			depth := calcDepth(line)
			// Adjust stack to correct depth
			if depth+1 <= len(stack) {
				stack = stack[:depth+1] // +1 because stack[0]=root
			} else {
				for len(stack) < depth+1 {
					phantom := &Dependency{}
					stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, *phantom)
					stack = append(stack, phantom)
				}
			}
			dep := parseDep(line)
			if dep.ArtifactID != "" {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, dep)
				stack = append(stack, &parent.Children[len(parent.Children)-1])
				coords := depCoord(dep)
				currentTree.ByArtifact[coords] = append(currentTree.ByArtifact[coords], dep)
				currentTree.ByGroup[dep.GroupID] = append(currentTree.ByGroup[dep.GroupID], dep)
			}
		} else if !strings.HasPrefix(line, "[") && !rootSet {
			// This is the root artifact line -- fill in only, do not append as child, and do this only once
			dep := parseDep(line)
			if dep.ArtifactID != "" {
				currentTree.Root.GroupID = dep.GroupID
				currentTree.Root.ArtifactID = dep.ArtifactID
				currentTree.Root.Version = dep.Version
				currentTree.Root.Scope = dep.Scope
				rootSet = true
			}
		}
	}

	result.buildAncestorMap()
	return result

}

func calcDepth(line string) int {
	// Maven tree uses 3-character blocks for each level: "|  " or "   "
	// Count how many such blocks exist before the tree connector (+- or \-)
	depth := 0
	i := 0
	for i+2 < len(line) {
		// Check for 3-character indentation block
		block := line[i : i+3]
		if block == "|  " || block == "   " {
			depth++
			i += 3
		} else {
			break
		}
	}
	return depth
}

func addDepToTree(tree *DependencyTree, dep Dependency, depth int) {
	if depth == 0 {
		tree.Root.Children = append(tree.Root.Children, dep)
		return
	}

	// Find the parent at depth-1 by walking the tree
	parent := findParentAtDepth(&tree.Root, depth-1)
	if parent != nil {
		parent.Children = append(parent.Children, dep)
	} else {
		// Fallback: add to root
		tree.Root.Children = append(tree.Root.Children, dep)
	}
}

func findParentAtDepth(root *Dependency, targetDepth int) *Dependency {
	return findParentRecursive(root, targetDepth, 0)
}

func findParentRecursive(node *Dependency, targetDepth, currentDepth int) *Dependency {
	if currentDepth == targetDepth {
		return node
	}
	// Check children in reverse order (most recently added is the likely parent)
	for i := len(node.Children) - 1; i >= 0; i-- {
		result := findParentRecursive(&node.Children[i], targetDepth, currentDepth+1)
		if result != nil {
			return result
		}
	}
	return nil
}

func isDependencyLine(line string) bool {
	return strings.HasPrefix(line, "\\- ") ||
		strings.HasPrefix(line, "+- ") ||
		strings.HasPrefix(line, "| ")
}

func parseDep(line string) Dependency {
	line = strings.TrimSpace(line)
	// Strip tree characters iteratively (for nested deps like "|  \- dep")
	for {
		oldLine := line
		line = strings.TrimPrefix(line, "\\- ")
		line = strings.TrimPrefix(line, "+- ")
		line = strings.TrimPrefix(line, "| ")
		line = strings.TrimSpace(line)
		if line == oldLine {
			break
		}
	}

	dep := Dependency{}

	if strings.HasSuffix(line, ":optional") {
		dep.Optional = true
		line = strings.TrimSuffix(line, ":optional")
	}

	// Maven dependency format: groupId:artifactId:packaging:version:scope
	// or groupId:artifactId:packaging:version
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		dep.GroupID = parts[0]
		dep.ArtifactID = parts[1]
	}
	// parts[2] is packaging (jar, war, etc.) - skip it
	if len(parts) >= 4 && parts[3] != "" {
		dep.Version = parts[3]
	}
	if len(parts) >= 5 && parts[4] != "" {
		dep.Scope = parts[4]
	}

	return dep
}

func depCoord(dep Dependency) string {
	if dep.Version != "" {
		return dep.GroupID + ":" + dep.ArtifactID + ":" + dep.Version
	}
	return dep.GroupID + ":" + dep.ArtifactID
}

func (d *DepsOutput) buildAncestorMap() {
	for _, tree := range d.Modules {
		d.buildAncestors(tree)
	}
}

func (d *DepsOutput) buildAncestors(tree *DependencyTree) {
	var recurse func(deps []Dependency, path []string)
	recurse = func(deps []Dependency, path []string) {
		for _, dep := range deps {
			coords := depCoord(dep)
			if len(path) > 0 {
				d.Ancestors[coords] = path
			}
			recurse(dep.Children, append(path, coords))
		}
	}
	// Start with the root module in the path
	rootCoords := depCoord(tree.Root)
	recurse(tree.Root.Children, []string{rootCoords})
}

func (d *DepsOutput) GetAncestors(dependency string) []string {
	return d.Ancestors[dependency]
}

func (d *DepsOutput) GetDependencyTree(module string) *DependencyTree {
	return d.Modules[module]
}

func (d *DepsOutput) FormatTree() string {
	var sb strings.Builder

	for _, tree := range d.Modules {
		sb.WriteString("\n=== ")
		sb.WriteString(tree.ModuleName)
		sb.WriteString(" ===\n")
		d.formatTreeRecursive(&sb, tree.Root.Children, 0)
	}

	return sb.String()
}

func (d *DepsOutput) formatTreeRecursive(sb *strings.Builder, deps []Dependency, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, dep := range deps {
		coords := dep.GroupID + ":" + dep.ArtifactID
		if dep.Version != "" {
			coords += ":" + dep.Version
		}
		if dep.Scope != "" && dep.Scope != "compile" {
			coords += ":" + dep.Scope
		}
		if dep.Optional {
			coords += " (optional)"
		}
		sb.WriteString(indent)
		sb.WriteString(coords)
		sb.WriteString("\n")
		d.formatTreeRecursive(sb, dep.Children, depth+1)
	}
}

func (d *DepsOutput) FormatAncestors(dependency string) string {
	// Try various key formats
	keysToTry := []string{dependency}
	parts := strings.Split(dependency, ":")
	if len(parts) >= 2 {
		keysToTry = append(keysToTry, parts[0]+":"+parts[1])
	}
	if len(parts) >= 3 {
		keysToTry = append(keysToTry, parts[0]+":"+parts[1]+":"+parts[2])
	}

	var ancestors []string
	for _, key := range keysToTry {
		if a, ok := d.Ancestors[key]; ok {
			ancestors = a
			break
		}
	}

	if ancestors == nil || len(ancestors) == 0 {
		return "No ancestors found for: " + dependency
	}

	var sb strings.Builder
	sb.WriteString("Ancestors of ")
	sb.WriteString(dependency)
	sb.WriteString(":\n")

	for i, ancestor := range ancestors {
		sb.WriteString(strings.Repeat("  ", i+1))
		sb.WriteString("└─ ")
		sb.WriteString(ancestor)
		sb.WriteString("\n")
	}

	return sb.String()
}
