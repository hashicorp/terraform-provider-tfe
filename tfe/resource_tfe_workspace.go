package tfe

import (
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var workspaceIdRegexp = regexp.MustCompile("^ws-[a-zA-Z0-9]{16}$")

func resourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceCreate,
		Read:   resourceTFEWorkspaceRead,
		Update: resourceTFEWorkspaceUpdate,
		Delete: resourceTFEWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeWorkspaceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeWorkspaceStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"auto_apply": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"file_triggers_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"operations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"trigger_prefixes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"vcs_repo": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Required: true,
						},

						"branch": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"ingress_submodules": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"oauth_token_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.WorkspaceCreateOptions{
		Name:                tfe.String(name),
		AutoApply:           tfe.Bool(d.Get("auto_apply").(bool)),
		FileTriggersEnabled: tfe.Bool(d.Get("file_triggers_enabled").(bool)),
		Operations:          tfe.Bool(d.Get("operations").(bool)),
		QueueAllRuns:        tfe.Bool(d.Get("queue_all_runs").(bool)),
		WorkingDirectory:    tfe.String(d.Get("working_directory").(string)),
	}

	// Process all configured options.
	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = tfe.String(tfVersion.(string))
	}

	if tps, ok := d.GetOk("trigger_prefixes"); ok {
		for _, tp := range tps.([]interface{}) {
			options.TriggerPrefixes = append(options.TriggerPrefixes, tp.(string))
		}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &tfe.VCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
			Branch:            tfe.String(vcsRepo["branch"].(string)),
			IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
			OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = tfe.String(branch)
		}
	}

	log.Printf("[DEBUG] Create workspace %s for organization: %s", name, organization)
	workspace, err := tfeClient.Workspaces.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for organization %s: %v", name, organization, err)
	}

	d.SetId(workspace.ID)

	if sshKeyID, ok := d.GetOk("ssh_key_id"); ok {
		_, err = tfeClient.Workspaces.AssignSSHKey(ctx, workspace.ID, tfe.WorkspaceAssignSSHKeyOptions{
			SSHKeyID: tfe.String(sshKeyID.(string)),
		})
		if err != nil {
			return fmt.Errorf("Error assigning SSH key to workspace %s: %v", name, err)
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of workspace: %s", id)
	workspace, err := tfeClient.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	// Update the config.
	d.Set("name", workspace.Name)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("external_id", workspace.ID)
	d.Set("organization", workspace.Organization.Name)

	var sshKeyID string
	if workspace.SSHKey != nil {
		sshKeyID = workspace.SSHKey.ID
	}
	d.Set("ssh_key_id", sshKeyID)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":         workspace.VCSRepo.Identifier,
			"branch":             workspace.VCSRepo.Branch,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":     workspace.VCSRepo.OAuthTokenID,
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			if vcsRepo, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				// Only set the branch if one is configured.
				if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
					vcsConfig["branch"] = workspace.VCSRepo.Branch
				}
			}
		}

		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)

	return nil
}

func resourceTFEWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") || d.HasChange("vcs_repo") ||
		d.HasChange("file_triggers_enabled") || d.HasChange("trigger_prefixes") ||
		d.HasChange("operations") {
		// Create a new options struct.
		options := tfe.WorkspaceUpdateOptions{
			Name:                tfe.String(d.Get("name").(string)),
			AutoApply:           tfe.Bool(d.Get("auto_apply").(bool)),
			FileTriggersEnabled: tfe.Bool(d.Get("file_triggers_enabled").(bool)),
			Operations:          tfe.Bool(d.Get("operations").(bool)),
			QueueAllRuns:        tfe.Bool(d.Get("queue_all_runs").(bool)),
			WorkingDirectory:    tfe.String(d.Get("working_directory").(string)),
		}

		// Process all configured options.
		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = tfe.String(tfVersion.(string))
		}

		if tps, ok := d.GetOk("trigger_prefixes"); ok {
			for _, tp := range tps.([]interface{}) {
				options.TriggerPrefixes = append(options.TriggerPrefixes, tp.(string))
			}
		} else {
			// Reset trigger prefixes when none are present in the config.
			options.TriggerPrefixes = []string{}
		}

		if workingDir, ok := d.GetOk("working_directory"); ok {
			options.WorkingDirectory = tfe.String(workingDir.(string))
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})

			options.VCSRepo = &tfe.VCSRepoOptions{
				Identifier:        tfe.String(vcsRepo["identifier"].(string)),
				Branch:            tfe.String(vcsRepo["branch"].(string)),
				IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
				OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
			}
		}

		log.Printf("[DEBUG] Update workspace %s", id)
		_, err := tfeClient.Workspaces.UpdateByID(ctx, id, options)
		if err != nil {
			return fmt.Errorf(
				"Error updating workspace %s: %v", id, err)
		}
	}

	// Remove vcs_repo from the workspace
	// if the value of vcs_repo has been changed
	// by removing it from the config
	if d.HasChange("vcs_repo") {
		_, ok := d.GetOk("vcs_repo")
		if !ok {
			_, err := tfeClient.Workspaces.RemoveVCSConnectionByID(ctx, id)
			if err != nil {
				return fmt.Errorf("Error removing VCS repo from workspace %s: %v", id, err)
			}
		}
	}

	// TODO: Why does this use the old value of external_id?
	// external_id shouldn't change so can we change externalID to just id/d.Id()?
	if d.HasChange("ssh_key_id") {
		sshKeyID := d.Get("ssh_key_id").(string)
		externalID, _ := d.GetChange("external_id")

		if sshKeyID != "" {
			_, err := tfeClient.Workspaces.AssignSSHKey(
				ctx,
				externalID.(string),
				tfe.WorkspaceAssignSSHKeyOptions{
					SSHKeyID: tfe.String(sshKeyID),
				},
			)
			if err != nil {
				return fmt.Errorf("Error assigning SSH key to workspace %s: %v", id, err)
			}
		} else {
			_, err := tfeClient.Workspaces.UnassignSSHKey(ctx, externalID.(string))
			if err != nil {
				return fmt.Errorf("Error unassigning SSH key from workspace %s: %v", id, err)
			}
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete workspace %s", id)
	err := tfeClient.Workspaces.DeleteByID(ctx, id)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting workspace %s: %v", id, err)
	}

	return nil
}
