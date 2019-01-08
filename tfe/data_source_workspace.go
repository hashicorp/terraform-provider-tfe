package tfe

import (
	"fmt"
	"log"

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
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": &schema.Schema{
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

	log.Printf("[DEBUG] Read configuration of workspace: %s", name)
	workspace, err := tfeClient.Workspaces.Read(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find workspace %s/%s", organization, name)
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config.
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("external_id", workspace.ID)

	if workspace.SSHKey != nil {
		d.Set("ssh_key_id", workspace.SSHKey.ID)
	}

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":         workspace.VCSRepo.Identifier,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":     workspace.VCSRepo.OAuthTokenID,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	d.SetId(workspace.ID)

	return nil
}
