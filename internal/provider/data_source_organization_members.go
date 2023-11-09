// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationMembersRead,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"organization_membership_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"members_waiting": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"organization_membership_id": {
							Type:     schema.TypeString,
							Computed: true,
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
