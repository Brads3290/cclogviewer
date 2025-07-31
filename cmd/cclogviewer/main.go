package main

import (
	"flag"
	"fmt"
	"github.com/brads3290/cclogviewer/internal/browser"
	"github.com/brads3290/cclogviewer/internal/constants"
	debugpkg "github.com/brads3290/cclogviewer/internal/debug"
	"github.com/brads3290/cclogviewer/internal/parser"
	"github.com/brads3290/cclogviewer/internal/processor"
	"github.com/brads3290/cclogviewer/internal/renderer"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

var (
	// Version can be set by ldflags during build
	Version = constants.DefaultVersion
	// BuildTime can be set by ldflags during build
	BuildTime = ""
)

func main() {
	var inputFile, outputFile string
	var openBrowser, showVersion bool
	flag.StringVar(&inputFile, "input", "", "Input JSONL file path")
	flag.StringVar(&outputFile, "output", "", "Output HTML file path (optional)")
	flag.BoolVar(&openBrowser, "open", false, "Open the generated HTML file in browser")
	flag.BoolVar(&debugpkg.Enabled, "debug", false, "Enable debug logging")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		version := Version
		if version == "" {
			// Try to get version from build info
			if info, ok := debug.ReadBuildInfo(); ok {
				version = info.Main.Version
				if version == constants.DevelopmentVersionString {
					version = constants.DevVersionString
				}
			} else {
				version = constants.UnknownVersionString
			}
		}

		fmt.Printf("cclogviewer version %s", version)
		if BuildTime != "" {
			fmt.Printf(" (built %s)", BuildTime)
		}
		fmt.Println()
		os.Exit(0)
	}

	if inputFile == "" {
		log.Fatal("Please provide an input file using -input flag")
	}

	// If no output file specified, create a temp file and auto-open it
	autoOpen := false
	if outputFile == "" {
		// Generate unique filename based on input file and timestamp
		baseName := filepath.Base(inputFile)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		timestamp := time.Now().Format(constants.TempFileTimestampFormat)
		outputFile = filepath.Join(os.TempDir(), fmt.Sprintf(constants.TempFileNameFormat, baseName, timestamp))
		autoOpen = true
	}

	entries, err := parser.ReadJSONLFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	processed := processor.ProcessEntries(entries)

	err = renderer.GenerateHTML(processed, outputFile, debugpkg.Enabled)
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
