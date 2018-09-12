package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFEOrganizationToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationTokenCreate,
		Read:   resourceTFEOrganizationTokenRead,
		Delete: resourceTFEOrganizationTokenDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFEOrganizationTokenImporter,
		},

		Schema: map[string]*schema.Schema{
			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"force_regenerate": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"token": &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceTFEOrganizationTokenCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	organization := d.Get("organization").(string)

	log.Printf("[DEBUG] Check if a token already exists for organization: %s", organization)
	_, err := tfeClient.OrganizationTokens.Read(ctx, organization)
	if err != nil && err != tfe.ErrResourceNotFound {
		return fmt.Errorf("Error checking if a token exists for organization %s: %v", organization, err)
	}

	// If error is nil, the token already exists.
	if err == nil {
		if !d.Get("force_regenerate").(bool) {
			return fmt.Errorf("A token already exists for organization: %s", organization)
		}
		log.Printf("[DEBUG] Regenerating existing token for organization: %s", organization)
	}

	token, err := tfeClient.OrganizationTokens.Generate(ctx, organization)
	if err != nil {
		return fmt.Errorf(
			"Error creating new token for organization %s: %v", organization, err)
	}

	d.SetId(organization)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	d.Set("token", token.Token)

	return resourceTFEOrganizationTokenRead(d, meta)
}

func resourceTFEOrganizationTokenRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read the token from organization: %s", d.Id())
	_, err := tfeClient.OrganizationTokens.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Token for organization %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading token from organization %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationTokenDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	organization := d.Get("organization").(string)

	log.Printf("[DEBUG] Delete token from organization: %s", organization)
	err := tfeClient.OrganizationTokens.Delete(ctx, organization)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting token from organization %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationTokenImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the organization field.
	d.Set("organization", d.Id())

	return []*schema.ResourceData{d}, nil
}
