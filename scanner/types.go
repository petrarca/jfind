package main

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
