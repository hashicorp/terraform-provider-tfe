// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				AtLeastOneOf: []string{"names", "tag_names"},
			},

			"tag_names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"exclude_tags": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"full_names": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func includedByName(names map[string]bool, workspaceName string) bool {
	for name := range names {
		switch {
		case len(name) == 0:
			continue
		case !strings.HasPrefix(name, "*") && !strings.HasSuffix(name, "*"):
			if name == workspaceName {
				return true
			}
		case strings.HasPrefix(name, "*") && strings.HasSuffix(name, "*"):
			if len(name) == 1 {
				return true
			}
			x := name[1 : len(name)-1]
			if strings.Contains(workspaceName, x) {
				return true
			}
		case strings.HasPrefix(name, "*"):
			x := name[1:]
			if strings.HasSuffix(workspaceName, x) {
				return true
			}
		case strings.HasSuffix(name, "*"):
			x := name[:len(name)-1]
			if strings.HasPrefix(workspaceName, x) {
				return true
			}
		}
	}
	return false
}

func dataSourceTFEWorkspaceIDsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a map with all the names we are looking for.
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		// ignore empty strings
		if name == nil {
			continue
		}

		id += name.(string)
		names[name.(string)] = true
	}

	// Create two maps to hold the results.
	fullNames := make(map[string]string, len(names))
	ids := make(map[string]string, len(names))

	options := &tfe.WorkspaceListOptions{}

	excludeTagLookupMap := make(map[string]bool)
	var excludeTagBuf strings.Builder
	for _, excludedTag := range d.Get("exclude_tags").(*schema.Set).List() {
		if exTag, ok := excludedTag.(string); ok && len(strings.TrimSpace(exTag)) != 0 {
			excludeTagLookupMap[exTag] = true

			if excludeTagBuf.Len() > 0 {
				excludeTagBuf.WriteByte(',')
			}
			excludeTagBuf.WriteString(exTag)
		}
	}

	if excludeTagBuf.Len() > 0 {
		options.ExcludeTags = excludeTagBuf.String()
	}

	// Create a search string with all the tag names we are looking for.
	var tagSearchParts []string
	for _, tagName := range d.Get("tag_names").([]interface{}) {
		if name, ok := tagName.(string); ok && len(strings.TrimSpace(name)) != 0 {
			id += name // add to the state id
			tagSearchParts = append(tagSearchParts, name)
		}
	}
	if len(tagSearchParts) > 0 {
		tagSearch := strings.Join(tagSearchParts, ",")
		options.Tags = tagSearch
	}

	hasOnlyTags := len(tagSearchParts) > 0 && len(names) == 0

	for {
		wl, err := config.Client.Workspaces.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %w", err)
		}

		for _, w := range wl.Items {
			// fallback for tfe instances that don't yet support exclude-tags
			hasExcludedTag := false
			for _, tag := range w.TagNames {
				if _, ok := excludeTagLookupMap[tag]; ok {
					hasExcludedTag = true
					break
				}
			}
			if (hasOnlyTags || includedByName(names, w.Name)) && !hasExcludedTag {
				fullNames[w.Name] = organization + "/" + w.Name
				ids[w.Name] = w.ID
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	d.Set("ids", ids)
	d.Set("full_names", fullNames)
	d.SetId(fmt.Sprintf("%s/%d", organization, schema.HashString(id)))

	return nil
}
