package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEVariableCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	// Create a new options struct.
	options := tfe.VariableCreateOptions{
		Key:       tfe.String(key),
		Value:     tfe.String(d.Get("value").(string)),
		Category:  tfe.Category(tfe.CategoryType(category)),
		HCL:       tfe.Bool(d.Get("hcl").(bool)),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
		Workspace: ws,
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := tfeClient.Variables.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := tfeClient.Variables.Read(ctx, d.Id())
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
	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.VariableUpdateOptions{
		Key:       tfe.String(d.Get("key").(string)),
		Value:     tfe.String(d.Get("value").(string)),
		HCL:       tfe.Bool(d.Get("hcl").(bool)),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err := tfeClient.Variables.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return resourceTFEVariableRead(d, meta)
}

func resourceTFEVariableDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err := tfeClient.Variables.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}

//TODO: Gotta fix this too
func resourceTFEVariableImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	workspace_id, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], tfeClient.Workspaces)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving workspace %s from organization %s: %v", s[0], s[1], err)
	}
	d.Set("workspace_id", workspace_id)
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
