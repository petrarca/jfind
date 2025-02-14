package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	defaultPostURL = "http://localhost:8000/api/jfind"
)

// JavaFinder represents a finder for Java executables
type JavaFinder struct {
	startPath string
	maxDepth  int // -1 means unlimited
	verbose   bool
	evaluate  bool
	scanned   atomic.Int64
	found     atomic.Int64
	done      chan struct{}
}

// JavaResult represents the result of evaluating a Java executable
type JavaResult struct {
	Path       string
	Properties *JavaProperties
	StdErr     string
	ReturnCode int
	Error      error
	Evaluated  bool
}

// JavaRuntimeJSON represents a single Java runtime for JSON output
type JavaRuntimeJSON struct {
	JavaExecutable string `json:"java_executable"`
	JavaRuntime    string `json:"java_runtime,omitempty"`
	JavaVendor     string `json:"java_vendor,omitempty"`
	IsOracle       bool   `json:"is_oracle,omitempty"`
	JavaVersion    string `json:"java_version,omitempty"`
	VersionMajor   int    `json:"java_version_major,omitempty"`
	VersionUpdate  int    `json:"java_version_update,omitempty"`
	ExecFailed     bool   `json:"exec_failed,omitempty"`
	RequireLicense *bool  `json:"require_license,omitempty"`
}

// MetaInfo represents metadata about the scan
type MetaInfo struct {
	ScanTimestamp       string `json:"scan_ts"`
	ComputerName        string `json:"computer_name"`
	UserName            string `json:"user_name"`
	ScanDuration        string `json:"scan_duration"`
	HasOracleJDK        bool   `json:"has_oracle_jdk"`
	CountResult         int    `json:"count_result"`
	CountRequireLicense int    `json:"count_require_license"`
	ScannedDirs         int    `json:"scanned_dirs"`
	ScanPath            string `json:"scan_path"`
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
		done:      make(chan struct{}),
	}
}

// logf prints to stderr
func logf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// log prints to stderr
func log(a ...interface{}) {
	fmt.Fprint(os.Stderr, a...)
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
		Path:      javaPath,
		Evaluated: true,
	}

	cmd := exec.Command(javaPath, "-XshowSettings:properties", "-version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	result.Error = cmd.Run()
	result.ReturnCode = 0
	if exitError, ok := result.Error.(*exec.ExitError); ok {
		result.ReturnCode = exitError.ExitCode()
	}

	result.StdErr = stderr.String()
	if result.Error == nil && result.ReturnCode == 0 {
		result.Properties = ParseJavaProperties(result.StdErr)
	}

	return result
}

// printResult prints the results of evaluating a Java executable
func printResult(result *JavaResult, runtime *JavaRuntimeJSON) {
	printf("Java executable: %s\n", result.Path)

	if !result.Evaluated {
		return
	}

	if result.Error != nil || result.ReturnCode != 0 {
		printf("Failed to execute: %v\n", result.Error)
		if result.ReturnCode != 0 {
			printf("Exit code: %d\n", result.ReturnCode)
		}
		return
	}

	if result.Properties != nil {
		printf("Java version: %s\n", result.Properties.Version)
		printf("Java vendor: %s\n", result.Properties.Vendor)
		printf("Java runtime name: %s\n", result.Properties.RuntimeName)
		printf("Java major version: %d\n", result.Properties.Major)
		printf("Java update version: %d\n", result.Properties.Update)
	}

	if runtime != nil {
		if runtime.IsOracle {
			printf("Info: Oracle JDK/JRE detected\n")
		}

		if runtime.RequireLicense != nil && *runtime.RequireLicense {
			printf("Warning: This Java runtime requires a commercial license\n")
		} else {
			printf("This Java runtime does not require a commercial license\n")
		}
	}
}

// startProgressReporting starts a goroutine to report progress periodically
func (f *JavaFinder) startProgressReporting() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var builder strings.Builder
				builder.WriteString(fmt.Sprintf("Scanned %s directories", humanize.Comma(f.scanned.Load())))
				if f.found.Load() > 0 {
					builder.WriteString(fmt.Sprintf(" (%s JDKs/JREs found)", humanize.Comma(f.found.Load())))
				}
				builder.WriteString("...\n")
				log(builder.String())
			case <-f.done:
				return
			}
		}
	}()
}

// evaluateFile checks if a file is a Java executable and evaluates it if required
func (f *JavaFinder) evaluateFile(path string, info os.FileInfo) *JavaResult {
	if info == nil {
		return nil
	}
	if !info.IsDir() && isJavaExecutable(info.Name()) && isExecutable(info) {
		if f.evaluate {
			result := f.evaluateJava(path)
			return &result
		}
		return &JavaResult{Path: path}
	}
	return nil
}

// handleDirectory processes a directory during the walk
func (f *JavaFinder) handleDirectory(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsPermission(err) {
			if f.verbose {
				logf("Permission denied: %s\n", path)
			}
			return filepath.SkipDir
		}
		// Skip other errors but log them in verbose mode
		if f.verbose {
			logf("Error accessing %s: %v\n", path, err)
		}
		return nil
	}

	// Print directory being scanned in verbose mode
	if f.verbose && info.IsDir() {
		logf("Scanning: %s\n", path)
	}

	// Count directories as we scan
	if info.IsDir() {
		f.scanned.Add(1)
	}

	// Check depth
	if f.maxDepth >= 0 && f.getPathDepth(path) > f.maxDepth {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	return nil
}

// Find searches for java executables starting from the specified path
func (f *JavaFinder) Find() ([]*JavaResult, error) {
	f.scanned.Store(0)
	f.found.Store(0)

	if f.verbose {
		logf("Start looking for java in %s (scanning subdirectories)\n", f.startPath)
	}
	var results []*JavaResult

	// Start progress reporting
	f.startProgressReporting()

	err := filepath.Walk(f.startPath, func(path string, info os.FileInfo, err error) error {
		// Handle directory first
		if err := f.handleDirectory(path, info, err); err != nil {
			return err
		}

		// Evaluate file if it exists
		if result := f.evaluateFile(path, info); result != nil {
			results = append(results, result)
			f.found.Add(1)
		}

		return nil
	})

	close(f.done)
	return results, err
}

// sendJSON sends the JSON payload to the specified URL via HTTP POST
func sendJSON(jsonData []byte, url string) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Check if it's a connection error
		if netErr, ok := err.(*net.OpError); ok {
			return fmt.Errorf("failed to connect to server at %s: %v", url, netErr)
		}
		return fmt.Errorf("failed to send JSON to %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if len(body) > 0 {
			return fmt.Errorf("server returned %s: %s", resp.Status, string(body))
		}
		return fmt.Errorf("server returned %s", resp.Status)
	}

	// Write response JSON directly to stdout
	if len(body) > 0 {
		os.Stdout.Write(body)
	}

	return nil
}

// parseFlags parses command line flags and returns the configuration
func parseFlags() (config struct {
	startPath      string
	maxDepth       int
	verbose        bool
	evaluate       bool
	jsonOutput     bool
	doPost         bool
	postURL        string
	requireLicense bool
	help           bool
}) {
	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Define flags
	flag.StringVar(&config.startPath, "path", ".", "Start path for searching")
	flag.IntVar(&config.maxDepth, "depth", -1, "Maximum depth to search (-1 for unlimited)")
	flag.BoolVar(&config.verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.evaluate, "eval", false, "Evaluate found java executables")
	flag.BoolVar(&config.jsonOutput, "json", false, "Output results in JSON format")
	flag.BoolVar(&config.doPost, "post", false, "Post JSON output to server (implies --json)")
	flag.StringVar(&config.postURL, "url", defaultPostURL, "URL to post JSON output to (only used with --post)")
	flag.BoolVar(&config.requireLicense, "require-license", false, "Filter only Java runtimes that require a commercial license")

	// Add help flags
	flag.BoolVar(&config.help, "help", false, "Show help message")

	flag.Parse()

	// If help is requested, print usage and exit
	if config.help {
		flag.Usage()
		os.Exit(0)
	}

	if config.doPost {
		config.jsonOutput = true
	}

	return config
}

// createMetaInfo creates the metadata information for JSON output
func createMetaInfo(startPath string, results []*JavaResult, finder *JavaFinder, startTime time.Time) MetaInfo {
	currentUser, _ := user.Current()
	username := "unknown"
	if currentUser != nil {
		username = currentUser.Username
	}

	duration := formatDurationISO8601(time.Since(startTime))
	return MetaInfo{
		ScanTimestamp:       time.Now().UTC().Format(time.RFC3339),
		ComputerName:        getComputerName(),
		UserName:            username,
		ScanDuration:        duration,
		HasOracleJDK:        false, // Will be updated later
		CountResult:         len(results),
		CountRequireLicense: 0, // Will be updated later
		ScannedDirs:        int(finder.scanned.Load()),
		ScanPath:           startPath,
	}
}

// createRuntimeJSON creates a JavaRuntimeJSON from a JavaResult
func createRuntimeJSON(result *JavaResult, evaluate bool) JavaRuntimeJSON {
	runtime := JavaRuntimeJSON{
		JavaExecutable: result.Path,
	}

	if evaluate && result.Properties != nil && result.Error == nil && result.ReturnCode == 0 {
		runtime.JavaVersion = result.Properties.Version
		runtime.JavaVendor = result.Properties.Vendor
		runtime.JavaRuntime = result.Properties.RuntimeName
		runtime.IsOracle = strings.Contains(result.Properties.Vendor, "Oracle")
		runtime.VersionMajor = result.Properties.Major
		runtime.VersionUpdate = result.Properties.Update
		runtime.checkLicenseRequirement()
	} else if evaluate && (result.Error != nil || result.ReturnCode != 0) {
		runtime.ExecFailed = true
	}

	return runtime
}

// handleJSONOutput processes results and outputs them in JSON format
func handleJSONOutput(results []*JavaResult, finder *JavaFinder, config struct {
	startPath      string
	maxDepth       int
	verbose        bool
	evaluate       bool
	jsonOutput     bool
	doPost         bool
	postURL        string
	requireLicense bool
	help           bool
}, startTime time.Time) error {
	output := JSONOutput{
		Meta:     createMetaInfo(config.startPath, results, finder, startTime),
		Runtimes: make([]JavaRuntimeJSON, 0, len(results)),
	}

	hasOracle := false
	countRequireLicense := 0

	// Process each result
	for _, result := range results {
		runtime := createRuntimeJSON(result, config.evaluate)

		if config.requireLicense && (runtime.RequireLicense == nil || !*runtime.RequireLicense) {
			continue
		}

		if runtime.IsOracle {
			hasOracle = true
		}

		if runtime.RequireLicense != nil && *runtime.RequireLicense {
			countRequireLicense++
		}

		output.Runtimes = append(output.Runtimes, runtime)
	}

	// Update meta information
	output.Meta.HasOracleJDK = hasOracle
	output.Meta.CountRequireLicense = countRequireLicense

	// Convert to JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Handle output
	if config.doPost {
		if err := sendJSON(jsonData, config.postURL); err != nil {
			return fmt.Errorf("error sending JSON: %v", err)
		}
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// handleRegularOutput processes results and outputs them in regular format
func handleRegularOutput(results []*JavaResult, config struct {
	startPath      string
	maxDepth       int
	verbose        bool
	evaluate       bool
	jsonOutput     bool
	doPost         bool
	postURL        string
	requireLicense bool
	help           bool
}) {
	for _, result := range results {
		var runtime *JavaRuntimeJSON
		if config.evaluate && result.Properties != nil && result.Error == nil && result.ReturnCode == 0 {
			runtime = &JavaRuntimeJSON{
				JavaExecutable: result.Path,
				JavaVersion:    result.Properties.Version,
				JavaVendor:     result.Properties.Vendor,
				JavaRuntime:    result.Properties.RuntimeName,
				IsOracle:       strings.Contains(result.Properties.Vendor, "Oracle"),
				VersionMajor:   result.Properties.Major,
				VersionUpdate:  result.Properties.Update,
			}
			runtime.checkLicenseRequirement()
		}
		printResult(result, runtime)
		printf("\n")
	}
}

func main() {
	config := parseFlags()

	// Convert relative path to absolute
	absPath, err := filepath.Abs(config.startPath)
	if err != nil {
		logf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		logf("Error: path '%s' does not exist\n", absPath)
		os.Exit(1)
	}

	logf("Start scanning (platform '%s') from path '%s'\n", runtime.GOOS, absPath)
	finder := NewJavaFinder(absPath, config.maxDepth, config.verbose, config.evaluate)
	startTime := time.Now()
	results, err := finder.Find()
	if err != nil {
		logf("Error during search: %v\n", err)
		os.Exit(1)
	}

	if config.jsonOutput {
		if err := handleJSONOutput(results, finder, config, startTime); err != nil {
			logf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		handleRegularOutput(results, config)
	}
}
