package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamCreate,
		Read:   resourceTFETeamRead,
		Delete: resourceTFETeamDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFETeamCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.TeamCreateOptions{
		Name: tfe.String(name),
	}

	log.Printf("[DEBUG] Create team %s for organization: %s", name, organization)
	team, err := tfeClient.Teams.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating team %s for organization %s: %v", name, organization, err)
	}

	d.SetId(team.ID)

	return resourceTFETeamRead(d, meta)
}

func resourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of team: %s", d.Id())
	_, err := tfeClient.Teams.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete team: %s", d.Id())
	err := tfeClient.Teams.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team %s: %v", d.Id(), err)
	}

	return nil
}
