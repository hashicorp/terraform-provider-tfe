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
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"auto_apply": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"terraform_version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"working_directory": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"vcs_repo": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"branch": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "master",
						},

						"ingress_submodules": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"oauth_token_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
		Name: tfe.String(name),
	}

	// Process all configured options.
	if autoApply, ok := d.GetOk("auto_apply"); ok {
		options.AutoApply = tfe.Bool(autoApply.(bool))
	}

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

	log.Printf("[DEBUG] Create workspace %s for organization: %s", name, organization)
	workspace, err := tfeClient.Workspaces.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for organization %s: %v", name, organization, err)
	}

	d.SetId(packWorkspaceID(workspace))

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name, organization := unpackWorkspaceID(d.Id())

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
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsRepo = append(vcsRepo, map[string]interface{}{
			"identifier":         workspace.VCSRepo.Identifier,
			"branch":             workspace.VCSRepo.Branch,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":     workspace.VCSRepo.OAuthTokenID,
		})
	}
	d.Set("vcs_repo", vcsRepo)

	return nil
}

func resourceTFEWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name, organization := unpackWorkspaceID(d.Id())

	// Create a new options struct.
	options := tfe.WorkspaceUpdateOptions{
		Name: tfe.String(d.Get("name").(string)),
	}

	// Process all configured options.
	if autoApply, ok := d.GetOk("auto_apply"); ok {
		options.AutoApply = tfe.Bool(autoApply.(bool))
	}

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

	d.SetId(packWorkspaceID(workspace))

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name, organization := unpackWorkspaceID(d.Id())

	log.Printf("[DEBUG] Delete workspace %s from organization: %s", name, organization)
	err := tfeClient.Workspaces.Delete(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting workspace %s from organization %s: %v", name, organization, err)
	}

	return nil
}

func packWorkspaceID(w *tfe.Workspace) string {
	return w.Name + "|" + w.Organization.Name
}

func unpackWorkspaceID(id string) (name, organization string) {
	s := strings.SplitN(id, "|", 2)
	return s[0], s[1]
}
