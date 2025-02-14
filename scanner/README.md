# jfind

A cross-platform command-line tool to find and evaluate Java installations on your system. It can locate Java executables and optionally evaluate their version information, providing output in both text and JSON formats.

## Features

- Find Java executables recursively in specified directories
- Detect Oracle JDKs (prints warning in text mode, sets is_oracle flag in JSON mode)
- Cross-platform support (Windows, Linux, macOS)
- Optional evaluation of Java version information
- JSON output format with metadata
- Configurable search depth
- Verbose mode for detailed scanning information
- License requirement detection for Java runtimes

### License Requirement Detection

When using the `-eval` flag, jfind determines if a Java runtime requires a commercial license based on the following rules:

1. **OpenJDK**: Never requires a license
   - Any runtime containing "openjdk" in its name (case-insensitive)
   - Specifically "OpenJDK Runtime Environment"

2. **Commercial Features**: Requires a license if
   - Runtime description contains "commercial" (case-insensitive)

3. **Oracle JDK Version Rules**:
   - JDK 7: Requires license for updates > 80
   - JDK 8: Requires license for updates > 202
   - JDK 11: Always requires license
   - JDK 17: Requires license for versions 17.0.13 and later
   - JDK 18-20: No license required
   - JDK 21+: No license required
   - Other versions: License required by default

Use the `-require-license` flag to filter and show only Java installations that require a commercial license.

## Installation

### Building from Source

1. Ensure you have Go installed on your system
2. Clone the repository
3. Build using task:

```bash
task build        # Build for current platform
task build:all    # Build for all supported platforms
```

## Usage

Basic usage:
```bash
jfind -path <search_path> [options]
```

### Options

- `-path string`: Starting path for search (required)
- `-depth int`: Maximum depth to search (-1 for unlimited)
- `-verbose`: Enable verbose output
- `-eval`: Evaluate found java executables
- `-json`: Output results in JSON format
- `-post`: Post JSON output to server (implies --json)
- `-url string`: URL to post JSON output to (only used with --post, default http://localhost:8000/api/jfind)
- `-require-license`: Filter only Java runtimes that require a commercial license (requires --eval)
- `-h, -help`: Show help message

Note: All options can be specified with either single dash (-) or double dash (--).

### Examples

Find Java installations in /usr/lib/jvm:
```bash
jfind -path /usr/lib/jvm
```

Find and evaluate Java installations with JSON output:
```bash
jfind -path /usr/lib/jvm -eval -json
```

Find Java installations requiring commercial license:
```bash
jfind -path /opt -eval -require-license
```

Find Java installations requiring license and post results to server:
```bash
jfind -path /usr/local -eval -require-license -post
```

### Output Formats

#### Text Output (default)
```
Java executable: /path/to/java
Java version: 11.0.12
Java vendor: Oracle Corporation
Java runtime name: Java(TM) SE Runtime Environment
```

#### JSON Output
When using the `-json` flag, output will be in JSON format:

```json
{
  "meta": {
    "scan_ts": "2025-02-04T15:12:01Z",      // Scan timestamp in UTC
    "computer_name": "hostname",             // Name of the computer
    "user_name": "username",                 // Name of the user
    "scan_duration": "PT5S",                 // Scan duration in ISO 8601
    "has_oracle_jdk": true,                  // Whether any Oracle JDK was found
    "count_result": 3,                       // Total number of Java executables found
    "count_require_license": 1,              // Number requiring commercial license
    "scanned_dirs": 150,                     // Number of directories scanned
    "scan_path": "/usr/lib/jvm"             // Starting path for scan
  },
  "result": [
    {
      "java_executable": "/path/to/java",    // Path to Java executable
      "java_runtime": "OpenJDK Runtime Environment", // Runtime name if evaluated
      "java_vendor": "Eclipse Adoptium",     // Vendor if evaluated
      "is_oracle": false,                    // Whether it's an Oracle JDK/JRE
      "java_version": "17.0.8.1",           // Version if evaluated
      "java_version_major": 17,             // Major version if evaluated
      "java_version_update": 8,             // Update version if evaluated
      "exec_failed": false,                 // True if evaluation failed
      "require_license": false              // Whether commercial license is required
    }
  ]
}
```

When using `-post`, this JSON will be sent to the specified server URL. The server will respond with:
```json
{
  "result": "ok",
  "scan_id": "<unique_scan_id>"
}
