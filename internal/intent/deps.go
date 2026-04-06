package intent

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentic-ai/mvn-llm/internal/maven"
)

type DepsHandler struct {
	Filter       string
	ShowAncestor string
	Verbose      bool
}

func HandleDeps(ctx context.Context, projectRoot string, opts maven.MavenOpts, handler DepsHandler) (string, error) {
	var args []string

	if handler.Verbose {
		args = append(args, "dependency:tree", "-Dverbose=true")
	} else {
		args = append(args, "dependency:tree")
	}

	if handler.Filter != "" {
		args = append(args, "-Dincludes="+handler.Filter)
	}

	output, err := maven.RunMaven(ctx, projectRoot, args, opts)
	if err != nil {
		return output, err
	}

	deps := maven.ParseDependencyTree(output)

	// Debug tree structure
	for _, tree := range deps.Modules {
		fmt.Printf("DEBUG: Module %s has %d root children\n", tree.ModuleName, len(tree.Root.Children))
		for i, dep := range tree.Root.Children {
			fmt.Printf("DEBUG:   Root dep[%d]: %s:%s (children: %d)\n", i, dep.GroupID, dep.ArtifactID, len(dep.Children))
		}
	}

	if handler.ShowAncestor != "" {
		return deps.FormatAncestors(handler.ShowAncestor), nil
	}

	return deps.FormatTree(), nil
}

func formatAncestorResult(deps maven.DepsOutput, dependency string) string {
	ancestors := deps.GetAncestors(dependency)
	if ancestors == nil {
		parts := strings.Split(dependency, ":")
		if len(parts) >= 2 {
			ancestors = deps.GetAncestors(parts[0] + ":" + parts[1])
		}
	}

	if ancestors == nil || len(ancestors) == 0 {
		return ""
	}

	return deps.FormatAncestors(dependency)
}
