package tfe

import (
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var variableSetIdRegexp = regexp.MustCompile("varset-[a-zA-Z0-9]{16}$")

func resourceTFEVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEVariableSetCreate,
		Read:   resourceTFEVariableSetRead,
		Update: resourceTFEVariableSetUpdate,
		Delete: resourceTFEVariableSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"global": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspaces"},
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspaces": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFEVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.VariableSetCreateOptions{
		Name:        tfe.String(name),
		Description: tfe.String(d.Get("description").(string)),
		Global:      tfe.Bool(d.Get("global").(bool)),
	}

	variableSet, err := tfeClient.VariableSets.Create(ctx, organization, &options)
	if err != nil {
		return fmt.Errorf(
			"Error creating variable set %s, for organization: %s: %v", name, organization, err)
	}

	if workspaceIDs, workspacesSet := d.GetOk("workspaces"); !*options.Global && workspacesSet {
		log.Printf("[DEBUG] Assign variable set %s to workspaces %v", name, workspaceIDs)

		assignOptions := tfe.VariableSetAssignOptions{}
		for _, workspaceID := range workspaceIDs.(*schema.Set).List() {
			assignOptions.Workspaces = append(assignOptions.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
		}

		variableSet, err = tfeClient.VariableSets.Assign(ctx, variableSet.ID, &assignOptions)
		if err != nil {
			return fmt.Errorf(
				"Error assigning variable set %s (%s) to given workspaces: %v", name, variableSet.ID, err)
		}
	}

	d.SetId(variableSet.ID)

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of variable set: %s", id)
	variableSet, err := tfeClient.VariableSets.Read(ctx, id, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOps{tfe.VariableSetWorkspaces},
	})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable set %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %v", id, err)
	}

	// Update the config.
	d.Set("name", variableSet.Name)
	d.Set("description", variableSet.Description)
	d.Set("global", variableSet.Global)
	d.Set("organization", variableSet.Organization)
	d.Set("workspaces", variableSet.Workspaces)

	return nil
}

func resourceTFEVariableSetUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("global") {
		options := tfe.VariableSetUpdateOptions{
			Name:        tfe.String(d.Get("name").(string)),
			Description: tfe.String(d.Get("description").(string)),
			Global:      tfe.Bool(d.Get("global").(bool)),
		}

		log.Printf("[DEBUG] Update variable set: %s", id)
		_, err := tfeClient.VariableSets.Update(ctx, id, &options)
		if err != nil {
			return fmt.Errorf("Error updateing variable %s: %v", id, err)
		}
	}

	if d.HasChanges("workspaces") {
		workspaceIDs := d.Get("workspaces")
		assignOptions := tfe.VariableSetAssignOptions{}
		for _, workspaceID := range workspaceIDs.(*schema.Set).List() {
			assignOptions.Workspaces = append(assignOptions.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
		}

		log.Printf("[DEBUG] Assign variable set %s to workspaces %v", id, workspaceIDs)
		vs, errr := tfeClient.VariableSets.Assign(ctx, id, &assignOptions)
		if errr != nil {
			return fmt.Errorf(
				"Error assigning variable set %s (%s) to given workspaces: %v", vs.Name, id, errr)
		}
	}

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete variable set: %s", id)
	err := tfeClient.VariableSets.Delete(ctx, id)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable set %s: %v", id, err)
	}

	return nil
}
