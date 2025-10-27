package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// This is an example of how another Go application can load
// the translation_map.json file into memory for fast lookups

// TranslationCache holds the in-memory map for O(1) lookups
type TranslationCache struct {
	data map[string]string
}

// NewTranslationCache loads the translation map from a JSON file
func NewTranslationCache(filepath string) (*TranslationCache, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open translation file: %w", err)
	}
	defer file.Close()

	var data map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &TranslationCache{data: data}, nil
}

// Get retrieves a translation by key (O(1) lookup)
func (tc *TranslationCache) Get(key string) (string, bool) {
	value, exists := tc.data[key]
	return value, exists
}

// Size returns the number of entries in the cache
func (tc *TranslationCache) Size() int {
	return len(tc.data)
}

// Example usage
func main() {
	fmt.Println("=== Translation Cache Loader Example ===\n")

	// Load the translation map into memory
	cache, err := NewTranslationCache("translation_map.json")
	if err != nil {
		fmt.Printf("Error loading cache: %v\n", err)
		return
	}

	fmt.Printf("✓ Loaded %d translations into memory\n\n", cache.Size())

	// Example lookups
	testKeys := []string{
		"ASD_Active,P",
		"ASD_Airlock_Title,P",
		"2019_Ann_Sale_Day1",
		"NonExistentKey",
	}

	fmt.Println("--- Lookup Examples ---")
	for _, key := range testKeys {
		if value, exists := cache.Get(key); exists {
			fmt.Printf("✓ %s = %s\n", key, value)
		} else {
			fmt.Printf("✗ %s = [NOT FOUND]\n", key)
		}
	}

	fmt.Println("\n=== Performance Test ===")
	fmt.Printf("Memory efficient: O(1) lookup time for %d entries\n", cache.Size())
	fmt.Println("Ready for processing millions of lines with fast lookups!")
}
