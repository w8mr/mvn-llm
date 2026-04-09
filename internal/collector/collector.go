package collector

import (
	"encoding/json"
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

const (
	maxExamples = 1000
	storagePath = "testdata/collector"
	searchDelay = 2 * time.Second
	httpTimeout = 30 * time.Second
)

type Collector struct {
	githubToken string
	storageDir  string
	count       int
	seenURLs    map[string]bool
	client      *http.Client
}

func New(githubToken, storageDir string) *Collector {
	if storageDir == "" {
		storageDir = storagePath
	}
	return &Collector{
		githubToken: githubToken,
		storageDir:  storageDir,
		seenURLs:    make(map[string]bool),
		client: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

func (c *Collector) ensureStorageDir() error {
	return os.MkdirAll(c.storageDir, 0755)
}

func (c *Collector) search(query string) ([]string, error) {
	// Simple search for BUILD SUCCESS in log files
	searchURL := fmt.Sprintf("https://api.github.com/search/code?q=%s+BUILD+SUCCESS&per_page=10", query)
	fmt.Printf("DEBUG: search URL: %s\n", searchURL)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.githubToken != "" {
		req.Header.Set("Authorization", "token "+c.githubToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Debug output
	fmt.Printf("DEBUG: status=%d, body_len=%d\n", resp.StatusCode, len(body))
	if len(body) < 500 {
		fmt.Printf("DEBUG: body=%s\n", string(body))
	}

	var result struct {
		Items []struct {
			URL  string `json:"html_url"`
			Name string `json:"name"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var links []string
	for _, item := range result.Items {
		// For repo search, construct URL to fetch raw content
		if item.URL != "" && !c.seenURLs[item.URL] {
			// Just store the repo URL for now
			links = append(links, item.URL)
		}
	}

	return links, nil
}

func (c *Collector) fetchContent(link string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.raw+json")
	if c.githubToken != "" {
		req.Header.Set("Authorization", "token "+c.githubToken)
	}

	resp, err := c.client.Do(req)
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

type ParseResult int

const (
	ParseResultNone    ParseResult = iota
	ParseResultFull                // has both module and summary
	ParseResultPartial             // has module OR summary OR initialization (partially parseable)
)

func (c *Collector) checkParse(content string) ParseResult {
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
	return ParseResultPartial
}

func (c *Collector) saveExample(content string, sourceURL string, result ParseResult) error {
	if err := c.ensureStorageDir(); err != nil {
		return err
	}

	subdir := "unknown"
	switch result {
	case ParseResultFull:
		subdir = "full"
	case ParseResultPartial:
		subdir = "partial"
	default:
		return nil // Don't save None
	}

	storageSubDir := path.Join(c.storageDir, subdir)
	if err := os.MkdirAll(storageSubDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("example_%d.txt", time.Now().UnixNano()+rand.Int63n(1000000))
	filepath := path.Join(storageSubDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	metadataPath := path.Join(storageSubDir, filename+".meta")
	metaFile, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer metaFile.Close()

	_, err = metaFile.WriteString(sourceURL)
	if err != nil {
		return err
	}

	c.count++
	return nil
}

func extractMavenLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[INFO]") || strings.HasPrefix(line, "[WARNING]") ||
			strings.HasPrefix(line, "[ERROR]") || strings.HasPrefix(line, "[DEBUG]") ||
			strings.HasPrefix(line, "[TRACE]") || strings.HasPrefix(line, "[FATAL]") {
			lines = append(lines, line)
		}
	}
	return lines
}

func (c *Collector) Collect(maxIterations int) error {
	if err := c.ensureStorageDir(); err != nil {
		return err
	}

	queries := []string{
		"maven build output",
		"maven test failure",
		"maven compile error",
		"maven package",
		"maven install",
	}

	for i := 0; i < maxIterations && c.count < maxExamples; i++ {
		query := queries[rand.Intn(len(queries))]
		fmt.Printf("Iteration %d: Searching for '%s'\n", i+1, query)

		links, err := c.search(query)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
			time.Sleep(searchDelay)
			continue
		}

		fmt.Printf("Found %d links\n", len(links))

		for _, link := range links {
			if c.count >= maxExamples {
				break
			}

			if c.seenURLs[link] {
				continue
			}
			c.seenURLs[link] = true

			fmt.Printf("Fetching: %s\n", link)
			content, err := c.fetchContent(link)
			if err != nil {
				fmt.Printf("Fetch error: %v\n", err)
				continue
			}

			result := c.checkParse(content)
			switch result {
			case ParseResultFull:
				fmt.Printf("Full Maven output! Saving...\n")
				c.saveExample(content, link, result)
			case ParseResultPartial:
				fmt.Printf("Partial Maven output! Saving...\n")
				c.saveExample(content, link, result)
			default:
				fmt.Printf("Not Maven output\n")
			}

			time.Sleep(searchDelay)
		}
	}

	fmt.Printf("\nCollected %d examples\n", c.count)
	return nil
}

func (c *Collector) Count() int {
	return c.count
}
