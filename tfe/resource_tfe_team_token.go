package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamTokenCreate,
		Read:   resourceTFETeamTokenRead,
		Delete: resourceTFETeamTokenDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamTokenImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"force_regenerate": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of resource.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFETeamTokenCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	log.Printf("[DEBUG] Check if a token already exists for team: %s", teamID)
	_, err := tfeClient.TeamTokens.Read(ctx, teamID)
	if err != nil && err != tfe.ErrResourceNotFound {
		return fmt.Errorf("Error checking if a token exists for team %s: %v", teamID, err)
	}

	// If error is nil, the token already exists.
	if err == nil {
		if !d.Get("force_regenerate").(bool) {
			return fmt.Errorf("A token already exists for team: %s", teamID)
		}
		log.Printf("[DEBUG] Regenerating existing token for team: %s", teamID)
	}

	log.Printf("[DEBUG] Create new token for team: %s", teamID)
	token, err := tfeClient.TeamTokens.Create(ctx, teamID)
	if err != nil {
		return fmt.Errorf(
			"Error creating new token for team %s: %v", teamID, err)
	}

	d.SetId(teamID)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	d.Set("token", token.Token)

	return resourceTFETeamTokenRead(d, meta)
}

func resourceTFETeamTokenRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read the token from team: %s", d.Id())
	_, err := tfeClient.TeamTokens.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Token for team %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading token from team %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamTokenDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete token from team: %s", d.Id())
	err := tfeClient.TeamTokens.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting token from team %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamTokenImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the team ID field.
	d.Set("team_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
