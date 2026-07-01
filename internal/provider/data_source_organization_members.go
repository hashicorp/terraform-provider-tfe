// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about members of an organization.",

		Read: dataSourceTFEOrganizationMembersRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"members": {
				Description: "A list of active members of the organization.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Description: "The ID of the user.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"organization_membership_id": {
							Description: "The ID of the organization membership.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},

			"members_waiting": {
				Description: "A list of members with pending invite to organization.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Description: "The ID of the user.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"organization_membership_id": {
							Description: "The ID of the organization membership.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTFEOrganizationMembersRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	organizationName, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	members, membersWaiting, err := fetchOrganizationMembers(config.Client, organizationName)
	if err != nil {
		return err
	}

	d.Set("members", members)
	d.Set("members_waiting", membersWaiting)
	d.SetId(organizationName)

	return nil
}
