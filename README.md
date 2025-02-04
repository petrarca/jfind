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
./jfind [options]
```

### Options

- `-path string`: Start path for searching (default ".")
- `-depth int`: Maximum depth to search (-1 for unlimited)
- `-verbose`: Enable verbose output
- `-eval`: Evaluate found Java executables by running java -version
- `-json`: Output results in JSON format on stdout

### Examples

Find Java installations in current directory:
```bash
./jfind
```

Find and evaluate Java installations with JSON output:
```bash
./jfind -eval -json
```

Search in specific directory with depth limit:
```bash
./jfind -path /usr/local -depth 2
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

#### JSON Output
```json
{
  "meta": {
    "scan_ts": "2025-02-04T13:27:25Z",
    "computer_name": "Stanford",
    "user_name": "username",
    "scan_duration": "PT2.345S",
    "has_oracle_jdk": true,
    "count_result": 1
  },
  "result": [
    {
      "java.executable": "/path/to/java",
      "java.version": "11.0.12",
      "java.vendor": "Oracle Corporation",
      "java.runtime.name": "Java(TM) SE Runtime Environment",
      "is_oracle": true    // Set to true when Oracle JDK is detected
    }
  ]
}
```

The `is_oracle` field in JSON output will be:
- `true` when an Oracle JDK is detected
- `false` for other JDK vendors
- Omitted when `-eval` is not specified

## Development

### Running Tests
```bash
make test
```

### Cleaning Build Artifacts
```bash
make clean
