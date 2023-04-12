// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEVariableOld() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEVariableCreate,
		Read:   resourceTFEVariableRead,
		Update: resourceTFEVariableUpdate,
		Delete: resourceTFEVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEVariableImporter,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeVariableResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeVariableStateUpgradeV0,
				Version: 0,
			},
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

			"category": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.CategoryEnv),
						string(tfe.CategoryTerraform),
					},
					false,
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"hcl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"workspace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Required:     false,
				ForceNew:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
				ValidateFunc: validation.StringMatch(
					workspaceIDRegexp,
					"must be a valid workspace ID (ws-<RANDOM STRING>)",
				),
			},

			"variable_set_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Required:     false,
				ForceNew:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
				ValidateFunc: validation.StringMatch(
					variableSetIDRegexp,
					"must be a valid variable set ID (varset-<RANDOM STRING>)",
				),
			},
		},

		CustomizeDiff: forceRecreateResourceIf(),
	}
}

func forceRecreateResourceIf() schema.CustomizeDiffFunc {
	/*
		Destroy and add a new resource when:
		1. the parameter key changed and the param sensitive is set to true
		2. the parameter sensitive changed from true to false
	*/
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		wasSensitiveVar := d.HasChange("sensitive") && !(d.Get("sensitive").(bool))

		if wasSensitiveVar {
			return d.ForceNew("sensitive")
		} else if d.HasChange("key") && d.Get("sensitive").(bool) {
			return d.ForceNew("key")
		}
		return nil
	}
}

func resourceTFEVariableCreate(d *schema.ResourceData, meta interface{}) error {
	// Switch to variable set variable logic if we need to
	_, variableSetIDProvided := d.GetOk("variable_set_id")
	if variableSetIDProvided {
		return resourceTFEVariableSetVariableCreate(d, meta)
	}

	config := meta.(ConfiguredClient)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Create a new options struct.
	options := tfe.VariableCreateOptions{
		Key:         tfe.String(key),
		Value:       tfe.String(d.Get("value").(string)),
		Category:    tfe.Category(tfe.CategoryType(category)),
		HCL:         tfe.Bool(d.Get("hcl").(bool)),
		Sensitive:   tfe.Bool(d.Get("sensitive").(bool)),
		Description: tfe.String(d.Get("description").(string)),
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := config.Client.Variables.Create(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %w", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableSetVariableCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get the variable set
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := config.Client.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %w", variableSetID, err)
	}

	// Create a new options struct.
	options := tfe.VariableSetVariableCreateOptions{
		Key:         tfe.String(key),
		Value:       tfe.String(d.Get("value").(string)),
		Category:    tfe.Category(tfe.CategoryType(category)),
		HCL:         tfe.Bool(d.Get("hcl").(bool)),
		Sensitive:   tfe.Bool(d.Get("sensitive").(bool)),
		Description: tfe.String(d.Get("description").(string)),
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := config.Client.VariableSetVariables.Create(ctx, vs.ID, &options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %w", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableRead(d *schema.ResourceData, meta interface{}) error {
	// Switch to variable set variable logic if we need to
	_, variableSetIDProvided := d.GetOk("variable_set_id")
	if variableSetIDProvided {
		return resourceTFEVariableSetVariableRead(d, meta)
	}

	config := meta.(ConfiguredClient)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s no longer exists", workspaceID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := config.Client.Variables.Read(ctx, ws.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %w", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("description", variable.Description)
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEVariableSetVariableRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the variable set
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := config.Client.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable set %s no longer exists", variableSetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf(
			"Error retrieving variable set %s: %w", variableSetID, err)
	}

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := config.Client.VariableSetVariables.Read(ctx, vs.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %w", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("description", variable.Description)
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	// Switch to variable set variable logic if we need to
	_, variableSetIDProvided := d.GetOk("variable_set_id")
	if variableSetIDProvided {
		return resourceTFEVariableSetVariableUpdate(d, meta)
	}

	config := meta.(ConfiguredClient)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Create a new options struct.
	options := tfe.VariableUpdateOptions{
		Key:         tfe.String(d.Get("key").(string)),
		Value:       tfe.String(d.Get("value").(string)),
		HCL:         tfe.Bool(d.Get("hcl").(bool)),
		Sensitive:   tfe.Bool(d.Get("sensitive").(bool)),
		Description: tfe.String(d.Get("description").(string)),
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err = config.Client.Variables.Update(ctx, ws.ID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %w", d.Id(), err)
	}

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableSetVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the variable set.
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := config.Client.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %w", variableSetID, err)
	}

	// Create a new options struct.
	options := tfe.VariableSetVariableUpdateOptions{
		Key:         tfe.String(d.Get("key").(string)),
		Value:       tfe.String(d.Get("value").(string)),
		HCL:         tfe.Bool(d.Get("hcl").(bool)),
		Sensitive:   tfe.Bool(d.Get("sensitive").(bool)),
		Description: tfe.String(d.Get("description").(string)),
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err = config.Client.VariableSetVariables.Update(ctx, vs.ID, d.Id(), &options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %w", d.Id(), err)
	}

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableDelete(d *schema.ResourceData, meta interface{}) error {
	// Switch to variable set variable logic if we need to
	_, variableSetIDProvided := d.GetOk("variable_set_id")
	if variableSetIDProvided {
		return resourceTFEVariableSetVariableDelete(d, meta)
	}

	config := meta.(ConfiguredClient)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err = config.Client.Variables.Delete(ctx, ws.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEVariableSetVariableDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the variable set.
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := config.Client.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %w", variableSetID, err)
	}

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err = config.Client.VariableSetVariables.Delete(ctx, vs.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEVariableImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <ORGANIZATION>/<WORKSPACE|VARIABLE SET>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	varsetIDUsed := variableSetIDRegexp.MatchString(s[1])
	if varsetIDUsed {
		d.Set("variable_set_id", s[1])
	} else {
		// Set the fields that are part of the import ID.
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], config.Client)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace %s from organization %s: %w", s[1], s[0], err)
		}
		d.Set("workspace_id", workspaceID)
	}

	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
