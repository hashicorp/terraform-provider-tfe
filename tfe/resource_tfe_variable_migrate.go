package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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

func resourceTfeVariableStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	tfeClient := meta.(*tfe.Client)
	workspaces := tfeClient.Workspaces

	humanID := rawState["workspace_id"].(string)
	id, err := fetchWorkspaceExternalID(humanID, workspaces)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s: %v", humanID, err)
	}

	rawState["workspace_id"] = id
	return rawState, nil
}
