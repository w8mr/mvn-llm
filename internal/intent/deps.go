package intent

import (
	"context"
	"github.com/agentic-ai/mvn-llm/internal/maven"
)

// HandleDeps runs 'mvn dependency:tree' in the given projectRoot and returns output/error.
func HandleDeps(ctx context.Context, projectRoot string, opts maven.MavenOpts) (string, error) {
	return maven.RunMaven(ctx, projectRoot, []string{"dependency:tree"}, opts)
}
