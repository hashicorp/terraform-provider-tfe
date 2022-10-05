package tfe

import (
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationMembersRead,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
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
	tfeClient := meta.(*tfe.Client)

	organizationName := d.Get("organization").(string)

	members, membersWaiting, err := fetchOrganizationMembers(tfeClient, organizationName)
	if err != nil {
		return err
	}

	d.Set("members", members)
	d.Set("members_waiting", membersWaiting)
	d.SetId(organizationName)

	return nil
}
