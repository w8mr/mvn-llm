package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/agentic-ai/mvn-llm/internal/collector"
)

func main() {
	apiKey := flag.String("api-key", "", "Google API key for search")
	storageDir := flag.String("storage-dir", "", "Directory to store examples")
	maxIterations := flag.Int("max-iterations", 10, "Maximum search iterations")
	maxExamples := flag.Int("max-examples", 100, "Maximum examples to collect")

	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Error: -api-key is required")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Starting collector with API key: %s...\n", (*apiKey)[:10]+"...")
	fmt.Printf("Storage directory: %s\n", *storageDir)
	fmt.Printf("Max iterations: %d, Max examples: %d\n", *maxIterations, *maxExamples)

	c := collector.New(*apiKey, *storageDir)
	if err := c.Collect(*maxIterations); err != nil {
		fmt.Printf("Collection error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done! Collected %d examples\n", c.Count())
}
