package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceVariables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVariableRead,

		Schema: map[string]*schema.Schema{
			"variables": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"category": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"hcl": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
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
	workspace_id := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace: %s", workspace_id)

	totalVariables := make([]interface{}, 0)
	variableList, err := tfeClient.Variables.List(ctx, workspace_id, tfe.VariableListOptions{})
	if err != nil {
		return fmt.Errorf("Error retrieving variable list: %v", err)
	}
	results := make([]interface{}, 0)
	for _, variable := range variableList.Items {
		result := make(map[string]interface{})
		result["name"] = variable.Key
		result["category"] = variable.Category
		result["hcl"] = variable.HCL
		if variable.Sensitive != true {
			result["value"] = variable.Value
		} else {
			result["value"] = "***"
		}

		results = append(results, result)
	}
	totalVariables = append(totalVariables, results...)
	d.SetId(fmt.Sprintf("variables/%v", workspace_id))
	d.Set("variables", totalVariables)
	log.Println(totalVariables)
	return nil
}
