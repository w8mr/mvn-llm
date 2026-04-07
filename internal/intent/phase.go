package intent

import (
	"context"
	"github.com/agentic-ai/mvn-llm/internal/maven"
)

// HandleMavenPhaseIntent runs any Maven phase/goal, parses output if recognized, otherwise returns raw output.
// Returns a structured result for known phases: dependency:tree, build, test. Otherwise returns raw output (as string).
// parseMode: "structured-json" only
func HandleMavenPhaseIntent(ctx context.Context, projectRoot string, phase string, opts maven.MavenOpts, parseMode ...string) (interface{}, error) {
	goal := phase

	args := []string{goal}

	// Special case verbose for dependency tree
	if goal == "dependency:tree" {
		args = append(args, "-Dverbose=true")
	}

	output, err := maven.RunMaven(ctx, projectRoot, args, opts)

	// Handle dependency:tree with special tree parsing
	if goal == "dependency:tree" {
		parsed := maven.ParseDependencyTree(output)
		return parsed, err
	}

	// Recognized build/test goals
	buildGoals := map[string]bool{
		"clean":        true,
		"validate":     true,
		"compile":      true,
		"test":         true,
		"test-compile": true,
		"package":      true,
		"verify":       true,
		"install":      true,
		"deploy":       true,
		"site":         true,
	}
	if buildGoals[goal] {
		return output, err // always return raw output for structured handling
	}

	// Unrecognized phase -- just return the output as raw string
	return output, err
}
