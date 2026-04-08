package errors

import (
	"fmt"
	"os"
)

const issueURL = "https://github.com/w8mr/mvn-llm/issues"

// Fatal prints an error message and exits with code 1.
// It does NOT include instructions to report an issue.
func Fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

// FatalWithIssue prints an error message and exits with code 1.
// It includes instructions to report the issue on GitHub.
func FatalWithIssue(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	fmt.Fprintf(os.Stderr, "\nPlease report this issue at: %s\n", issueURL)
	fmt.Fprintf(os.Stderr, "Include the Maven log output that triggered this error.\n")
	os.Exit(1)
}
