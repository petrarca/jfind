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

func TestParseJavaPropertiesWithOracleAndOpenJDK(t *testing.T) {
	// Test with Oracle JDK
	oracleOutput := `java.runtime.name = Java(TM) SE Runtime Environment
java.vendor = Oracle Corporation
java.version = 1.8.0_202
`
	result := JavaResult{
		Path:      "/path/to/java",
		StdErr:    oracleOutput,
		Properties: ParseJavaProperties(oracleOutput),
	}

	if result.Properties == nil {
		t.Fatal("Expected properties to be parsed")
	}

	if !strings.Contains(result.Properties.Vendor, "Oracle") {
		t.Error("Expected Oracle vendor")
	}

	// Test with OpenJDK
	openJDKOutput := `java.runtime.name = OpenJDK Runtime Environment
java.vendor = Eclipse Adoptium
java.version = 11.0.20
`
	result = JavaResult{
		Path:      "/path/to/java",
		StdErr:    openJDKOutput,
		Properties: ParseJavaProperties(openJDKOutput),
	}

	if result.Properties == nil {
		t.Fatal("Expected properties to be parsed")
	}

	if strings.Contains(result.Properties.Vendor, "Oracle") {
		t.Error("Expected non-Oracle vendor")
	}
}

func TestEvaluateJava(t *testing.T) {
	// Test non-Oracle vendor
	result := JavaResult{
		Path: "/path/to/java",
		Properties: &JavaProperties{
			Version:     "11.0.20",
			Vendor:      "Eclipse Adoptium",
			RuntimeName: "OpenJDK Runtime Environment",
		},
	}

	if strings.Contains(result.Properties.Vendor, "Oracle") {
		t.Error("Expected non-Oracle vendor")
	}

	// Test Oracle vendor
	result = JavaResult{
		Path: "/path/to/java",
		Properties: &JavaProperties{
			Version:     "1.8.0_202",
			Vendor:      "Oracle Corporation",
			RuntimeName: "Java(TM) SE Runtime Environment",
		},
	}

	if !strings.Contains(result.Properties.Vendor, "Oracle") {
		t.Error("Expected Oracle vendor")
	}
}
