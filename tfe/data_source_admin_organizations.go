package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEAdminOrganizations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEAdminOrganizationList,

		Schema: map[string]*schema.Schema{
			"organizations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEAdminOrganizationList(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	options := tfe.AdminOrganizationListOptions{}
	log.Printf("[DEBUG] OMAR Listing all organizations")
	orgs, err := tfeClient.Admin.Organizations.List(ctx, options)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not list organizations")
		}
		return fmt.Errorf("Error retrieving organizations: %v", err)
	}
	orgNames := []string{}
	log.Printf("[DEBUG] OMAR org count: ", len(orgs.Items))
	for _, org := range orgs.Items {
		orgNames = append(orgNames, org.Name)
	}

	log.Printf("[DEBUG] OMAR org names: ", orgNames)
	log.Printf("[DEBUG] KEY: ", d.Get("organizations"))
	// Update the config.
	d.Set("organizations", orgNames)
	log.Printf("[DEBUG] KEY: ", d.Get("organizations"))

	return nil
}
