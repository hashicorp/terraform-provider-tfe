// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

func isLegacyVersionFormat(v string) bool {
	dateVersionRegex := regexp.MustCompile(`^v(\d{6})-(\d+)$`)
	return dateVersionRegex.MatchString(v)
}

func isModernVersionFormat(v string) bool {
	dottedVersionRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	return dottedVersionRegex.MatchString(v)
}

func parseLegacyVersion(v string) (int, int, bool) {
	dateVersionRegex := regexp.MustCompile(`^v(\d{6})-(\d+)$`)
	matches := dateVersionRegex.FindStringSubmatch(v)

	if matches == nil {
		return 0, 0, false
	}

	var yyyymm, releaseNum int
	fmt.Sscanf(matches[1], "%d", &yyyymm)
	fmt.Sscanf(matches[2], "%d", &releaseNum)
	return yyyymm, releaseNum, true
}

func compareLegacyVersions(a, b string) int {
	aYYYYMM, aRelease, aOk := parseLegacyVersion(a)
	bYYYYMM, bRelease, bOk := parseLegacyVersion(b)

	if !aOk || !bOk {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}

	if aYYYYMM != bYYYYMM {
		if aYYYYMM < bYYYYMM {
			return -1
		}
		return 1
	}

	if aRelease < bRelease {
		return -1
	} else if aRelease > bRelease {
		return 1
	}
	return 0
}

func validateMinVersion(minVersion string) error {
	if !isLegacyVersionFormat(minVersion) && !isModernVersionFormat(minVersion) {
		return fmt.Errorf("invalid TFE version format %q: must be v{YYYYMM}-{N} or X.Y.Z", minVersion)
	}
	return nil
}

func checkTFEVersion(remoteVersion, minVersion string) (bool, error) {
	if err := validateMinVersion(minVersion); err != nil {
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
		return compareLegacyVersions(remoteVersion, minVersion) >= 0, nil
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

		return remoteVer.Compare(minVer) >= 0, nil
	}

	return false, nil
}
