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

	"github.com/dustin/go-humanize"
)

// JavaFinder represents a finder for Java executables
type JavaFinder struct {
	startPath string
	maxDepth  int
	evaluate  bool
	scanned   atomic.Int64
	found     atomic.Int64
	ticker    atomic.Bool
	done      chan struct{}
}

// NewJavaFinder creates a new JavaFinder instance
func NewJavaFinder(startPath string, maxDepth int, evaluate bool) *JavaFinder {
	f := &JavaFinder{
		startPath: startPath,
		maxDepth:  maxDepth,
		evaluate:  evaluate,
		done:      make(chan struct{}),
	}
	f.scanned.Store(0)
	f.found.Store(0)
	f.ticker.Store(false)
	return f
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
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				f.ticker.Store(true)
				scanned := f.scanned.Load()
				found := f.found.Load()
				// no linefeed, so progress report stay on same output line
				logf("\rScanned %s directories, found %d java executables.", humanize.Comma(scanned), found)
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
			if f.ticker.Load() {
				logf("\n")
			}
			logf("Permission denied: %s\n", path)
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
	fmt.Printf("Java executable: %s\n", result.Path)

	if !result.Evaluated {
		return
	}

	if result.Error != nil || result.ReturnCode != 0 {
		fmt.Printf("Failed to execute: %v\n", result.Error)
		if result.ReturnCode != 0 {
			fmt.Printf("Exit code: %d\n", result.ReturnCode)
		}
		return
	}

	if result.Properties != nil {
		fmt.Printf("Java version: %s\n", result.Properties.Version)
		fmt.Printf("Java vendor: %s\n", result.Properties.Vendor)
		fmt.Printf("Java runtime name: %s\n", result.Properties.RuntimeName)
		if result.Properties.Major > 0 {
			fmt.Printf("Java major version: %d\n", result.Properties.Major)
		}
		if result.Properties.Update > 0 {
			fmt.Printf("Java update version: %d\n", result.Properties.Update)
		}
	}

	if runtime != nil {
		if runtime.IsOracle {
			fmt.Printf("Info: Oracle JDK/JRE detected\n")
		}

		if runtime.RequireLicense != nil && *runtime.RequireLicense {
			fmt.Printf("Warning: This Java runtime requires a commercial license\n")
		} else {
			fmt.Printf("This Java runtime does not require a commercial license\n")
		}
	}
}
