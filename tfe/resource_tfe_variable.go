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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceTFEVariableCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the workspace.
	ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		log.Printf("[DEBUG] Workspace %s no longer exists, so variable doesn't exist either", workspace)
		d.SetId("")
		return nil
	}

	// Check if variable with key already exists. Overwrite if necessary
	vl, _ := tfeClient.Variables.List(ctx, ws.ID, tfe.VariableListOptions{})
	for _, v := range vl.Items {
		if v.Key == key {
			overwrite := d.Get("overwrite").(bool)
			if !overwrite {
				return fmt.Errorf("Error creating %s variable %s: variable already exists", category, key)
			} else {
				// overwrite existing variable and assume its ID
				options := tfe.VariableUpdateOptions{
					Key:         tfe.String(d.Get("key").(string)),
					Value:       tfe.String(d.Get("value").(string)),
					HCL:         tfe.Bool(d.Get("hcl").(bool)),
					Sensitive:   tfe.Bool(d.Get("sensitive").(bool)),
					Description: tfe.String(d.Get("description").(string)),
				}

				log.Printf("[DEBUG] Update variable: %s", v.ID)
				_, err = tfeClient.Variables.Update(ctx, ws.ID, v.ID, options)
				if err != nil {
					return fmt.Errorf("Error updating variable %s: %v", v.ID, err)
				}
				d.SetId(v.ID)
				return resourceTFEVariableRead(d, meta)
			}
			break
		}
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

func resourceTFEVariableRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the workspace.
	ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s from organization %s: %v", workspace, organization, err)
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

func resourceTFEVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the workspace.
	ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s from organization %s: %v", workspace, organization, err)
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

func resourceTFEVariableDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get organization and workspace.
	organization, workspace, err := unpackWorkspaceID(d.Get("workspace_id").(string))
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	// Get the workspace.
	ws, err := tfeClient.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s from organization %s: %v", workspace, organization, err)
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

func resourceTFEVariableImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("workspace_id", s[0]+"/"+s[1])
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
