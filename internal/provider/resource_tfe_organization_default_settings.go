// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOrganizationDefaultSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationDefaultSettingsCreate,
		Read:   resourceTFEOrganizationDefaultSettingsRead,
		Delete: resourceTFEOrganizationDefaultSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOrganizationDefaultSettingsImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"default_execution_mode": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"agent",
						"local",
						"remote",
					},
					false,
				),
				ForceNew: true,
			},

			"default_agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"default_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEOrganizationDefaultSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return fmt.Errorf("error getting organization name: %w", err)
	}

	// If the "default_agent_pool_id" was provided, get the agent pool
	var agentPool *tfe.AgentPool
	if v, ok := d.GetOk("default_agent_pool_id"); ok && v.(string) != "" {
		agentPool = &tfe.AgentPool{
			ID: v.(string),
		}
	}

	// If the default project id was provided, get the default project
	var project *tfe.Project
	if v, ok := d.GetOk("default_project_id"); ok && v.(string) != "" {
		project = &tfe.Project{
			ID: v.(string),
		}
	}

	defaultExecutionMode := ""
	if v, ok := d.GetOk("default_execution_mode"); ok {
		defaultExecutionMode = v.(string)
	} else {
		return fmt.Errorf("default_execution_mode was missing from tfstate, please create an issue to report this error")
	}

	// set organization default execution mode
	_, err = config.Client.Organizations.Update(context.Background(), organization, tfe.OrganizationUpdateOptions{
		DefaultExecutionMode: tfe.String(defaultExecutionMode),
		DefaultAgentPool:     agentPool,
		DefaultProject:       project,
	})
	if err != nil {
		return fmt.Errorf("error setting default execution mode of organization %s: %w", d.Id(), err)
	}

	d.SetId(organization)

	return resourceTFEOrganizationDefaultSettingsRead(d, meta)
}

func resourceTFEOrganizationDefaultSettingsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read the organization: %s", d.Id())
	organization, err := config.Client.Organizations.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading organization %s: %w", d.Id(), err)
	}

	defaultExecutionMode := ""
	if v, ok := d.GetOk("default_execution_mode"); ok {
		defaultExecutionMode = v.(string)
	} else {
		return fmt.Errorf("default_execution_mode was missing from tfstate, please create an issue to report this error")
	}
	if organization.DefaultExecutionMode != defaultExecutionMode {
		// set id to empty string so that the provider knows it needs to set the default execution mode again
		d.SetId("")
	}

	return nil
}

func resourceTFEOrganizationDefaultSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return fmt.Errorf("error getting organization name: %w", err)
	}

	log.Printf("[DEBUG] Reseting default execution mode of organization: %s", organization)
	// reset organization default execution mode
	_, err = config.Client.Organizations.Update(context.Background(), organization, tfe.OrganizationUpdateOptions{
		DefaultExecutionMode: tfe.String("remote"),
		DefaultAgentPool:     nil,
		DefaultProject:       nil,
	})
	if err != nil {
		return fmt.Errorf("error updating organization default execution mode: %w", err)
	}

	return nil
}

func resourceTFEOrganizationDefaultSettingsImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read the organization: %s", d.Id())
	organization, err := config.Client.Organizations.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] organization %s no longer exists", d.Id())
			d.SetId("")
		}
		return nil, fmt.Errorf("error reading organization %s: %w", d.Id(), err)
	}

	// Set the organization field.
	d.Set("organization", d.Id())
	d.Set("default_execution_mode", organization.DefaultExecutionMode)
	if organization.DefaultAgentPool != nil {
		d.Set("default_agent_pool_id", organization.DefaultAgentPool.ID)
	}
	if organization.DefaultProject != nil {
		d.Set("default_project_id", organization.DefaultProject.ID)
	}

	return []*schema.ResourceData{d}, nil
}
