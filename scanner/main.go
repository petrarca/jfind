package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultPostURL = "http://localhost:8000/api/jfind"

type config struct {
	startPath      string
	maxDepth       int
	verbose        bool
	evaluate       bool
	jsonOutput     bool
	doPost         bool
	postURL        string
	requireLicense bool
	help           bool
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

func parseFlags() config {
	var config config

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
	flag.BoolVar(&config.help, "h", false, "Show help message")
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
		ScannedDirs:         int(finder.scanned.Load()),
		ScanPath:            startPath,
	}
}

func handleJSONOutput(results []*JavaResult, finder *JavaFinder, config config, startTime time.Time) error {
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

func handleRegularOutput(results []*JavaResult, config config) {
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
