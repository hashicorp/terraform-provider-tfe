// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentToken() *schema.Resource {
	return &schema.Resource{
		Description: "Manages agent tokens." +
			"\n\nEach agent pool has its own set of tokens which are not shared across pools. These tokens allow agents to communicate securely with HCP Terraform.",

		Create: resourceTFEAgentTokenCreate,
		Read:   resourceTFEAgentTokenRead,
		Delete: resourceTFEAgentTokenDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the agent token.",
				Type:        schema.TypeString,
				Computed:    true,
			},

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
	config := meta.(ConfiguredClient)

	// Get the agent pool ID
	agentPoolID := d.Get("agent_pool_id").(string)

	// Get the description.
	description := d.Get("description").(string)

	// Build the v2 request body.
	body := models.NewAuthenticationTokensEnvelope()
	token := models.NewAuthenticationTokens()
	tokenType := models.AUTHENTICATIONTOKENS_AUTHENTICATIONTOKENS_TYPE
	token.SetTypeEscaped(&tokenType)
	attrs := models.NewAuthenticationTokens_attributes()
	attrs.SetDescription(&description)
	token.SetAttributes(attrs)
	body.SetData(token)

	log.Printf("[DEBUG] Create new agent token for agent pool ID: %s", agentPoolID)
	env, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(agentPoolID).AuthenticationTokens().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf("Error creating agent token for agent pool ID %s: %w", agentPoolID, err)
	}

	agentToken := env.GetData()
	if agentToken == nil {
		return fmt.Errorf("Error creating agent token for agent pool ID %s: API returned no data", agentPoolID)
	}

	d.SetId(valueOrZero(agentToken.GetId()))

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	if attrs := agentToken.GetAttributes(); attrs != nil {
		d.Set("token", valueOrZero(attrs.GetToken()))
	}

	return resourceTFEAgentTokenRead(d, meta)
}

func resourceTFEAgentTokenRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of agent token: %s", d.Id())
	env, err := config.ClientV2.API.AuthenticationTokens().ById(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] agent token %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent token %s: %w", d.Id(), err)
	}

	agentToken := env.GetData()
	if agentToken == nil {
		log.Printf("[DEBUG] agent token %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	// Update the config
	if attrs := agentToken.GetAttributes(); attrs != nil {
		d.Set("description", valueOrZero(attrs.GetDescription()))
	}

	return nil
}

func resourceTFEAgentTokenDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete agent token: %s", d.Id())
	err := config.ClientV2.API.AuthenticationTokens().ById(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting agent token %s: %w", d.Id(), err)
	}

	return nil
}
