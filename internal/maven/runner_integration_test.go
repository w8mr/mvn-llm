package maven

import (
	"context"
	"strings"
	"testing"
)

func TestRunMavenImpl_E2E(t *testing.T) {
	projectRoot := "../../testdata/sample-e2e"
	ctx := context.Background()
	t.Run("no duplicate clean with explicit clean", func(t *testing.T) {
		out, err := runMavenImpl(ctx, projectRoot, []string{"clean", "validate"}, MavenOpts{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !(strings.Contains(out, "--- maven-clean-plugin:") || strings.Contains(out, "--- clean:")) {
			t.Errorf("expected a clean plugin block in output, got: %s", out)
		}
		if strings.Count(out, "--- maven-clean-plugin:") > 1 {
			t.Errorf("should only see one clean plugin block")
		}
		if strings.Contains(out, "clean clean") {
			t.Errorf("should not have duplicate 'clean' in command")
		}
	})

	t.Run("implicit clean is added", func(t *testing.T) {
		out, err := runMavenImpl(ctx, projectRoot, []string{"validate"}, MavenOpts{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !(strings.Contains(out, "--- maven-clean-plugin:") || strings.Contains(out, "--- clean:")) {
			t.Errorf("expected a clean plugin block in output, got: %s", out)
		}
		if strings.Contains(out, "clean clean") {
			t.Errorf("should not have duplicate 'clean' in command")
		}
	})

	t.Run("NoClean: true disables clean", func(t *testing.T) {
		out, err := runMavenImpl(ctx, projectRoot, []string{"validate"}, MavenOpts{NoClean: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(out, "--- maven-clean-plugin:") || strings.Contains(out, "--- clean:") {
			t.Errorf("did not expect clean plugin block, got: %s", out)
		}
	})
}
