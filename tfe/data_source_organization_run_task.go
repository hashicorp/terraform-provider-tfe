package tfe

import (
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOrganizationRunTask() *schema.Resource {
	return &schema.Resource{
		Description: "[Run tasks](https://www.terraform.io/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization." +
			"\n\nUse this data source to get information about an [Organization Run tasks](https://www.terraform.io/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task).",
		Read: dataSourceTFEOrganizationRunTaskRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the Run task.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"url": {
				Description: "URL to send a run task payload.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"category": {
				Description: "The type of task.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"enabled": {
				Description: "Whether the task will be run.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func dataSourceTFEOrganizationRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	task, err := fetchOrganizationRunTask(name, organization, tfeClient)
	if err != nil {
		return err
	}

	d.Set("url", task.URL)
	d.Set("category", task.Category)
	d.Set("enabled", task.Enabled)
	d.SetId(task.ID)

	return nil
}
