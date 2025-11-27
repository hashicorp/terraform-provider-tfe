// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEStackVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEStackVariableSetCreate,
		Read:   resourceTFEStackVariableSetRead,
		Delete: resourceTFEStackVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEStackVariableSetImporter,
		},

		Schema: map[string]*schema.Schema{
			"variable_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stack_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEStackVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	vSID := d.Get("variable_set_id").(string)
	stkID := d.Get("stack_id").(string)

	applyOptions := tfe.VariableSetApplyToStacksOptions{}
	applyOptions.Stacks = append(applyOptions.Stacks, &tfe.Stack{ID: stkID})

	err := config.Client.VariableSets.ApplyToStacks(ctx, vSID, &applyOptions)
	if err != nil {
		return fmt.Errorf(
			"Error applying variable set id %s to stack %s: %w", vSID, stkID, err)
	}

	id := encodeVariableSetStackAttachment(stkID, vSID)
	d.SetId(id)

	return resourceTFEStackVariableSetRead(d, meta)
}

func resourceTFEStackVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	stkID := d.Get("stack_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of stack variable set: %s", d.Id())
	vS, err := config.Client.VariableSets.Read(ctx, vSID, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetStacks},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Verify stack listed in variable set
	check := false
	for _, stack := range vS.Stacks {
		if stack.ID == stkID {
			check = true
			d.Set("stack_id", stkID)
			break
		}
	}
	if !check {
		log.Printf("[DEBUG] Stack %s not attached to variable set %s. Removing from state.", stkID, vSID)
		d.SetId("")
		return nil
	}

	d.Set("variable_set_id", vSID)
	return nil
}

func resourceTFEStackVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	stkID := d.Get("stack_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Delete stack (%s) from variable set (%s)", stkID, vSID)
	removeOptions := tfe.VariableSetRemoveFromStacksOptions{}
	removeOptions.Stacks = append(removeOptions.Stacks, &tfe.Stack{ID: stkID})

	err := config.Client.VariableSets.RemoveFromStacks(ctx, vSID, &removeOptions)
	if err != nil {
		return fmt.Errorf(
			"Error removing stack %s from variable set %s: %w", stkID, vSID, err)
	}

	return nil
}

func encodeVariableSetStackAttachment(stkID, vSID string) string {
	return fmt.Sprintf("%s_%s", stkID, vSID)
}

func resourceTFEStackVariableSetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <STACK ID/VARSET NAME>
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format, expected STACK_ID/VARIABLE_SET_NAME (got %q)", d.Id())
	}

	stkID := parts[0]
	vSName := parts[1]

	config := meta.(ConfiguredClient)

	// Ensure a stack with this ID exists
	_, err := config.Client.Stacks.Read(ctx, stkID)
	if err != nil {
		return nil, fmt.Errorf("error reading stack %s: %w", stkID, err)
	}

	// List all variable sets and find the one with the matching name
	options := &tfe.VariableSetListOptions{}
	for {
		list, err := config.Client.VariableSets.List(ctx, "", options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving variable sets: %w", err)
		}

		for _, vSet := range list.Items {
			if vSet.Name == vSName {
				d.SetId(encodeVariableSetStackAttachment(stkID, vSet.ID))
				d.Set("stack_id", stkID)
				d.Set("variable_set_id", vSet.ID)
				return []*schema.ResourceData{d}, nil
			}
		}

		if list.CurrentPage >= list.TotalPages {
			break
		}

		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("could not find variable set %s", vSName)
}
