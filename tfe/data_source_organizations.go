package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOrganizationList,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEOrganizationList(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	fmt.Println("DEBUG: DATA ORGS")
	options := tfe.OrganizationListOptions{}
	log.Printf("[DEBUG] Listing all organizations")
	orgs, err := tfeClient.Organizations.List(ctx, options)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not list organizations")
		}
		return fmt.Errorf("Error retrieving organizations: %v", err)
	}
	resourceID := ""
	names := []string{}
	ids := map[string]string{}

	for _, org := range orgs.Items {
		resourceID += org.Name
		ids[org.Name] = org.ExternalID
		names = append(names, org.Name)
	}

	log.Printf("[DEBUG] Setting Organizations Attributes")
	d.Set("names", names)
	d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%d", schema.HashString(resourceID)))

	return nil
}
