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

// Helper to split string into lines, trimming trailing empty lines
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
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
	output := flag.String("o", "json-full-with-lines", "Output format(s): comma-separated list of text, json, json-full, json-full-with-lines, maven-output")
	depFilter := flag.String("dep-filter", "", "Filter dependencies (e.g., 'junit')")
	depAncestor := flag.String("dep-ancestor", "", "Show ancestors for this dependency")
	depVerbose := flag.Bool("dep-verbose", false, "Show verbose dependency tree")
	flag.Parse()

	parseConfig := structured.ParseConfig{
		"depFilter":   *depFilter,
		"depAncestor": *depAncestor,
		"depVerbose":  *depVerbose,
	}

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
	mvnOut, mvnErr = intent.HandleMavenGoal(ctx, *projectRoot, *goal, opts)

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
		if outType == "json-full-with-lines" {
			if outStr, ok := mvnOut.(string); ok {
				parser := structured.NewOutputParser()
				structuredOut := parser.ParseOutput(splitLines(outStr), mvnErr, parseConfig)
				jsonBytes, err := marshalStructuredJSON(structuredOut)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode structured JSON: %v\n", err)
					os.Exit(1)
				}
				os.Stdout.Write(jsonBytes)
			}
		}
		if outType == "text" {
			if outStr, ok := mvnOut.(string); ok {
				fmt.Println(outStr)
			}
		}
		if outType == "json" || outType == "json-full" {
			if outStr, ok := mvnOut.(string); ok {
				parser := structured.NewOutputParser()
				structuredOut := parser.ParseOutput(splitLines(outStr), mvnErr, parseConfig)
				if outType == "json-full" {
					structured.StripLines(structuredOut)
				}
				jsonBytes, err := marshalStructuredJSON(structuredOut)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
					os.Exit(1)
				}
				os.Stdout.Write(jsonBytes)
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
