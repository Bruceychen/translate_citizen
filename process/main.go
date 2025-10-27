package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	// File paths
	outputFilePath   = "output/global.ini"
	backupFilePath   = "bak/global_bak.ini"
	sourceFilePath   = "../source/global.ini"
	translationMapTC = "../init/translation_map_tc.json"
)

func main() {
	fmt.Println("=== Translation Processor ===")
	fmt.Println()

	// Step 1: Move output/global.ini to bak/global_bak.ini
	if err := backupFile(outputFilePath, backupFilePath); err != nil {
		log.Printf("Warning: Could not backup file: %v (continuing anyway)", err)
	} else {
		fmt.Printf("✓ Step 1: Backed up %s to %s\n", outputFilePath, backupFilePath)
	}

	// Step 2: Copy source/global.ini to output/global.ini
	if err := copyFile(sourceFilePath, outputFilePath); err != nil {
		log.Fatalf("Error copying source file: %v", err)
	}
	fmt.Printf("✓ Step 2: Copied %s to %s\n", sourceFilePath, outputFilePath)

	// Step 3: Load translation map from JSON
	translationMap, err := loadTranslationMap(translationMapTC)
	if err != nil {
		log.Fatalf("Error loading translation map: %v", err)
	}
	fmt.Printf("✓ Step 3: Loaded %d translations from %s\n", len(translationMap), translationMapTC)

	// Step 4: Translate output/global.ini
	stats, err := translateFile(outputFilePath, translationMap)
	if err != nil {
		log.Fatalf("Error translating file: %v", err)
	}
	fmt.Printf("✓ Step 4: Translated %s\n", outputFilePath)
	fmt.Println()

	// Print statistics
	fmt.Println("=== Translation Complete ===")
	fmt.Printf("Total lines processed: %d\n", stats.TotalLines)
	fmt.Printf("Lines translated: %d\n", stats.Translated)
	fmt.Printf("Lines unchanged: %d\n", stats.Unchanged)
	fmt.Printf("Lines skipped (empty/comment): %d\n", stats.Skipped)
	fmt.Printf("Keys not found in map: %d\n", stats.NotFound)
	fmt.Println()
	fmt.Printf("Output file: %s\n", outputFilePath)
	fmt.Printf("Backup file: %s\n", backupFilePath)
}

// TranslationStats holds statistics about the translation process
type TranslationStats struct {
	TotalLines int
	Translated int
	Unchanged  int
	Skipped    int
	NotFound   int
}

// backupFile moves a file from src to dst (essentially a rename)
func backupFile(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", src)
	}

	// Remove destination if it exists
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("failed to remove existing backup: %w", err)
		}
	}

	// Move the file
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// loadTranslationMap loads the translation map from a JSON file
func loadTranslationMap(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open translation map file: %w", err)
	}
	defer file.Close()

	var translationMap map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&translationMap); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return translationMap, nil
}

// translateFile translates the INI file using the provided translation map
func translateFile(filepath string, translationMap map[string]string) (*TranslationStats, error) {
	// Read the file
	inputFile, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer inputFile.Close()

	// Create temporary file for writing
	tempFile, err := os.CreateTemp("", "translation_*.ini")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file if something goes wrong

	stats := &TranslationStats{}
	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(tempFile)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		stats.TotalLines++
		line := scanner.Text()
		originalLine := line

		// Handle UTF-8 BOM if present (at the start of file)
		if lineNumber == 1 && strings.HasPrefix(line, "\ufeff") {
			// Keep the BOM and process the rest
			bomPrefix := "\ufeff"
			line = strings.TrimPrefix(line, bomPrefix)

			// Process the line without BOM
			processedLine := processLine(line, translationMap, stats)

			// Write back with BOM
			fmt.Fprintf(writer, "%s%s\n", bomPrefix, processedLine)
			continue
		}

		// Process normal lines
		processedLine := processLine(line, translationMap, stats)
		fmt.Fprintf(writer, "%s\n", processedLine)

		// Track if line was actually changed
		if processedLine != originalLine && !isEmptyOrComment(originalLine) {
			// Line was translated (already counted in processLine)
		}
	}

	if err := scanner.Err(); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Flush the writer
	if err := writer.Flush(); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("error flushing writer: %w", err)
	}
	tempFile.Close()

	// Replace original file with translated version
	if err := os.Rename(tempPath, filepath); err != nil {
		return nil, fmt.Errorf("failed to replace original file: %w", err)
	}

	return stats, nil
}

// processLine processes a single line and returns the translated version
func processLine(line string, translationMap map[string]string, stats *TranslationStats) string {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines and comments
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
		stats.Skipped++
		return line
	}

	// Find the '=' separator
	index := strings.Index(line, "=")
	if index == -1 {
		// No '=' found, return line as-is
		stats.Unchanged++
		return line
	}

	// Extract key and value
	key := strings.TrimSpace(line[:index])
	_ = strings.TrimSpace(line[index+1:]) // value unused in lookup, only for validation

	if key == "" {
		stats.Unchanged++
		return line
	}

	// Look up translation
	if translatedValue, found := translationMap[key]; found {
		stats.Translated++
		// Preserve the original spacing around '='
		// Reconstruct the line with translated value
		return key + "=" + translatedValue
	}

	// Key not found in translation map
	stats.NotFound++
	stats.Unchanged++
	return line
}

// isEmptyOrComment checks if a line is empty or a comment
func isEmptyOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";")
}