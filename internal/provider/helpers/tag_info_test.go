// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestNewTagInfoSeparatesDirectAndInheritedBindings(t *testing.T) {
	bindings := []*tfe.EffectiveTagBinding{
		{Key: "environment", Value: "project", Links: map[string]interface{}{"inherited-from": "/projects/prj-123"}},
		{Key: "team", Value: "platform", Links: map[string]interface{}{"inherited-from": "/projects/prj-123"}},
		{Key: "environment", Value: "workspace"},
		{Key: "application", Value: "api", Links: map[string]interface{}{}},
	}

	info := NewTagInfo(nil, bindings, false)

	assert.Equal(t, map[string]interface{}{
		"application": "api",
		"environment": "workspace",
		"team":        "platform",
	}, info.EffectiveTags)
	assert.Equal(t, map[string]interface{}{
		"application": "api",
		"environment": "workspace",
	}, info.SelfTags)
}

func TestNewTagInfoIgnoresUnmanagedDirectBindings(t *testing.T) {
	bindings := []*tfe.EffectiveTagBinding{
		{Key: "managed", Value: "yes"},
		{Key: "external", Value: "yes"},
	}

	info := NewTagInfo(map[string]interface{}{"managed": "yes"}, bindings, true)

	assert.Equal(t, map[string]interface{}{"managed": "yes", "external": "yes"}, info.EffectiveTags)
	assert.Equal(t, map[string]interface{}{"managed": "yes"}, info.SelfTags)
}
