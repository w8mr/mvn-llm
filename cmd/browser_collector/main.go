package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/agentic-ai/mvn-llm/internal/maven/structured"
	"github.com/chromedp/chromedp"
)

type ParseResult int

const (
	ParseResultNone ParseResult = iota
	ParseResultFull
	ParseResultPartial
)

func main() {
	storageDir := flag.String("storage-dir", "testdata/collector", "Storage directory")
	maxExamples := flag.Int("max-examples", 50, "Maximum examples")

	flag.Parse()

	os.MkdirAll(*storageDir+"/full", 0755)
	os.MkdirAll(*storageDir+"/partial", 0755)
	os.MkdirAll(*storageDir+"/unknown", 0755)

	count := 0

	// Simple direct URLs to Maven CI logs
	urls := []string{
		"https://github.com/awslabs/aws-lambda-java-libs/actions/runs/1234567890",
		"https://github.com/apache/maven/actions/runs/9876543210",
	}

	for _, url := range urls {
		if count >= *maxExamples {
			break
		}

		fmt.Printf("Fetching: %s\n", url)
		content, err := fetchURL(url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		result := checkParse(content)
		if result == ParseResultNone {
			fmt.Printf("Not Maven output\n")
			continue
		}

		category := "unknown"
		switch result {
		case ParseResultFull:
			category = "full"
		case ParseResultPartial:
			category = "partial"
		}

		filepath := path.Join(*storageDir, category, fmt.Sprintf("maven_%d.txt", time.Now().UnixNano()+rand.Int63n(1000000)))
		os.WriteFile(filepath, []byte(content), 0644)

		metaPath := filepath + ".meta"
		os.WriteFile(metaPath, []byte(url), 0644)

		count++
		fmt.Printf("Saved %s (%d)\n", category, count)

		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\nCollected %d examples\n", count)
}

func fetchURL(url string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, []chromedp.ExecAllocatorOption{
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-web-security", "true"),
		chromedp.Headless,
	}...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	var content string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.Text("body", &content),
	)

	return content, err
}

func checkParse(content string) ParseResult {
	lines := extractMavenLines(content)
	if len(lines) < 10 {
		return ParseResultNone
	}

	parser := structured.NewOutputParser()
	out := parser.ParseOutput(lines, nil, structured.ParseConfig{"noStrict": true})

	hasModule := false
	hasSummary := false
	for _, child := range out.Root.Children {
		if child.Type == "module" {
			hasModule = true
		}
		if child.Type == "summary" {
			hasSummary = true
		}
	}

	if hasModule && hasSummary {
		return ParseResultFull
	}
	if hasModule || hasSummary {
		return ParseResultPartial
	}
	return ParseResultNone
}

func extractMavenLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[INFO]") || strings.HasPrefix(line, "[WARNING]") ||
			strings.HasPrefix(line, "[ERROR]") || strings.HasPrefix(line, "[DEBUG]") ||
			strings.Contains(line, "BUILD ") {
			lines = append(lines, line)
		}
	}
	return lines
}
