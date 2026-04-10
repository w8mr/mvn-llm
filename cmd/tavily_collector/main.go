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
	"path/filepath"
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
	seenURLs := make(map[string]bool)

	// Load existing URLs to avoid duplicates
	filepath.Walk(*storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".meta") {
			if data, err := os.ReadFile(path); err == nil {
				seenURLs[string(data)] = true
			}
		}
		return nil
	})
	fmt.Printf("Found %d existing URLs to skip\n", len(seenURLs))

	queries := []string{
		"site:build.alpinelinux.org BUILD SUCCESS",
		"site:thebusybiscuit.github.io BUILD SUCCESS",
		"site:kojipkgs.fedoraproject.org BUILD SUCCESS",
		"site:download.eclipse.org maven build log",
		"site:download.oracle.com glassfish maven",
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

			// Skip duplicates
			if seenURLs[link.URL] {
				fmt.Printf("Skipping duplicate: %s\n", link.URL)
				continue
			}
			seenURLs[link.URL] = true

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

// extractMavenLines extracts Maven log section from raw content (which may include HTML)
// Uses start/end markers to find the Maven block and returns all lines in that range
func extractMavenLines(content string) []string {
	lines := strings.Split(content, "\n")

	// Find start: look for Maven build start patterns
	startIdx := -1
	for i, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "[info] scanning for projects") ||
			strings.Contains(lower, "---<") ||
			strings.Contains(lower, "reactor build order") ||
			strings.Contains(lower, "building ") && strings.Contains(lower, " from ") {
			// Keep some lines before the start (up to 5 lines)
			startIdx = i - 5
			if startIdx < 0 {
				startIdx = 0
			}
			break
		}
	}

	// Find end: look for BUILD SUCCESS/FAILURE
	endIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		lower := strings.ToLower(lines[i])
		if strings.Contains(lower, "build success") ||
			strings.Contains(lower, "build failure") ||
			strings.Contains(lower, "reactor summary") {
			// Keep some lines after the end (up to 10 lines)
			endIdx = i + 10
			if endIdx >= len(lines) {
				endIdx = len(lines) - 1
			}
			break
		}
	}

	// Extract the relevant section
	if startIdx >= 0 && endIdx >= 0 && startIdx < endIdx {
		lines = lines[startIdx : endIdx+1]
	} else if startIdx >= 0 {
		lines = lines[startIdx:]
	} else if endIdx >= 0 {
		lines = lines[:endIdx+1]
	}

	// Limit size - take first 5000 lines
	if len(lines) > 5000 {
		lines = lines[:5000]
	}

	return lines
}
