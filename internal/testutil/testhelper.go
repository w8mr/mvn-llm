package testutil

import (
	"os"
	"path/filepath"
)

// FindRepoRoot walks upward from the current working directory to find go.mod, signifying the repo root.
func FindRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	panic("repo root (go.mod) not found")
}
