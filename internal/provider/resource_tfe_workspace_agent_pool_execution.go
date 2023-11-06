// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceAgentPoolExecution() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceAgentPoolExecutionCreate,
		Read:   resourceTFEWorkspaceAgentPoolExecutionRead,
		Update: resourceTFEWorkspaceAgentPoolExecutionUpdate,
		Delete: resourceTFEWorkspaceAgentPoolExecutionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEWorkspaceAgentPoolExecutionCreate(d *schema.ResourceData, meta interface{}) error {
	// config := meta.(ConfiguredClient)

	// poolID := d.Get("agent_pool_id").(string)
	// workpaceID := d.Get("workspace_id").(string)

	return resourceTFEWorkspaceAgentPoolExecutionRead(d, meta)
}

func resourceTFEWorkspaceAgentPoolExecutionRead(d *schema.ResourceData, meta interface{}) error {
	// config := meta.(ConfiguredClient)

	return nil
}

func resourceTFEWorkspaceAgentPoolExecutionUpdate(d *schema.ResourceData, meta interface{}) error {
	// config := meta.(ConfiguredClient)

	// poolID := d.Get("agent_pool_id").(string)
	// workpaceID := d.Get("workspace_id").(string)

	return resourceTFEWorkspaceAgentPoolExecutionRead(d, meta)
}

func resourceTFEWorkspaceAgentPoolExecutionDelete(d *schema.ResourceData, meta interface{}) error {
	// config := meta.(ConfiguredClient)

	// poolID := d.Get("agent_pool_id").(string)
	// workpaceID := d.Get("workspace_id").(string)

	return nil
}
