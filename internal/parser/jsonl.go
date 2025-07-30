package parser

import (
	"bufio"
	"encoding/json"
	"github.com/Brads3290/cclogviewer/internal/debug"
	"github.com/Brads3290/cclogviewer/internal/models"
	"log"
	"os"
)

// ReadJSONLFile reads a JSONL file and returns a slice of LogEntry
func ReadJSONLFile(filename string) ([]models.LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []models.LogEntry
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max line size

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		var entry models.LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			if debug.Enabled {
				log.Printf("Error parsing line %d: %v", lineNum, err)
			}
			continue
		}

		// Skip summary messages
		if entry.Type == "summary" {
			if debug.Enabled {
				log.Printf("Skipping summary message at line %d", lineNum)
			}
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
