package tfe

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAgentTokenCreate,
		Read:   resourceTFEAgentTokenRead,
		Delete: resourceTFEAgentTokenDelete,

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceTFEAgentTokenCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the agent pool ID
	agentPoolID := d.Get("agent_pool_id").(string)

	// Get the description.
	description := d.Get("description").(string)

	// Create a new options struct
	options := tfe.AgentTokenGenerateOptions{
		Description: tfe.String(description),
	}

	log.Printf("[DEBUG] Create new agent token for agent pool ID: %s", agentPoolID)
	agentToken, err := tfeClient.AgentTokens.Generate(ctx, agentPoolID, options)
	if err != nil {
		return fmt.Errorf("Error creating agent token for agent pool ID %s: %v", agentPoolID, err)

	}

	d.SetId(agentToken.ID)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	d.Set("token", agentToken.Token)

	return resourceTFEAgentTokenRead(d, meta)
}

func resourceTFEAgentTokenRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of agent token: %s", d.Id())
	agentToken, err := tfeClient.AgentTokens.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] agent token %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent token %s: %v", d.Id(), err)
	}

	// Update the config
	d.Set("description", agentToken.Description)

	return nil
}

func resourceTFEAgentTokenDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete agent token: %s", d.Id())
	err := tfeClient.AgentTokens.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting agent token %s: %v", d.Id(), err)
	}

	return nil
}
