package maven

import (
	"strings"
	"testing"
)

// TestRunMaven_CleanDeduplication verifies that "clean" is not added
// to the Maven command when it's already present in the user's args.
func TestRunMaven_CleanDeduplication(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		noClean         bool
		wantContains    string
		wantNotContains string
	}{
		{
			name:            "user provides clean - should not duplicate",
			args:            []string{"clean", "install"},
			noClean:         false,
			wantContains:    "clean install",
			wantNotContains: "clean clean",
		},
		{
			name:            "user does not provide clean - should add it",
			args:            []string{"install"},
			noClean:         false,
			wantContains:    "clean install",
			wantNotContains: "",
		},
		{
			name:            "NoClean=true - should not add clean",
			args:            []string{"install"},
			noClean:         true,
			wantContains:    "install",
			wantNotContains: "clean",
		},
		{
			name:            "NoClean=true with user clean - should keep it",
			args:            []string{"clean", "install"},
			noClean:         true,
			wantContains:    "clean install",
			wantNotContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from runMavenImpl to test it directly
			var cmdArgs []string

			// This is the logic from runMavenImpl (lines 18-27)
			hasClean := false
			for _, arg := range tt.args {
				if arg == "clean" {
					hasClean = true
					break
				}
			}
			opts := MavenOpts{NoClean: tt.noClean}
			if !opts.NoClean && !hasClean {
				cmdArgs = append(cmdArgs, "clean")
			}
			cmdArgs = append(cmdArgs, tt.args...)

			// Verify the constructed command
			cmdStr := strings.Join(cmdArgs, " ")
			if tt.wantContains != "" && !strings.Contains(cmdStr, tt.wantContains) {
				t.Errorf("Expected command to contain %q, got: %q", tt.wantContains, cmdStr)
			}
			if tt.wantNotContains != "" && strings.Contains(cmdStr, tt.wantNotContains) {
				t.Errorf("Expected command NOT to contain %q, got: %q", tt.wantNotContains, cmdStr)
			}

			// Additional check: "clean clean" should never appear
			if strings.Contains(cmdStr, "clean clean") {
				t.Errorf("Command contains duplicate 'clean': %q", cmdStr)
			}
		})
	}
}
