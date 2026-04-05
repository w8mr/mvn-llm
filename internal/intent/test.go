package intent

import (
	"context"
	"github.com/agentic-ai/mvn-llm/internal/maven"
)

// HandleTest runs 'mvn test' in the given projectRoot and returns output/error.
func HandleTest(ctx context.Context, projectRoot string, opts maven.MavenOpts) (string, error) {
	return maven.RunMaven(ctx, projectRoot, []string{"test"}, opts)
}
