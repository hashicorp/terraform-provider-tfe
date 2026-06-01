// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestFilterExactSCIMGroups(t *testing.T) {
	groups := []*tfe.AdminSCIMGroup{
		{ID: "sgr-1", Name: "platform-ops-idp"},
		{ID: "sgr-2", Name: "platform-ops-idp-admin-group"},
		{ID: "sgr-3", Name: "platform-ops-idp-eng-group"},
		{ID: "sgr-4", Name: "platform-ops-idp-audit-group"},
	}

	t.Run("case-insensitive exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "Platform-Ops-Idp-Admin-Group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-2", matched[0].ID)
	})

	t.Run("fuzzy substring siblings are rejected", func(t *testing.T) {
		// ?q=platform-ops-idp matches all four as a substring; we keep only
		// the exact name.
		matched := filterExactSCIMGroups(groups, "platform-ops-idp")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-1", matched[0].ID)
	})

	t.Run("no match returns nil", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "nonexistent")
		assert.Nil(t, matched)
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		matched := filterExactSCIMGroups(nil, "platform-ops-idp-admin-group")
		assert.Nil(t, matched)
	})

	t.Run("nil entries are skipped", func(t *testing.T) {
		matched := filterExactSCIMGroups([]*tfe.AdminSCIMGroup{nil}, "anything")
		assert.Nil(t, matched)
	})

	t.Run("single exact match", func(t *testing.T) {
		matched := filterExactSCIMGroups(groups, "platform-ops-idp-audit-group")
		assert.Len(t, matched, 1)
		assert.Equal(t, "sgr-4", matched[0].ID)
	})
}
