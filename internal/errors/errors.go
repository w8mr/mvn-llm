package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// FatalWithMavenLog saves Maven output to a temp file and exits with error message.
// The temp file is created once with a single timestamp, and the path is included in the error message.
func FatalWithMavenLog(lines []string, format string, args ...any) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("mvn-llm_error_%s.log", timestamp)
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	// Write Maven output to temp file
	content := ""
	for _, line := range lines {
		content += line + "\n"
	}
	err := os.WriteFile(filePath, []byte(content), 0644)

	// Print error message
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)

	if err == nil {
		fmt.Fprintf(os.Stderr, "\nMaven output saved to: %s\n", filePath)
		fmt.Fprintf(os.Stderr, "\nPlease report this issue at: %s\n", issueURL)
		fmt.Fprintf(os.Stderr, "Attach the log file: %s\n", filePath)
	} else {
		fmt.Fprintf(os.Stderr, "\nPlease report this issue at: %s\n", issueURL)
		fmt.Fprintf(os.Stderr, "Include the Maven log output that triggered this error.\n")
	}

	os.Exit(1)
}
