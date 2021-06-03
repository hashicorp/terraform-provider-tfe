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

	names := []string{}
	ids := map[string]string{}

	adminOrgs, err := listAdminOrgs(tfeClient)
	if err != nil {
		return err
	}

	if adminOrgs != nil {
		for _, org := range adminOrgs.Items {
			ids[org.Name] = org.ExternalID
			names = append(names, org.Name)
		}
	} else {
		orgs, err := listOrgs(tfeClient)
		if err != nil {
			return err
		}
		for _, org := range orgs.Items {
			ids[org.Name] = org.ExternalID
			names = append(names, org.Name)
		}
	}

	log.Printf("[DEBUG] Setting Organizations Attributes")
	d.SetId("organizations")
	d.Set("names", names)
	d.Set("ids", ids)

	return nil
}

func listAdminOrgs(client *tfe.Client) (*tfe.AdminOrganizationList, error) {
	log.Printf("[DEBUG] Listing all organizations (admin)")
	orgs, err := client.Admin.Organizations.List(ctx, tfe.AdminOrganizationListOptions{})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving organizations: %v.", err)
	}

	return orgs, nil
}

func listOrgs(client *tfe.Client) (*tfe.OrganizationList, error) {
	log.Printf("[DEBUG] Listing all organizations (non-admin)")
	orgs, err := client.Organizations.List(ctx, tfe.OrganizationListOptions{})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil, fmt.Errorf("Could not list organizations.")
		}
		return nil, fmt.Errorf("Error retrieving organizations: %v.", err)
	}

	return orgs, nil
}
