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

func resourceTFEPolicySetParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetParameterCreate,
		Read:   resourceTFEPolicySetParameterRead,
		Update: resourceTFEPolicySetParameterUpdate,
		Delete: resourceTFEPolicySetParameterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEPolicySetParameterImporter,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"value": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"policy_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEPolicySetParameterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get key
	key := d.Get("key").(string)

	ps := d.Get("policy_set_id").(string)
	policySet, err := config.Client.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %w", ps, err)
	}

	// Create a new options struct.
	options := tfe.PolicySetParameterCreateOptions{
		Key:       tfe.String(key),
		Value:     tfe.String(d.Get("value").(string)),
		Category:  tfe.Category(tfe.CategoryPolicySet),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
	}

	log.Printf("[DEBUG] Create %s parameter: %s", tfe.CategoryPolicySet, key)
	parameter, err := config.Client.PolicySetParameters.Create(ctx, policySet.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating %s parameter %s %w", tfe.CategoryPolicySet, key, err)
	}

	d.SetId(parameter.ID)

	return resourceTFEPolicySetParameterRead(d, meta)
}

func resourceTFEPolicySetParameterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	ps := d.Get("policy_set_id").(string)
	policySet, err := config.Client.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %w", ps, err)
	}

	log.Printf("[DEBUG] Read parameter: %s", d.Id())
	parameter, err := config.Client.PolicySetParameters.Read(ctx, policySet.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] PolicySetParameter %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading parameter %s: %w", d.Id(), err)
	}

	// Update config.
	d.Set("key", parameter.Key)
	d.Set("sensitive", parameter.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !parameter.Sensitive {
		d.Set("value", parameter.Value)
	}

	return nil
}

func resourceTFEPolicySetParameterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	ps := d.Get("policy_set_id").(string)
	policySet, err := config.Client.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %w", ps, err)
	}

	// Create a new options struct.
	options := tfe.PolicySetParameterUpdateOptions{
		Key:       tfe.String(d.Get("key").(string)),
		Value:     tfe.String(d.Get("value").(string)),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
	}

	log.Printf("[DEBUG] Update parameter: %s", d.Id())
	_, err = config.Client.PolicySetParameters.Update(ctx, policySet.ID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating parameter %s: %w", d.Id(), err)
	}

	return resourceTFEPolicySetParameterRead(d, meta)
}

func resourceTFEPolicySetParameterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	ps := d.Get("policy_set_id").(string)
	policySet, err := config.Client.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %w", ps, err)
	}

	log.Printf("[DEBUG] Delete parameter: %s", d.Id())
	err = config.Client.PolicySetParameters.Delete(ctx, policySet.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting parameter %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEPolicySetParameterImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid parameter import format: %s (expected <POLICY SET ID>/<PARAMETER ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("policy_set_id", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
