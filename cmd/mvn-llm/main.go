package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/agentic-ai/mvn-llm/internal/errors"
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

// safeMain wraps the main logic with panic recovery for unknown errors.
func safeMain() {
	defer func() {
		if r := recover(); r != nil {
			errors.FatalWithIssue("Unexpected error: %v", r)
		}
	}()

	mainLogic()
}

// mainLogic contains the original main function body.
func mainLogic() {
	goal := flag.String("goal", "", "Maven goal (e.g., install, test, compile)")
	projectRoot := flag.String("project-root", ".", "Project root directory")
	noClean := flag.Bool("no-clean", false, "Skip mvn clean before build")
	resumeFrom := flag.String("rf", "", "Resume build from specified module")
	output := flag.String("o", "structured-json", "Output format(s): comma-separated list of text, json, structured-json, maven-output")
	outputFile := flag.String("output-file", "", "Optional file path for JSON output")
	depFilter := flag.String("dep-filter", "", "Filter dependencies (e.g., 'junit')")
	depAncestor := flag.String("dep-ancestor", "", "Show ancestors for this dependency")
	depVerbose := flag.Bool("dep-verbose", false, "Show verbose dependency tree")
	noStrict := flag.Bool("no-strict", false, "Disable strict parsing")
	flag.Parse()

	if *goal == "" {
		fmt.Fprintln(os.Stderr, "Usage: mvn-llm -goal <goal> [flags]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	ctx := context.Background()
	var mvnOut interface{}
	var mvnErr error
	opts := maven.MavenOpts{
		NoClean:    *noClean,
		ResumeFrom: *resumeFrom,
	}
	if *goal == "deps" {
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
	mvnOut, mvnErr = intent.HandleMavenGoal(ctx, *projectRoot, *goal, opts, "structured-json")

	outputTypes := strings.Split(*output, ",")

	for _, outType := range outputTypes {
		outType = strings.TrimSpace(outType)
		if outType == "maven-output" {
			if outStr, ok := mvnOut.(string); ok {
				fmt.Print(outStr)
			}
			if mvnErr != nil {
				fmt.Print(mvnErr)
			}
		}
		if outType == "structured-json" {
			if outStr, ok := mvnOut.(string); ok {
				parser := structured.NewOutputParser()
				structuredOut := parser.ParseOutputStrict(splitLines(outStr), !*noStrict)
				if mvnErr != nil {
					if structuredOut.Root.Meta == nil {
						structuredOut.Root.Meta = make(map[string]any)
					}
					structuredOut.Root.Meta["error"] = mvnErr
				}
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
			}
		}
		if outType == "text" {
			if outStr, ok := mvnOut.(string); ok {
				fmt.Println(outStr)
			}
		}
		if outType == "json" {
			if outStr, ok := mvnOut.(string); ok {
				parser := structured.NewOutputParser()
				structuredOut := parser.ParseOutputStrict(splitLines(outStr), !*noStrict)
				if mvnErr != nil {
					if structuredOut.Root.Meta == nil {
						structuredOut.Root.Meta = make(map[string]any)
					}
					structuredOut.Root.Meta["error"] = mvnErr
				}
				jsonBytes, err := marshalStructuredJSON(structuredOut)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
					os.Exit(1)
				}
				if *outputFile != "" {
					os.WriteFile(*outputFile, jsonBytes, 0644)
				} else {
					os.Stdout.Write(jsonBytes)
				}
			}
		}
	}

	if mvnErr != nil {
		os.Exit(1)
	}
}

// main is now a wrapper that handles panics.
func main() {
	safeMain()
}
