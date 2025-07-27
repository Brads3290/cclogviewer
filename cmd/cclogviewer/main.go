package main

import (
	"cclogviewer/internal/browser"
	"cclogviewer/internal/parser"
	"cclogviewer/internal/processor"
	"cclogviewer/internal/renderer"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var inputFile, outputFile string
	var openBrowser bool
	flag.StringVar(&inputFile, "input", "", "Input JSONL file path")
	flag.StringVar(&outputFile, "output", "", "Output HTML file path (optional)")
	flag.BoolVar(&openBrowser, "open", false, "Open the generated HTML file in browser")
	flag.Parse()

	if inputFile == "" {
		log.Fatal("Please provide an input file using -input flag")
	}

	// If no output file specified, create a temp file and auto-open it
	autoOpen := false
	if outputFile == "" {
		// Generate unique filename based on input file and timestamp
		baseName := filepath.Base(inputFile)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		timestamp := time.Now().Format("20060102-150405")
		outputFile = filepath.Join(os.TempDir(), fmt.Sprintf("cclog-%s-%s.html", baseName, timestamp))
		autoOpen = true
	}

	entries, err := parser.ReadJSONLFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	processed := processor.ProcessEntries(entries)
	
	err = renderer.GenerateHTML(processed, outputFile)
	if err != nil {
		log.Fatalf("Error generating HTML: %v", err)
	}

	fmt.Printf("Successfully generated %s\n", outputFile)
	
	// Open browser if -open flag was set OR if output was auto-generated
	if openBrowser || autoOpen {
		if err := browser.OpenInBrowser(outputFile); err != nil {
			log.Printf("Warning: Could not open browser: %v", err)
		}
	}
}