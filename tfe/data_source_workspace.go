package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
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

			"global_remote_state": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"remote_state_consumer_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"assessments_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"operations": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"policy_check_failures": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"resource_count": {
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

			"speculative_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ssh_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"structured_run_output_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"tag_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

			"trigger_patterns": {
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

						"branch": {
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

						"tags_regex": {
							Type:     schema.TypeString,
							Computed: true,
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
	d.Set("assessments_enabled", workspace.AssessmentsEnabled)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("policy_check_failures", workspace.PolicyCheckFailures)

	// If target tfe instance predates projects, then workspace.Project will be nil
	if workspace.Project != nil {
		d.Set("project_id", workspace.Project.ID)
	}

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
			"tags_regex":         workspace.VCSRepo.TagsRegex,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	d.SetId(workspace.ID)

	return nil
}
