package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTFERunTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERunTriggerCreate,
		Read:   resourceTFERunTriggerRead,
		Delete: resourceTFERunTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace_external_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	// Get workspace
	workspaceID := d.Get("workspace_external_id").(string)

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

	// Update config
	d.Set("workspace_external_id", runTrigger.Workspace.ID)
	d.Set("sourceable_id", runTrigger.Sourceable.ID)

	return nil
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
