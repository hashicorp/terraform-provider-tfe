// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTfeTeamAccessResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.AccessAdmin),
						string(tfe.AccessRead),
						string(tfe.AccessPlan),
						string(tfe.AccessWrite),
					},
					false,
				),
			},

			"team_id": {
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

func resourceTfeTeamAccessStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	config := meta.(ConfiguredClient)

	// This state upgrader (schema version 0 -> 1, migrating the legacy
	// "<ORG>/<WORKSPACE NAME>" workspace_id format to the external workspace
	// ID) remains on the go-tfe v1 client. Its existing unit test
	// (TestResourceTfeTeamAccessStateUpgradeV0) relies on go-tfe v1's
	// swappable client.Workspaces service-interface field to mock workspace
	// reads; go-tfe v2's generated client.API tree has no equivalent
	// mockable seam. Every other operation on the tfe_team_access resource
	// (create, read, update, delete, and the current-version import path)
	// uses the go-tfe v2 client.
	humanID := rawState["workspace_id"].(string)
	id, err := fetchWorkspaceExternalID(humanID, config.Client)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s: %w", humanID, err)
	}

	rawState["workspace_id"] = id
	return rawState, nil
}
