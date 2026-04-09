package intent

import (
	"context"

	"github.com/agentic-ai/mvn-llm/internal/maven"
)

// HandleMavenGoal runs any Maven goal and returns the output
// It automatically parses the output based on the goal type
func HandleMavenGoal(ctx context.Context, projectRoot string, goal string, opts maven.MavenOpts) (interface{}, error) {
	return HandleMavenPhaseIntent(ctx, projectRoot, goal, opts)
}

func isBuildGoal(goal string) bool {
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
	return buildGoals[goal]
}
