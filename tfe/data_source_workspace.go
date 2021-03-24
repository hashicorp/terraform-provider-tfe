package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "Data source \"tfe_workspace\"\n\n\"external_id\": [DEPRECATED] Use id instead. The external_id attribute will be removed in the future. See the CHANGELOG to learn more: https://github.com/hashicorp/terraform-provider-tfe/blob/v0.24.0/CHANGELOG.md",
		Read:               dataSourceTFEWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"allow_destroy_plan": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"auto_apply": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"file_triggers_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"operations": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"speculative_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ssh_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"trigger_prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"working_directory": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vcs_repo": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ingress_submodules": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"oauth_token_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"resource_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"policy_check_failures": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"run_failures": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"runs_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"external_id": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use id instead. The external_id attribute will be removed in the future. See the CHANGELOG to learn more: https://github.com/hashicorp/terraform-provider-tfe/blob/v0.24.0/CHANGELOG.md",
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
	d.Set("allow_destroy_plan", workspace.AllowDestroyPlan)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("speculative_enabled", workspace.SpeculativeEnabled)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("resource_count", workspace.ResourceCount)
	d.Set("policy_check_failures", workspace.PolicyCheckFailures)
	d.Set("run_failures", workspace.RunFailures)
	d.Set("runs_count", workspace.RunsCount)
	// TODO: remove when external_id is removed
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
