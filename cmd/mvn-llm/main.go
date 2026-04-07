package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/agentic-ai/mvn-llm/internal/intent"
	"github.com/agentic-ai/mvn-llm/internal/maven"
	structured "github.com/agentic-ai/mvn-llm/internal/maven/structured"
)

// Helper to split string into lines
func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

// Helper to marshal any struct to pretty JSON
func marshalStructuredJSON(out interface{}) ([]byte, error) {
	return json.MarshalIndent(out, "", "  ")
}

func main() {
	projectRoot := flag.String("project-root", ".", "Project root directory")
	noClean := flag.Bool("no-clean", false, "Skip mvn clean before build")
	resumeFrom := flag.String("rf", "", "Resume build from specified module")
	output := flag.String("o", "structured-json", "Output format: structured-json, text, or json")
	outputFile := flag.String("output-file", "", "Optional file path for JSON output")
	depFilter := flag.String("dep-filter", "", "Filter dependencies (e.g., 'junit')")
	depAncestor := flag.String("dep-ancestor", "", "Show ancestors for this dependency")
	depVerbose := flag.Bool("dep-verbose", false, "Show verbose dependency tree")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: mvn-llm <goal> [flags]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	goal := args[0]
	ctx := context.Background()
	var mvnOut interface{}
	var mvnErr error
	opts := maven.MavenOpts{
		NoClean:    *noClean,
		ResumeFrom: *resumeFrom,
	}
	if goal == "deps" {
		depsHandler := intent.DepsHandler{
			Filter:       *depFilter,
			ShowAncestor: *depAncestor,
			Verbose:      *depVerbose,
		}
		mvnOut, mvnErr = intent.HandleDeps(ctx, *projectRoot, opts, depsHandler)
		fmt.Println(mvnOut)
		if mvnErr != nil {
			os.Exit(1)
		}
		return
	}
	mvnOut, mvnErr = intent.HandleMavenGoal(ctx, *projectRoot, goal, opts, "structured-json")

	// If mvnOut is a string (raw Maven output, best for structured parser)
	if *output == "structured-json" {
		if outStr, ok := mvnOut.(string); ok {
			reg := structured.NewDefaultRegistry()
			structuredOut := reg.ParseOutput(splitLines(outStr))
			jsonBytes, err := marshalStructuredJSON(structuredOut)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode structured JSON: %v\n", err)
				os.Exit(1)
			}
			if *outputFile != "" {
				os.WriteFile(*outputFile, jsonBytes, 0644)
			} else {
				os.Stdout.Write(jsonBytes)
			}
			if mvnErr != nil {
				os.Exit(1)
			}
			return
		}
	}
	if outStr, ok := mvnOut.(string); ok {
		fmt.Println(outStr)
		if mvnErr != nil {
			os.Exit(1)
		}
		return
	}
}
