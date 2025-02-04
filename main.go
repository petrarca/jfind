package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// JavaFinder represents a finder for Java executables
type JavaFinder struct {
	startPath string
	maxDepth  int // -1 means unlimited
	verbose   bool
	evaluate  bool
}

// JavaResult represents the result of evaluating a Java executable
type JavaResult struct {
	Path       string
	Properties *JavaProperties
	Warnings   []string
	StdErr     string
	ReturnCode int
	Error      error
}

// NewJavaFinder creates a new JavaFinder instance
func NewJavaFinder(startPath string, maxDepth int, verbose bool, evaluate bool) *JavaFinder {
	return &JavaFinder{
		startPath: startPath,
		maxDepth:  maxDepth,
		verbose:   verbose,
		evaluate:  evaluate,
	}
}

// logf prints to stderr
func logf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// printf prints to stdout
func printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// isExecutable checks if a file is executable based on the operating system
func isExecutable(info os.FileInfo) bool {
	if runtime.GOOS == "windows" {
		// On Windows, we only check if it's a regular file
		return !info.IsDir()
	}
	// On Unix-like systems, check for executable permission
	return info.Mode()&0111 != 0
}

// isJavaExecutable checks if the filename matches java executable patterns
func isJavaExecutable(name string) bool {
	if runtime.GOOS == "windows" {
		return name == "java.exe"
	}
	return name == "java"
}

// getPathDepth returns the depth of a path relative to the start path
func (f *JavaFinder) getPathDepth(path string) int {
	relPath, err := filepath.Rel(f.startPath, path)
	if err != nil {
		return 0
	}
	if relPath == "." {
		return 0
	}
	return len(strings.Split(relPath, string(os.PathSeparator)))
}

// evaluateJava runs java -version and returns the result
func (f *JavaFinder) evaluateJava(javaPath string) JavaResult {
	result := JavaResult{
		Path: javaPath,
	}

	cmd := exec.Command(javaPath, "-XshowSettings:properties", "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ReturnCode = exitError.ExitCode()
		}
		result.Error = err
	} else {
		result.ReturnCode = 0
	}

	// Java outputs properties and version info to stderr
	result.StdErr = stderr.String()
	result.Properties = ParseJavaProperties(stderr.String())

	// Check for Oracle vendor
	if result.Properties != nil && strings.Contains(result.Properties.Vendor, "Oracle") {
		result.Warnings = append(result.Warnings, "Warning: Oracle vendor detected")
	}

	return result
}

// printResult prints the results of evaluating a Java executable
func printResult(result *JavaResult) {
	if result.Error != nil {
		printf("Failed to execute: %v\n", result.Error)
		return
	}

	if result.ReturnCode != 0 {
		printf("Command failed with return code: %d\n", result.ReturnCode)
		return
	}

	if result.Properties != nil {
		printf("Java version: %s\n", result.Properties.Version)
		printf("Java vendor: %s\n", result.Properties.Vendor)
		if result.Properties.RuntimeName != "" {
			printf("Java runtime name: %s\n", result.Properties.RuntimeName)
		}
	}
	
	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			printf("%s\n", warning)
		}
	}
}

// Find searches for java executables starting from the specified path
func (f *JavaFinder) Find() error {
	logf("Start looking for java in %s (scanning subdirectories)\n", f.startPath)

	return filepath.Walk(f.startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors, continue walking
		}

		// Print directory being scanned in verbose mode
		if f.verbose && info.IsDir() {
			logf("Scanning directory: %s\n", path)
		}

		// Check depth limit if set
		if f.maxDepth >= 0 {
			depth := f.getPathDepth(path)
			if depth > f.maxDepth {
				if info.IsDir() {
					if f.verbose {
						logf("Skipping directory (max depth reached): %s\n", path)
					}
					return filepath.SkipDir // skip this directory and its children
				}
				return nil
			}
		}

		// Check if it's a file and has the name we're looking for
		if !info.IsDir() && isJavaExecutable(info.Name()) {
			// Check if the file is executable based on the OS
			if isExecutable(info) {
				if f.evaluate {
					// If evaluating, print path as part of evaluation result
					printf("\nJava executable: %s\n", path)
					result := f.evaluateJava(path)

					printResult(&result)
				} else {
					// If not evaluating, just print the path
					fmt.Println(path)
				}
			} else if f.verbose {
				logf("Found non-executable java file: %s\n", path)
			}
		}
		return nil
	})
}

func main() {
	// Parse command line arguments
	startPath := flag.String("path", ".", "Starting path for the search (--path /some/path)")
	maxDepth := flag.Int("depth", -1, "Maximum directory depth to search, -1 for unlimited (--depth 2)")
	verbose := flag.Bool("v", false, "Verbose mode - print directories being scanned (--v)")
	evaluate := flag.Bool("eval", false, "Evaluate found Java executables by running java -version (--eval)")

	// Add custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Find Java executables and optionally evaluate their version.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample usage:\n")
		fmt.Fprintf(os.Stderr, "  %s --path /usr/local --depth 2 --eval\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --v --eval\n", os.Args[0])
	}

	flag.Parse()

	// Get absolute path
	absPath, err := filepath.Abs(*startPath)
	if err != nil {
		logf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Create and run finder
	finder := NewJavaFinder(absPath, *maxDepth, *verbose, *evaluate)
	if err := finder.Find(); err != nil {
		logf("Error during search: %v\n", err)
		os.Exit(1)
	}
}
