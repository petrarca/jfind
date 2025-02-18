package main

import (
	"fmt"
	"strings"
)

// checkOpenJDK checks if the runtime is OpenJDK
func (j *JavaRuntimeJSON) checkOpenJDK() bool {
	if j.JavaRuntime == "" {
		return false
	}

	if strings.Contains(strings.ToLower(j.JavaRuntime), "openjdk") {
		return true
	}

	return j.JavaRuntime == "OpenJDK Runtime Environment"
}

// checkCommercialFeatures checks if the runtime has commercial features
func (j *JavaRuntimeJSON) checkCommercialFeatures() bool {
	return j.JavaRuntime != "" && strings.Contains(strings.ToLower(j.JavaRuntime), "commercial")
}

// checkVersionSpecificRules checks version-specific license requirements
func (j *JavaRuntimeJSON) checkVersionSpecificRules() (bool, bool) {
	switch j.VersionMajor {
	case 7:
		return true, j.VersionUpdate > 80
	case 8:
		return true, j.VersionUpdate > 202
	case 11:
		return true, true
	case 17:
		return true, j.VersionUpdate >= 13
	}

	// For versions 18-20 and 21+
	if j.VersionMajor >= 18 && j.VersionMajor <= 20 {
		return true, false
	}
	if j.VersionMajor >= 21 {
		return true, false
	}

	return false, false
}

// checkLicenseRequirement determines if a commercial license is required for the Java runtime
func (j *JavaRuntimeJSON) checkLicenseRequirement() {
	j.RequireLicense = new(bool)

	// Non-Oracle JDKs never require a license
	if !j.IsOracle {
		*j.RequireLicense = false
		return
	}

	// OpenJDK never requires a license
	if j.checkOpenJDK() {
		*j.RequireLicense = false
		return
	}

	// Check for commercial features
	if j.checkCommercialFeatures() {
		*j.RequireLicense = true
		return
	}

	// Check version-specific rules
	if hasRule, requiresLicense := j.checkVersionSpecificRules(); hasRule {
		*j.RequireLicense = requiresLicense
		return
	}

	// Default case: require license for any other Oracle JDK version
	*j.RequireLicense = true
}

// Must be aligned with the codified rules
func showRules() {
	fmt.Println("Java License Check Rules:")
	fmt.Println("\nOracle JDK License Requirements:")
	fmt.Println("- OpenJDK: Never requires a commercial license")
	fmt.Println("- Oracle JDK 7: Free for updates <= 80, requires license for later versions")
	fmt.Println("- Oracle JDK 8: Free for updates <= 202, requires license for later versions")
	fmt.Println("- Oracle JDK 11: Always requires a commercial license")
	fmt.Println("- Oracle JDK 17: Requires commercial license for version 17.0.13 and later")
	fmt.Println("- Oracle JDK 18-20: No commercial license required")
	fmt.Println("- Oracle JDK 21+: No commercial license required")
	fmt.Println("\nNotes:")
	fmt.Println("- Non-Oracle JDKs never require a commercial license")
	fmt.Println("- Any Oracle JDK version not listed above requires a commercial license by default")
}
