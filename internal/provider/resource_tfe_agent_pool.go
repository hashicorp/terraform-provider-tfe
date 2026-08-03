// Copyright IBM Corp. 2018, 2026
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
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func resourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "Manages agent pools." +
			"\n\nAn agent pool represents a group of agents, often related to one another by sharing a common network segment or purpose. A workspace may be configured to use one of the organization's agent pools to run remote operations with isolated, private, or on-premises infrastructure.",

		Create: resourceTFEAgentPoolCreate,
		Read:   resourceTFEAgentPoolRead,
		Update: resourceTFEAgentPoolUpdate,
		Delete: resourceTFEAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEAgentPoolImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"hostname": {
						Type:              schema.TypeString,
						OptionalForImport: true,
					},
				}
			},
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the agent pool.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"organization_scoped": {
				Description: "Whether or not the agent pool is scoped to all workspaces in the organization. Defaults to true. Should be false when limiting workspaces that can use the agent pool with the [tfe_agent_pool_allowed_workspaces](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/agent_pool_allowed_workspaces) resource.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
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

	// Build the v2 request body.
	body := models.NewAgentPoolsEnvelope()
	pool := models.NewAgentPools()
	poolType := models.AGENTPOOLS_AGENTPOOLS_TYPE
	pool.SetTypeEscaped(&poolType)
	attrs := models.NewAgentPools_attributes()
	attrs.SetName(&name)
	orgScoped := d.Get("organization_scoped").(bool)
	attrs.SetOrganizationScoped(&orgScoped)
	pool.SetAttributes(attrs)
	body.SetData(pool)

	log.Printf("[DEBUG] Create new agent pool for organization: %s", organization)
	env, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).AgentPools().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating agent pool %s for organization %s: %w", name, organization, err)
	}

	agentPool := env.GetData()
	if agentPool == nil {
		return fmt.Errorf("Error creating agent pool %s for organization %s: API returned no data", name, organization)
	}

	poolID := valueOrZero(agentPool.GetId())
	d.SetId(poolID)
	// Set organization from the input since v2 doesn't return it in the response.
	d.Set("organization", organization)

	err = helpers.WriteTFEIdentity(d, poolID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of agent pool: %s", d.Id())
	env, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of agent pool %s: %w", d.Id(), err)
	}

	agentPool := env.GetData()
	if agentPool == nil {
		log.Printf("[DEBUG] agent pool %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	// Update the config.
	if attrs := agentPool.GetAttributes(); attrs != nil {
		d.Set("name", valueOrZero(attrs.GetName()))
		d.Set("organization_scoped", valueOrZero(attrs.GetOrganizationScoped()))
	}
	// The organization JSON:API id is the organization name in Atlas. Populate
	// it from the relationship when present so that import-by-ID works correctly.
	if rels := agentPool.GetRelationships(); rels != nil {
		if org := rels.GetOrganization(); org != nil && org.GetData() != nil && valueOrZero(org.GetData().GetId()) != "" {
			d.Set("organization", valueOrZero(org.GetData().GetId()))
		}
	}

	err = helpers.WriteTFEIdentity(d, d.Id(), config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFEAgentPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Build the v2 request body.
	body := models.NewAgentPoolsEnvelope()
	pool := models.NewAgentPools()
	poolType := models.AGENTPOOLS_AGENTPOOLS_TYPE
	pool.SetTypeEscaped(&poolType)
	attrs := models.NewAgentPools_attributes()
	name := d.Get("name").(string)
	attrs.SetName(&name)
	orgScoped := d.Get("organization_scoped").(bool)
	attrs.SetOrganizationScoped(&orgScoped)
	pool.SetAttributes(attrs)
	body.SetData(pool)

	log.Printf("[DEBUG] Update agent pool: %s", d.Id())
	_, err := config.ClientV2.API.AgentPools().ByAgent_pool_id(d.Id()).Patch(ctx, body, nil)
	if err != nil {
		return fmt.Errorf("Error updating agent pool %s: %w", d.Id(), err)
	}

	return resourceTFEAgentPoolRead(d, meta)
}

func resourceTFEAgentPoolDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete agent pool: %s", d.Id())
	err := config.ClientV2.API.AgentPools().ByAgent_pool_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting agent pool %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEAgentPoolImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	if d.Id() != "" {
		// Import using the import prefix instead of identity
		return resourceTFEAgentPoolImporterLegacy(d, config)
	}

	// We are using an identity
	identity, err := d.Identity()
	if err != nil {
		return nil, fmt.Errorf("error reading workspace identity: %w", err)
	}

	d.SetId(identity.Get("id").(string))

	return []*schema.ResourceData{d}, nil
}

func resourceTFEAgentPoolImporterLegacy(d *schema.ResourceData, cfg ConfiguredClient) ([]*schema.ResourceData, error) {
	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid agent pool input format: %s (expected <ORGANIZATION>/<AGENT POOL NAME> or <AGENT POOL ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		org := s[0]
		poolName := s[1]
		pool, err := fetchAgentPool(org, poolName, cfg.ClientV2)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving agent pool with name %s from organization %s %w", poolName, org, err)
		}

		d.SetId(valueOrZero(pool.GetId()))
		// Pre-populate organization so Read can preserve it; v2 doesn't return
		// the org name in the pool response.
		d.Set("organization", org)
	}

	return []*schema.ResourceData{d}, nil
}
