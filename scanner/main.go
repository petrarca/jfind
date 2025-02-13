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
func printResult(result *JavaResult) {
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

		if strings.Contains(result.Properties.Vendor, "Oracle") {
			printf("Warning: Oracle JDK detected\n")
		}
	}
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
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var builder strings.Builder
				builder.WriteString(fmt.Sprintf("Scanned %s directories", humanize.Comma(f.scanned.Load())))
				if f.found.Load() > 0 {
					builder.WriteString(fmt.Sprintf(" (%s JDKs/JREs found) ", humanize.Comma(f.found.Load())))
				}
				builder.WriteString("...\n")
				log(builder.String())
			case <-f.done:
				return
			}
		}
	}()

	err := filepath.Walk(f.startPath, func(path string, info os.FileInfo, err error) error {
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

		// Check if file is executable and named 'java' or 'java.exe' depending on OS
		if !info.IsDir() && isJavaExecutable(info.Name()) && isExecutable(info) {
			if f.evaluate {
				result := f.evaluateJava(path)
				results = append(results, &result)
			} else {
				results = append(results, &JavaResult{Path: path})
			}

			f.found.Add(1)
		}

		return nil
	})

	close(f.done)
	return results, err
}

// formatDurationISO8601 formats a duration according to ISO8601 with millisecond precision
func formatDurationISO8601(d time.Duration) string {
	d = d.Round(time.Millisecond)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond

	var result strings.Builder
	result.WriteString("PT")
	if h > 0 {
		result.WriteString(fmt.Sprintf("%dH", h))
	}
	if m > 0 {
		result.WriteString(fmt.Sprintf("%dM", m))
	}
	if s > 0 || ms > 0 || (h == 0 && m == 0) {
		if ms > 0 {
			result.WriteString(fmt.Sprintf("%d.%03dS", s, ms))
		} else {
			result.WriteString(fmt.Sprintf("%dS", s))
		}
	}
	return result.String()
}

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

func main() {
	var startPath string
	var maxDepth int
	var verbose bool
	var evaluate bool
	var jsonOutput bool
	var doPost bool
	var postURL string
	var requireLicense bool

	flag.StringVar(&startPath, "path", ".", "Start path for searching")
	flag.IntVar(&maxDepth, "depth", -1, "Maximum depth to search (-1 for unlimited)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&evaluate, "eval", false, "Evaluate found java executables")
	flag.BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	flag.BoolVar(&doPost, "post", false, "Post JSON output to server (implies --json)")
	flag.StringVar(&postURL, "url", defaultPostURL, "URL to post JSON output to (only used with --post)")
	flag.BoolVar(&requireLicense, "require-license", false, "Filter only Java runtimes that require a commercial license")
	flag.Parse()

	if doPost {
		jsonOutput = true
	}

	// Convert relative path to absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		logf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	logf("Start scanning (platform '%s') from path '%s'\n", runtime.GOOS, absPath)
	finder := NewJavaFinder(absPath, maxDepth, verbose, evaluate)
	startTime := time.Now()
	results, err := finder.Find()
	if err != nil {
		logf("Error during search: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		// Get meta information
		currentUser, _ := user.Current()
		username := "unknown"
		if currentUser != nil {
			username = currentUser.Username
		}

		hasOracle := false
		countRequireLicense := 0
		duration := formatDurationISO8601(time.Since(startTime))
		output := JSONOutput{
			Meta: MetaInfo{
				ScanTimestamp:       time.Now().UTC().Format(time.RFC3339),
				ComputerName:        getComputerName(),
				UserName:            username,
				ScanDuration:        duration,
				HasOracleJDK:        false,
				CountResult:         len(results),
				CountRequireLicense: countRequireLicense,
				ScannedDirs:         int(finder.scanned.Load()),
			},
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
				runtime.VersionMajor = result.Properties.Major
				runtime.VersionUpdate = result.Properties.Update
				if runtime.IsOracle {
					hasOracle = true
				}
			} else if evaluate && (result.Error != nil || result.ReturnCode != 0) {
				runtime.ExecFailed = true
			}

			runtime.checkLicenseRequirement()

			// Skip if require-license is set and either the runtime doesn't require a license or the license requirement is unknown
			if requireLicense && (runtime.RequireLicense == nil || !*runtime.RequireLicense) {
				continue
			}

			output.Runtimes = append(output.Runtimes, runtime)
			if runtime.RequireLicense != nil && *runtime.RequireLicense {
				countRequireLicense++
			}
		}

		// Update hasOracle after scanning all results
		output.Meta.HasOracleJDK = hasOracle
		output.Meta.CountRequireLicense = countRequireLicense

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			logf("Error generating JSON output: %v\n", err)
			os.Exit(1)
		}

		if doPost {
			logf("Posting JSON to %s...\n", postURL)
			if err := sendJSON(jsonData, postURL); err != nil {
				logf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			os.Stdout.Write(jsonData)
		}
	} else {
		for _, result := range results {
			printResult(result)
			printf("\n")
		}
	}
}
