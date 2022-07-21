package tfe

import (
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
			State: resourceTFEAgentPoolImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEAgentPoolCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.AgentPoolCreateOptions{
		Name: tfe.String(name),
	}

	log.Printf("[DEBUG] Create new agent pool for organization: %s", organization)
	agentPool, err := tfeClient.AgentPools.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating agent pool %s for organization %s: %w", name, organization, err)
	}

	d.SetId(agentPool.ID)

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of agent pool: %s", d.Id())
	agentPool, err := tfeClient.AgentPools.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] agent pool %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent pool %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", agentPool.Name)
	d.Set("organization", agentPool.Organization.Name)

	return nil
}

func resourceTFEAgentPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.AgentPoolUpdateOptions{
		Name: tfe.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Update agent pool: %s", d.Id())
	_, err := tfeClient.AgentPools.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", d.Id(), err)
	}

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete agent pool: %s", d.Id())
	err := tfeClient.AgentPools.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting agent pool %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEAgentPoolImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid agent pool input format: %s (expected <ORGANIZATION>/<AGENT POOL NAME> or <AGENT POOL ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		org := s[0]
		poolName := s[1]
		poolID, err := fetchtAgentPoolID(org, poolName, tfeClient)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving agent pool with name %s from organization %s %w", poolName, org, err)
		}

		d.SetId(poolID)
	}

	return []*schema.ResourceData{d}, nil
}
