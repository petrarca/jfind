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
3. Build using make:

```bash
make build        # Build for current platform
make build-all    # Build for all supported platforms
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
      "java_version": "1.8.0_292",          // Java version (if -eval used)
      "java_vendor": "Oracle Corporation",   // Java vendor (if -eval used)
      "java_runtime": "Java(TM) SE Runtime", // Runtime name (if -eval used)
      "is_oracle": true                      // Whether it's Oracle Java
    }
  ]
}
```

## Development

### Running Tests
```bash
make test
```

### Cleaning Build Artifacts
```bash
make clean
