// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helpers

import "github.com/hashicorp/go-tfe"

// TagInfo contains information about the different tags associated with a resource.
type TagInfo struct {
	// Effective tags are the calculated tags for the resource, including inherited and inherited but overridden tags.
	EffectiveTags map[string]interface{}

	// SelfTags are a map of tag binding keys to their values that are set directly on the resource.
	SelfTags map[string]interface{}
}

func NewTagInfo(configBindings map[string]interface{}, calculatedEffective []*tfe.EffectiveTagBinding, ignoreAdditionalTags bool) TagInfo {
	effectiveTagMap := make(map[string]interface{})
	tagBindingsMap := make(map[string]interface{})

	// Set only inherited tags in the effective map first, that way all non-inherited tags will override them.
	for _, binding := range calculatedEffective {
		if binding.Links != nil && binding.Links["inherited-from"] != nil {
			effectiveTagMap[binding.Key] = binding.Value
		}
	}

	// Now we can focus only on the non-inherited tags, overriding effective tags and setting self tags.
	for _, binding := range calculatedEffective {
		if binding.Links != nil && binding.Links["inherited-from"] != nil {
			continue
		}

		effectiveTagMap[binding.Key] = binding.Value
		if _, ok := configBindings[binding.Key]; ok || !ignoreAdditionalTags {
			tagBindingsMap[binding.Key] = binding.Value
		}
	}

	return TagInfo{
		EffectiveTags: effectiveTagMap,
		SelfTags:      tagBindingsMap,
	}
}
