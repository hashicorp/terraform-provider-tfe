package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceVariableSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceVariableSetCreate,
		Read:   resourceTFEWorkspaceVariableSetRead,
		Delete: resourceTFEWorkspaceVariableSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"variable_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEWorkspaceVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	vSId := d.Get("variable_set_id").(string)
	wId := d.Get("workspace_id").(string)

	applyOptions := tfe.VariableSetApplyToWorkspacesOptions{}
	applyOptions.Workspaces = append(applyOptions.Workspaces, &tfe.Workspace{ID: wId})

	err := tfeClient.VariableSets.ApplyToWorkspaces(ctx, vSId, &applyOptions)
	if err != nil {
		return fmt.Errorf(
			"Error applying variable set id %s to workspace %s: %w", vSId, wId, err)
	}

	id := encodeVariableSetWorkspaceAttachment(vSId, wId)
	d.SetId(id)

	return resourceTFEWorkspaceVariableSetRead(d, meta)
}

func resourceTFEWorkspaceVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	wId := d.Get("workspace_id").(string)
	vSId := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace variable set: %s", d.Id())
	vS, err := tfeClient.VariableSets.Read(ctx, vSId, &tfe.VariableSetReadOptions{
		Include: &[]tfe.VariableSetIncludeOpt{tfe.VariableSetWorkspaces},
	})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Verify workspace listed in variable set
	check := false
	for _, workspace := range vS.Workspaces {
		if workspace.ID == wId {
			check = true
			d.Set("workspace_id", wId)
		}
	}
	if !check {
		log.Printf("[DEBUG] Workspace %s not attached to variable set %s. Removing from state.", wId, vSId)
		d.SetId("")
		return nil
	}

	d.Set("variable_set_id", vSId)
	return nil
}

func resourceTFEWorkspaceVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	wId := d.Get("workspace_id").(string)
	vSId := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Delete workspace (%s) from variable set (%s)", wId, vSId)
	removeOptions := tfe.VariableSetRemoveFromWorkspacesOptions{}
	removeOptions.Workspaces = append(removeOptions.Workspaces, &tfe.Workspace{ID: wId})

	err := tfeClient.VariableSets.RemoveFromWorkspaces(ctx, vSId, &removeOptions)
	if err != nil {
		return fmt.Errorf(
			"Error removing workspace %s from variable set %s: %w", wId, vSId, err)
	}

	return nil
}

func encodeVariableSetWorkspaceAttachment(vSId, wId string) string {
	return fmt.Sprintf("%s_%s", vSId, wId)
}
