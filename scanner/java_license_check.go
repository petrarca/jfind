package main

import "strings"

// checkLicenseRequirement determines if a commercial license is required for the Java runtime
func (j *JavaRuntimeJSON) checkLicenseRequirement() {
	// Only set RequireLicense for Oracle JDKs
	if !j.IsOracle {
		j.RequireLicense = nil
		return
	}

	// Initialize the pointer if needed
	if j.RequireLicense == nil {
		j.RequireLicense = new(bool)
	}

	// OpenJDK never requires a license
	if j.JavaRuntime != "" && strings.Contains(strings.ToLower(j.JavaRuntime), "openjdk") {
		*j.RequireLicense = false
		return
	}

	if j.JavaRuntime == "OpenJDK Runtime Environment" {
		*j.RequireLicense = false
		return
	}

	// Check for "commercial features" in JavaRuntime string
	if j.JavaRuntime != "" && strings.Contains(strings.ToLower(j.JavaRuntime), "commercial") {
		*j.RequireLicense = true
		return
	}

	// Oracle JDK 7: Free for updates <= 80, requires license for later versions
	if j.VersionMajor == 7 {
		if j.VersionUpdate <= 80 {
			*j.RequireLicense = false
			return
		}
		*j.RequireLicense = true
		return
	}

	// Oracle JDK 8: Free for updates <= 202, requires license for later versions
	if j.VersionMajor == 8 {
		if j.VersionUpdate <= 202 {
			*j.RequireLicense = false
			return
		}
		*j.RequireLicense = true
		return
	}

	// Oracle JDK 11: Commercial license required
	if j.VersionMajor == 11 {
		*j.RequireLicense = true
		return
	}

	// Oracle JDK 17: Commercial license required for 17.0.13 and later
	if j.VersionMajor == 17 {
		if j.VersionUpdate >= 13 {
			*j.RequireLicense = true
			return
		}
		*j.RequireLicense = false
		return
	}

	// For versions 18-20: No commercial license required
	if j.VersionMajor >= 18 && j.VersionMajor <= 20 {
		*j.RequireLicense = false
		return
	}

	// Oracle JDK 21 and later: No commercial license required
	if j.VersionMajor >= 21 {
		*j.RequireLicense = false
		return
	}

	// Default case: require license for any other Oracle JDK version
	*j.RequireLicense = true
}
