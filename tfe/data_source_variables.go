package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceVariables() *schema.Resource {
	varSchema := map[string]*schema.Schema{
		"category": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"hcl": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"sensitive": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"value": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},
	}
	return &schema.Resource{
		Read: dataSourceVariableRead,

		Schema: map[string]*schema.Schema{
			"environment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: varSchema,
				},
			},
			"terraform": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: varSchema,
				},
			},
			"variables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: varSchema,
				},
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceVariableRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace: %s", workspaceID)

	totalEnvVariables := make([]interface{}, 0)
	totalTerraformVariables := make([]interface{}, 0)

	options := tfe.VariableListOptions{}

	for {
		variableList, err := tfeClient.Variables.List(ctx, workspaceID, options)
		if err != nil {
			return fmt.Errorf("Error retrieving variable list: %w", err)
		}
		terraformVars := make([]interface{}, 0)
		envVars := make([]interface{}, 0)
		for _, variable := range variableList.Items {
			result := make(map[string]interface{})
			result["id"] = variable.ID
			result["category"] = variable.Category
			result["hcl"] = variable.HCL
			result["name"] = variable.Key
			result["sensitive"] = variable.Sensitive
			result["value"] = variable.Value
			if variable.Category == "terraform" {
				terraformVars = append(terraformVars, result)
			} else if variable.Category == "env" {
				envVars = append(envVars, result)
			}
		}

		totalEnvVariables = append(totalEnvVariables, envVars...)
		totalTerraformVariables = append(totalTerraformVariables, terraformVars...)

		// Exit the loop when we've seen all pages.
		if variableList.CurrentPage >= variableList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = variableList.NextPage
	}

	d.SetId(fmt.Sprintf("variables/%v", workspaceID))
	d.Set("variables", append(totalTerraformVariables, totalEnvVariables...))
	d.Set("terraform", totalTerraformVariables)
	d.Set("environment", totalEnvVariables)
	return nil
}
