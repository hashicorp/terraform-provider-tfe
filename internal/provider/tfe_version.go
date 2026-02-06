// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

var legacyVersionRegex = regexp.MustCompile(`^v(\d{6})-(\d+)$`)
var modernVersionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func isLegacyVersionFormat(v string) bool {
	return legacyVersionRegex.MatchString(v)
}

func isModernVersionFormat(v string) bool {
	return modernVersionRegex.MatchString(v)
}

func parseLegacyVersion(v string) (int, int, bool) {
	matches := legacyVersionRegex.FindStringSubmatch(v)

	if matches == nil {
		return 0, 0, false
	}

	var yyyymm, releaseNum int
	fmt.Sscanf(matches[1], "%d", &yyyymm)
	fmt.Sscanf(matches[2], "%d", &releaseNum)
	return yyyymm, releaseNum, true
}

// compareLegacyVersions compares two legacy version strings and
// returns -1 if a < b, 0 if a == b, and 1 if a > b.
func compareLegacyVersions(a, b string) (int, error) {
	aYYYYMM, aRelease, aOk := parseLegacyVersion(a)
	if !aOk {
		return 0, fmt.Errorf("invalid legacy version format: %q", a)
	}
	bYYYYMM, bRelease, bOk := parseLegacyVersion(b)
	if !bOk {
		return 0, fmt.Errorf("invalid legacy version format: %q", b)
	}

	if aYYYYMM != bYYYYMM {
		if aYYYYMM < bYYYYMM {
			return -1, nil
		}
		return 1, nil
	}

	if aRelease < bRelease {
		return -1, nil
	} else if aRelease > bRelease {
		return 1, nil
	}
	return 0, nil
}

// validateVersion checks if the given version string is valid.
func validateVersion(v string) error {
	if !isLegacyVersionFormat(v) && !isModernVersionFormat(v) {
		return fmt.Errorf("invalid TFE version format %q: must be v{YYYYMM}-{N} or X.Y.Z", v)
	}
	return nil
}

// checkTFEVersion checks if the remoteVersion meets the minVersion requirement.
func checkTFEVersion(remoteVersion, minVersion string) (bool, error) {
	if err := validateVersion(minVersion); err != nil {
		return false, err
	}

	minIsLegacy := isLegacyVersionFormat(minVersion)
	minIsModern := isModernVersionFormat(minVersion)
	remoteIsLegacy := isLegacyVersionFormat(remoteVersion)
	remoteIsModern := isModernVersionFormat(remoteVersion)

	if minIsModern && remoteIsLegacy {
		return false, nil
	}

	if minIsLegacy && remoteIsModern {
		return true, nil
	}

	if minIsLegacy && remoteIsLegacy {
		cmp, err := compareLegacyVersions(remoteVersion, minVersion)
		if err != nil {
			return false, fmt.Errorf("comparing versions %q and %q: %w", remoteVersion, minVersion, err)
		}
		return cmp >= 0, nil
	}

	if minIsModern && remoteIsModern {
		minNormalized := strings.TrimPrefix(minVersion, "v")
		remoteNormalized := strings.TrimPrefix(remoteVersion, "v")

		remoteVer, err := version.NewVersion(remoteNormalized)
		if err != nil {
			return false, fmt.Errorf("parsing remote version %q: %w", remoteVersion, err)
		}
		minVer, err := version.NewVersion(minNormalized)
		if err != nil {
			return false, fmt.Errorf("parsing minimum version %q: %w", minVersion, err)
		}

		return remoteVer.GreaterThanOrEqual(minVer), nil
	}

	return false, nil
}
