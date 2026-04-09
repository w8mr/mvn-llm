package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/agentic-ai/mvn-llm/internal/maven/structured"
)

type ParseResult int

const (
	ParseResultNone ParseResult = iota
	ParseResultFull
	ParseResultPartial
)

type SearchResult struct {
	URL string
}

func main() {
	apiKey := flag.String("api-key", "", "Tavily API key")
	storageDir := flag.String("storage-dir", "testdata/collector", "Storage directory")
	maxExamples := flag.Int("max-examples", 50, "Maximum examples")
	maxIterations := flag.Int("max-iterations", 5, "Maximum iterations")

	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Error: -api-key is required")
		flag.Usage()
		os.Exit(1)
	}

	os.MkdirAll(*storageDir+"/full", 0755)
	os.MkdirAll(*storageDir+"/partial", 0755)
	os.MkdirAll(*storageDir+"/unknown", 0755)

	client := &http.Client{Timeout: 60 * time.Second}
	count := 0

	queries := []string{
		"maven build output \"BUILD SUCCESS\" log",
		"maven test failure \"BUILD FAILURE\" CI log",
		"maven compile error build log",
		"Travis CI maven build output",
		"Jenkins maven build log",
		"GitHub Actions maven build log",
	}

	for iteration := 0; iteration < *maxIterations && count < *maxExamples; iteration++ {
		query := queries[rand.Intn(len(queries))]
		fmt.Printf("Iteration %d: %s\n", iteration+1, query)

		links, err := searchTavily(client, *apiKey, query)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
			continue
		}

		fmt.Printf("Found %d results\n", len(links))

		for _, link := range links {
			if count >= *maxExamples {
				break
			}

			fmt.Printf("Fetching: %s\n", link.URL)
			content, err := fetchURL(client, link.URL)
			if err != nil {
				fmt.Printf("Fetch error: %v\n", err)
				continue
			}

			result := checkParse(content)
			if result == ParseResultNone {
				continue
			}

			category := "unknown"
			switch result {
			case ParseResultFull:
				category = "full"
			case ParseResultPartial:
				category = "partial"
			}

			filename := fmt.Sprintf("maven_%d.txt", time.Now().UnixNano()+rand.Int63n(1000000))
			filepath := path.Join(*storageDir, category, filename)
			os.WriteFile(filepath, []byte(content), 0644)

			metaPath := filepath + ".meta"
			os.WriteFile(metaPath, []byte(link.URL), 0644)

			count++
			fmt.Printf("Saved %s (%d)\n", category, count)

			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("\nCollected %d examples\n", count)
}

type TavilySearchResponse struct {
	Results []struct {
		URL string `json:"url"`
	} `json:"results"`
}

func searchTavily(client *http.Client, apiKey, query string) ([]SearchResult, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"query":       query,
		"max_results": "10",
	})

	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TavilySearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse error: %v, body: %s", err, string(body))
	}

	var links []SearchResult
	for _, r := range result.Results {
		links = append(links, SearchResult{URL: r.URL})
	}

	return links, nil
}

func fetchURL(client *http.Client, url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
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
