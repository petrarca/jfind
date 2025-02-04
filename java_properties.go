package main

import (
	"bufio"
	"strconv"
	"strings"
)

// JavaProperties represents properties parsed from java -version output
type JavaProperties struct {
	Version     string
	Vendor      string
	RuntimeName string
	Major       int
	Update      int
}

// ParseJavaProperties parses the output of java -XshowSettings:properties -version
func ParseJavaProperties(input string) *JavaProperties {
	props := &JavaProperties{}

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Look for property lines that contain "="
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			switch key {
			case "java.version":
				props.Version = value
			case "java.vendor":
				props.Vendor = value
			case "java.runtime.name":
				props.RuntimeName = value
			}
		}
	}

	// Parse version components
	if props.Version != "" {
		props.Major, props.Update = parseJavaVersion(props.Version)
	}

	return props
}

// parseJavaVersion extracts major and update versions from Java version string
func parseJavaVersion(version string) (major, update int) {
	// Handle pre-Java 9 versions (1.8.0_202)
	if strings.HasPrefix(version, "1.") {
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			major, _ = strconv.Atoi(parts[1])
		}
		// Find update version after "_"
		if idx := strings.Index(version, "_"); idx != -1 {
			update, _ = strconv.Atoi(version[idx+1:])
		}
		return
	}

	// Handle Java 9+ versions (11.0.20, 17.0.1, etc.)
	parts := strings.Split(version, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 3 {
		update, _ = strconv.Atoi(parts[2])
	}
	return
}
