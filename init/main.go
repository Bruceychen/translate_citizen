package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	// Input file path
	sourceFile = "../source/global.ini"
	// Output file path - JSON format for easy loading
	outputJSON = "translation_map.json"
)

// TranslationEntry represents a key-value pair from the INI file
type TranslationEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	fmt.Println("=== INI Key-Value Extractor ===")
	fmt.Printf("Reading from: %s\n", sourceFile)

	// Read and parse the INI file
	entries, err := parseINIFile(sourceFile)
	if err != nil {
		log.Fatalf("Error parsing INI file: %v", err)
	}

	fmt.Printf("Extracted %d key-value pairs\n", len(entries))

	// Write to JSON file (recommended for loading into memory)
	if err := writeJSON(outputJSON, entries); err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}
	fmt.Printf("✓ Written to: %s\n", outputJSON)

	fmt.Println("\n=== Extraction Complete ===")
	fmt.Printf("Total entries processed: %d\n", len(entries))
}

// parseINIFile reads the INI file and extracts key-value pairs
func parseINIFile(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	entries := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Skip UTF-8 BOM if present (at the start of file)
		if lineNumber == 1 && strings.HasPrefix(line, "\ufeff") {
			line = strings.TrimPrefix(line, "\ufeff")
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Find the first '=' as the separator
		index := strings.Index(line, "=")
		if index == -1 {
			// No '=' found, skip this line
			fmt.Printf("Warning: Line %d has no '=' separator, skipping: %s\n", lineNumber, line)
			continue
		}

		// Split into key and value
		key := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])

		if key == "" {
			fmt.Printf("Warning: Line %d has empty key, skipping\n", lineNumber)
			continue
		}

		// Store in map
		entries[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

// writeJSON writes the entries to a JSON file
func writeJSON(filepath string, entries map[string]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print with 2-space indentation

	if err := encoder.Encode(entries); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// writePlainText writes the entries to a plain text file (one entry per line)
func writePlainText(filepath string, entries map[string]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create text file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintf(writer, "# Translation Map - Plain Text Format\n")
	fmt.Fprintf(writer, "# Total entries: %d\n", len(entries))
	fmt.Fprintf(writer, "# Format: KEY=VALUE\n\n")

	// Write each entry
	for key, value := range entries {
		fmt.Fprintf(writer, "%s=%s\n", key, value)
	}

	return nil
}
