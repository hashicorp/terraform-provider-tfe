package tfe

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTFERunTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERunTriggerCreate,
		Read:   resourceTFERunTriggerRead,
		Update: resourceTFERunTriggerUpdate,
		Delete: resourceTFERunTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			// ForceNew if workspace_external_id is changing from a non-empty value to another non-empty value
			// If workspace_external_id is changing from an empty value to a non-empty value or a non-empty value
			// to an empty value, we know we are switching between workspace_external_id and workspace_id because
			// we ensure later that one of them has to be set.
			customdiff.ForceNewIfChange("workspace_external_id", func(old, new, meta interface{}) bool {
				oldWorkspaceExternalID := old.(string)
				newWorkspaceExternalID := new.(string)
				return oldWorkspaceExternalID != "" && newWorkspaceExternalID != ""
			}),
			// ForceNew if workspace_id is changing from a non-empty value to another non-empty value
			// If workspace_id is changing from an empty value to a non-empty value or a non-empty value
			// to an empty value, we know we are switching between workspace_external_id and workspace_id because
			// we ensure later that one of them has to be set.
			customdiff.ForceNewIfChange("workspace_id", func(old, new, meta interface{}) bool {
				oldWorkspaceID := old.(string)
				newWorkspaceID := new.(string)
				return oldWorkspaceID != "" && newWorkspaceID != ""
			}),
		),

		Schema: map[string]*schema.Schema{
			"workspace_external_id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"workspace_id"},
				Deprecated:    "Use workspace_id instead. The workspace_external_id attribute will be removed in the future. See the CHANGELOG to learn more: https://github.com/terraform-providers/terraform-provider-tfe/blob/v0.18.0/CHANGELOG.md",
			},
			"workspace_id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"workspace_external_id"},
			},
			"sourceable_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFERunTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get workspace ID
	workspaceExternalIDValue, workspaceExternalIDValueOk := d.GetOk("workspace_external_id")
	workspaceIDValue, workspaceIDValueOk := d.GetOk("workspace_id")
	if !workspaceExternalIDValueOk && !workspaceIDValueOk {
		return fmt.Errorf("One of workspace_id or workspace_external_id must be set")
	}

	var workspaceID string
	if workspaceExternalIDValueOk {
		workspaceID = workspaceExternalIDValue.(string)
	} else {
		workspaceID = workspaceIDValue.(string)
	}

	// Get attributes
	sourceableID := d.Get("sourceable_id").(string)

	// Create a new options struct
	options := tfe.RunTriggerCreateOptions{
		Sourceable: &tfe.Workspace{
			ID: sourceableID,
		},
	}

	log.Printf("[DEBUG] Create run trigger on workspace %s with sourceable %s", workspaceID, sourceableID)
	runTrigger, err := tfeClient.RunTriggers.Create(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("Error creating run trigger on workspace %s with sourceable %s: %v", workspaceID, sourceableID, err)
	}

	d.SetId(runTrigger.ID)

	return resourceTFERunTriggerRead(d, meta)
}

func resourceTFERunTriggerRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read run trigger: %s", d.Id())
	runTrigger, err := tfeClient.RunTriggers.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] run trigger %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading run trigger %s: %v", d.Id(), err)
	}

	// TODO: remove once workspace_external_id has been removed
	d.Set("workspace_external_id", runTrigger.Workspace.ID)

	d.Set("workspace_id", runTrigger.Workspace.ID)
	d.Set("sourceable_id", runTrigger.Sourceable.ID)

	return nil
}

// TODO: remove once workspace_external_id has been removed
// This update function is here because if you don't have an update function,
// you must set ForceNew. We can't set ForceNew right now because we need to use
// a customDiff to force new only if the value in workspace_id or workspace_external_id
// changes.
func resourceTFERunTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceTFERunTriggerRead(d, meta)
}

func resourceTFERunTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete run trigger: %s", d.Id())
	err := tfeClient.RunTriggers.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting run trigger %s: %v", d.Id(), err)
	}

	return nil
}
