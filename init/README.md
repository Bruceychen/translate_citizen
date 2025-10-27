# INI Key-Value Extractor

This tool extracts key-value pairs from `source/global.ini` and generates data files that can be loaded into memory by other Go applications.

## What It Does

1. Reads `source/global.ini` line by line
2. Parses each line using `=` as the separator
3. Extracts key-value pairs
4. Outputs two files:
   - `translation_map.json` - JSON format (recommended for loading into memory)
   - `translation_map.txt` - Plain text format (for human readability)

## Features

- Handles UTF-8 BOM markers (common in INI files)
- Skips empty lines and comments
- Preserves keys with `,P` suffix
- Validates each line has proper key=value format
- Reports warnings for malformed lines

## Usage

### Run the extractor:

```bash
cd init
go run main.go
```

### Build and run:

```bash
cd init
go build -o extractor
./extractor
```

## Output Files

### translation_map.json
JSON format, easy to load into Go maps:
```json
{
  "ASD_Active": "Active",
  "ASD_Airlock_Title,P": "Airlock Control"
}
```

### translation_map.txt
Plain text format for manual inspection:
```
KEY=VALUE
ASD_Active=Active
ASD_Airlock_Title,P=Airlock Control
```

## Loading the Map in Another Go App

Example code to load the JSON file:

```go
package main

import (
    "encoding/json"
    "os"
)

func loadTranslationMap(filepath string) (map[string]string, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var translations map[string]string
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&translations); err != nil {
        return nil, err
    }

    return translations, nil
}

func main() {
    translations, err := loadTranslationMap("init/translation_map.json")
    if err != nil {
        panic(err)
    }

    // Use the map for O(1) lookups
    value := translations["ASD_Active"]
    println(value) // Output: Active
}
```

## Performance

- Time Complexity: O(n) where n is the number of lines
- Space Complexity: O(n) to store all key-value pairs
- Suitable for files with millions of lines
- JSON loading is fast with O(1) map lookups