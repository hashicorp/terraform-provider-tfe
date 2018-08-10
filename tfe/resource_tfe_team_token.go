package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFETeamToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamTokenCreate,
		Read:   resourceTFETeamTokenRead,
		Delete: resourceTFETeamTokenDelete,

		Schema: map[string]*schema.Schema{
			"team_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"token": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFETeamTokenCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	log.Printf("[DEBUG] Create new token for team: %s", teamID)
	token, err := tfeClient.TeamTokens.Generate(ctx, teamID)
	if err != nil {
		return fmt.Errorf(
			"Error creating new token for team %s: %v", teamID, err)
	}

	d.Set("token", token.Token)
	d.SetId(teamID)

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

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	log.Printf("[DEBUG] Delete token from team: %s", teamID)
	err := tfeClient.TeamTokens.Delete(ctx, teamID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting token from team %s: %v", d.Id(), err)
	}

	return nil
}
