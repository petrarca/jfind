package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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

// JavaRuntimeJSON represents a single Java runtime for JSON output
type JavaRuntimeJSON struct {
	JavaExecutable string `json:"java.executable"`
	JavaVersion    string `json:"java.version,omitempty"`
	JavaVendor     string `json:"java.vendor,omitempty"`
	JavaRuntime    string `json:"java.runtime.name,omitempty"`
	IsOracle       bool   `json:"is_oracle,omitempty"`
}

// MetaInfo represents metadata about the scan
type MetaInfo struct {
	ScanTimestamp string `json:"scan_ts"`
	ComputerName  string `json:"computer_name"`
	UserName      string `json:"user_name"`
}

// JSONOutput represents the root JSON output structure
type JSONOutput struct {
	Meta     MetaInfo          `json:"meta"`
	Runtimes []JavaRuntimeJSON `json:"result"`
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

	printf("Java executable: %s\n", result.Path)

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

// printJSONResult prints the results in JSON format
func printJSONResult(results []*JavaResult) {
	output := JSONOutput{
		Runtimes: make([]JavaRuntimeJSON, 0),
	}

	for _, result := range results {
		runtime := JavaRuntimeJSON{
			JavaExecutable: result.Path,
		}

		if result.Properties != nil && result.Error == nil && result.ReturnCode == 0 {
			runtime.JavaVersion = result.Properties.Version
			runtime.JavaVendor = result.Properties.Vendor
			runtime.JavaRuntime = result.Properties.RuntimeName
			runtime.IsOracle = strings.Contains(result.Properties.Vendor, "Oracle")
		}

		output.Runtimes = append(output.Runtimes, runtime)
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logf("Error generating JSON output: %v\n", err)
		return
	}
	printf("%s\n", jsonData)
}

// Find searches for java executables starting from the specified path
func (f *JavaFinder) Find() ([]*JavaResult, error) {
	if f.verbose {
		logf("Start looking for java in %s (scanning subdirectories)\n", f.startPath)
	}

	var results []*JavaResult

	err := filepath.Walk(f.startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors, continue walking
		}

		// Print directory being scanned in verbose mode
		if f.verbose && info.IsDir() {
			logf("Scanning: %s\n", path)
		}

		// Check depth
		if f.maxDepth >= 0 && f.getPathDepth(path) > f.maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is executable and matches java pattern
		if !info.IsDir() && isExecutable(info) && isJavaExecutable(filepath.Base(path)) {
			// Always log the executable path to stderr when found
			logf("%s\n", path)

			if f.evaluate {
				result := f.evaluateJava(path)
				results = append(results, &result)
			} else {
				// For non-evaluated executables, create a basic result
				result := JavaResult{
					Path: path,
				}
				results = append(results, &result)
			}
		}
		return nil
	})

	return results, err
}

// getComputerName returns the computer name based on the operating system
func getComputerName() string {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("scutil", "--get", "ComputerName")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "windows":
		cmd := exec.Command("cmd", "/c", "hostname")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "linux":
		// Try to read from /etc/hostname first
		if data, err := os.ReadFile("/etc/hostname"); err == nil {
			return strings.TrimSpace(string(data))
		}
		// Fallback to hostname command
		cmd := exec.Command("hostname")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return "unknown"
}

func getMachineInfo() (MetaInfo, error) {
	info := MetaInfo{}

	// Get computer name using OS-specific method
	info.ComputerName = getComputerName()

	// Get username (works on all platforms)
	currentUser, err := user.Current()
	if err != nil {
		info.UserName = "unknown"
	} else {
		info.UserName = currentUser.Username
	}

	// Get current timestamp in ISO 8601 format
	info.ScanTimestamp = time.Now().UTC().Format(time.RFC3339)

	return info, nil
}

func main() {
	var startPath string
	var maxDepth int
	var verbose bool
	var evaluate bool
	var jsonOutput bool

	flag.StringVar(&startPath, "path", ".", "Start path for searching")
	flag.IntVar(&maxDepth, "depth", -1, "Maximum depth to search (-1 for unlimited)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&evaluate, "eval", false, "Evaluate found java executables")
	flag.BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	flag.Parse()

	// Convert relative path to absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		logf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	finder := NewJavaFinder(absPath, maxDepth, verbose, evaluate)
	results, err := finder.Find()
	if err != nil {
		logf("Error during search: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		// Get meta information
		meta, err := getMachineInfo()
		if err != nil {
			logf("Warning: Could not get complete machine info: %v\n", err)
		}

		output := JSONOutput{
			Meta:     meta,
			Runtimes: make([]JavaRuntimeJSON, 0),
		}

		for _, result := range results {
			runtime := JavaRuntimeJSON{
				JavaExecutable: result.Path,
			}

			if evaluate && result.Properties != nil && result.Error == nil && result.ReturnCode == 0 {
				runtime.JavaVersion = result.Properties.Version
				runtime.JavaVendor = result.Properties.Vendor
				runtime.JavaRuntime = result.Properties.RuntimeName
				runtime.IsOracle = strings.Contains(result.Properties.Vendor, "Oracle")
			}

			output.Runtimes = append(output.Runtimes, runtime)
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			logf("Error generating JSON output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		for _, result := range results {
			printResult(result)
			printf("\n")
		}
	}
}
