package maven

import (
	"bytes"
	"context"
	"os/exec"
)

// RunMaven invokes Maven with the given arguments from projectRoot.
// Returns collected output as a single string, or error.
// Always adds '-B', '--no-transfer-progress', '-Dstyle.color=never' at the start.
var RunMaven = runMavenImpl

func runMavenImpl(ctx context.Context, projectRoot string, args []string, opts MavenOpts) (string, error) {
	var cmdArgs []string

	// Only prepend clean if not already in args and NoClean is false
	hasClean := false
	for _, arg := range args {
		if arg == "clean" {
			hasClean = true
			break
		}
	}
	if !opts.NoClean && !hasClean {
		cmdArgs = append(cmdArgs, "clean")
	}

	cmdArgs = append(cmdArgs, args...)

	if opts.ResumeFrom != "" {
		cmdArgs = append(cmdArgs, "-rf", opts.ResumeFrom)
	}

	cmdArgs = append(cmdArgs, "-B", "--no-transfer-progress", "-Dstyle.color=never")
	cmd := exec.CommandContext(ctx, "mvn", cmdArgs...)
	cmd.Dir = projectRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}
