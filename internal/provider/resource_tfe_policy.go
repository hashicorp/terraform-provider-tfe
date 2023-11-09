// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicyCreate,
		Read:   resourceTFEPolicyRead,
		Update: resourceTFEPolicyUpdate,
		Delete: resourceTFEPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEPolicyImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the policy",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"description": {
				Description: "Text describing the policy's purpose",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"organization": {
				Description: "Name of the organization that this policy belongs to",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"kind": {
				Description: "The policy-as-code framework for the policy. Valid values are sentinel and opa",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     string(tfe.Sentinel),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.OPA),
						string(tfe.Sentinel),
					}, false),
			},

			"query": {
				Description: "The OPA query to run. Required for OPA policies",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"policy": {
				Description: "Text of a valid Sentinel or OPA policy",
				Type:        schema.TypeString,
				Required:    true,
			},

			"enforce_mode": {
				Type: schema.TypeString,
				Description: fmt.Sprintf(
					"The enforcement configuration of the policy. For Sentinel, valid values are %s. For OPA, Valid values are `%s`", sentenceList(
						sentinelPolicyEnforcementLevels(),
						"`",
						"`",
						"and"),
					sentenceList(
						opaPolicyEnforcementLevels(),
						"`",
						"`",
						"and")),
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.EnforcementAdvisory),
						string(tfe.EnforcementHard),
						string(tfe.EnforcementSoft),
						string(tfe.EnforcementMandatory),
					},
					false,
				),
			},
		},
	}
}

func sentinelPolicyEnforcementLevels() []string {
	return []string{
		string(tfe.EnforcementHard),
		string(tfe.EnforcementSoft),
		string(tfe.EnforcementAdvisory),
	}
}

func opaPolicyEnforcementLevels() []string {
	return []string{
		string(tfe.EnforcementMandatory),
		string(tfe.EnforcementAdvisory),
	}
}

func resourceTFEPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	var kind string
	if vKind, ok := d.GetOk("kind"); ok {
		kind = vKind.(string)
	}

	// Setup common policy options
	options := &tfe.PolicyCreateOptions{
		Name: tfe.String(name),
		Kind: tfe.PolicyKind(kind),
	}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = tfe.String(desc.(string))
	}

	//  Setup per-kind policy options
	switch tfe.PolicyKind(kind) {
	case tfe.Sentinel:
		options = createSentinelPolicyOptions(options, d)
	case tfe.OPA:
		options, err = createOPAPolicyOptions(options, d)
	default:
		err = fmt.Errorf(
			"unsupported policy kind %s: has to be one of [%s, %s]", kind, string(tfe.Sentinel), string(tfe.OPA))
	}
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Create %s policy %s for organization: %s", kind, name, organization)
	policy, err := config.Client.Policies.Create(ctx, organization, *options)
	if err != nil {
		return fmt.Errorf(
			"Error creating %s policy %s for organization %s: %w", kind, name, organization, err)
	}

	d.SetId(policy.ID)

	log.Printf("[DEBUG] Upload %s policy %s for organization: %s", kind, name, organization)
	err = config.Client.Policies.Upload(ctx, policy.ID, []byte(d.Get("policy").(string)))
	if err != nil {
		return fmt.Errorf(
			"Error uploading %s policy %s for organization %s: %w", kind, name, organization, err)
	}

	return resourceTFEPolicyRead(d, meta)
}

func createOPAPolicyOptions(options *tfe.PolicyCreateOptions, d *schema.ResourceData) (*tfe.PolicyCreateOptions, error) {
	name := d.Get("name").(string)
	path := name + ".rego"
	enforceOpts := &tfe.EnforcementOptions{
		Path: tfe.String(path),
	}

	if v, ok := d.GetOk("enforce_mode"); !ok {
		enforceOpts.Mode = tfe.EnforcementMode(getDefaultEnforcementMode(tfe.OPA))
	} else {
		enforceOpts.Mode = tfe.EnforcementMode(tfe.EnforcementLevel(v.(string)))
	}

	options.Enforce = []*tfe.EnforcementOptions{enforceOpts}

	vQuery, ok := d.GetOk("query")
	if !ok {
		return options, fmt.Errorf("missing query for OPA policy")
	}
	options.Query = tfe.String(vQuery.(string))

	return options, nil
}

func createSentinelPolicyOptions(options *tfe.PolicyCreateOptions, d *schema.ResourceData) *tfe.PolicyCreateOptions {
	name := d.Get("name").(string)
	path := name + ".sentinel"
	enforceOpts := &tfe.EnforcementOptions{
		Path: tfe.String(path),
	}

	if v, ok := d.GetOk("enforce_mode"); !ok {
		enforceOpts.Mode = tfe.EnforcementMode(getDefaultEnforcementMode(tfe.Sentinel))
	} else {
		enforceOpts.Mode = tfe.EnforcementMode(tfe.EnforcementLevel(v.(string)))
	}

	options.Enforce = []*tfe.EnforcementOptions{enforceOpts}
	return options
}

func getDefaultEnforcementMode(kind tfe.PolicyKind) tfe.EnforcementLevel {
	switch kind {
	case tfe.Sentinel:
		return tfe.EnforcementSoft

	case tfe.OPA:
		return tfe.EnforcementAdvisory

	default:
		return ""
	}
}

func resourceTFEPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read policy: %s", d.Id())
	policy, err := config.Client.Policies.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Policy %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("kind", policy.Kind)

	if len(policy.Enforce) == 1 {
		d.Set("enforce_mode", string(policy.Enforce[0].Mode))
	}

	content, err := config.Client.Policies.Download(ctx, policy.ID)
	if err != nil {
		return fmt.Errorf("Error downloading policy %s: %w", d.Id(), err)
	}
	d.Set("policy", string(content))

	return nil
}

func resourceTFEPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// nolint:nestif
	if d.HasChange("description") || d.HasChange("enforce_mode") {
		// Create a new options struct.
		options := tfe.PolicyUpdateOptions{}

		if desc, ok := d.GetOk("description"); ok {
			options.Description = tfe.String(desc.(string))
		}

		path := d.Get("name").(string) + ".sentinel"
		vKind, ok := d.GetOk("kind")
		if ok {
			if vKind == tfe.OPA {
				path = d.Get("name").(string) + ".rego"
			}
		}
		if d.HasChange("enforce_mode") {
			options.Enforce = []*tfe.EnforcementOptions{
				{
					Path: tfe.String(path),
					Mode: tfe.EnforcementMode(tfe.EnforcementLevel(d.Get("enforce_mode").(string))),
				},
			}
		}

		log.Printf("[DEBUG] Update configuration for %s policy: %s", vKind, d.Id())
		_, err := config.Client.Policies.Update(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for %s policy %s: %w", vKind, d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		vKind := d.Get("kind").(string)
		log.Printf("[DEBUG] Update %s policy: %s", vKind, d.Id())
		err := config.Client.Policies.Upload(ctx, d.Id(), []byte(d.Get("policy").(string)))
		if err != nil {
			return fmt.Errorf("Error updating %s policy %s: %w", vKind, d.Id(), err)
		}
	}

	return resourceTFEPolicyRead(d, meta)
}

func resourceTFEPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete policy: %s", d.Id())
	err := config.Client.Policies.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting policy %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEPolicyImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid policy import format: %s (expected <ORGANIZATION>/<POLICY ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("organization", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
