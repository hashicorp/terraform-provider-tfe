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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFESentinelPolicy() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "tfe_sentinel_policy is deprecated, please use tfe_policy instead",
		Create:             resourceTFESentinelPolicyCreate,
		Read:               resourceTFESentinelPolicyRead,
		Update:             resourceTFESentinelPolicyUpdate,
		Delete:             resourceTFESentinelPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFESentinelPolicyImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enforce_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(tfe.EnforcementSoft),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.EnforcementAdvisory),
						string(tfe.EnforcementHard),
						string(tfe.EnforcementSoft),
					},
					false,
				),
			},
		},
	}
}

func resourceTFESentinelPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.PolicyCreateOptions{
		Name: tfe.String(name),
		Enforce: []*tfe.EnforcementOptions{
			{
				Path: tfe.String(name + ".sentinel"),
				Mode: tfe.EnforcementMode(tfe.EnforcementLevel(d.Get("enforce_mode").(string))),
			},
		},
	}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = tfe.String(desc.(string))
	}

	log.Printf("[DEBUG] Create sentinel policy %s for organization: %s", name, organization)
	policy, err := config.Client.Policies.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating sentinel policy %s for organization %s: %w", name, organization, err)
	}

	d.SetId(policy.ID)

	log.Printf("[DEBUG] Upload sentinel policy %s for organization: %s", name, organization)
	err = config.Client.Policies.Upload(ctx, policy.ID, []byte(d.Get("policy").(string)))
	if err != nil {
		return fmt.Errorf(
			"Error uploading sentinel policy %s for organization %s: %w", name, organization, err)
	}

	return resourceTFESentinelPolicyRead(d, meta)
}

func resourceTFESentinelPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read sentinel policy: %s", d.Id())
	policy, err := config.Client.Policies.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Sentinel policy %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading sentinel policy %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)

	if len(policy.Enforce) == 1 {
		d.Set("enforce_mode", string(policy.Enforce[0].Mode))
	}

	content, err := config.Client.Policies.Download(ctx, policy.ID)
	if err != nil {
		return fmt.Errorf("Error downloading sentinel policy %s: %w", d.Id(), err)
	}
	d.Set("policy", string(content))

	return nil
}

func resourceTFESentinelPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	if d.HasChange("description") || d.HasChange("enforce_mode") {
		// Create a new options struct.
		options := tfe.PolicyUpdateOptions{}

		if desc, ok := d.GetOk("description"); ok {
			options.Description = tfe.String(desc.(string))
		}

		if d.HasChange("enforce_mode") {
			options.Enforce = []*tfe.EnforcementOptions{
				{
					Path: tfe.String(d.Get("name").(string) + ".sentinel"),
					Mode: tfe.EnforcementMode(tfe.EnforcementLevel(d.Get("enforce_mode").(string))),
				},
			}
		}

		log.Printf("[DEBUG] Update configuration for sentinel policy: %s", d.Id())
		_, err := config.Client.Policies.Update(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for sentinel policy %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		log.Printf("[DEBUG] Update sentinel policy: %s", d.Id())
		err := config.Client.Policies.Upload(ctx, d.Id(), []byte(d.Get("policy").(string)))
		if err != nil {
			return fmt.Errorf("Error updating sentinel policy %s: %w", d.Id(), err)
		}
	}

	return resourceTFESentinelPolicyRead(d, meta)
}

func resourceTFESentinelPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete sentinel policy: %s", d.Id())
	err := config.Client.Policies.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting sentinel policy %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFESentinelPolicyImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid Sentinel policy import format: %s (expected <ORGANIZATION>/<POLICY ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("organization", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
