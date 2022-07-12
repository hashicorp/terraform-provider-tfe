package tfe

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var workspaceIdRegexp = regexp.MustCompile("^ws-[a-zA-Z0-9]{16}$")

func resourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceCreate,
		Read:   resourceTFEWorkspaceRead,
		Update: resourceTFEWorkspaceUpdate,
		Delete: resourceTFEWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFEWorkspaceImporter,
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
			err := validateAgentExecution(c, d)
			if err != nil {
				return err
			}

			err = validateRemoteState(c, d)
			if err != nil {
				return err
			}

			validateVcsTriggers(d)

			return nil
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

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"agent_pool_id": {
				Type:          schema.TypeString,
				Optional:      true,
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

			"operations": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"execution_mode", "agent_pool_id"},
			},

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"trigger_patterns"},
			},

			"trigger_patterns": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
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
							Type:     schema.TypeString,
							Required: true,
						},

						"tags_regex": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"trigger_patterns", "trigger_prefixes"},
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
		Name:                       tfe.String(name),
		AllowDestroyPlan:           tfe.Bool(d.Get("allow_destroy_plan").(bool)),
		AutoApply:                  tfe.Bool(d.Get("auto_apply").(bool)),
		Description:                tfe.String(d.Get("description").(string)),
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

	if v, ok := d.GetOk("operations"); ok {
		options.Operations = tfe.Bool(v.(bool))
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
		options.TriggerPrefixes = []string{}
	}

	if tps, ok := d.GetOk("trigger_patterns"); ok {
		for _, tp := range tps.([]interface{}) {
			options.TriggerPatterns = append(options.TriggerPatterns, tp.(string))
		}
	} else {
		options.TriggerPatterns = []string{}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &tfe.VCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
			IngressSubmodules: tfe.Bool(vcsRepo["ingress_submodules"].(bool)),
			OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
			TagsRegex:         tfe.String(vcsRepo["tags_regex"].(string)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = tfe.String(branch)
		}
	}

	for _, tagName := range d.Get("tag_names").(*schema.Set).List() {
		name := tagName.(string)
		if len(strings.TrimSpace(name)) != 0 {
			if tagContainsUppercase(name) {
				warnWorkspaceTagsCasing(ctx, name)
			}
			options.Tags = append(options.Tags, &tfe.Tag{Name: name})
		}
	}

	log.Printf("[DEBUG] Create workspace %s for organization: %s", name, organization)
	workspace, err := tfeClient.Workspaces.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for organization %s: %w", name, organization, err)
	}

	d.SetId(workspace.ID)

	if sshKeyID, ok := d.GetOk("ssh_key_id"); ok {
		_, err = tfeClient.Workspaces.AssignSSHKey(ctx, workspace.ID, tfe.WorkspaceAssignSSHKeyOptions{
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
		err = tfeClient.Workspaces.AddRemoteStateConsumers(ctx, workspace.ID, options)
		if err != nil {
			return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", name, err)
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
		return fmt.Errorf("Error reading configuration of workspace %s: %w", id, err)
	}

	// Update the config.
	d.Set("name", workspace.Name)
	d.Set("allow_destroy_plan", workspace.AllowDestroyPlan)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("description", workspace.Description)
	d.Set("file_triggers_enabled", workspace.FileTriggersEnabled)
	d.Set("operations", workspace.Operations)
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("speculative_enabled", workspace.SpeculativeEnabled)
	d.Set("structured_run_output_enabled", workspace.StructuredRunOutputEnabled)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("trigger_prefixes", workspace.TriggerPrefixes)
	d.Set("trigger_patterns", workspace.TriggerPatterns)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("organization", workspace.Organization.Name)

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
			"identifier":         workspace.VCSRepo.Identifier,
			"branch":             workspace.VCSRepo.Branch,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
			"oauth_token_id":     workspace.VCSRepo.OAuthTokenID,
			"tags_regex":         workspace.VCSRepo.TagsRegex,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)

	if workspace.GlobalRemoteState {
		d.Set("global_remote_state", true)
	} else {
		globalRemoteState, remoteStateConsumerIDs, err := readWorkspaceStateConsumers(id, tfeClient)
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
	tfeClient := meta.(*tfe.Client)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") ||
		d.HasChange("vcs_repo") || d.HasChange("file_triggers_enabled") ||
		d.HasChange("trigger_prefixes") || d.HasChange("trigger_patterns") ||
		d.HasChange("allow_destroy_plan") || d.HasChange("speculative_enabled") ||
		d.HasChange("operations") || d.HasChange("execution_mode") ||
		d.HasChange("description") || d.HasChange("agent_pool_id") ||
		d.HasChange("global_remote_state") || d.HasChange("structured_run_output_enabled") {
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
				options.TriggerPatterns = append(options.TriggerPatterns, tp.(string))
			}
		} else {
			options.TriggerPatterns = []string{}
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
				TagsRegex:         tfe.String(vcsRepo["tags_regex"].(string)),
			}
		}

		log.Printf("[DEBUG] Update workspace %s", id)
		_, err := tfeClient.Workspaces.UpdateByID(ctx, id, options)
		if err != nil {
			d.Partial(true)
			return fmt.Errorf(
				"Error updating workspace %s: %w", id, err)
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
				d.Partial(true)
				return fmt.Errorf("Error removing VCS repo from workspace %s: %w", id, err)
			}
		}
	}

	if d.HasChange("ssh_key_id") {
		sshKeyID := d.Get("ssh_key_id").(string)

		if sshKeyID != "" {
			_, err := tfeClient.Workspaces.AssignSSHKey(
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
			_, err := tfeClient.Workspaces.UnassignSSHKey(ctx, id)
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
				if tagContainsUppercase(name) {
					warnWorkspaceTagsCasing(ctx, name)
				}
				addTags = append(addTags, &tfe.Tag{Name: name})
			}

			log.Printf("[DEBUG] Adding tags to workspace: %s", d.Id())
			err := tfeClient.Workspaces.AddTags(ctx, d.Id(), tfe.WorkspaceAddTagsOptions{Tags: addTags})
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
			err := tfeClient.Workspaces.RemoveTags(ctx, d.Id(), tfe.WorkspaceRemoveTagsOptions{Tags: removeTags})
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
			err := tfeClient.Workspaces.AddRemoteStateConsumers(ctx, d.Id(), options)
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
			err := tfeClient.Workspaces.RemoveRemoteStateConsumers(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error removing remote state consumers from workspace %s: %w", d.Id(), err)
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
			"Error deleting workspace %s: %w", id, err)
	}

	return nil
}

// An agent pool can only be specified when execution_mode is set to "agent". You currently cannot specify a
// schema validation based on a different argument's value, so we do so here at plan time instead.
func validateAgentExecution(_ context.Context, d *schema.ResourceDiff) error {
	if executionMode, ok := d.GetOk("execution_mode"); ok {
		if executionMode.(string) != "agent" && d.Get("agent_pool_id") != "" {
			return fmt.Errorf("execution_mode must be set to 'agent' to assign agent_pool_id")
		} else if executionMode.(string) == "agent" && d.NewValueKnown("agent_pool_id") && d.Get("agent_pool_id") == "" {
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

func validateRemoteState(_ context.Context, d *schema.ResourceDiff) error {
	// If remote state consumers aren't set, the global setting can be either value and it
	// doesn't matter.
	_, ok := d.GetOk("remote_state_consumer_ids")
	if !ok {
		return nil
	}

	if globalRemoteState, ok := d.GetOk("global_remote_state"); ok {
		if globalRemoteState.(bool) == true {
			return fmt.Errorf("global_remote_state must be 'false' when setting remote_state_consumer_ids")
		}
	}

	return nil
}

func validateVcsTriggers(d *schema.ResourceDiff) {
	if d.HasChange("trigger_patterns") {
		d.SetNewComputed("trigger_prefixes")
	} else if d.HasChange("trigger_prefixes") {
		d.SetNewComputed("trigger_patterns")
	}
}

func resourceTFEWorkspaceImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], tfeClient)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace with name %s from organization %s %w", s[1], s[0], err)
		}

		d.SetId(workspaceID)
	}

	return []*schema.ResourceData{d}, nil
}

// Warns the user that a workspace tag containing uppercase characters will be downcased.
func warnWorkspaceTagsCasing(ctx context.Context, tag string) {
	log.Printf("[WARN] The tag \"%s\" contains uppercase characters that will be transformed by the API. Please update your configuration to lowercase tags in order to avoid conflicts with state.", tag)
}
