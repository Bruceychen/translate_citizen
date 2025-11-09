# Simplified Chinese Character Detector (findSC)

This tool scans INI files to detect Simplified Chinese characters that should be Traditional Chinese. It uses encoding detection (BIG5, GB2312/GBK) to identify problematic characters.

## Purpose

Ensures that translation files use Traditional Chinese (Taiwan) instead of Simplified Chinese (China) by:
1. Detecting the file's character encoding
2. Scanning character-by-character for Simplified Chinese
3. Printing out any problematic characters with line numbers and context

## Usage

### Default Mode (scans `../source/global.ini`)

```bash
cd findSC
go run main.go
# or
./findSC.exe
```

### Custom File Mode

```bash
cd findSC
go run main.go <path_to_file>
# or
./findSC.exe <path_to_file>

# Examples:
./findSC.exe ../source/global.ini
./findSC.exe ../process/output/global.ini
./findSC.exe D:\path\to\any\file.ini
```

### Build Executable

```bash
cd findSC
go build -o findSC
```

## How It Works

### 1. Encoding Detection

The tool first detects the overall file encoding:

- **UTF-8**: Checks for BOM marker and validates UTF-8 sequences
- **BIG5**: Traditional Chinese encoding (Taiwan) - **GOOD** ✓
- **GB2312/GBK**: Simplified Chinese encoding (China) - **BAD** ⚠️

### 2. Character-by-Character Scanning

Depending on the detected encoding:

#### If GB2312/GBK (Simplified Chinese):
- **WARNING**: Entire file is in Simplified Chinese encoding
- Decodes and prints all Chinese characters with:
  - Line number
  - Character and Unicode code point
  - Context (surrounding text)

#### If UTF-8:
- Scans for known Simplified-only character variants
- Checks against a mapping of Simplified → Traditional pairs:
  - 国 → 國
  - 门 → 門
  - 长 → 長
  - 开 → 開
  - 车 → 車
  - (and more)
- Prints line number, character, Unicode, and context

#### If BIG5 (Traditional Chinese):
- Validates the encoding is correct ✓
- Reports success

#### If Mixed/Unknown:
- Performs byte-level scan for GBK signatures
- Reports any GB2312/GBK byte sequences found

## Output Format

### For GB2312/GBK Files:

```
=== Scanning for Simplified Chinese Characters ===
File: ../source/global.ini

Overall file encoding: GB2312/GBK

⚠️  WARNING: Entire file is encoded in GB2312/GBK (Simplified Chinese)!
This file should be re-encoded to BIG5 or UTF-8 with Traditional Chinese characters.

Simplified Chinese characters found:

Line | Character | Context
-----|-----------|--------
   5 | 国 (U+56FD) | ASD_Country=国家
  12 | 门 (U+95E8) | ASD_Door=大门
...

Total Simplified Chinese characters found: 150
```

### For UTF-8 Files with Simplified Characters:

```
=== Scanning for Simplified Chinese Characters ===
File: ../source/global.ini

Overall file encoding: UTF-8

Scanning UTF-8 file for Simplified Chinese characters...

Scanning for characters that are Simplified-only variants...

Line | Character | Unicode  | Context
-----|-----------|----------|--------
  42 | 国         | U+56FD   | ASD_Country=国家设置
  88 | 车         | U+8F66   | ASD_Vehicle=车辆管理

⚠️  Found 2 Simplified-only characters
```

### For BIG5 Files (Correct):

```
=== Scanning for Simplified Chinese Characters ===
File: ../source/global.ini

Overall file encoding: BIG5

✓ File is encoded in BIG5 (Traditional Chinese)
Scanning for any anomalies...
✓ Successfully decoded as BIG5
Total characters: 83278

No Simplified Chinese encoding detected.
```

## Encoding Reference

### Traditional Chinese (Taiwan) - Correct:
- **BIG5** (繁體中文)
- **ISO-2022-TW**
- UTF-8 with Traditional variants

### Simplified Chinese (China) - Incorrect:
- **GB2312** (简体中文)
- **GBK** (扩展)
- **GB18030**
- UTF-8 with Simplified variants

## Technical Details

### BIG5 Byte Ranges:
- High byte: `0xA1` - `0xF9`
- Low byte: `0x40` - `0x7E` or `0xA1` - `0xFE`

### GBK/GB2312 Byte Ranges:
- High byte: `0x81` - `0xFE`
- Low byte: `0x40` - `0xFE`

### CJK Unicode Range:
- `0x4E00` - `0x9FFF` (Common Chinese characters)

## Limitations

For UTF-8 files, the tool uses a **limited set** of common Simplified-only characters. For comprehensive detection:

1. Consider using a full Traditional/Simplified mapping database
2. Use specialized NLP libraries like `opencc` (Open Chinese Convert)
3. Manually verify critical translations

## Integration with Translation Workflow

1. **Before extraction** (`init/`):
   ```bash
   cd findSC
   ./findSC.exe ../source/global.ini
   ```
   Ensure source file has no Simplified Chinese characters

2. **After translation** (`process/`):
   ```bash
   cd findSC
   ./findSC.exe ../process/output/global.ini
   ```
   Verify output file maintained Traditional Chinese

## Dependencies

- Go 1.25.0+
- `golang.org/x/text v0.30.0` (for encoding/decoding)