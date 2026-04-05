package intent

import (
	"context"
	"github.com/agentic-ai/mvn-llm/internal/maven"
)

// HandleInstall runs 'mvn install' in the given projectRoot and returns output/error.
func HandleInstall(ctx context.Context, projectRoot string, opts maven.MavenOpts) (string, error) {
	return maven.RunMaven(ctx, projectRoot, []string{"install"}, opts)
}
