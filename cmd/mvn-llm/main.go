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
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: mvn-llm <install|test|deps> [flags]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	intentArg := args[0]
	validIntents := map[string]bool{"install": true, "test": true, "deps": true}
	if !validIntents[intentArg] {
		fmt.Fprintf(os.Stderr, "Invalid intent: %s. Valid intents: install, test, deps\n", intentArg)
		os.Exit(2)
	}

	ctx := context.Background()
	var mvnOut string
	var mvnErr error

	opts := maven.MavenOpts{
		NoClean:    *noClean,
		ResumeFrom: *resumeFrom,
	}

	switch intentArg {
	case "install":
		mvnOut, mvnErr = intent.HandleInstall(ctx, *projectRoot, opts)
	case "test":
		mvnOut, mvnErr = intent.HandleTest(ctx, *projectRoot, opts)
	case "deps":
		mvnOut, mvnErr = intent.HandleDeps(ctx, *projectRoot, opts)
	}

	result := maven.ParseMavenOutput(mvnOut, *projectRoot)

	if *output == "json" {
		jsonOutput := result.ToJSON()
		if *outputFile != "" {
			os.WriteFile(*outputFile, []byte(jsonOutput), 0644)
		} else {
			fmt.Println(jsonOutput)
		}
	} else {
		// Clean single-line summary for agents
		summary := result.GetAgentSummary()
		fmt.Println(summary)
	}

	if mvnErr != nil || result.Status == "BUILD FAILURE" {
		os.Exit(1)
	}
}
