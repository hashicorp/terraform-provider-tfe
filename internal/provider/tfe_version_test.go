// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestIsLegacyVersionFormat(t *testing.T) {
	cases := map[string]bool{
		"v202404-1":  true,
		"v202505-1":  true,
		"v202312-10": true,
		"1.0.0":      false,
		"v1.0.0":     false,
		"":           false,
		"v20240401":  false,
	}

	for version, expected := range cases {
		t.Run(version, func(t *testing.T) {
			if got := isLegacyVersionFormat(version); got != expected {
				t.Errorf("isLegacyVersionFormat(%q) = %v, want %v", version, got, expected)
			}
		})
	}
}

func TestIsModernVersionFormat(t *testing.T) {
	cases := map[string]bool{
		"1.0.0":     true,
		"1.0.1":     true,
		"10.20.30":  true,
		"v1.0.0":    true,
		"v202404-1": false,
		"":          false,
		"1.0":       false,
	}

	for version, expected := range cases {
		t.Run(version, func(t *testing.T) {
			if got := isModernVersionFormat(version); got != expected {
				t.Errorf("isModernVersionFormat(%q) = %v, want %v", version, got, expected)
			}
		})
	}
}

func TestCompareLegacyVersions(t *testing.T) {
	cases := map[string]struct {
		a, b     string
		expected int
	}{
		"equal versions":                {"v202404-1", "v202404-1", 0},
		"a newer month":                 {"v202405-1", "v202404-1", 1},
		"a older month":                 {"v202404-1", "v202405-1", -1},
		"a higher release same month":   {"v202404-2", "v202404-1", 1},
		"a lower release same month":    {"v202404-1", "v202404-2", -1},
		"a higher release prior month":  {"v202404-2", "v202404-5", -1},
		"numeric not string comparison": {"v202404-10", "v202404-2", 1},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := compareLegacyVersions(tc.a, tc.b); got != tc.expected {
				t.Errorf("compareLegacyVersions(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}

func TestCheckTFEVersion(t *testing.T) {
	cases := map[string]struct {
		remoteVersion string
		minVersion    string
		expected      bool
		wantError     bool
	}{
		// Modern min + legacy remote = FAIL
		"modern min, legacy remote fails":       {"v202404-1", "1.0.0", false, false},
		"modern min, older legacy remote fails": {"v202301-1", "1.0.0", false, false},

		// Legacy min + modern remote = PASS
		"legacy min, modern remote passes":       {"1.0.0", "v202404-1", true, false},
		"legacy min, newer modern remote passes": {"1.1.0", "v202501-1", true, false},

		// Both legacy
		"both legacy, remote newer":              {"v202505-1", "v202404-1", true, false},
		"both legacy, remote equal":              {"v202404-1", "v202404-1", true, false},
		"both legacy, remote older":              {"v202401-1", "v202404-1", false, false},
		"both legacy, same month higher release": {"v202404-2", "v202404-1", true, false},
		"both legacy, same month lower release":  {"v202404-1", "v202404-2", false, false},

		// Both modern
		"both modern, remote newer":  {"1.1.0", "1.0.0", true, false},
		"both modern, remote equal":  {"1.0.0", "1.0.0", true, false},
		"both modern, remote older":  {"1.0.0", "1.1.0", false, false},
		"both modern, patch version": {"1.0.2", "1.0.1", true, false},

		// Unknown remote = fail closed
		"empty remote with modern min":   {"", "1.0.0", false, false},
		"empty remote with legacy min":   {"", "v202404-1", false, false},
		"unknown remote with modern min": {"unknown", "1.0.0", false, false},

		// Invalid min version = error
		"invalid min version": {"1.0.0", "invalid", false, true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := checkTFEVersion(tc.remoteVersion, tc.minVersion)
			if (err != nil) != tc.wantError {
				t.Errorf("checkTFEVersion() error = %v, wantError %v", err, tc.wantError)
				return
			}
			if !tc.wantError && got != tc.expected {
				t.Errorf("checkTFEVersion(%q, %q) = %v, want %v", tc.remoteVersion, tc.minVersion, got, tc.expected)
			}
		})
	}
}

func TestValidateMinVersion(t *testing.T) {
	cases := map[string]bool{
		"v202404-1": false,
		"1.0.0":     false,
		"v1.0.0":    false,
		"":          true,
		"invalid":   true,
		"202404-1":  true,
	}

	for version, wantError := range cases {
		t.Run(version, func(t *testing.T) {
			err := validateMinVersion(version)
			if (err != nil) != wantError {
				t.Errorf("validateMinVersion(%q) error = %v, wantError %v", version, err, wantError)
			}
		})
	}
}
