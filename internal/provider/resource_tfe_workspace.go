// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var workspaceIDRegexp = regexp.MustCompile("^ws-[a-zA-Z0-9]{16}$")

func resourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceCreate,
		Read:   resourceTFEWorkspaceRead,
		Update: resourceTFEWorkspaceUpdate,
		Delete: resourceTFEWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceImporter,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeWorkspaceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeWorkspaceStateUpgradeV0,
				Version: 0,
			},
		},

		CustomizeDiff: func(c context.Context, d *schema.ResourceDiff, meta interface{}) error {
			// NOTE: execution mode must be set to default first before calling the validation functions
			if err := setExecutionModeDefault(c, d); err != nil {
				return err
			}

			if err := validateAgentExecution(c, d); err != nil {
				return err
			}

			if err := validateRemoteState(c, d); err != nil {
				return err
			}

			if err := validateTagNames(c, d); err != nil {
				return err
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"agent_pool_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				ConflictsWith: []string{"operations"},
			},

			"allow_destroy_plan": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"auto_apply": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"execution_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"operations"},
				ValidateFunc: validation.StringInSlice(
					[]string{
						"agent",
						"local",
						"remote",
					},
					false,
				),
			},

			"file_triggers_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"global_remote_state": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"remote_state_consumer_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"assessments_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"operations": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Use execution_mode instead.",
				ConflictsWith: []string{"execution_mode", "agent_pool_id"},
			},

			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"source_url"},
			},

			"source_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				RequiredWith: []string{"source_name"},
			},

			"speculative_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"structured_run_output_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"tag_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"trigger_prefixes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"trigger_patterns"},
			},

			"trigger_patterns": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"trigger_prefixes"},
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
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.github_app_installation_id"},
						},

						"tags_regex": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"trigger_patterns", "trigger_prefixes"},
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
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"resource_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.WorkspaceCreateOptions{
		Name:                       tfe.String(name),
		AllowDestroyPlan:           tfe.Bool(d.Get("allow_destroy_plan").(bool)),
		AutoApply:                  tfe.Bool(d.Get("auto_apply").(bool)),
		Description:                tfe.String(d.Get("description").(string)),
		AssessmentsEnabled:         tfe.Bool(d.Get("assessments_enabled").(bool)),
		FileTriggersEnabled:        tfe.Bool(d.Get("file_triggers_enabled").(bool)),
		QueueAllRuns:               tfe.Bool(d.Get("queue_all_runs").(bool)),
		SpeculativeEnabled:         tfe.Bool(d.Get("speculative_enabled").(bool)),
		StructuredRunOutputEnabled: tfe.Bool(d.Get("structured_run_output_enabled").(bool)),
		WorkingDirectory:           tfe.String(d.Get("working_directory").(string)),
	}

	// Send global_remote_state if it's set; otherwise, let it be computed.
	globalRemoteState, ok := d.GetOkExists("global_remote_state")
	if ok {
		options.GlobalRemoteState = tfe.Bool(globalRemoteState.(bool))
	}

	if v, ok := d.GetOk("agent_pool_id"); ok && v.(string) != "" {
		options.AgentPoolID = tfe.String(v.(string))
	}

	if v, ok := d.GetOk("execution_mode"); ok {
		options.ExecutionMode = tfe.String(v.(string))
	}

	if v, ok := d.GetOkExists("operations"); ok {
		options.Operations = tfe.Bool(v.(bool))
	}

	if v, ok := d.GetOk("source_url"); ok {
		options.SourceURL = tfe.String(v.(string))
	}
	if v, ok := d.GetOk("source_name"); ok {
		options.SourceName = tfe.String(v.(string))
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
		options.TriggerPrefixes = nil
	}

	if tps, ok := d.GetOk("trigger_patterns"); ok {
		for _, tp := range tps.([]interface{}) {
			options.TriggerPatterns = append(options.TriggerPatterns, tp.(string))
		}
	} else {
		options.TriggerPatterns = nil
	}

	if d.HasChange("project_id") {
		if v, ok := d.GetOk("project_id"); ok && v.(string) != "" {
			options.Project = &tfe.Project{ID: *tfe.String(v.(string))}
		}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &tfe.VCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
			IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
			TagsRegex:         tfe.String(vcsRepo["tags_regex"].(string)),
		}

		// Only set the oauth_token_id if it is configured.
		if oauthTokenID, ok := vcsRepo["oauth_token_id"].(string); ok && oauthTokenID != "" {
			options.VCSRepo.OAuthTokenID = tfe.String(oauthTokenID)
		}

		// Only set the github_app_installation_id if it is configured.
		if ghaInstallationID, ok := vcsRepo["github_app_installation_id"].(string); ok && ghaInstallationID != "" {
			options.VCSRepo.GHAInstallationID = tfe.String(ghaInstallationID)
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = tfe.String(branch)
		}
	}

	for _, tagName := range d.Get("tag_names").(*schema.Set).List() {
		name := tagName.(string)
		options.Tags = append(options.Tags, &tfe.Tag{Name: name})
	}

	log.Printf("[DEBUG] Create workspace %s for organization: %s", name, organization)
	workspace, err := config.Client.Workspaces.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for organization %s: %w", name, organization, err)
	}

	d.SetId(workspace.ID)

	if sshKeyID, ok := d.GetOk("ssh_key_id"); ok {
		_, err = config.Client.Workspaces.AssignSSHKey(ctx, workspace.ID, tfe.WorkspaceAssignSSHKeyOptions{
			SSHKeyID: tfe.String(sshKeyID.(string)),
		})
		if err != nil {
			return fmt.Errorf("Error assigning SSH key to workspace %s: %w", name, err)
		}
	}

	remoteStateConsumerIDs, ok := d.GetOk("remote_state_consumer_ids")
	if ok && globalRemoteState.(bool) == false {
		options := tfe.WorkspaceAddRemoteStateConsumersOptions{}
		for _, remoteStateConsumerID := range remoteStateConsumerIDs.(*schema.Set).List() {
			options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: remoteStateConsumerID.(string)})
		}
		err = config.Client.Workspaces.AddRemoteStateConsumers(ctx, workspace.ID, options)
		if err != nil {
			return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", name, err)
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of workspace: %s", id)
	workspace, err := config.Client.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Workspace %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %w", id, err)
	}

	// Update the config.
	d.Set("name", workspace.Name)
	d.Set("allow_destroy_plan", workspace.AllowDestroyPlan)

	// TFE (onprem) does not currently have this feature and this value won't be returned in those cases.
	// workspace.AssessmentsEnabled will default to false
	d.Set("assessments_enabled", workspace.AssessmentsEnabled)

	d.Set("auto_apply", workspace.AutoApply)
	d.Set("description", workspace.Description)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("source_name", workspace.SourceName)
	d.Set("source_url", workspace.SourceURL)
	d.Set("speculative_enabled", workspace.SpeculativeEnabled)
	d.Set("structured_run_output_enabled", workspace.StructuredRunOutputEnabled)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("trigger_patterns", workspace.TriggerPatterns)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("organization", workspace.Organization.Name)
	d.Set("resource_count", workspace.ResourceCount)

	if workspace.Links["self-html"] != nil {
		baseAPI := config.Client.BaseURL()
		htmlURL := url.URL{
			Scheme: baseAPI.Scheme,
			Host:   baseAPI.Host,
			Path:   workspace.Links["self-html"].(string),
		}

		d.Set("html_url", htmlURL.String())
	}

	// Project will be nil for versions of TFE that predate projects
	if workspace.Project != nil {
		d.Set("project_id", workspace.Project.ID)
	}

	var sshKeyID string
	if workspace.SSHKey != nil {
		sshKeyID = workspace.SSHKey.ID
	}
	d.Set("ssh_key_id", sshKeyID)

	var agentPoolID string
	if workspace.AgentPool != nil {
		agentPoolID = workspace.AgentPool.ID
	}
	d.Set("agent_pool_id", agentPoolID)

	var tagNames []interface{}
	for _, tagName := range workspace.TagNames {
		tagNames = append(tagNames, tagName)
	}
	d.Set("tag_names", tagNames)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":                 workspace.VCSRepo.Identifier,
			"branch":                     workspace.VCSRepo.Branch,
			"ingress_submodules":         workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":             workspace.VCSRepo.OAuthTokenID,
			"github_app_installation_id": workspace.VCSRepo.GHAInstallationID,
			"tags_regex":                 workspace.VCSRepo.TagsRegex,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)

	if workspace.GlobalRemoteState {
		d.Set("global_remote_state", true)
	} else {
		globalRemoteState, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(id, config.Client)
		if err != nil {
			return fmt.Errorf(
				"Error reading remote state consumers for workspace %s: %w", id, err)
		}

		d.Set("global_remote_state", globalRemoteState)
		d.Set("remote_state_consumer_ids", remoteStateConsumerIDs)
	}

	return nil
}

func resourceTFEWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") ||
		d.HasChange("vcs_repo") || d.HasChange("file_triggers_enabled") ||
		d.HasChange("trigger_prefixes") || d.HasChange("trigger_patterns") ||
		d.HasChange("allow_destroy_plan") || d.HasChange("speculative_enabled") ||
		d.HasChange("operations") || d.HasChange("execution_mode") ||
		d.HasChange("description") || d.HasChange("agent_pool_id") ||
		d.HasChange("global_remote_state") || d.HasChange("structured_run_output_enabled") ||
		d.HasChange("assessments_enabled") || d.HasChange("project_id") {
		// Create a new options struct.
		options := tfe.WorkspaceUpdateOptions{
			Name:                       tfe.String(d.Get("name").(string)),
			AllowDestroyPlan:           tfe.Bool(d.Get("allow_destroy_plan").(bool)),
			AutoApply:                  tfe.Bool(d.Get("auto_apply").(bool)),
			Description:                tfe.String(d.Get("description").(string)),
			FileTriggersEnabled:        tfe.Bool(d.Get("file_triggers_enabled").(bool)),
			GlobalRemoteState:          tfe.Bool(d.Get("global_remote_state").(bool)),
			QueueAllRuns:               tfe.Bool(d.Get("queue_all_runs").(bool)),
			SpeculativeEnabled:         tfe.Bool(d.Get("speculative_enabled").(bool)),
			StructuredRunOutputEnabled: tfe.Bool(d.Get("structured_run_output_enabled").(bool)),
			WorkingDirectory:           tfe.String(d.Get("working_directory").(string)),
		}

		if d.HasChange("project_id") {
			if v, ok := d.GetOk("project_id"); ok && v.(string) != "" {
				options.Project = &tfe.Project{ID: *tfe.String(v.(string))}
			}
		}

		if d.HasChange("assessments_enabled") {
			if v, ok := d.GetOkExists("assessments_enabled"); ok {
				options.AssessmentsEnabled = tfe.Bool(v.(bool))
			}
		}

		if d.HasChange("agent_pool_id") {
			if v, ok := d.GetOk("agent_pool_id"); ok && v.(string) != "" {
				options.AgentPoolID = tfe.String(v.(string))
			}
		}

		if d.HasChange("execution_mode") {
			if v, ok := d.GetOk("execution_mode"); ok {
				options.ExecutionMode = tfe.String(v.(string))
			}
		}

		if d.HasChange("operations") {
			if v, ok := d.GetOkExists("operations"); ok {
				options.Operations = tfe.Bool(v.(bool))
			}
		}

		// Process all configured options.
		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = tfe.String(tfVersion.(string))
		}

		if tps, ok := d.GetOk("trigger_prefixes"); ok {
			for _, tp := range tps.([]interface{}) {
				if val, ok := tp.(string); ok {
					options.TriggerPrefixes = append(options.TriggerPrefixes, val)
				}
			}
		} else {
			options.TriggerPrefixes = []string{}
		}

		if tps, ok := d.GetOk("trigger_patterns"); ok {
			for _, tp := range tps.([]interface{}) {
				if val, ok := tp.(string); ok {
					options.TriggerPatterns = append(options.TriggerPatterns, val)
				}
			}
		} else {
			options.TriggerPatterns = []string{}
		}

		if d.GetRawConfig().GetAttr("trigger_patterns").IsNull() {
			options.TriggerPatterns = nil
		} else if d.GetRawConfig().GetAttr("trigger_prefixes").IsNull() {
			options.TriggerPrefixes = nil
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
				GHAInstallationID: tfe.String(vcsRepo["github_app_installation_id"].(string)),
				TagsRegex:         tfe.String(vcsRepo["tags_regex"].(string)),
			}
		}

		// Remove vcs_repo from the workspace
		// if the value of vcs_repo has been changed
		// by removing it from the config
		if d.HasChange("vcs_repo") {
			_, ok := d.GetOk("vcs_repo")
			if !ok {
				_, err := config.Client.Workspaces.RemoveVCSConnectionByID(ctx, id)
				if err != nil {
					d.Partial(true)
					return fmt.Errorf("Error removing VCS repo from workspace %s: %w", id, err)
				}
			}
		}

		log.Printf("[DEBUG] Update workspace %s", id)
		_, err := config.Client.Workspaces.UpdateByID(ctx, id, options)
		if err != nil {
			d.Partial(true)
			return fmt.Errorf(
				"Error updating workspace %s: %w", id, err)
		}
	}

	if d.HasChange("ssh_key_id") {
		sshKeyID := d.Get("ssh_key_id").(string)

		if sshKeyID != "" {
			_, err := config.Client.Workspaces.AssignSSHKey(
				ctx,
				id,
				tfe.WorkspaceAssignSSHKeyOptions{
					SSHKeyID: tfe.String(sshKeyID),
				},
			)
			if err != nil {
				return fmt.Errorf("Error assigning SSH key to workspace %s: %w", id, err)
			}
		} else {
			_, err := config.Client.Workspaces.UnassignSSHKey(ctx, id)
			if err != nil {
				return fmt.Errorf("Error unassigning SSH key from workspace %s: %w", id, err)
			}
		}
	}

	if d.HasChange("tag_names") {
		oldTagNameValues, newTagNameValues := d.GetChange("tag_names")
		newTagNamesSet := newTagNameValues.(*schema.Set)
		oldTagNamesSet := oldTagNameValues.(*schema.Set)

		newTagNames := newTagNamesSet.Difference(oldTagNamesSet)
		oldTagNames := oldTagNamesSet.Difference(newTagNamesSet)

		// First add the new tags
		if newTagNames.Len() > 0 {
			var addTags []*tfe.Tag

			for _, tagName := range newTagNames.List() {
				name := tagName.(string)
				addTags = append(addTags, &tfe.Tag{Name: name})
			}

			log.Printf("[DEBUG] Adding tags to workspace: %s", d.Id())
			err := config.Client.Workspaces.AddTags(ctx, d.Id(), tfe.WorkspaceAddTagsOptions{Tags: addTags})
			if err != nil {
				return fmt.Errorf("Error adding tags to workspace %s: %w", d.Id(), err)
			}
		}

		// Then remove all the old tags
		if oldTagNames.Len() > 0 {
			var removeTags []*tfe.Tag

			for _, tagName := range oldTagNames.List() {
				removeTags = append(removeTags, &tfe.Tag{Name: tagName.(string)})
			}

			log.Printf("[DEBUG] Removing tags from workspace: %s", d.Id())
			err := config.Client.Workspaces.RemoveTags(ctx, d.Id(), tfe.WorkspaceRemoveTagsOptions{Tags: removeTags})
			if err != nil {
				return fmt.Errorf("Error removing tags from workspace %s: %w", d.Id(), err)
			}
		}
	}

	globalRemoteState := d.Get("global_remote_state").(bool)
	if !globalRemoteState && d.HasChange("remote_state_consumer_ids") {
		oldWorkspaceIDValues, newWorkspaceIDValues := d.GetChange("remote_state_consumer_ids")
		newWorkspaceIDsSet := newWorkspaceIDValues.(*schema.Set)
		oldWorkspaceIDsSet := oldWorkspaceIDValues.(*schema.Set)

		newWorkspaceIDs := newWorkspaceIDsSet.Difference(oldWorkspaceIDsSet)
		oldWorkspaceIDs := oldWorkspaceIDsSet.Difference(newWorkspaceIDsSet)

		// First add the new consumerss
		if newWorkspaceIDs.Len() > 0 {
			options := tfe.WorkspaceAddRemoteStateConsumersOptions{}

			for _, workspaceID := range newWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Adding remote state consumers to workspace: %s", d.Id())
			err := config.Client.Workspaces.AddRemoteStateConsumers(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", d.Id(), err)
			}
		}

		// Then remove all the old consumers.
		if oldWorkspaceIDs.Len() > 0 {
			options := tfe.WorkspaceRemoveRemoteStateConsumersOptions{}

			for _, workspaceID := range oldWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Removing remote state consumers from workspace: %s", d.Id())
			err := config.Client.Workspaces.RemoveRemoteStateConsumers(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error removing remote state consumers from workspace %s: %w", d.Id(), err)
			}
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func safeWorkspaceDelete(ctx context.Context, config ConfiguredClient, id string) error {
	return retry.RetryContext(ctx, time.Duration(5)*time.Minute, func() *retry.RetryError {
		err := config.Client.Workspaces.SafeDeleteByID(ctx, id)
		if errors.Is(err, tfe.ErrWorkspaceStillProcessing) {
			return retry.RetryableError(err)
		} else if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

func resourceTFEWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	id := d.Id()

	log.Printf("[DEBUG] Delete workspace %s", id)

	ws, err := config.Client.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error reading workspace %s: %w", id, err)
	}

	forceDelete := d.Get("force_delete").(bool)

	// presence of Permissions.CanForceDelete will determine if current version of TFE supports safe deletes
	if ws.Permissions.CanForceDelete == nil {
		if forceDelete {
			err = config.Client.Workspaces.DeleteByID(ctx, id)
		} else {
			return fmt.Errorf(
				"Error deleting workspace %s: This version of Terraform Enterprise does not support workspace safe-delete. Workspaces must be force deleted by setting force_delete=true", id)
		}
	} else if *ws.Permissions.CanForceDelete {
		if forceDelete {
			err = config.Client.Workspaces.DeleteByID(ctx, id)
		} else {
			err = errWorkspaceResourceCountCheck(id, ws.ResourceCount)
			if err != nil {
				return err
			}

			err = safeWorkspaceDelete(ctx, config, id)
			return errWorkspaceSafeDeleteWithPermission(id, err)
		}
	} else {
		if forceDelete {
			return fmt.Errorf(
				"Error deleting workspace %s: missing required permissions to set force delete workspaces in the organization", id)
		}
		err = errWorkspaceResourceCountCheck(id, ws.ResourceCount)
		if err != nil {
			return err
		}
		err = safeWorkspaceDelete(ctx, config, id)
	}

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting workspace %s: %w", id, err)
	}
	return nil
}

// since execution_mode is marked as Optional: true, and Computed: true,
// unsetting the execution_mode in the config after it's been set to a valid
// value is not detected by ResourceDiff so read value from RawConfig instead
func setExecutionModeDefault(_ context.Context, d *schema.ResourceDiff) error {
	configMap := d.GetRawConfig().AsValueMap()
	operations, operationsReadOk := configMap["operations"]
	executionMode, executionModeReadOk := configMap["execution_mode"]
	executionModeState := d.Get("execution_mode")
	if operationsReadOk && executionModeReadOk {
		if operations.IsNull() && executionMode.IsNull() && executionModeState != "remote" {
			err := d.SetNew("execution_mode", "remote")
			if err != nil {
				return fmt.Errorf("failed to set execution_mode: %w", err)
			}
		}
	}

	return nil
}

// An agent pool can only be specified when execution_mode is set to "agent". You currently cannot specify a
// schema validation based on a different argument's value, so we do so here at plan time instead.
func validateAgentExecution(_ context.Context, d *schema.ResourceDiff) error {
	if executionMode, ok := d.GetOk("execution_mode"); ok {
		executionModeIsAgent := executionMode.(string) == "agent"
		if !executionModeIsAgent && d.Get("agent_pool_id") != "" {
			return fmt.Errorf("execution_mode must be set to 'agent' to assign agent_pool_id")
		} else if executionModeIsAgent && d.NewValueKnown("agent_pool_id") && d.Get("agent_pool_id") == "" {
			return fmt.Errorf("agent_pool_id must be provided when execution_mode is 'agent'")
		}
	}

	if d.HasChange("execution_mode") {
		d.SetNewComputed("operations")
	} else if d.HasChange("operations") {
		d.SetNewComputed("execution_mode")
	}

	return nil
}

func validTagName(tag string) bool {
	// Tags are re-validated here because the API will accept uppercase letters and automatically
	// downcase them, causing resource drift. It's better to catch this issue during the plan phase
	//
	//     \A            match beginning of string
	//     [a-z0-9]      match a letter or number for the first char; case insensitive
	//     (?:           start non-capture group; used to group sub-expressions; will not capture/store, interally
	//       [a-z0-9_:-]*     match 0 or more letter, number, colon, or hyphen
	//       [a-z0-9]    match a letter or number as the final character when this group is present
	//     )?            end non-capture group; ? is quantifier; matches 0 or 1 instances of the non-capture group in preceding set
	//     \z            match end of string; requires last char to match preceding subset; in this case, an alphanumeric char
	tagPattern := regexp.MustCompile(`\A[a-z0-9](?:[a-z0-9_:-]*[a-z0-9])?\z`)
	return tagPattern.MatchString(tag)
}

func validateTagNames(_ context.Context, d *schema.ResourceDiff) error {
	names, ok := d.GetOk("tag_names")
	if !ok {
		return nil
	}

	for _, t := range names.(*schema.Set).List() {
		tagName := t.(string)
		if !validTagName(tagName) {
			return fmt.Errorf("%q is not a valid tag name. Tag must be one or more characters; can include lowercase letters, numbers, colons, hyphens, and underscores; and must begin and end with a letter or number", tagName)
		}
	}
	return nil
}

func validateRemoteState(_ context.Context, d *schema.ResourceDiff) error {
	// If remote state consumers aren't set, the global setting can be either value and it
	// doesn't matter.
	_, ok := d.GetOk("remote_state_consumer_ids")
	if !ok {
		return nil
	}

	if globalRemoteState, ok := d.GetOk("global_remote_state"); ok {
		if globalRemoteState.(bool) {
			return fmt.Errorf("global_remote_state must be 'false' when setting remote_state_consumer_ids")
		}
	}

	return nil
}

func resourceTFEWorkspaceImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], config.Client)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace with name %s from organization %s %w", s[1], s[0], err)
		}

		d.SetId(workspaceID)
	}

	return []*schema.ResourceData{d}, nil
}

func errWorkspaceSafeDeleteWithPermission(workspaceID string, err error) error {
	if err != nil {
		if strings.HasPrefix(err.Error(), "conflict") {
			return fmt.Errorf("error deleting workspace %s: %w\nTo delete this workspace without destroying the managed resources, add force_delete = true to the resource config", workspaceID, err)
		}
		return err
	}
	return nil
}

func errWorkspaceResourceCountCheck(workspaceID string, resourceCount int) error {
	if resourceCount > 0 {
		return fmt.Errorf(
			"error deleting workspace %s: This workspace has %v resources under management and must be force deleted by setting force_delete = true", workspaceID, resourceCount)
	}
	return nil
}
