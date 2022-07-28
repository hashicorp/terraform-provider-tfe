package tfe

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentToken() *schema.Resource {
	return &schema.Resource{
		Description: "Each agent pool has its own set of tokens which are not shared across pools. These tokens allow agents to communicate securely with Terraform Cloud." +
			"\n\n ~> **NOTE:** This resource requires using the provider with Terraform Cloud and a Terraform Cloud for Business account. [Learn more about Terraform Cloud pricing here](https://www.hashicorp.com/products/terraform/pricing?_ga=2.56441195.1392855715.1658762101-1323299352.1652184430).",

		Create: resourceTFEAgentTokenCreate,
		Read:   resourceTFEAgentTokenRead,
		Delete: resourceTFEAgentTokenDelete,

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Description: "ID of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description of the agent token.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"token": {
				Description: "The generated token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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
	options := tfe.AgentTokenCreateOptions{
		Description: tfe.String(description),
	}

	log.Printf("[DEBUG] Create new agent token for agent pool ID: %s", agentPoolID)
	agentToken, err := tfeClient.AgentTokens.Create(ctx, agentPoolID, options)
	if err != nil {
		return fmt.Errorf("Error creating agent token for agent pool ID %s: %w", agentPoolID, err)
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
			log.Printf("[DEBUG] agent token %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent token %s: %w", d.Id(), err)
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
		return fmt.Errorf("Error deleting agent token %s: %w", d.Id(), err)
	}

	return nil
}
