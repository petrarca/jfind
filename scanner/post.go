package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
)

// sendJSON sends the JSON payload to the specified URL via HTTP POST
func sendJSON(jsonData []byte, urlStr string) error {
	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil || !parsedURL.IsAbs() {
		return fmt.Errorf("invalid URL %s: %v", urlStr, err)
	}

	resp, err := http.Post(urlStr, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Check if it's a connection error
		if netErr, ok := err.(*net.OpError); ok {
			return fmt.Errorf("failed to connect to server at %s: %v", urlStr, netErr)
		}
		return fmt.Errorf("failed to send JSON to %s: %v", urlStr, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", err)
		}
	}()

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
		if _, err := os.Stdout.Write(body); err != nil {
			return fmt.Errorf("failed to write response: %v", err)
		}
		fmt.Println()
	}

	return nil
}
