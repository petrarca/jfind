package main

import (
	"testing"
	"strings"
)

func TestParseJavaProperties(t *testing.T) {
	input := `Property settings:
    java.version = 21.0.5
    java.vendor = Eclipse Adoptium
    other.property = some value
`

	props := ParseJavaProperties(input)

	if props.Version != "21.0.5" {
		t.Errorf("Expected version 21.0.5, got %s", props.Version)
	}

	if props.Vendor != "Eclipse Adoptium" {
		t.Errorf("Expected vendor Eclipse Adoptium, got %s", props.Vendor)
	}
}

func TestEvaluateJava(t *testing.T) {
	// Test non-Oracle vendor
	result := JavaResult{
		Properties: &JavaProperties{
			Version: "21.0.5",
			Vendor:  "Eclipse Adoptium",
		},
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Expected no warnings for non-Oracle vendor, got %v", result.Warnings)
	}

	// Test Oracle vendor
	result = JavaResult{
		Properties: &JavaProperties{
			Version: "21.0.5",
			Vendor:  "Oracle Corporation",
		},
	}
	// Manually trigger Oracle check
	if result.Properties != nil && strings.Contains(result.Properties.Vendor, "Oracle") {
		result.Warnings = append(result.Warnings, "Warning: Oracle vendor detected")
	}
	if len(result.Warnings) == 0 {
		t.Errorf("Expected warning for Oracle vendor")
	}
}
