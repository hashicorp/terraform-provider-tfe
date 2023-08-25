// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
)

// Checks if a given string matches the typical ID format for a TFC/E ressource
// <resource specific prefix>-<16 base58 characters  >
func isResourceIDFormat(resourcePrefix string, id string) bool {
	base58Regex, err := regexp.Compile(fmt.Sprintf("^%s-[1-9A-HJ-NP-Za-km-z]{16}$", resourcePrefix))
	if err != nil {
		return false
	}
	return base58Regex.MatchString(id)
}
