// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTfeVariableResourceV0() *schema.Resource {
	return &schema.Resource{
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
		},
	}
}

func resourceTfeVariableStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	config := meta.(ConfiguredClient)

	humanID := rawState["workspace_id"].(string)
	id, err := fetchWorkspaceExternalID(humanID, config.Client)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s: %w", humanID, err)
	}

	rawState["workspace_id"] = id
	return rawState, nil
}
