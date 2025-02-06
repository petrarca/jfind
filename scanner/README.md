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
- `-post`: Post JSON output to server (implies --json). With optional URL parameter (default http://localhost:8000/jfind)

### Examples

Find Java installations in current directory:
```bash
jfind
```

Find and evaluate Java installations with JSON output:
```bash
jfind -eval -json
```

Search in specific directory with depth limit:
```bash
jfind -path /usr/local -depth 2
```

Find Java installations and post results to default server:
```bash
jfind -eval -post
```

Find Java installations and post results to custom server:
```bash
jfind -eval -post=http://myserver.com:8080/java
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
      "exec_failed": true                    // Present and true if java -version execution failed
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
