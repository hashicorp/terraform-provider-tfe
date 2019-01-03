package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"auto_apply": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ssh_key_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"queue_all_runs": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"terraform_version": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"working_directory": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"vcs_repo": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"branch": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"ingress_submodules": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},

						"oauth_token_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"external_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	_, err := tfeClient.Workspaces.Read(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find workspace %s/%s", organization, name)
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	d.SetId(organization + "/" + name)
	return resourceTFEWorkspaceRead(d, meta)
}
