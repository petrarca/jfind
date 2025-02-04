package main

import (
	"bufio"
	"strings"
)

type JavaProperties struct {
	Version string
	Vendor  string
}

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
			}
		}
	}

	return props
}
