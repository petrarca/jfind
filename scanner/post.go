package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

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
