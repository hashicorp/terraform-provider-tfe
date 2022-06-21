package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEVariableSetWorkspaceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEVariableSetWorkspaceAttachmentCreate,
		Read:   resourceTFEVariableSetWorkspaceAttachmentRead,
		Delete: resourceTFEVariableSetWorkspaceAttachmentDelete,
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

func resourceTFEVariableSetWorkspaceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
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

	return resourceTFEVariableSetWorkspaceAttachmentRead(d, meta)
}

func resourceTFEVariableSetWorkspaceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	vSId, wId, err := DecodeVariableSetWorkspaceAttachment(d.Id())

	if err != nil {
		return fmt.Errorf("error decoding ID (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Read configuration of variable set workplace attachment: %s", d.Id())
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

func resourceTFEVariableSetWorkspaceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	vSId, wId, err := DecodeVariableSetWorkspaceAttachment(d.Id())

	if err != nil {
		return fmt.Errorf("error decoding ID (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Delete workspace (%s) from variable set (%s)", wId, vSId)
	removeOptions := tfe.VariableSetRemoveFromWorkspacesOptions{}
	removeOptions.Workspaces = append(removeOptions.Workspaces, &tfe.Workspace{ID: wId})

	err = tfeClient.VariableSets.RemoveFromWorkspaces(ctx, vSId, &removeOptions)
	if err != nil {
		return fmt.Errorf(
			"Error removing workspace %s from variable set %s: %w", wId, vSId, err)
	}

	return nil
}

func encodeVariableSetWorkspaceAttachment(vSId, wId string) string {
	return fmt.Sprintf("%s_%s", vSId, wId)
}

func DecodeVariableSetWorkspaceAttachment(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of variable-set-id_workspace-id, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
