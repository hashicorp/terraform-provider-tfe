// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceLock() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceLockCreate,
		Read:   resourceTFEWorkspaceLockRead,
		Delete: resourceTFEWorkspaceLockDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEWorkspaceLockCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	wID := d.Get("workspace_id").(string)

	applyOptions := tfe.WorkspaceLockOptions{}

	if v, ok := d.GetOk("reason"); ok {
		applyOptions.Reason = tfe.String(v.(string))
	}

	_, err := config.Client.Workspaces.Lock(ctx, wID, applyOptions)
	if err != nil {
		return fmt.Errorf(
			"Error locking workspace %s: %w", wID, err)
	}

	d.SetId(wID)

	return resourceTFEWorkspaceLockRead(d, meta)
}

func resourceTFEWorkspaceLockRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	wID := d.Id()

	log.Printf("[DEBUG] Read configuration of workspace %s", d.Id())
	ws, err := config.Client.Workspaces.ReadByID(ctx, wID)

	if err != nil {
		return fmt.Errorf("Error reading Workspace %s: %w", d.Id(), err)
	}

	if !ws.Locked {
		log.Printf("[DEBUG] Workspace ID %s is no longer locked", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(wID)

	return nil
}

func resourceTFEWorkspaceLockDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	wID := d.Id()

	log.Printf("[DEBUG] Unlock workspace (%s)", wID)

	_, err := config.Client.Workspaces.Unlock(ctx, wID)
	if err != nil {
		return fmt.Errorf(
			"Error unlocking workspace %s: %w", wID, err)
	}

	return nil
}
