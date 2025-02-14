package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// JavaFinder represents a finder for Java executables
type JavaFinder struct {
	startPath string
	maxDepth  int
	verbose   bool
	evaluate  bool
	scanned   atomic.Int64
	found     atomic.Int64
	done      chan struct{}
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

// getPathDepth returns the depth of a path relative to the start path
func (f *JavaFinder) getPathDepth(path string) int {
	if !strings.HasPrefix(path, f.startPath) {
		return -1
	}

	relPath := strings.TrimPrefix(path, f.startPath)
	if relPath == "" {
		return 0
	}

	return len(strings.Split(strings.Trim(relPath, string(os.PathSeparator)), string(os.PathSeparator)))
}

// evaluateJava runs java -version and returns the result
func (f *JavaFinder) evaluateJava(javaPath string) JavaResult {
	result := JavaResult{
		Path: javaPath,
	}

	if !f.evaluate {
		return result
	}

	cmd := exec.Command(javaPath, "-XshowSettings:properties", "-version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	result.StdErr = stderr.String()
	result.Error = err
	result.ReturnCode = 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ReturnCode = exitError.ExitCode()
		}
	}

	if err == nil && result.ReturnCode == 0 {
		result.Properties = ParseJavaProperties(result.StdErr)
	}

	result.Evaluated = true
	return result
}

// startProgressReporting starts a goroutine to report progress periodically
func (f *JavaFinder) startProgressReporting() {
	if !f.verbose {
		return
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				scanned := f.scanned.Load()
				found := f.found.Load()
				logf("\rScanned %d directories, found %d java executables", scanned, found)
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
		return err
	}

	// Skip .git directories
	if info.IsDir() && info.Name() == ".git" {
		return filepath.SkipDir
	}

	// Check depth
	if f.maxDepth >= 0 {
		depth := f.getPathDepth(path)
		if depth > f.maxDepth {
			return filepath.SkipDir
		}
	}

	// Update progress
	if info.IsDir() {
		f.scanned.Add(1)
	}

	return nil
}

// Find searches for java executables starting from the specified path
func (f *JavaFinder) Find() ([]*JavaResult, error) {
	results := make([]*JavaResult, 0)

	f.startProgressReporting()
	defer close(f.done)

	err := filepath.Walk(f.startPath, func(path string, info os.FileInfo, err error) error {
		if err := f.handleDirectory(path, info, err); err != nil {
			return err
		}

		if result := f.evaluateFile(path, info); result != nil {
			f.found.Add(1)
			results = append(results, result)
		}

		return nil
	})

	if f.verbose {
		fmt.Println()
	}

	return results, err
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
		if result.Properties.Major > 0 {
			printf("Java major version: %d\n", result.Properties.Major)
		}
		if result.Properties.Update > 0 {
			printf("Java update version: %d\n", result.Properties.Update)
		}
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
