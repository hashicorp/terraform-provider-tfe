// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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
	config := meta.(ConfiguredClient)

	// Get the agent pool ID
	agentPoolID := d.Get("agent_pool_id").(string)

	// Get the description.
	description := d.Get("description").(string)

	// Create a new options struct
	options := tfe.AgentTokenCreateOptions{
		Description: tfe.String(description),
	}

	log.Printf("[DEBUG] Create new agent token for agent pool ID: %s", agentPoolID)
	agentToken, err := config.Client.AgentTokens.Create(ctx, agentPoolID, options)
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
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of agent token: %s", d.Id())
	agentToken, err := config.Client.AgentTokens.Read(ctx, d.Id())
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
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete agent token: %s", d.Id())
	err := config.Client.AgentTokens.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting agent token %s: %w", d.Id(), err)
	}

	return nil
}
