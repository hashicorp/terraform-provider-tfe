package tfe

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEAgentPoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	id, err := fetchAgentPoolID(organization, name, config.Client)
	if err != nil {
		return err
	}
	d.SetId(id)
	return nil
}
