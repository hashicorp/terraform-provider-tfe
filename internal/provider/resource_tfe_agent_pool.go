// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAgentPoolCreate,
		Read:   resourceTFEAgentPoolRead,
		Update: resourceTFEAgentPoolUpdate,
		Delete: resourceTFEAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEAgentPoolImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"organization_scoped": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceTFEAgentPoolCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.AgentPoolCreateOptions{
		Name:               tfe.String(name),
		OrganizationScoped: tfe.Bool(d.Get("organization_scoped").(bool)),
	}

	log.Printf("[DEBUG] Create new agent pool for organization: %s", organization)
	agentPool, err := config.Client.AgentPools.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating agent pool %s for organization %s: %w", name, organization, err)
	}

	d.SetId(agentPool.ID)

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of agent pool: %s", d.Id())
	agentPool, err := config.Client.AgentPools.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent pool %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", agentPool.Name)
	d.Set("organization", agentPool.Organization.Name)
	d.Set("organization_scoped", agentPool.OrganizationScoped)

	return nil
}

func resourceTFEAgentPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Create a new options struct.
	options := tfe.AgentPoolUpdateOptions{
		Name:               tfe.String(d.Get("name").(string)),
		OrganizationScoped: tfe.Bool(d.Get("organization_scoped").(bool)),
	}

	log.Printf("[DEBUG] Update agent pool: %s", d.Id())
	_, err := config.Client.AgentPools.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", d.Id(), err)
	}

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete agent pool: %s", d.Id())
	err := config.Client.AgentPools.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting agent pool %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEAgentPoolImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid agent pool input format: %s (expected <ORGANIZATION>/<AGENT POOL NAME> or <AGENT POOL ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		org := s[0]
		poolName := s[1]
		pool, err := fetchAgentPool(org, poolName, config.Client)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving agent pool with name %s from organization %s %w", poolName, org, err)
		}

		d.SetId(pool.ID)
	}

	return []*schema.ResourceData{d}, nil
}
