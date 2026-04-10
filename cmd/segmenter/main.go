package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	structured "github.com/agentic-ai/mvn-llm/internal/maven/structured"
)

// inputDirs holds all input directories for log files
var inputDirs = []string{
	"testdata/collector/full",
	"testdata/collector/partial",
}

// outputDir is base for all segmented outputs
const outputDir = "testdata/segmented"

func main() {
	args := os.Args[1:]
	var targets []string
	if len(args) == 0 {
		// No file specified: collect all .txt files from inputDirs
		for _, dir := range inputDirs {
			matches, err := filepath.Glob(filepath.Join(dir, "*.txt"))
			if err != nil {
				log.Fatalf("Error globbing in %s: %v", dir, err)
			}
			for _, f := range matches {
				targets = append(targets, f)
			}
		}
		if len(targets) == 0 {
			log.Fatalf("No log files found in full or partial subfolders")
		}
	} else if len(args) == 1 {
		targets = append(targets, args[0])
	} else {
		log.Fatalf("Usage: %s <input-log-file>", os.Args[0])
	}
	for _, src := range targets {
		log.Printf("Processing: %s", src)
		lines, err := readLines(src)
		if err != nil {
			log.Printf("Error reading %s: %v", src, err)
			continue
		}
		parser := structured.NewOutputParser()
		parsed := parser.ParseOutput(lines, nil, map[string]any{})
		segmentTree(parsed.Root, src)
	}
}

// readLines reads all the lines from a .txt file
func readLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// segmentTree recursively visits nodes to segment/emit output per type
func segmentTree(node structured.Node, srcFile string) {
	if len(node.Lines) > 0 {
		dir, fname := outputPathForNode(node, srcFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("Error creating directory %s: %v", dir, err)
			return
		}
		writeUniqueSegment(dir, fname, node.Lines)
	}
	for _, child := range node.Children {
		segmentTree(child, srcFile)
	}
}

// shouldEmitNode returns true if this node type should be written to output
func shouldEmitNode(t string) bool {
	switch t {
	case "initialization", "summary", "module", "build-block":
		return true
	default:
		return false
	}
}

// outputPathForNode determines the output dir and base filename for a node
func outputPathForNode(node structured.Node, srcFile string) (string, string) {
	base := filepath.Base(srcFile)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	var dir string
	switch node.Type {
	case "initialization", "summary", "module":
		dir = filepath.Join(outputDir, node.Type)
	case "build-block":
		plugin, _ := node.Meta["plugin"].(string)
		goal, _ := node.Meta["goal"].(string)
		if plugin == "" {
			plugin = "unknown"
		}
		if goal == "" {
			goal = "unknown"
		}
		dir = filepath.Join(outputDir, "build", plugin, goal)
	default:
		dir = filepath.Join(outputDir, "other", node.Type)
	}
	return dir, name
}

// writeUniqueSegment finds next sequence number, checks for dup content, writes if new.
func writeUniqueSegment(dir, name string, lines []string) {
	seq := 1
	maxSeq := 10000
	for ; seq < maxSeq; seq++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d.txt", name, seq))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			// File does not exist, write it.
			writeLines(candidate, lines)
			log.Printf("Wrote: %s", candidate)
			return
		}
		// Exists - check contents
		existing, err := readLines(candidate)
		if err != nil {
			log.Printf("Error reading existing %s: %v", candidate, err)
			continue
		}
		if sliceEqual(existing, lines) {
			log.Printf("Duplicate segment matches existing: %s, skipping", candidate)
			return
		}
	}
	log.Printf("Too many segments in %s for %s, giving up!", dir, name)
}

// sliceEqual compares two string slices
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// writeLines writes lines (with newlines) to a file
func writeLines(filename string, lines []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Error writing %s: %v", filename, err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		w.WriteString(l + "\n")
	}
	w.Flush()
}
