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
jfind [options] [-post[=URL]]
```

### Options

- `-path string`: Start path for searching (default ".")
- `-depth int`: Maximum depth to search (-1 for unlimited)
- `-verbose`: Enable verbose output
- `-eval`: Evaluate found java executables
- `-json`: Output results in JSON format
- `-post`: Post JSON output to server (implies --json)
- `-url string`: URL to post JSON output to (only used with --post, default http://localhost:8000/api/jfind)
- `-require-license`: Filter only Java runtimes that require a commercial license (requires -eval)

### Examples

Find Java installations in current directory:
```bash
jfind
```

Find and evaluate Java installations with JSON output:
```bash
jfind -path /usr/lib/jvm -eval -json
```

Find Java installations requiring commercial license:
```bash
jfind -path /opt -eval -require-license
```

Find Java installations requiring license and post to server:
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
Warning: Oracle vendor detected
```

When an Oracle JDK is detected, a warning message is printed to alert the user.

#### JSON Output (-json or -post)

When using `-json`, the output is written to stdout in JSON format, making it easy to pipe the results to other JSON processing tools like `jq`. When using `-post`, the JSON is sent to the specified server instead.

Examples of JSON processing:
```bash
# Find Oracle JDKs and extract their paths
jfind -eval -json | jq -r '.result[] | select(.is_oracle==true) | .java_executable'

# Count Java installations requiring license
jfind -eval -json | jq '.result[] | select(.require_license==true) | length'

# Get all Java major versions found
jfind -eval -json | jq -r '.result[].java_version_major' | sort -u
```

The JSON output includes metadata about the scan and the results:

```json
{
  "meta": {
    "scan_ts": "2025-02-04T15:12:01Z",      // Scan timestamp in UTC
    "computer_name": "hostname",             // Name of the computer
    "user_name": "username",                 // Name of the user
    "scan_duration": "PT2.345S",            // Duration in ISO8601 format
    "has_oracle_jdk": false,                // Whether Oracle JDK was found
    "count_result": 2,                      // Number of Java installations found
    "count_require_license": 1,             // Number of Java installations requiring license
    "scanned_dirs": 56                      // Number of directories scanned
  },
  "result": [
    {
      "java_executable": "/path/to/java",    // Path to Java executable
      "java_version": "11.0.20",            // Full Java version string (if -eval used)
      "java_vendor": "Oracle Corporation",   // Java vendor (if -eval used)
      "java_runtime": "Java(TM) SE Runtime", // Runtime name (if -eval used)
      "is_oracle": true,                     // Whether it's Oracle Java
      "java_version_major": 11,              // Major version number (8 for 1.8.0, 11 for 11.0.20)
      "java_version_update": 20,             // Update version number (202 for 1.8.0_202, 20 for 11.0.20)
      "exec_failed": true,                   // Present and true if java -version execution failed
      "require_license": true                // Present if license requirement is determined (true/false)
    }
  ]
}
```

The version fields follow Java's version scheme:
- For Java 8 and earlier (e.g., "1.8.0_202"):
  - `java_version_major` = 8
  - `java_version_update` = 202
- For Java 9+ (e.g., "11.0.20"):
  - `java_version_major` = 11
  - `java_version_update` = 20

## Development

### Running Tests
```bash
task test
```

### Cleaning Build Artifacts
```bash
task clean
