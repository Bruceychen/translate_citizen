package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

const (
	defaultSourceFile = "../source/global.ini"
)

func main() {
	// Determine which file to scan
	filePath := defaultSourceFile
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	fmt.Printf("=== Scanning for Simplified Chinese Characters ===\n")
	fmt.Printf("File: %s\n\n", filePath)

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Detect overall encoding
	encoding := detectFileEncoding(content)
	fmt.Printf("Overall file encoding: %s\n\n", encoding)

	// If file is entirely GBK/GB2312, it's all Simplified Chinese
	if encoding == "GB2312/GBK" {
		fmt.Println("⚠️  WARNING: Entire file is encoded in GB2312/GBK (Simplified Chinese)!")
		fmt.Println("This file should be re-encoded to BIG5 or UTF-8 with Traditional Chinese characters.")
		fmt.Println()
		scanAndPrintSimplifiedChars(content, true)
		return
	}

	// For UTF-8 or BIG5, scan character by character
	if encoding == "UTF-8" {
		fmt.Println("Scanning UTF-8 file for Simplified Chinese characters...")
		scanUTF8ForSimplified(string(content))
	} else if encoding == "BIG5" {
		fmt.Println("✓ File is encoded in BIG5 (Traditional Chinese)")
		fmt.Println("Scanning for any anomalies...")
		scanBIG5File(content)
	} else {
		fmt.Println("Unknown encoding - attempting byte-level scan...")
		scanMixedEncodingFile(content)
	}
}

// detectFileEncoding detects the primary encoding of the file
func detectFileEncoding(content []byte) string {
	// Check for UTF-8 BOM
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return "UTF-8"
	}

	// Check if valid UTF-8
	if utf8.Valid(content) {
		// Could be UTF-8, but check for Chinese byte ranges
		hasGBKRange := false
		hasBIG5Range := false

		for i := 0; i < len(content); i++ {
			b := content[i]
			// GBK high byte range: 0x81-0xFE
			if b >= 0x81 && b <= 0xFE {
				hasGBKRange = true
			}
			// BIG5 high byte range: 0xA1-0xF9
			if b >= 0xA1 && b <= 0xF9 {
				hasBIG5Range = true
			}
		}

		if !hasGBKRange && !hasBIG5Range {
			return "UTF-8"
		}
	}

	// Try BIG5 decode
	if isBIG5Encoded(content) {
		return "BIG5"
	}

	// Try GBK decode
	if isGBKEncoded(content) {
		return "GB2312/GBK"
	}

	return "UTF-8"
}

// isBIG5Encoded checks if content is BIG5 encoded
func isBIG5Encoded(content []byte) bool {
	decoder := traditionalchinese.Big5.NewDecoder()
	decoded := make([]byte, len(content)*3)
	n, _, err := decoder.Transform(decoded, content, true)
	if err != nil {
		return false
	}
	// Check if we have valid BIG5 byte ranges
	for i := 0; i < len(content)-1; i++ {
		if content[i] >= 0xA1 && content[i] <= 0xF9 {
			if (content[i+1] >= 0x40 && content[i+1] <= 0x7E) ||
				(content[i+1] >= 0xA1 && content[i+1] <= 0xFE) {
				return n > 0
			}
		}
	}
	return false
}

// isGBKEncoded checks if content is GBK/GB2312 encoded
func isGBKEncoded(content []byte) bool {
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded := make([]byte, len(content)*3)
	n, _, err := decoder.Transform(decoded, content, true)
	if err != nil {
		return false
	}
	// Check if we have valid GBK byte ranges
	for i := 0; i < len(content)-1; i++ {
		if content[i] >= 0x81 && content[i] <= 0xFE {
			if content[i+1] >= 0x40 && content[i+1] <= 0xFE {
				return n > 0
			}
		}
	}
	return false
}

// scanAndPrintSimplifiedChars prints all characters from a GBK-encoded file
func scanAndPrintSimplifiedChars(content []byte, isGBK bool) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded := make([]byte, len(content)*3)
	n, _, err := decoder.Transform(decoded, content, true)
	if err != nil {
		fmt.Printf("Error decoding GBK: %v\n", err)
		return
	}

	text := string(decoded[:n])
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNum := 0
	foundCount := 0

	fmt.Println("Simplified Chinese characters found:")
	fmt.Println()
	fmt.Println("Line | Character | Context")
	fmt.Println("-----|-----------|--------")

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Find Chinese characters in the line
		runes := []rune(line)
		for i, r := range runes {
			// Check if it's a CJK character
			if r >= 0x4E00 && r <= 0x9FFF {
				foundCount++
				// Get context (10 chars before and after)
				start := i - 10
				if start < 0 {
					start = 0
				}
				end := i + 10
				if end > len(runes) {
					end = len(runes)
				}
				context := string(runes[start:end])
				fmt.Printf("%4d | %s (U+%04X) | %s\n", lineNum, string(r), r, context)
			}
		}
	}

	fmt.Printf("\nTotal Simplified Chinese characters found: %d\n", foundCount)
}

// scanUTF8ForSimplified scans UTF-8 content for Simplified Chinese characters
func scanUTF8ForSimplified(content string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	foundCount := 0

	fmt.Println()
	fmt.Println("Scanning for characters that are Simplified-only variants...")
	fmt.Println()

	// Known Simplified-only characters (a small sample)
	// In practice, you'd need a comprehensive mapping
	simplifiedOnly := map[rune]bool{
		'国': true, // 國 in Traditional
		'门': true, // 門 in Traditional
		'长': true, // 長 in Traditional
		'开': true, // 開 in Traditional
		'车': true, // 車 in Traditional
		'贝': true, // 貝 in Traditional
		'见': true, // 見 in Traditional
		'气': true, // 氣 in Traditional
		'无': true, // 無 in Traditional
		'专': true, // 專 in Traditional
	}

	fmt.Println("Line | Character | Unicode  | Context")
	fmt.Println("-----|-----------|----------|--------")

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		runes := []rune(line)

		for i, r := range runes {
			if simplifiedOnly[r] {
				foundCount++
				// Get context
				start := i - 15
				if start < 0 {
					start = 0
				}
				end := i + 15
				if end > len(runes) {
					end = len(runes)
				}
				context := string(runes[start:end])
				fmt.Printf("%4d | %c         | U+%04X   | %s\n", lineNum, r, r, context)
			}
		}
	}

	if foundCount == 0 {
		fmt.Println("✓ No common Simplified-only characters found")
		fmt.Println("\nNote: This check uses a limited set of Simplified-only characters.")
		fmt.Println("For comprehensive detection, consider using a full Traditional/Simplified mapping.")
	} else {
		fmt.Printf("\n⚠️  Found %d Simplified-only characters\n", foundCount)
	}
}

// scanBIG5File scans a BIG5 encoded file
func scanBIG5File(content []byte) {
	decoder := traditionalchinese.Big5.NewDecoder()
	decoded := make([]byte, len(content)*3)
	n, _, err := decoder.Transform(decoded, content, true)
	if err != nil {
		fmt.Printf("Error decoding BIG5: %v\n", err)
		return
	}

	text := string(decoded[:n])
	fmt.Println("✓ Successfully decoded as BIG5")
	fmt.Printf("Total characters: %d\n", utf8.RuneCountInString(text))
	fmt.Println("\nNo Simplified Chinese encoding detected.")
}

// scanMixedEncodingFile scans for mixed encoding issues
func scanMixedEncodingFile(content []byte) {
	fmt.Println()
	fmt.Println("Scanning for encoding inconsistencies...")
	fmt.Println()

	lineNum := 1
	pos := 0
	foundIssues := 0

	fmt.Println("Line | Byte Pos | Bytes      | Issue")
	fmt.Println("-----|----------|------------|-------")

	for pos < len(content) {
		if content[pos] == '\n' {
			lineNum++
			pos++
			continue
		}

		// Check for GBK signature
		if pos < len(content)-1 {
			if content[pos] >= 0x81 && content[pos] <= 0xFE {
				if content[pos+1] >= 0x40 && content[pos+1] <= 0xFE {
					foundIssues++
					fmt.Printf("%4d | %8d | %02X %02X     | Possible GBK (Simplified)\n",
						lineNum, pos, content[pos], content[pos+1])
					pos += 2
					continue
				}
			}
		}

		pos++
	}

	if foundIssues == 0 {
		fmt.Println("✓ No encoding issues detected")
	} else {
		fmt.Printf("\n⚠️  Found %d potential Simplified Chinese byte sequences\n", foundIssues)
	}
}
