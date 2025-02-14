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
