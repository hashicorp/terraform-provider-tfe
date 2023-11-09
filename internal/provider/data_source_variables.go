// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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
			"env": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
			},
			"variable_set_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
			},
		},
	}
}

func dataSourceVariableRead(d *schema.ResourceData, meta interface{}) error {
	// Switch to variable set variable logic
	_, variableSetIDProvided := d.GetOk("variable_set_id")
	if variableSetIDProvided {
		return dataSourceVariableSetVariableRead(d, meta)
	}

	config := meta.(ConfiguredClient)

	// Get the name and organization.
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace: %s", workspaceID)

	totalEnvVariables := make([]interface{}, 0)
	totalTerraformVariables := make([]interface{}, 0)

	options := &tfe.VariableListOptions{}

	for {
		variableList, err := config.Client.Variables.List(ctx, workspaceID, options)
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
	d.Set("env", totalEnvVariables)
	return nil
}

func dataSourceVariableSetVariableRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the id.
	variableSetID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of variable set: %s", variableSetID)

	totalEnvVariables := make([]interface{}, 0)
	totalTerraformVariables := make([]interface{}, 0)

	options := tfe.VariableSetVariableListOptions{}

	for {
		variableList, err := config.Client.VariableSetVariables.List(ctx, variableSetID, &options)
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

	d.SetId(fmt.Sprintf("variables/%v", variableSetID))
	d.Set("variables", append(totalTerraformVariables, totalEnvVariables...))
	d.Set("terraform", totalTerraformVariables)
	d.Set("env", totalEnvVariables)
	return nil
}
