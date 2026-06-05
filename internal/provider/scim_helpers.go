// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"

	tfe "github.com/hashicorp/go-tfe"
)

// filterExactSCIMGroups returns the groups whose name matches the given name,
// case-insensitively. The List API only does substring matching, so we filter
// for exact matches here. Safe to call per page while paginating.
func filterExactSCIMGroups(groups []*tfe.AdminSCIMGroup, name string) []*tfe.AdminSCIMGroup {
	var matched []*tfe.AdminSCIMGroup
	for _, g := range groups {
		if g == nil {
			continue
		}
		if strings.EqualFold(g.Name, name) {
			matched = append(matched, g)
		}
	}
	return matched
}
