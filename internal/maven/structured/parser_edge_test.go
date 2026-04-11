package structured

import (
	"strings"
	"testing"
)

// Test cases for parser edge cases
var parserEdgeCases = []struct {
	name     string
	input    string
	wantType []string // expected child types in order
	wantErr  bool
}{
	{
		name:     "empty input",
		input:    "",
		wantType: []string{},
	},
	{
		name:     "single info line",
		input:    "[INFO] just one line",
		wantType: []string{"unparsable"},
	},
	{
		name: "initialization only",
		input: `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]`,
		wantType: []string{"initialization"},
	},
	{
		name: "module only",
		input: `[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT
[INFO] --- compiler:3.1.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file`,
		wantType: []string{"module"},
	},
	{
		name: "summary only",
		input: `[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS`,
		wantType: []string{"summary"},
	},
	{
		name: "initialization then module",
		input: `[INFO] Scanning for projects...
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT`,
		wantType: []string{"initialization", "module"},
	},
	{
		name: "module then summary",
		input: `[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]`,
		wantType: []string{"module", "summary"},
	},
	{
		name: "all three: initialization, module, summary",
		input: `[INFO] Scanning for projects...
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] BUILD SUCCESS`,
		wantType: []string{"initialization", "module", "summary"},
	},
	{
		name: "two modules",
		input: `[INFO] ---< com.example:module-a >---
[INFO] Building module-a 1.0-SNAPSHOT
[INFO] ---< com.example:module-b >---
[INFO] Building module-b 1.0-SNAPSHOT`,
		wantType: []string{"module", "module"},
	},
	{
		name: "module with build block",
		input: `[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT
[INFO] --- compiler:3.1.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file`,
		wantType: []string{"module"},
	},
	{
		name: "unparsable before module creates two blocks",
		input: `[INFO] Some random output
[WARN] Something unusual
[DEBUG] Debug message
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT`,
		wantType: []string{"unparsable", "module"},
	},
	{
		name: "summary then initialization",
		input: `[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] BUILD SUCCESS
[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]`,
		wantType: []string{"summary", "initialization"},
	},
}

func TestParserEdgeCases(t *testing.T) {
	for _, tc := range parserEdgeCases {
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(strings.ReplaceAll(tc.input, "\r\n", "\n"), "\n")
			// Filter empty lines that would result from splitting empty string
			if tc.input == "" {
				lines = []string{}
			}

			parsed := NewOutputParser().ParseOutput(lines, nil, ParseConfig{})

			gotTypes := make([]string, len(parsed.Root.Children))
			for i, child := range parsed.Root.Children {
				gotTypes[i] = child.Type
			}

			// Verify we got the expected number of children
			if len(gotTypes) != len(tc.wantType) {
				t.Errorf("got %d children, want %d", len(gotTypes), len(tc.wantType))
				t.Logf("got types: %v", gotTypes)
				t.Logf("want types: %v", tc.wantType)
				return
			}

			// Verify each child type in order
			for i, want := range tc.wantType {
				if gotTypes[i] != want {
					t.Errorf("child %d: got %q, want %q", i, gotTypes[i], want)
				}
			}
		})
	}
}

// Test edge cases specifically for line number tracking
var lineNumberEdgeCases = []struct {
	name       string
	input      string
	wantRanges []struct {
		startLine int
		endLine   int
	}
}{
	{
		name:       "single line",
		input:      "[INFO] test",
		wantRanges: []struct{ startLine, endLine int }{{1, 1}},
	},
	{
		name:       "three lines",
		input:      "[INFO] line 1\n[INFO] line 2\n[INFO] line 3",
		wantRanges: []struct{ startLine, endLine int }{{1, 3}},
	},
	{
		name:       "two blocks - lines split differently than expected",
		input:      "[INFO] A\n[INFO] B\n[INFO] X\n[INFO] Y",
		wantRanges: []struct{ startLine, endLine int }{{1, 4}}, // may be combined
	},
}

func TestLineNumbersEdgeCases(t *testing.T) {
	for _, tc := range lineNumberEdgeCases {
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(strings.ReplaceAll(tc.input, "\r\n", "\n"), "\n")
			if tc.input == "" {
				lines = []string{}
			}

			parsed := NewOutputParser().ParseOutput(lines, nil, ParseConfig{})

			if len(parsed.Root.Children) != len(tc.wantRanges) {
				t.Errorf("got %d children, want %d", len(parsed.Root.Children), len(tc.wantRanges))
				return
			}

			for i, want := range tc.wantRanges {
				got := parsed.Root.Children[i]
				if got.StartLine != want.startLine || got.EndLine != want.endLine {
					t.Errorf("child %d: got (%d-%d), want (%d-%d)",
						i, got.StartLine, got.EndLine, want.startLine, want.endLine)
				}
			}
		})
	}
}
