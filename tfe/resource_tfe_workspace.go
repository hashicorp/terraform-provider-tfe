package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceCreate,
		Read:   resourceTFEWorkspaceRead,
		Update: resourceTFEWorkspaceUpdate,
		Delete: resourceTFEWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

			"ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"vcs_repo": {
				Type:     schema.TypeSet,
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
		Name:         tfe.String(name),
		AutoApply:    tfe.Bool(d.Get("auto_apply").(bool)),
		QueueAllRuns: tfe.Bool(d.Get("queue_all_runs").(bool)),
	}

	// Process all configured options.
	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = tfe.String(tfVersion.(string))
	}

	if workingDir, ok := d.GetOk("working_directory"); ok {
		options.WorkingDirectory = tfe.String(workingDir.(string))
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.(*schema.Set).List()[0].(map[string]interface{})

		options.VCSRepo = &tfe.VCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
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

	id, err := packWorkspaceID(workspace)
	if err != nil {
		return fmt.Errorf("Error creating ID for workspace %s: %v", name, err)
	}

	d.SetId(id)

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

	// Get the organization and workspace name.
	organization, name, err := unpackWorkspaceID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	log.Printf("[DEBUG] Read configuration of workspace: %s", name)
	workspace, err := tfeClient.Workspaces.Read(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s does no longer exist", name)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %v", name, err)
	}

	// Update the config.
	d.Set("name", workspace.Name)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("external_id", workspace.ID)

	if workspace.Organization != nil {
		d.Set("organization", workspace.Organization.Name)
	}

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

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.(*schema.Set).List()[0].(map[string]interface{})

			// Only set the branch if one is configured.
			if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
				vcsConfig["branch"] = workspace.VCSRepo.Branch
			}
		}

		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)

	// We do this here as a means to convert the internal ID,
	// in case anyone still uses the old format.
	id, err := packWorkspaceID(workspace)
	if err != nil {
		return err
	}
	d.SetId(id)

	return nil
}

func resourceTFEWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization and workspace name.
	organization, name, err := unpackWorkspaceID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") || d.HasChange("vcs_repo") {
		// Create a new options struct.
		options := tfe.WorkspaceUpdateOptions{
			Name:         tfe.String(d.Get("name").(string)),
			AutoApply:    tfe.Bool(d.Get("auto_apply").(bool)),
			QueueAllRuns: tfe.Bool(d.Get("queue_all_runs").(bool)),
		}

		// Process all configured options.
		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = tfe.String(tfVersion.(string))
		}

		if workingDir, ok := d.GetOk("working_directory"); ok {
			options.WorkingDirectory = tfe.String(workingDir.(string))
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.(*schema.Set).List()[0].(map[string]interface{})

			options.VCSRepo = &tfe.VCSRepoOptions{
				Identifier:        tfe.String(vcsRepo["identifier"].(string)),
				Branch:            tfe.String(vcsRepo["branch"].(string)),
				IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
				OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
			}
		}

		log.Printf("[DEBUG] Update workspace %s for organization: %s", name, organization)
		workspace, err := tfeClient.Workspaces.Update(ctx, organization, name, options)
		if err != nil {
			return fmt.Errorf(
				"Error updating workspace %s for organization %s: %v", name, organization, err)
		}

		id, err := packWorkspaceID(workspace)
		if err != nil {
			return fmt.Errorf("Error creating ID for workspace %s: %v", name, err)
		}

		d.SetId(id)
	}

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
				return fmt.Errorf("Error assigning SSH key to workspace %s: %v", name, err)
			}
		} else {
			_, err := tfeClient.Workspaces.UnassignSSHKey(ctx, externalID.(string))
			if err != nil {
				return fmt.Errorf("Error unassigning SSH key from workspace %s: %v", name, err)
			}
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization and workspace name.
	organization, name, err := unpackWorkspaceID(d.Id())
	if err != nil {
		return fmt.Errorf("Error unpacking workspace ID: %v", err)
	}

	log.Printf("[DEBUG] Delete workspace %s from organization: %s", name, organization)
	err = tfeClient.Workspaces.Delete(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting workspace %s from organization %s: %v", name, organization, err)
	}

	return nil
}

func packWorkspaceID(w *tfe.Workspace) (id string, err error) {
	if w.Organization == nil {
		return "", fmt.Errorf("no organization in workspace response")
	}
	return w.Organization.Name + "/" + w.Name, nil
}

func unpackWorkspaceID(id string) (organization, name string, err error) {
	// Support the old ID format for backwards compatibitily.
	if s := strings.SplitN(id, "|", 2); len(s) == 2 {
		return s[1], s[0], nil
	}

	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid workspace ID format: %s (expected <ORGANIZATION>/<WORKSPACE>)", id)
	}

	return s[0], s[1], nil
}
