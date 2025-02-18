package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

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

// getComputerName returns the computer name, or "unknown" if it cannot be determined
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

func isExecutable(info os.FileInfo) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}

func isJavaExecutable(name string) bool {
	name = strings.ToLower(name)
	return name == "java" || name == "java.exe"
}

func printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// logf writes formatted output to stderr
func logf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// log writes output to stderr
/*
func log(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}
*/

// getPlatformInfo returns platform information in a parsable string format
// Format: OS=<os>;Version=<version>;Arch=<arch>;[Extra=<extra>]
func getPlatformInfo() string {
	var info []string
	info = append(info, fmt.Sprintf("OS=%s", runtime.GOOS))
	info = append(info, fmt.Sprintf("Arch=%s", runtime.GOARCH))

	// Get OS version based on platform
	switch runtime.GOOS {
	case "darwin":
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			version := strings.TrimSpace(string(out))
			info = append(info, fmt.Sprintf("Version=%s", version))
		}
		// Get macOS codename (e.g., Ventura)
		if out, err := exec.Command("sw_vers", "-productName").Output(); err == nil {
			name := strings.TrimSpace(string(out))
			info = append(info, fmt.Sprintf("Name=%s", name))
		}
	case "linux":
		// Try to get Linux distribution info
		if out, err := exec.Command("lsb_release", "-d").Output(); err == nil {
			desc := strings.TrimSpace(string(out))
			if parts := strings.SplitN(desc, ":", 2); len(parts) == 2 {
				info = append(info, fmt.Sprintf("Name=%s", strings.TrimSpace(parts[1])))
			}
		}
		// Try to get Linux version
		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			version := strings.TrimSpace(string(out))
			info = append(info, fmt.Sprintf("Version=%s", version))
		}
	case "windows":
		// Get Windows version using PowerShell
		cmd := exec.Command("powershell", "-Command", "(Get-WmiObject -class Win32_OperatingSystem).Version")
		if out, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(out))
			info = append(info, fmt.Sprintf("Version=%s", version))
		}
		// Get Windows edition
		cmd = exec.Command("powershell", "-Command", "(Get-WmiObject -class Win32_OperatingSystem).Caption")
		if out, err := cmd.Output(); err == nil {
			name := strings.TrimSpace(string(out))
			info = append(info, fmt.Sprintf("Name=%s", name))
		}
	}

	return strings.Join(info, ";")
}
