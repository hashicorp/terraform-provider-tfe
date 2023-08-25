// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetCreate,
		Read:   resourceTFEPolicySetRead,
		Update: resourceTFEPolicySetUpdate,
		Delete: resourceTFEPolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\A[\w\_\-]+\z`), "can only include letters, numbers, -, and _."),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"global": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_ids"},
			},

			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(tfe.Sentinel),
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.OPA),
						string(tfe.Sentinel),
					}, false),
			},

			"overridable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"policies_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"policy_ids"},
			},

			"slug": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"policy_ids", "vcs_repo"},
			},

			"policy_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vcs_repo", "policies_path"},
			},

			"vcs_repo": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"policy_ids"},
				MinItems:      1,
				MaxItems:      1,
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
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.github_app_installation_id"},
						},

						"github_app_installation_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.oauth_token_id"},
							AtLeastOneOf:  []string{"vcs_repo.0.oauth_token_id", "vcs_repo.0.github_app_installation_id"},
						},
					},
				},
			},

			"workspace_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"global"},
			},
		},
	}
}

func resourceTFEPolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.PolicySetCreateOptions{
		Name:   tfe.String(name),
		Global: tfe.Bool(d.Get("global").(bool)),
	}

	// Process all configured options.
	if vKind, ok := d.GetOk("kind"); ok {
		options.Kind = tfe.PolicyKind(vKind.(string))
	}

	if vOverridable, ok := d.GetOk("overridable"); ok {
		options.Overridable = tfe.Bool(vOverridable.(bool))
	}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = tfe.String(desc.(string))
	}

	if policiesPath, ok := d.GetOk("policies_path"); ok {
		options.PoliciesPath = tfe.String(policiesPath.(string))
	}

	for _, policyID := range d.Get("policy_ids").(*schema.Set).List() {
		options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &tfe.VCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
			IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
			OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
			GHAInstallationID: tfe.String(vcsRepo["github_app_installation_id"].(string)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = tfe.String(branch)
		}
	}

	for _, workspaceID := range d.Get("workspace_ids").(*schema.Set).List() {
		options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
	}

	log.Printf("[DEBUG] Create policy set %s for organization: %s", name, organization)
	policySet, err := config.Client.PolicySets.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating policy set %s for organization %s: %w", name, organization, err)
	}
	_, hasVCSRepo := d.GetOk("vcs_repo")
	_, hasSlug := d.GetOk("slug")
	if hasSlug && !hasVCSRepo {
		err := resourceTFEPolicySetUploadVersion(config.Client, d, policySet.ID)
		if err != nil {
			return err
		}
	}

	d.SetId(policySet.ID)

	return resourceTFEPolicySetRead(d, meta)
}

func resourceTFEPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read policy set: %s", d.Id())
	policySet, err := config.Client.PolicySets.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Policy set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading policy set %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", policySet.Name)
	d.Set("description", policySet.Description)
	d.Set("global", policySet.Global)
	d.Set("policies_path", policySet.PoliciesPath)

	if policySet.Organization != nil {
		d.Set("organization", policySet.Organization.Name)
	}

	// Note: Old API endpoints return an empty string, so use the default in the schema
	if policySet.Kind != "" {
		d.Set("kind", policySet.Kind)
	}

	if policySet.Overridable != nil {
		d.Set("overridable", policySet.Overridable)
	}

	// Set VCS policy set options.
	var vcsRepo []interface{}
	if policySet.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":                 policySet.VCSRepo.Identifier,
			"ingress_submodules":         policySet.VCSRepo.IngressSubmodules,
			"oauth_token_id":             policySet.VCSRepo.OAuthTokenID,
			"github_app_installation_id": policySet.VCSRepo.GHAInstallationID,
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			if vcsRepo, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				// Only set the branch if one is configured.
				if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
					vcsConfig["branch"] = policySet.VCSRepo.Branch
				}
			}
		}

		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)

	// Update the policies.
	var policyIDs []interface{}
	for _, policy := range policySet.Policies {
		policyIDs = append(policyIDs, policy.ID)
	}
	d.Set("policy_ids", policyIDs)

	// Update the workspaces.
	var workspaceIDs []interface{}
	if !policySet.Global {
		for _, workspace := range policySet.Workspaces {
			workspaceIDs = append(workspaceIDs, workspace.ID)
		}
	}
	d.Set("workspace_ids", workspaceIDs)

	return nil
}

func resourceTFEPolicySetUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name := d.Get("name").(string)
	global := d.Get("global").(bool)

	// If a user is setting the policy set to "global", make sure the workspaces
	// that _had_ been set are explicitly removed. This helps keep the policy
	// set's state in check
	if global && d.HasChange("global") {
		// The new set of workspaces will be an empty set, so we don't need it
		oldWorkspaceIDs, _ := d.GetChange("workspace_ids")

		if oldWorkspaceIDs.(*schema.Set).Len() > 0 {
			options := tfe.PolicySetRemoveWorkspacesOptions{}

			for _, workspaceID := range oldWorkspaceIDs.(*schema.Set).List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Removing previous workspaces from now-global policy set: %s", d.Id())
			err := config.Client.PolicySets.RemoveWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error detaching policy set %s from workspaces: %w", d.Id(), err)
			}
		}
	}

	// Don't bother updating the policy set's attributes if they haven't changed
	if d.HasChange("name") || d.HasChange("description") || d.HasChange("global") || d.HasChange("vcs_repo") || d.HasChange("overridable") {
		// Create a new options struct.
		options := tfe.PolicySetUpdateOptions{
			Name:   tfe.String(name),
			Global: tfe.Bool(global),
		}

		if desc, ok := d.GetOk("description"); ok {
			options.Description = tfe.String(desc.(string))
		}

		if d.HasChange("overridable") {
			o := d.Get("overridable").(bool)
			options.Overridable = tfe.Bool(o)
		}

		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})

			options.VCSRepo = &tfe.VCSRepoOptions{
				Identifier:        tfe.String(vcsRepo["identifier"].(string)),
				Branch:            tfe.String(vcsRepo["branch"].(string)),
				IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
				OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
				GHAInstallationID: tfe.String(vcsRepo["github_app_installation_id"].(string)),
			}
		}

		log.Printf("[DEBUG] Update configuration for policy set: %s", d.Id())
		_, err := config.Client.PolicySets.Update(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for policy set %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("policy_ids") {
		oldSet, newSet := d.GetChange("policy_ids")
		oldPolicyIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newPolicyIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		// First add the new policies.
		if newPolicyIDs.Len() > 0 {
			options := tfe.PolicySetAddPoliciesOptions{}

			for _, policyID := range newPolicyIDs.List() {
				options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
			}

			log.Printf("[DEBUG] Add policies to policy set: %s", d.Id())
			err := config.Client.PolicySets.AddPolicies(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error adding policies to policy set %s: %w", d.Id(), err)
			}
		}

		// Then remove all the old policies.
		if oldPolicyIDs.Len() > 0 {
			options := tfe.PolicySetRemovePoliciesOptions{}

			for _, policyID := range oldPolicyIDs.List() {
				options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
			}

			log.Printf("[DEBUG] Remove policies from policy set: %s", d.Id())
			err := config.Client.PolicySets.RemovePolicies(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error removing policies from policy set %s: %w", d.Id(), err)
			}
		}
	}

	_, hasVCSRepo := d.GetOk("vcs_repo")
	if d.HasChange("slug") && !hasVCSRepo {
		err := resourceTFEPolicySetUploadVersion(config.Client, d, d.Id())
		if err != nil {
			return err
		}
	}

	if !global && d.HasChange("workspace_ids") {
		oldWorkspaceIDValues, newWorkspaceIDValues := d.GetChange("workspace_ids")
		newWorkspaceIDsSet := newWorkspaceIDValues.(*schema.Set)
		oldWorkspaceIDsSet := oldWorkspaceIDValues.(*schema.Set)

		newWorkspaceIDs := newWorkspaceIDsSet.Difference(oldWorkspaceIDsSet)
		oldWorkspaceIDs := oldWorkspaceIDsSet.Difference(newWorkspaceIDsSet)

		// First add the new workspaces.
		if newWorkspaceIDs.Len() > 0 {
			options := tfe.PolicySetAddWorkspacesOptions{}

			for _, workspaceID := range newWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Attach policy set to workspaces: %s", d.Id())
			err := config.Client.PolicySets.AddWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error attaching policy set %s to workspaces: %w", d.Id(), err)
			}
		}

		// Then remove all the old workspaces.
		if oldWorkspaceIDs.Len() > 0 {
			options := tfe.PolicySetRemoveWorkspacesOptions{}

			for _, workspaceID := range oldWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Detach policy set from workspaces: %s", d.Id())
			err := config.Client.PolicySets.RemoveWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error detaching policy set %s from workspaces: %w", d.Id(), err)
			}
		}
	}

	return resourceTFEPolicySetRead(d, meta)
}

func resourceTFEPolicySetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete policy set: %s", d.Id())
	err := config.Client.PolicySets.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting policy set %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEPolicySetUploadVersion(client *tfe.Client, d *schema.ResourceData, policySetID string) error {
	log.Printf("[DEBUG] Create policy set version for policy set %s.", policySetID)
	psv, err := client.PolicySetVersions.Create(ctx, policySetID)
	if err != nil {
		return fmt.Errorf("Error creating policy set version for policy set %s: %w", policySetID, err)
	}

	slug := d.Get("slug").(map[string]interface{})
	path := slug["source_path"].(string)

	log.Printf("[DEBUG] Upload policy set version %s.", psv.ID)
	err = client.PolicySetVersions.Upload(ctx, *psv, path)
	if err != nil {
		return fmt.Errorf("Error uploading policies for policy set version %s: %w", psv.ID, err)
	}

	return nil
}
