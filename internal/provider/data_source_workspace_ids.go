// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	v2orgs "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on workspace IDs." +
			"\n\n-> **Note:** At least one of `names` or `tag_names` must be provided.",

		Read: dataSourceTFEWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Derived from the organization name and a hash of the matched workspace IDs. Do not rely on this value.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"names": {
				Description:  "A list of workspace names to search for. Names that don't match a valid workspace will be omitted from the results, but are not an error. To select _all_ workspaces for an organization, provide a list with a single asterisk, like `[\"*\"]`. The asterisk also supports partial matching on prefix and/or suffix, like `[*-prod]`, `[test-*]`, `[*dev*]`.",
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				AtLeastOneOf: []string{"tag_filters", "names", "tag_names"},
			},

			"tag_names": {
				Description: "A set of key-value tag filters to search for workspaces. At least one of this or `names` must be present.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Deprecated:  "Use `tag_filters.include` instead. This attribute will be removed in a future release of the provider.",
			},

			"exclude_tags": {
				Description: "A list of tag names to exclude when searching.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Deprecated:  "Use `tag_filters.exclude` instead. This attribute will be removed in a future release of the provider.",
			},

			"tag_filters": {
				Description: "A set of key-value tag filters to search for workspaces.",
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include": {
							Description: "A map of key-value tags the workspaces must contain. Each tag included here will be combined using a logical AND when filtering results.",
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"exclude": {
							Description: "A map of key-value tags to exclude workspaces from the returned list. To exclude all workspaces containing a specific key, use `\"*\"` as the value.",
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"organization": {
				Description: "The name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"ids": {
				Description: "A map of workspace names and their opaque, immutable IDs, which look like `ws-<RANDOM STRING>`.",
				Type:        schema.TypeMap,
				Computed:    true,
			},

			"full_names": {
				Description: "A map of workspace names and their full names, which look like `<ORGANIZATION>/<WORKSPACE>`.",
				Type:        schema.TypeMap,
				Computed:    true,
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
	api := config.ClientV2.API

	// Get the organization.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}
	if _, ok := d.GetOk("tag_filters"); ok {
		// Kiota cannot encode the filter[tagged][n][key|value] deep-object
		// parameters generated from the OpenAPI schema.
		return dataSourceTFEWorkspaceIDsReadTagBindingsV1(d, config, organization)
	}

	// Create a map with all the names we are looking for.
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		if name == nil {
			continue
		}
		id += name.(string)
		names[name.(string)] = true
	}

	// Create two maps to hold the results.
	fullNames := make(map[string]string, len(names))
	ids := make(map[string]string, len(names))

	// Build the query params for workspace listing.
	pageSize := int32(100)
	queryParams := &v2orgs.ItemWorkspacesRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}

	// Build exclude tag lookup map (old-style tag names).
	excludeTagLookupMap := make(map[string]bool)
	var excludeTagParts []string
	for _, excludedTag := range d.Get("exclude_tags").(*schema.Set).List() {
		if exTag, ok := excludedTag.(string); ok && len(strings.TrimSpace(exTag)) != 0 {
			excludeTagLookupMap[exTag] = true
			excludeTagParts = append(excludeTagParts, exTag)
		}
	}
	if len(excludeTagParts) > 0 {
		excludeTags := strings.Join(excludeTagParts, ",")
		queryParams.SearchexcludeTags = &excludeTags
	}

	// Old-style tag name include filtering.
	var tagSearchParts []string
	for _, tagName := range d.Get("tag_names").([]interface{}) {
		if name, ok := tagName.(string); ok && len(strings.TrimSpace(name)) != 0 {
			id += name
			tagSearchParts = append(tagSearchParts, name)
		}
	}
	if len(tagSearchParts) > 0 {
		tagSearch := strings.Join(tagSearchParts, ",")
		queryParams.Searchtags = &tagSearch
	}

	hasLegacyTags := len(tagSearchParts) > 0
	hasOnlyTags := hasLegacyTags && len(names) == 0

	for {
		wl, err := api.Organizations().ByOrganization_name(organization).Workspaces().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			if errors.Is(err, tfev2.ErrNotFound) {
				return fmt.Errorf("Error retrieving workspaces: organization %s not found", organization)
			}
			return fmt.Errorf("Error retrieving workspaces: %w", err)
		}

		for _, w := range wl.GetData() {
			wsAttrs := w.GetAttributes()
			if wsAttrs == nil {
				continue
			}
			wsName := valueOrZero(wsAttrs.GetName())
			wsID := valueOrZero(w.GetId())

			// Client-side exclude by old-style tag names.
			hasExcludedTag := false
			for _, tag := range wsAttrs.GetTagNames() {
				if excludeTagLookupMap[tag] {
					hasExcludedTag = true
					break
				}
			}

			if (hasOnlyTags || includedByName(names, wsName)) && !hasExcludedTag {
				fullNames[wsName] = organization + "/" + wsName
				ids[wsName] = wsID
			}
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageFromMeta(wl.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	d.Set("ids", ids)
	d.Set("full_names", fullNames)
	d.SetId(fmt.Sprintf("%s/%d", organization, schema.HashString(id)))

	return nil
}

func dataSourceTFEWorkspaceIDsReadTagBindingsV1(d *schema.ResourceData, config ConfiguredClient, organization string) error {
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		if name == nil {
			continue
		}
		id += name.(string)
		names[name.(string)] = true
	}

	fullNames := make(map[string]string, len(names))
	ids := make(map[string]string, len(names))
	options := &tfe.WorkspaceListOptions{}
	excludeTagLookupMap := make(map[string]bool)
	var excludeTagParts []string
	for _, excludedTag := range d.Get("exclude_tags").(*schema.Set).List() {
		if exTag, ok := excludedTag.(string); ok && len(strings.TrimSpace(exTag)) != 0 {
			excludeTagLookupMap[exTag] = true
			excludeTagParts = append(excludeTagParts, exTag)
		}
	}
	options.ExcludeTags = strings.Join(excludeTagParts, ",")

	var tagSearchParts []string
	for _, tagName := range d.Get("tag_names").([]interface{}) {
		if name, ok := tagName.(string); ok && len(strings.TrimSpace(name)) != 0 {
			id += name
			tagSearchParts = append(tagSearchParts, name)
		}
	}
	options.Tags = strings.Join(tagSearchParts, ",")

	excludeTagBindings := make(map[string]string)
	tagFilters := d.Get("tag_filters").([]interface{})[0].(map[string]interface{})
	options.Include = []tfe.WSIncludeOpt{tfe.WSEffectiveTagBindings}
	if include, ok := tagFilters["include"].(map[string]interface{}); ok {
		for key, val := range include {
			options.TagBindings = append(options.TagBindings, &tfe.TagBinding{Key: key, Value: val.(string)})
		}
	}
	if exclude, ok := tagFilters["exclude"].(map[string]interface{}); ok {
		for key, val := range exclude {
			excludeTagBindings[key] = val.(string)
		}
	}

	hasOnlyTags := (len(tagSearchParts) > 0 || len(options.TagBindings) > 0 || len(excludeTagBindings) > 0) && len(names) == 0
	for {
		wl, err := config.Client.Workspaces.List(ctx, organization, options)
		if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
			options.Include = nil
			wl, err = config.Client.Workspaces.List(ctx, organization, options)
		}
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %w", err)
		}

		for _, workspace := range wl.Items {
			hasExcludedTag := false
			for _, tag := range workspace.TagNames {
				if excludeTagLookupMap[tag] {
					hasExcludedTag = true
					break
				}
			}
			for _, binding := range workspace.EffectiveTagBindings {
				if value, ok := excludeTagBindings[binding.Key]; ok && (value == binding.Value || value == "*") {
					hasExcludedTag = true
				}
			}
			if (hasOnlyTags || includedByName(names, workspace.Name)) && !hasExcludedTag {
				fullNames[workspace.Name] = organization + "/" + workspace.Name
				ids[workspace.Name] = workspace.ID
			}
		}

		if wl.CurrentPage >= wl.TotalPages {
			break
		}
		options.PageNumber = wl.NextPage
	}

	d.Set("ids", ids)
	d.Set("full_names", fullNames)
	d.SetId(fmt.Sprintf("%s/%d", organization, schema.HashString(id)))
	return nil
}
