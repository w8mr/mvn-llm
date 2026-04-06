package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/agentic-ai/mvn-llm/internal/intent"
	"github.com/agentic-ai/mvn-llm/internal/maven"
)

func main() {
	projectRoot := flag.String("project-root", ".", "Project root directory")
	noClean := flag.Bool("no-clean", false, "Skip mvn clean before build")
	resumeFrom := flag.String("rf", "", "Resume build from specified module")
	output := flag.String("o", "text", "Output format: text or json")
	outputFile := flag.String("output-file", "", "Optional file path for JSON output")

	// Deps-specific flags
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

	// Handle deps specially
	if goal == "deps" {
		depsHandler := intent.DepsHandler{
			Filter:       *depFilter,
			ShowAncestor: *depAncestor,
			Verbose:      *depVerbose,
		}
		mvnOut, mvnErr = intent.HandleDeps(ctx, *projectRoot, opts, depsHandler)

		// Print deps output directly
		fmt.Println(mvnOut)
		if mvnErr != nil {
			os.Exit(1)
		}
		return
	}

	// For all other goals, use generalized handler
	mvnOut, mvnErr = intent.HandleMavenGoal(ctx, *projectRoot, goal, opts)

	// If mvnOut is a string (raw output, not a known build/test phase), just print it
	if outStr, ok := mvnOut.(string); ok {
		fmt.Println(outStr)
		if mvnErr != nil {
			os.Exit(1)
		}
		return
	}

	// If mvnOut is a parsed result (MavenOutput) or similar, cast and handle output format
	if result, ok := mvnOut.(maven.MavenOutput); ok {
		if *output == "json" {
			jsonOutput := result.ToJSON()
			if *outputFile != "" {
				os.WriteFile(*outputFile, []byte(jsonOutput), 0644)
			} else {
				fmt.Println(jsonOutput)
			}
		} else {
			summary := result.GetAgentSummary()
			fmt.Println(summary)
		}
		if mvnErr != nil || result.Status == "BUILD FAILURE" {
			os.Exit(1)
		}
		return
	}

	// Fallback: print Go-style value output
	fmt.Printf("%+v\n", mvnOut)
	if mvnErr != nil {
		os.Exit(1)
	}
}
