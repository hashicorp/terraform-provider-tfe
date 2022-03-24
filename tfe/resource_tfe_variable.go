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

func resourceTFEVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEVariableCreate,
		Read:   resourceTFEVariableRead,
		Update: resourceTFEVariableUpdate,
		Delete: resourceTFEVariableDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFEVariableImporter,
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
					workspaceIdRegexp,
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
					variableSetIdRegexp,
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
	//Switch to variable set variable logic if we need to
	_, variableSetIdProvided := d.GetOk("variable_set_id")
	if variableSetIdProvided {
		return resourceTFEVariableSetVariableCreate(d, meta)
	}

	tfeClient := meta.(*tfe.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get the workspace if workspace_id present
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
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
	variable, err := tfeClient.Variables.Create(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableSetVariableCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get the variable set
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := tfeClient.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %v", variableSetID, err)
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
	variable, err := tfeClient.VariableSetVariables.Create(ctx, vs.ID, &options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableRead(d *schema.ResourceData, meta interface{}) error {
	//Switch to variable set variable logic if we need to
	_, variableSetIdProvided := d.GetOk("variable_set_id")
	if variableSetIdProvided {
		return resourceTFEVariableSetVariableRead(d, meta)
	}

	tfeClient := meta.(*tfe.Client)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s no longer exists", workspaceID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := tfeClient.Variables.Read(ctx, ws.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("description", string(variable.Description))
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEVariableSetVariableRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the variable set
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := tfeClient.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable set %s no longer exists", variableSetID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf(
			"Error retrieving variable set %s: %v", variableSetID, err)
	}

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := tfeClient.VariableSetVariables.Read(ctx, vs.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("description", string(variable.Description))
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	//Switch to variable set variable logic if we need to
	_, variableSetIdProvided := d.GetOk("variable_set_id")
	if variableSetIdProvided {
		return resourceTFEVariableSetVariableUpdate(d, meta)
	}

	tfeClient := meta.(*tfe.Client)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
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
	_, err = tfeClient.Variables.Update(ctx, ws.ID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableSetVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the variable set.
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := tfeClient.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %v", variableSetID, err)
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
	_, err = tfeClient.VariableSetVariables.Update(ctx, vs.ID, d.Id(), &options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableDelete(d *schema.ResourceData, meta interface{}) error {
	//Switch to variable set variable logic if we need to
	_, variableSetIdProvided := d.GetOk("variable_set_id")
	if variableSetIdProvided {
		return resourceTFEVariableSetVariableDelete(d, meta)
	}

	tfeClient := meta.(*tfe.Client)

	// Get the workspace.
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err = tfeClient.Variables.Delete(ctx, ws.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEVariableSetVariableDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the variable set.
	variableSetID := d.Get("variable_set_id").(string)
	vs, err := tfeClient.VariableSets.Read(ctx, variableSetID, nil)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving variable set %s: %v", variableSetID, err)
	}

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err = tfeClient.VariableSetVariables.Delete(ctx, vs.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEVariableImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <ORGANIZATION>/<WORKSPACE|VARIABLE SET>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	varsetIdUsed := variableSetIdRegexp.MatchString(s[1])
	if varsetIdUsed {
		d.Set("variable_set_id", s[1])
	} else {
		// Set the fields that are part of the import ID.
		workspace_id, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], tfeClient)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace %s from organization %s: %v", s[1], s[0], err)
		}
		d.Set("workspace_id", workspace_id)
	}

	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
