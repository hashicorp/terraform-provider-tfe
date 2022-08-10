package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a workspace." +
			"\n\n ~> **NOTE:** Using `global_remote_state` or `remote_state_consumer_ids` requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202104-1.",

		Read: dataSourceTFEWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"allow_destroy_plan": {
				Description: "Indicates whether destroy plans can be queued on the workspace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_apply": {
				Description: "Indicates whether to automatically apply changes when a Terraform plan is successful.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"file_triggers_enabled": {
				Description: "Indicates whether runs are triggered based on the changed files in a VCS push (if `true`) or always triggered on every push (if `false`).",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"global_remote_state": {
				Description: "Whether the workspace should allow all workspaces in the organization to access its state data during runs. If false, then only specifically approved workspaces can access its state (determined by the `remote_state_consumer_ids` argument).",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"remote_state_consumer_ids": {
				Description: "A set of workspace IDs that will be set as the remote state consumers for the given workspace. Cannot be used if `global_remote_state` is set to `true`.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"operations": {
				Description: "Indicates whether the workspace is using remote execution mode. Set to `false` to switch execution mode to local. `true` by default.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"policy_check_failures": {
				Description: "The number of policy check failures from the latest run.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"queue_all_runs": {
				Description: "Indicates whether the workspace will automatically perform runs in response to webhooks immediately after its creation. If `false`, an initial run must be manually queued to enable future automatic runs.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"resource_count": {
				Description: "The number of resources managed by the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"run_failures": {
				Description: "The number of run failures on the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"runs_count": {
				Description: "The number of runs on the workspace.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"speculative_enabled": {
				Description: "Indicates whether this workspace allows speculative plans.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"ssh_key_id": {
				Description: "The ID of an SSH key assigned to the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"structured_run_output_enabled": {
				Description: "Indicates whether runs in this workspace use the enhanced apply UI.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"tag_names": {
				Description: "The names of tags added to this workspace.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"terraform_version": {
				Description: "The version (or version constraint) of Terraform used for this workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"trigger_prefixes": {
				Description: "List of trigger prefixes that describe the paths Terraform Cloud monitors for changes, in addition to the working directory. Trigger prefixes are always appended to the root directory of the repository. Terraform Cloud or Terraform Enterprise will start a run when files are changed in any directory path matching the provided set of prefixes.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"trigger_patterns": {
				Description: "List of [glob patterns](https://www.terraform.io/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) that describe the files Terraform Cloud monitors for changes. Trigger patterns are always appended to the root directory of the repository. Only available for Terraform Cloud.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"working_directory": {
				Description: "A relative path that Terraform will execute within.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"vcs_repo": {
				Description: "Settings for the workspace's VCS repository.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to your VCS repository in the format `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"branch": {
							Description: "The repository branch that Terraform will execute from.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"ingress_submodules": {
							Description: "Indicates whether submodules should be fetched when cloning the VCS repository.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"oauth_token_id": {
							Description: "OAuth token ID of the configured VCS connection.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
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
			return fmt.Errorf("could not find workspace %s/%s", organization, name)
		}
		return fmt.Errorf("Error retrieving workspace: %w", err)
	}
	// Update the config.
	d.Set("allow_destroy_plan", workspace.AllowDestroyPlan)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("description", workspace.Description)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("policy_check_failures", workspace.PolicyCheckFailures)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("resource_count", workspace.ResourceCount)
	d.Set("run_failures", workspace.RunFailures)
	d.Set("runs_count", workspace.RunsCount)
	d.Set("speculative_enabled", workspace.SpeculativeEnabled)
	d.Set("structured_run_output_enabled", workspace.StructuredRunOutputEnabled)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("trigger_patterns", workspace.TriggerPatterns)
	d.Set("working_directory", workspace.WorkingDirectory)

	// Set remote_state_consumer_ids if global_remote_state is false
	globalRemoteState := workspace.GlobalRemoteState
	if globalRemoteState {
		if err := d.Set("remote_state_consumer_ids", []string{}); err != nil {
			return err
		}
	} else {
		legacyGlobalState, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(workspace.ID, tfeClient)

		if err != nil {
			return fmt.Errorf(
				"Error reading remote state consumers for workspace %s: %w", workspace.ID, err)
		}

		if legacyGlobalState {
			globalRemoteState = true
		}
		d.Set("remote_state_consumer_ids", remoteStateConsumerIDs)
	}
	d.Set("global_remote_state", globalRemoteState)

	if workspace.SSHKey != nil {
		d.Set("ssh_key_id", workspace.SSHKey.ID)
	}

	// Update the tag names
	var tagNames []interface{}
	for _, tagName := range workspace.TagNames {
		tagNames = append(tagNames, tagName)
	}
	d.Set("tag_names", tagNames)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":         workspace.VCSRepo.Identifier,
			"branch":             workspace.VCSRepo.Branch,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":     workspace.VCSRepo.OAuthTokenID,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	d.SetId(workspace.ID)

	return nil
}
