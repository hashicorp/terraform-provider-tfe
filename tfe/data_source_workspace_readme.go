package tfe

import (
	"fmt"
	"io/ioutil"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceReadme() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceReadReadme,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"raw_markdown": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEWorkspaceReadReadme(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	readme, err := tfeClient.Workspaces.Readme(ctx, workspaceID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find workspace with ID %s", workspaceID)
		}
		return fmt.Errorf(
			"Error retrieving readme %s: %v", workspaceID, err)
	}

	rs, err := ioutil.ReadAll(readme)
	if err != nil {
		return fmt.Errorf("Error reading readme: %s", err)
	}

	d.Set("raw_markdown", rs)

	d.SetId(workspaceID)

	return nil
}
