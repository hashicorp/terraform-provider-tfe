// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFERunTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERunTriggerCreate,
		Read:   resourceTFERunTriggerRead,
		Delete: resourceTFERunTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
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
	config := meta.(ConfiguredClient)

	// Get attributes
	workspaceID := d.Get("workspace_id").(string)
	sourceableID := d.Get("sourceable_id").(string)

	// Create a new options struct
	options := tfe.RunTriggerCreateOptions{
		Sourceable: &tfe.Workspace{
			ID: sourceableID,
		},
	}

	log.Printf("[DEBUG] Create run trigger on workspace %s with sourceable %s", workspaceID, sourceableID)
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		runTrigger, err := config.Client.RunTriggers.Create(ctx, workspaceID, options)
		if err == nil {
			d.SetId(runTrigger.ID)
			return nil
		}

		if strings.Contains(err.Error(), "Run Trigger creation locked") {
			log.Printf("[DEBUG] Run triggers are locked for workspace %s, will retry", workspaceID)
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if err != nil {
		return fmt.Errorf("Error creating run trigger on workspace %s with sourceable %s: %w", workspaceID, sourceableID, err)
	}

	return resourceTFERunTriggerRead(d, meta)
}

func resourceTFERunTriggerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read run trigger: %s", d.Id())
	runTrigger, err := config.Client.RunTriggers.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] run trigger %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading run trigger %s: %w", d.Id(), err)
	}

	d.Set("workspace_id", runTrigger.Workspace.ID)
	d.Set("sourceable_id", runTrigger.Sourceable.ID)

	return nil
}

func resourceTFERunTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete run trigger: %s", d.Id())
	err := config.Client.RunTriggers.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting run trigger %s: %w", d.Id(), err)
	}

	return nil
}
