// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

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
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
	"github.com/microsoft/kiota-abstractions-go/serialization"
)

var workspaceIDRegexp = regexp.MustCompile("^ws-[a-zA-Z0-9]{16}$")

func resourceTFEWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a workspace resource." +
			"\n\n~> **Note:** Setting the execution mode and agent pool affinity directly on the workspace is deprecated in favor of using both [tfe_workspace_settings](workspace_settings) and [tfe_organization_default_settings](organization_default_settings), since they allow more precise control and fully support [agent_pool_allowed_workspaces](agent_pool_allowed_workspaces). Use caution when unsetting `execution_mode`, as it now leaves any prior value unmanaged instead of reverting to the old default value of `\"remote\"`." +
			"\n\n-> **Note:** `auto_destroy_at` is not intended for workspaces containing production resources or long-lived workspaces. Since this attribute is in-part managed by HCP Terraform, using `ignore_changes` for this attribute may be preferred.",

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
			if err := validateAgentExecution(c, d); err != nil {
				return err
			}

			if err := validateTagNames(c, d); err != nil {
				return err
			}

			if err := customizeDiffIfProviderDefaultOrganizationChanged(c, d, meta); err != nil {
				return err
			}

			if err := customizeDiffAutoDestroyAt(c, d); err != nil {
				return err
			}

			if err := customizeDiffAutoDestroyActivityDuration(c, d); err != nil {
				return err
			}

			if d.HasChange("name") {
				if err := d.SetNewComputed("html_url"); err != nil {
					return err
				}
			}

			return nil
		},

		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"hostname": {
						Type:              schema.TypeString,
						OptionalForImport: true,
					},
				}
			},
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The workspace ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the workspace.",
			},

			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "A description for the workspace.",
			},

			"agent_pool_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"operations"},
				Deprecated:    "Use resource `tfe_workspace_settings` to modify the workspace execution settings. This attribute will be removed in a future release of the provider.",
				Description:   "The ID of an agent pool to assign to the workspace.",
			},

			"allow_destroy_plan": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether destroy plans can be queued on the workspace.",
			},

			"auto_apply": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to automatically apply changes when a Terraform plan is successful. Defaults to `false`.",
			},

			"auto_apply_run_trigger": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to automatically apply changes for runs that were created by run triggers from another workspace. Defaults to `false`.",
			},

			"auto_destroy_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "A future date/time string at which point all resources in a workspace will be scheduled for deletion. Must be a string in RFC3339 format (e.g. \"2100-01-01T00:00:00Z\"). Conflicts with `auto_destroy_activity_duration`.",
			},

			"auto_destroy_activity_duration": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"auto_destroy_at"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^\d{1,4}[dh]$`), "must be 1-4 digits followed by d or h"),
				Description:   "A duration string of the period of time after workspace activity to automatically schedule an auto-destroy run. Must be of the form `<number><unit>` where allowed unit values are \"d\" and \"h\". Conflicts with `auto_destroy_at`.",
			},

			"execution_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"operations"},
				Deprecated:    "Use resource `tfe_workspace_settings` to modify the workspace execution settings. This attribute will be removed in a future release of the provider.",
				ValidateFunc: validation.StringInSlice(
					[]string{
						"agent",
						"local",
						"remote",
					},
					false,
				),
				Description: "Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) to use.",
			},

			"file_triggers_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to filter runs based on the changed files in a VCS push. Defaults to `true`. If enabled, the working directory and trigger prefixes describe a set of paths which must contain changes for a VCS push to trigger a run. If disabled, any push will trigger a run.",
			},

			"global_remote_state": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Deprecated:  "Use resource `tfe_workspace_settings` to modify the workspace `global_remote_state`. `global_remote_state` on `tfe_workspace` is no longer validated properly and will be removed in a future release of the provider.",
				Description: "Whether the workspace allows all workspaces in the organization to access its state data during runs.",
			},

			"inherits_project_auto_destroy": {
				Type:        schema.TypeBool,
				Optional:    false,
				Computed:    true,
				Required:    false,
				Description: "Indicates whether this workspace inherits project auto destroy settings.",
			},

			"remote_state_consumer_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Deprecated:  "Use resource `tfe_workspace_settings` to modify the workspace `remote_state_consumer_ids`. `remote_state_consumer_ids` on `tfe_workspace` is no longer validated properly and will be removed in a future release of the provider.",
				Description: "The set of workspace IDs set as explicit remote state consumers for the given workspace.",
			},

			"assessments_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to regularly run health assessments such as drift detection on the workspace. Defaults to `false`.",
			},

			"operations": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Use resource `tfe_workspace_settings` to modify the workspace execution settings. This attribute will be removed in a future release of the provider.",
				ConflictsWith: []string{"execution_mode", "agent_pool_id"},
				Description:   "Whether to use remote execution mode. Defaults to `true`. When set to `false`, the workspace will be used for state storage only. This value **must not** be provided if `execution_mode` is provided.",
			},

			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the project where the workspace should be created.",
			},

			"queue_all_runs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the workspace should start automatically performing runs immediately after its creation. Defaults to `true`. When set to `false`, runs triggered by a webhook (such as a commit in VCS) will not be queued until at least one run has been manually queued. **Note** that this default differs from the HCP Terraform API default, which is `false`. The provider uses `true` as any workspace provisioned with `false` would need to then have a run manually queued out-of-band before accepting webhooks.",
			},

			"source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"source_url"},
				Description:  "A friendly name for the application or client creating this workspace. If set, this will be displayed on the workspace as \"Created via <SOURCE NAME>\". This value cannot be updated after initial creation. Use `terraform apply -replace` to update this value. Requires `source_url` to also be set.",
			},

			"source_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				RequiredWith: []string{"source_name"},
				Description:  "A URL for the application or client creating this workspace. This can be the URL of a related resource in another app, or a link to documentation or other info about the client. Requires `source_name` to also be set. This value cannot be updated after initial creation. Use `terraform apply -replace` to update this value. **Note:** The API does not (currently) allow this to be updated after a workspace has been created, so modifying this value will result in the workspace being replaced. To disable this, use an [ignore changes](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) lifecycle meta-argument.",
			},

			"speculative_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether this workspace allows speculative plans. Defaults to `true`. Setting this to `false` prevents HCP Terraform or the Terraform Enterprise instance from running plans on pull requests, which can improve security if the VCS repository is public or includes untrusted contributors.",
			},

			"ssh_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The ID of an SSH key to assign to the workspace.",
			},

			"structured_run_output_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether this workspace should show output from Terraform runs using the enhanced UI when available. Defaults to `true`. Setting this to `false` ensures that all runs in this workspace will display their output as text logs.",
			},

			"tag_names": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of tag names for this workspace. Note that tags must only contain lowercase letters, numbers, colons, or hyphens.",
			},

			"ignore_additional_tag_names": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Explicitly ignores `tag_names` _not_ defined by config so they will not be overwritten by the configured tags. This creates exceptional behavior in terraform with respect to `tag_names` and is not recommended. This value must be applied before it will be used.",
			},

			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of key value tags for this workspace.",
			},

			"ignore_additional_tags": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Explicitly ignores `tags` _not_ defined by config so they will not be overwritten by the configured tags. This creates exceptional behavior in terraform with respect to `tags` and is not recommended. This value must be applied before it will be used.",
			},

			"effective_tags": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of key value tags for this workspace, including any tags inherited from the parent project.",
			},

			"terraform_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The version of Terraform to use for this workspace. This can be either an exact version or a [version constraint](https://developer.hashicorp.com/terraform/language/expressions/version-constraints) (like `~> 1.0.0`); if you specify a constraint, the workspace will always use the newest release that meets that constraint. Defaults to the latest available version.",
			},

			"trigger_prefixes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"trigger_patterns"},
				Description:   "List of repository-root-relative paths which describe all locations to be tracked for changes.",
			},

			"trigger_patterns": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"trigger_prefixes"},
				Description:   "List of [glob patterns](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) that describe the files HCP Terraform monitors for changes. Trigger patterns are always appended to the root directory of the repository. Mutually exclusive with `trigger_prefixes`.",
			},

			"working_directory": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A relative path that Terraform will execute within. Defaults to the root of your repository.",
			},

			"vcs_repo": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Description: "Settings for the workspace's VCS repository, enabling the [UI/VCS-driven run workflow](https://developer.hashicorp.com/terraform/cloud-docs/run/ui). Omit this argument to utilize the [CLI-driven](https://developer.hashicorp.com/terraform/cloud-docs/run/cli) and [API-driven](https://developer.hashicorp.com/terraform/cloud-docs/run/api) workflows, where runs are not driven by webhooks on your VCS provider.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider. The format for Azure DevOps is `<ado organization>/<ado project>/_git/<ado repository>`.",
						},

						"branch": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The repository branch that Terraform will execute from. Defaults to the repository's default branch (e.g. main).",
						},

						"ingress_submodules": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether submodules should be fetched when cloning the VCS repository. Defaults to `false`.",
						},

						"oauth_token_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.github_app_installation_id"},
							Description:   "The VCS Connection (OAuth Connection + Token) to use. This ID can be obtained from a `tfe_oauth_client` resource. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.",
						},

						"tags_regex": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"trigger_patterns", "trigger_prefixes"},
							Description:   "A regular expression used to trigger a workspace run for matching Git tags. This option conflicts with `trigger_patterns` and `trigger_prefixes`. Should only set this value if the former is not being used.",
						},

						"github_app_installation_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.oauth_token_id"},
							AtLeastOneOf:  []string{"vcs_repo.0.oauth_token_id", "vcs_repo.0.github_app_installation_id"},
							Description:   "The installation ID of the GitHub App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.",
						},
					},
				},
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, the workspace will be force deleted when destroyed via this provider, even if the workspace contains resources managed by Terraform. If this is false or omitted, it will safe delete the workspace.",
			},
			"resource_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of resources managed by the workspace.",
			},
			"html_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to the browsable HTML overview of the workspace.",
			},
			"hyok_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "(Available only in HCP Terraform) Whether HYOK (Hold Your Own Key) is enabled for the workspace.",
			},
		},
	}
}

func resourceTFEWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Build workspace attributes.
	attrs := models.NewWorkspaces_attributes()
	attrs.SetName(ptr(name))
	attrs.SetAllowDestroyPlan(ptr(d.Get("allow_destroy_plan").(bool)))
	attrs.SetAutoApplyRunTrigger(ptr(d.Get("auto_apply_run_trigger").(bool)))
	attrs.SetFileTriggersEnabled(ptr(d.Get("file_triggers_enabled").(bool)))
	attrs.SetQueueAllRuns(ptr(d.Get("queue_all_runs").(bool)))
	attrs.SetSpeculativeEnabled(ptr(d.Get("speculative_enabled").(bool)))
	attrs.SetStructuredRunOutputEnabled(ptr(d.Get("structured_run_output_enabled").(bool)))
	attrs.SetWorkingDirectory(ptr(d.Get("working_directory").(string)))

	// Send global_remote_state if it's set; otherwise, let it be computed.
	if v, ok := d.GetOkExists("global_remote_state"); ok { //nolint:staticcheck
		attrs.SetGlobalRemoteState(ptr(v.(bool)))
	}

	if v, ok := d.GetOkExists("auto_apply"); ok { //nolint:staticcheck
		attrs.SetAutoApply(ptr(v.(bool)))
	}

	if v, ok := d.GetOkExists("assessments_enabled"); ok { //nolint:staticcheck
		attrs.SetAssessmentsEnabled(ptr(v.(bool)))
	}

	if v, ok := d.GetOk("description"); ok {
		attrs.SetDescription(ptr(v.(string)))
	}

	// Build setting overwrites and relationships.
	settingOverwrites := models.NewWorkspaces_attributes_settingOverwrites()
	settingOverwritesExplicit := false

	var rels *models.Workspaces_relationships

	if v, ok := d.GetOk("agent_pool_id"); ok && v.(string) != "" {
		settingOverwrites.SetExecutionMode(ptr(true))
		settingOverwrites.SetAgentPool(ptr(true))
		settingOverwritesExplicit = true

		apType := models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE
		apData := models.NewAgentPoolsHasOne_data()
		apData.SetId(ptr(v.(string)))
		apData.SetTypeEscaped(&apType)
		apHasOne := models.NewAgentPoolsHasOne()
		apHasOne.SetData(apData)
		rels = models.NewWorkspaces_relationships()
		rels.SetAgentPool(apHasOne)
	}

	if _, ok := d.GetOk("auto_destroy_at"); ok {
		rawV := d.GetRawConfig().GetAttr("auto_destroy_at")
		if !rawV.IsNull() {
			t, parseErr := time.Parse(time.RFC3339, rawV.AsString())
			if parseErr != nil {
				return fmt.Errorf("Error expanding auto destroy during create: %w", parseErr)
			}
			attrs.SetAutoDestroyAt(ptr(t))
		}
	}

	if v, ok := d.GetOk("auto_destroy_activity_duration"); ok {
		attrs.SetAutoDestroyActivityDuration(ptr(v.(string)))
	}

	if v, ok := d.GetOk("execution_mode"); ok {
		modeAny, _ := models.ParseWorkspaces_attributes_executionMode(v.(string))
		if modeAny != nil {
			mode := modeAny.(*models.Workspaces_attributes_executionMode)
			attrs.SetExecutionMode(mode)
		}
		settingOverwrites.SetExecutionMode(ptr(true))
		settingOverwrites.SetAgentPool(ptr(true))
		settingOverwritesExplicit = true
	}

	if v, ok := d.GetOkExists("operations"); ok { //nolint:staticcheck
		attrs.SetOperations(ptr(v.(bool)))
		settingOverwrites.SetExecutionMode(ptr(true))
		settingOverwrites.SetAgentPool(ptr(true))
		settingOverwritesExplicit = true
	}

	if !settingOverwritesExplicit {
		settingOverwrites.SetExecutionMode(ptr(false))
		settingOverwrites.SetAgentPool(ptr(false))
	}
	attrs.SetSettingOverwrites(settingOverwrites)

	if v, ok := d.GetOk("source_url"); ok {
		attrs.SetSourceUrl(ptr(v.(string)))
	}
	if v, ok := d.GetOk("source_name"); ok {
		attrs.SetSourceName(ptr(v.(string)))
	}

	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		attrs.SetTerraformVersion(ptr(tfVersion.(string)))
	}

	if tps, ok := d.GetOk("trigger_prefixes"); ok {
		var prefixes []string
		for _, tp := range tps.([]interface{}) {
			if val, ok := tp.(string); ok {
				prefixes = append(prefixes, val)
			}
		}
		attrs.SetTriggerPrefixes(prefixes)
	}

	if tps, ok := d.GetOk("trigger_patterns"); ok {
		var patterns []string
		for _, tp := range tps.([]interface{}) {
			patterns = append(patterns, tp.(string))
		}
		attrs.SetTriggerPatterns(patterns)
	}

	if d.HasChange("project_id") {
		if v, ok := d.GetOk("project_id"); ok && v.(string) != "" {
			prjData := models.NewProjectsHasOne_data()
			prjData.SetId(ptr(v.(string)))
			prjHasOne := models.NewProjectsHasOne()
			prjHasOne.SetData(prjData)
			if rels == nil {
				rels = models.NewWorkspaces_relationships()
			}
			rels.SetProject(prjHasOne)
		}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})
		vcsAttrs := models.NewWorkspaces_attributes_vcsRepo()
		vcsAttrs.SetIdentifier(ptr(vcsRepo["identifier"].(string)))
		vcsAttrs.SetIngressSubmodules(ptr(vcsRepo["ingress_submodules"].(bool)))
		if tagsRegex, ok := vcsRepo["tags_regex"].(string); ok && tagsRegex != "" {
			vcsAttrs.SetTagsRegex(ptr(tagsRegex))
		}
		if oauthTokenID, ok := vcsRepo["oauth_token_id"].(string); ok && oauthTokenID != "" {
			vcsAttrs.SetOauthTokenId(ptr(oauthTokenID))
		}
		if ghaInstallationID, ok := vcsRepo["github_app_installation_id"].(string); ok && ghaInstallationID != "" {
			vcsAttrs.SetGithubAppInstallationId(ptr(ghaInstallationID))
		}
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			vcsAttrs.SetBranch(ptr(branch))
		}
		attrs.SetVcsRepo(vcsAttrs)
	}

	// Build and send the workspace creation envelope.
	wsType := models.WORKSPACES_WORKSPACES_TYPE
	wsData := models.NewWorkspaces()
	wsData.SetTypeEscaped(&wsType)
	wsData.SetAttributes(attrs)
	if rels != nil {
		wsData.SetRelationships(rels)
	}
	envelope := models.NewWorkspacesEnvelope()
	envelope.SetData(wsData)

	log.Printf("[DEBUG] Create workspace %s for organization: %s", name, organization)
	resp, err := api.Organizations().ByOrganization_name(organization).Workspaces().Post(ctx, envelope, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for organization %s: %w", name, organization, err)
	}

	wsID := valueOrZero(resp.GetData().GetId())
	d.SetId(wsID)

	err = helpers.WriteTFEIdentity(d, wsID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	// tag_names (old-style flat tag names) — the workspace create POST body
	// attributes.tag-names field is not processed by Atlas; tags must be set via
	// the tags relationship endpoint after the workspace exists.
	if tagNamesSet := d.Get("tag_names").(*schema.Set); tagNamesSet.Len() > 0 {
		var addData []models.TagsCreateArrayDocument_dataable
		for _, tagName := range tagNamesSet.List() {
			tagAttrs := models.NewTagsCreateArrayDocument_data_attributes()
			tagAttrs.SetName(ptr(tagName.(string)))
			tagData := models.NewTagsCreateArrayDocument_data()
			tagType := models.TAGS_TAGSCREATEARRAYDOCUMENT_DATA_TYPE
			tagData.SetTypeEscaped(&tagType)
			tagData.SetAttributes(tagAttrs)
			addData = append(addData, tagData)
		}
		tagBody := models.NewTagsCreateArrayDocument()
		tagBody.SetData(addData)
		log.Printf("[DEBUG] Adding tag_names to workspace: %s", wsID)
		if tagErr := api.Workspaces().ByWorkspace_id(wsID).Relationships().Tags().Post(ctx, tagBody, nil); tagErr != nil {
			return fmt.Errorf("Error adding tag_names to workspace %s: %w", name, tagErr)
		}
	}

	// tag_bindings (new-style key/value tags) — posted after create since the
	// workspace create POST body does not directly support tag bindings in v2.
	if tagBindings, ok := d.Get("tags").(map[string]interface{}); ok && len(tagBindings) > 0 {
		var tbData []models.TagBindingsable
		for key, val := range tagBindings {
			tb := models.NewTagBindings()
			tbType := models.TAGBINDINGS_TAGBINDINGS_TYPE
			tb.SetTypeEscaped(&tbType)
			tbAttrs := models.NewTagBindings_attributes()
			tbAttrs.SetKey(ptr(key))
			tbAttrs.SetValue(ptr(val.(string)))
			tb.SetAttributes(tbAttrs)
			tbData = append(tbData, tb)
		}
		collection := models.NewTagBindingsCollection()
		collection.SetData(tbData)
		if tbErr := api.Workspaces().ByWorkspace_id(wsID).Relationships().TagBindings().Post(ctx, collection, nil); tbErr != nil {
			return fmt.Errorf("Error adding tag bindings to workspace %s: %w", name, tbErr)
		}
	}

	// SSH key assignment.
	if sshKeyID, ok := d.GetOk("ssh_key_id"); ok {
		if sshErr := assignSSHKeyV2(ctx, config.ClientV2, wsID, sshKeyID.(string)); sshErr != nil {
			return fmt.Errorf("Error assigning SSH key to workspace %s: %w", name, sshErr)
		}
	}

	// Remote state consumers.
	globalRemoteState, grOk := d.GetOkExists("global_remote_state") //nolint:staticcheck
	remoteStateConsumerIDs, rscOk := d.GetOk("remote_state_consumer_ids")
	if rscOk && grOk && !globalRemoteState.(bool) {
		var consumerData []models.WorkspacesIdentifierArrayDocument_dataable
		wsIdentType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE
		for _, consumerID := range remoteStateConsumerIDs.(*schema.Set).List() {
			wsIdent := models.NewWorkspacesIdentifierArrayDocument_data()
			wsIdent.SetId(ptr(consumerID.(string)))
			wsIdent.SetTypeEscaped(&wsIdentType)
			consumerData = append(consumerData, wsIdent)
		}
		body := models.NewWorkspacesIdentifierArrayDocument()
		body.SetData(consumerData)
		if rscErr := api.Workspaces().ByWorkspace_id(wsID).Relationships().RemoteStateConsumers().Post(ctx, body, nil); rscErr != nil {
			return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", name, rscErr)
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

func resourceTFEWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of workspace: %s", id)

	resp, err := api.Workspaces().ByWorkspace_id(id).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Workspace %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %w", id, err)
	}

	wsData := resp.GetData()
	attrs := wsData.GetAttributes()
	rels := wsData.GetRelationships()
	wsID := valueOrZero(wsData.GetId())

	err = helpers.WriteTFEIdentity(d, wsID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	// Fetch effective tag bindings via the dedicated v2 endpoint.
	// If the server does not support this endpoint (e.g. older TFE), log and
	// continue with an empty list, matching the ErrInvalidIncludeValue fallback
	// behavior from the previous v1 implementation.
	var effectiveTagBindings []*tfe.EffectiveTagBinding
	etbResp, etbErr := config.ClientV2.API.Workspaces().ByWorkspace_id(id).EffectiveTagBindings().Get(ctx, nil)
	if etbErr != nil {
		log.Printf("[DEBUG] Workspace %s effective-tag-bindings unavailable, skipping: %v", id, etbErr)
	} else if etbResp != nil {
		effectiveTagBindings = v2EffectiveTagBindings(etbResp.GetData())
	}

	// Given this computed attribute will be null when tag bindings are not
	// supported, directly set to an empty map to avoid continuous planned
	// changes on this attribute.
	d.Set("effective_tags", map[string]interface{}{})

	tagInfo := helpers.NewTagInfo(d.Get("tags").(map[string]interface{}), effectiveTagBindings, d.Get("ignore_additional_tags").(bool))

	// Update the config.
	d.Set("name", valueOrZero(attrs.GetName()))
	d.Set("allow_destroy_plan", valueOrZero(attrs.GetAllowDestroyPlan()))

	// TFE (onprem) does not currently have this feature and this value won't be returned in those cases.
	d.Set("assessments_enabled", valueOrZero(attrs.GetAssessmentsEnabled()))

	d.Set("auto_apply", valueOrZero(attrs.GetAutoApply()))
	d.Set("auto_apply_run_trigger", valueOrZero(attrs.GetAutoApplyRunTrigger()))
	d.Set("description", valueOrZero(attrs.GetDescription()))
	d.Set("file_triggers_enabled", valueOrZero(attrs.GetFileTriggersEnabled()))
	d.Set("operations", valueOrZero(attrs.GetOperations()))
	d.Set("execution_mode", enumStringOrEmpty(attrs.GetExecutionMode()))
	d.Set("effective_tags", tagInfo.EffectiveTags)
	d.Set("queue_all_runs", valueOrZero(attrs.GetQueueAllRuns()))
	d.Set("source_name", valueOrZero(attrs.GetSourceName()))
	d.Set("source_url", valueOrZero(attrs.GetSourceUrl()))
	d.Set("speculative_enabled", valueOrZero(attrs.GetSpeculativeEnabled()))
	d.Set("structured_run_output_enabled", valueOrZero(attrs.GetStructuredRunOutputEnabled()))
	d.Set("tags", tagInfo.SelfTags)
	d.Set("terraform_version", valueOrZero(attrs.GetTerraformVersion()))
	d.Set("trigger_prefixes", attrs.GetTriggerPrefixes())
	d.Set("trigger_patterns", attrs.GetTriggerPatterns())
	d.Set("working_directory", valueOrZero(attrs.GetWorkingDirectory()))
	d.Set("resource_count", int(valueOrZero(attrs.GetResourceCount())))
	d.Set("inherits_project_auto_destroy", valueOrZero(attrs.GetInheritsProjectAutoDestroy()))
	d.Set("hyok_enabled", valueOrZero(attrs.GetHyokEnabled()))

	// Organization name from the relationship.
	if orgRel := rels.GetOrganization(); orgRel != nil && orgRel.GetData() != nil {
		d.Set("organization", valueOrZero(orgRel.GetData().GetId()))
	}

	// html_url from links.
	if links := wsData.GetLinks(); links != nil && links.GetSelfHtml() != nil {
		baseAPI := config.Client.BaseURL()
		htmlURL := url.URL{
			Scheme: baseAPI.Scheme,
			Host:   baseAPI.Host,
			Path:   valueOrZero(links.GetSelfHtml()),
		}
		d.Set("html_url", htmlURL.String())
	}

	// Project from the relationship.
	if prjRel := rels.GetProject(); prjRel != nil && prjRel.GetData() != nil {
		d.Set("project_id", valueOrZero(prjRel.GetData().GetId()))
	}

	// SSH key from the relationship.
	sshKeyID := ""
	if sshRel := rels.GetSshKey(); sshRel != nil && sshRel.GetData() != nil {
		sshKeyID = valueOrZero(sshRel.GetData().GetId())
	}
	d.Set("ssh_key_id", sshKeyID)

	// Agent pool from the relationship.
	agentPoolID := ""
	if apRel := rels.GetAgentPool(); apRel != nil && apRel.GetData() != nil {
		agentPoolID = valueOrZero(apRel.GetData().GetId())
	}
	d.Set("agent_pool_id", agentPoolID)

	// auto_destroy_at: v2 returns *time.Time.
	if t := attrs.GetAutoDestroyAt(); t != nil {
		d.Set("auto_destroy_at", t.Format(time.RFC3339))
	} else {
		d.Set("auto_destroy_at", nil)
	}

	// auto_destroy_activity_duration: v2 returns *string.
	if dur := attrs.GetAutoDestroyActivityDuration(); dur != nil {
		d.Set("auto_destroy_activity_duration", *dur)
	} else {
		d.Set("auto_destroy_activity_duration", nil)
	}

	// tag_names. Newer responses expose these through the tags relationship.
	workspaceTagNames := attrs.GetTagNames()
	if tagsRel := rels.GetTags(); tagsRel != nil {
		workspaceTagNames = workspaceTagNames[:0]
		for _, tag := range tagsRel.GetData() {
			if tagAttrs := tag.GetAttributes(); tagAttrs != nil {
				workspaceTagNames = append(workspaceTagNames, valueOrZero(tagAttrs.GetName()))
			}
		}
	}
	var tagNames []interface{}
	managedTags := d.Get("tag_names").(*schema.Set)
	for _, tagName := range workspaceTagNames {
		if managedTags.Contains(tagName) || !d.Get("ignore_additional_tag_names").(bool) {
			tagNames = append(tagNames, tagName)
		}
	}
	d.Set("tag_names", tagNames)

	// vcs_repo.
	var vcsRepo []interface{}
	if vcsAttrs := attrs.GetVcsRepo(); vcsAttrs != nil && valueOrZero(vcsAttrs.GetIdentifier()) != "" {
		vcsConfig := map[string]interface{}{
			"identifier":                 valueOrZero(vcsAttrs.GetIdentifier()),
			"branch":                     valueOrZero(vcsAttrs.GetBranch()),
			"ingress_submodules":         valueOrZero(vcsAttrs.GetIngressSubmodules()),
			"oauth_token_id":             valueOrZero(vcsAttrs.GetOauthTokenId()),
			"github_app_installation_id": valueOrZero(vcsAttrs.GetGithubAppInstallationId()),
			"tags_regex":                 valueOrZero(vcsAttrs.GetTagsRegex()),
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	if valueOrZero(attrs.GetGlobalRemoteState()) {
		d.Set("global_remote_state", true)
	} else {
		globalRemoteState, remoteStateConsumerIDs, rscErr := readWorkspaceStateConsumersV2(id, config.ClientV2)
		if rscErr != nil {
			return fmt.Errorf(
				"Error reading remote state consumers for workspace %s: %w", id, rscErr)
		}
		d.Set("global_remote_state", globalRemoteState)
		d.Set("remote_state_consumer_ids", remoteStateConsumerIDs)
	}

	return nil
}

func resourceTFEWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API
	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("auto_apply_run_trigger") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") ||
		d.HasChange("vcs_repo") || d.HasChange("file_triggers_enabled") ||
		d.HasChange("trigger_prefixes") || d.HasChange("trigger_patterns") ||
		d.HasChange("allow_destroy_plan") || d.HasChange("speculative_enabled") ||
		d.HasChange("operations") || d.HasChange("execution_mode") ||
		d.HasChange("description") || d.HasChange("agent_pool_id") ||
		d.HasChange("global_remote_state") || d.HasChange("structured_run_output_enabled") ||
		d.HasChange("assessments_enabled") || d.HasChange("project_id") ||
		hasAutoDestroyAtChange(d) || d.HasChange("auto_destroy_activity_duration") {
		// Build workspace update attributes.
		attrs := models.NewWorkspaces_attributes()
		attrs.SetName(ptr(d.Get("name").(string)))
		attrs.SetAllowDestroyPlan(ptr(d.Get("allow_destroy_plan").(bool)))
		attrs.SetAutoApplyRunTrigger(ptr(d.Get("auto_apply_run_trigger").(bool)))
		attrs.SetFileTriggersEnabled(ptr(d.Get("file_triggers_enabled").(bool)))
		attrs.SetGlobalRemoteState(ptr(d.Get("global_remote_state").(bool)))
		attrs.SetQueueAllRuns(ptr(d.Get("queue_all_runs").(bool)))
		attrs.SetSpeculativeEnabled(ptr(d.Get("speculative_enabled").(bool)))
		attrs.SetStructuredRunOutputEnabled(ptr(d.Get("structured_run_output_enabled").(bool)))
		attrs.SetWorkingDirectory(ptr(d.Get("working_directory").(string)))

		var rels *models.Workspaces_relationships

		if d.HasChange("project_id") {
			if v, ok := d.GetOk("project_id"); ok && v.(string) != "" {
				prjData := models.NewProjectsHasOne_data()
				prjData.SetId(ptr(v.(string)))
				prjHasOne := models.NewProjectsHasOne()
				prjHasOne.SetData(prjData)
				if rels == nil {
					rels = models.NewWorkspaces_relationships()
				}
				rels.SetProject(prjHasOne)
			}
		}

		if d.HasChange("assessments_enabled") {
			if v, ok := d.GetOkExists("assessments_enabled"); ok { //nolint:staticcheck
				attrs.SetAssessmentsEnabled(ptr(v.(bool)))
			}
		}

		if d.HasChange("auto_apply") {
			if v, ok := d.GetOkExists("auto_apply"); ok { //nolint:staticcheck
				attrs.SetAutoApply(ptr(v.(bool)))
			}
		}

		if d.HasChange("description") {
			if v, ok := d.GetOk("description"); ok {
				attrs.SetDescription(ptr(v.(string)))
			}
		}

		// NOTE: since agent_pool_id and execution_mode are both deprecated on
		// tfe_workspace and we want tfe_workspace_settings to be authoritative,
		// we must not set the overwrites values to false in the checks below.
		if d.HasChange("agent_pool_id") {
			agentPoolID := d.GetRawConfig().GetAttr("agent_pool_id")
			if !agentPoolID.IsNull() {
				apType := models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE
				apData := models.NewAgentPoolsHasOne_data()
				apData.SetId(ptr(agentPoolID.AsString()))
				apData.SetTypeEscaped(&apType)
				apHasOne := models.NewAgentPoolsHasOne()
				apHasOne.SetData(apData)
				if rels == nil {
					rels = models.NewWorkspaces_relationships()
				}
				rels.SetAgentPool(apHasOne)

				settingOverwrites := models.NewWorkspaces_attributes_settingOverwrites()
				settingOverwrites.SetAgentPool(ptr(true))
				attrs.SetSettingOverwrites(settingOverwrites)
			}
		}

		if hasAutoDestroyAtChange(d) {
			rawV := d.GetRawConfig().GetAttr("auto_destroy_at")
			if rawV.IsNull() {
				// Clear auto_destroy_at: v2 omits nil pointers, so we use additional data
				// to explicitly send null. This preserves the v1 NullNullableAttr behavior.
				addlData := attrs.GetAdditionalData()
				if addlData == nil {
					addlData = make(map[string]any)
				}
				addlData["auto-destroy-at"] = nil
				attrs.SetAdditionalData(addlData)
			} else {
				t, parseErr := time.Parse(time.RFC3339, rawV.AsString())
				if parseErr != nil {
					return fmt.Errorf("Error expanding auto destroy during update: %w", parseErr)
				}
				attrs.SetAutoDestroyAt(ptr(t))
			}
		}

		if d.HasChange("auto_destroy_activity_duration") {
			duration, durOk := d.GetOk("auto_destroy_activity_duration")
			if !durOk {
				// Clear auto_destroy_activity_duration via additionalData null.
				addlData := attrs.GetAdditionalData()
				if addlData == nil {
					addlData = make(map[string]any)
				}
				addlData["auto-destroy-activity-duration"] = nil
				attrs.SetAdditionalData(addlData)
			} else {
				attrs.SetAutoDestroyActivityDuration(ptr(duration.(string)))
			}
		}

		if d.HasChange("execution_mode") {
			if v, ok := d.GetOk("execution_mode"); ok {
				modeAny, _ := models.ParseWorkspaces_attributes_executionMode(v.(string))
				if modeAny != nil {
					mode := modeAny.(*models.Workspaces_attributes_executionMode)
					attrs.SetExecutionMode(mode)
				}
				settingOverwrites := models.NewWorkspaces_attributes_settingOverwrites()
				settingOverwrites.SetExecutionMode(ptr(true))
				attrs.SetSettingOverwrites(settingOverwrites)
			}
		}

		if d.HasChange("operations") {
			if v, ok := d.GetOkExists("operations"); ok { //nolint:staticcheck
				attrs.SetOperations(ptr(v.(bool)))
			}
		}

		// tag_bindings (new-style key/value tags) — sent via the dedicated
		// tag-bindings PATCH endpoint (replaces all existing bindings).
		if tagBindings, ok := d.Get("tags").(map[string]interface{}); ok {
			var tbData []models.TagBindingsable
			for key, val := range tagBindings {
				tb := models.NewTagBindings()
				tbType := models.TAGBINDINGS_TAGBINDINGS_TYPE
				tb.SetTypeEscaped(&tbType)
				tbAttrs := models.NewTagBindings_attributes()
				tbAttrs.SetKey(ptr(key))
				tbAttrs.SetValue(ptr(val.(string)))
				tb.SetAttributes(tbAttrs)
				tbData = append(tbData, tb)
			}

			if len(tbData) == 0 && !d.Get("ignore_additional_tags").(bool) {
				// Explicitly clear all tag bindings.
				collection := models.NewTagBindingsCollection()
				collection.SetData([]models.TagBindingsable{})
				if tbErr := api.Workspaces().ByWorkspace_id(id).Relationships().TagBindings().Patch(ctx, collection, nil); tbErr != nil {
					d.Partial(true)
					return fmt.Errorf("Error removing tag bindings from workspace %s: %w", id, tbErr)
				}
			} else if len(tbData) > 0 {
				// Replace all tag bindings with the new set.
				collection := models.NewTagBindingsCollection()
				collection.SetData(tbData)
				if tbErr := api.Workspaces().ByWorkspace_id(id).Relationships().TagBindings().Patch(ctx, collection, nil); tbErr != nil {
					d.Partial(true)
					return fmt.Errorf("Error updating tag bindings for workspace %s: %w", id, tbErr)
				}
			}
		}

		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			attrs.SetTerraformVersion(ptr(tfVersion.(string)))
		}

		if tps, ok := d.GetOk("trigger_prefixes"); ok {
			var prefixes []string
			for _, tp := range tps.([]interface{}) {
				if val, ok := tp.(string); ok {
					prefixes = append(prefixes, val)
				}
			}
			attrs.SetTriggerPrefixes(prefixes)
		} else {
			attrs.SetTriggerPrefixes([]string{})
		}

		if tps, ok := d.GetOk("trigger_patterns"); ok {
			var patterns []string
			for _, tp := range tps.([]interface{}) {
				if val, ok := tp.(string); ok {
					patterns = append(patterns, val)
				}
			}
			attrs.SetTriggerPatterns(patterns)
		} else {
			attrs.SetTriggerPatterns([]string{})
		}

		if d.GetRawConfig().GetAttr("trigger_patterns").IsNull() {
			attrs.SetTriggerPatterns(nil)
		} else if d.GetRawConfig().GetAttr("trigger_prefixes").IsNull() {
			attrs.SetTriggerPrefixes(nil)
		}

		if workingDir, ok := d.GetOk("working_directory"); ok {
			attrs.SetWorkingDirectory(ptr(workingDir.(string)))
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})
			vcsAttrs := models.NewWorkspaces_attributes_vcsRepo()
			vcsAttrs.SetIdentifier(ptr(vcsRepo["identifier"].(string)))
			vcsAttrs.SetBranch(ptr(vcsRepo["branch"].(string)))
			vcsAttrs.SetIngressSubmodules(ptr(vcsRepo["ingress_submodules"].(bool)))
			vcsAttrs.SetOauthTokenId(ptr(vcsRepo["oauth_token_id"].(string)))
			vcsAttrs.SetGithubAppInstallationId(ptr(vcsRepo["github_app_installation_id"].(string)))
			vcsAttrs.SetTagsRegex(ptr(vcsRepo["tags_regex"].(string)))
			attrs.SetVcsRepo(vcsAttrs)
		}

		// Remove vcs_repo from the workspace if the value of vcs_repo has been
		// changed by removing it from the config.
		//
		// go-tfe v2 migration: RemoveVCSConnectionByID remains on go-tfe v1.
		// Reason: Removing VCS connection requires sending PATCH /workspaces/{id}
		// with "vcs-repo": null in the attributes. The go-tfe v2 generated client
		// does not support serializing nil values as JSON null for typed
		// attributes. The Kiota serializer omits nil fields rather than
		// serializing them as null, and there is no safe way to force-serialize
		// null via additionalData for this endpoint.
		// Removal condition: when go-tfe/v2 exposes a dedicated VCS-removal
		// operation or supports nullable attribute semantics for the workspace
		// PATCH body.
		if d.HasChange("vcs_repo") {
			_, ok := d.GetOk("vcs_repo")
			if !ok {
				_, vcsErr := config.Client.Workspaces.RemoveVCSConnectionByID(ctx, id)
				if vcsErr != nil {
					d.Partial(true)
					return fmt.Errorf("Error removing VCS repo from workspace %s: %w", id, vcsErr)
				}
			}
		}

		// Build the workspace update envelope.
		wsType := models.WORKSPACES_WORKSPACES_TYPE
		wsData := models.NewWorkspaces()
		wsData.SetTypeEscaped(&wsType)
		wsData.SetAttributes(attrs)
		if rels != nil {
			wsData.SetRelationships(rels)
		}
		updateEnvelope := models.NewWorkspacesEnvelope()
		updateEnvelope.SetData(wsData)

		log.Printf("[DEBUG] Update workspace %s", id)
		_, updateErr := api.Workspaces().ByWorkspace_id(id).Patch(ctx, updateEnvelope, nil)
		if updateErr != nil {
			d.Partial(true)
			return fmt.Errorf(
				"Error updating workspace %s: %w%s", id, updateErr, v2ErrorDetails(updateErr))
		}
	}

	if d.HasChange("ssh_key_id") {
		sshKeyID := d.Get("ssh_key_id").(string)
		if sshKeyID != "" {
			if sshErr := assignSSHKeyV2(ctx, config.ClientV2, id, sshKeyID); sshErr != nil {
				return fmt.Errorf("Error assigning SSH key to workspace %s: %w", id, sshErr)
			}
		} else {
			if sshErr := unassignSSHKeyV2(ctx, config.ClientV2, id); sshErr != nil {
				return fmt.Errorf("Error unassigning SSH key from workspace %s: %w", id, sshErr)
			}
		}
	}

	if d.HasChange("tag_names") {
		oldTagNameValues, newTagNameValues := d.GetChange("tag_names")
		newTagNamesSet := newTagNameValues.(*schema.Set)
		oldTagNamesSet := oldTagNameValues.(*schema.Set)

		newTagNames := newTagNamesSet.Difference(oldTagNamesSet)
		oldTagNames := oldTagNamesSet.Difference(newTagNamesSet)

		// First add the new tags.
		if newTagNames.Len() > 0 {
			var addData []models.TagsCreateArrayDocument_dataable
			for _, tagName := range newTagNames.List() {
				tagAttrs := models.NewTagsCreateArrayDocument_data_attributes()
				tagAttrs.SetName(ptr(tagName.(string)))
				tagData := models.NewTagsCreateArrayDocument_data()
				tagType := models.TAGS_TAGSCREATEARRAYDOCUMENT_DATA_TYPE
				tagData.SetTypeEscaped(&tagType)
				tagData.SetAttributes(tagAttrs)
				addData = append(addData, tagData)
			}
			tagBody := models.NewTagsCreateArrayDocument()
			tagBody.SetData(addData)

			log.Printf("[DEBUG] Adding tags to workspace: %s", d.Id())
			if addErr := api.Workspaces().ByWorkspace_id(d.Id()).Relationships().Tags().Post(ctx, tagBody, nil); addErr != nil {
				return fmt.Errorf("Error adding tags to workspace %s: %w", d.Id(), addErr)
			}
		}

		// Then remove all the old tags: GET IDs first, then DELETE.
		if oldTagNames.Len() > 0 {
			if rmErr := removeTagNamesByNameV2(ctx, config.ClientV2, d.Id(), oldTagNames.List()); rmErr != nil {
				return fmt.Errorf("Error removing tags from workspace %s: %w", d.Id(), rmErr)
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

		wsIdentType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE

		// First add the new consumers.
		if newWorkspaceIDs.Len() > 0 {
			var addData []models.WorkspacesIdentifierArrayDocument_dataable
			for _, wsID := range newWorkspaceIDs.List() {
				wsIdent := models.NewWorkspacesIdentifierArrayDocument_data()
				wsIdent.SetId(ptr(wsID.(string)))
				wsIdent.SetTypeEscaped(&wsIdentType)
				addData = append(addData, wsIdent)
			}
			body := models.NewWorkspacesIdentifierArrayDocument()
			body.SetData(addData)

			log.Printf("[DEBUG] Adding remote state consumers to workspace: %s", d.Id())
			if addErr := api.Workspaces().ByWorkspace_id(d.Id()).Relationships().RemoteStateConsumers().Post(ctx, body, nil); addErr != nil {
				return fmt.Errorf("Error adding remote state consumers to workspace %s: %w", d.Id(), addErr)
			}
		}

		// Then remove all the old consumers.
		if oldWorkspaceIDs.Len() > 0 {
			var rmData []models.WorkspacesIdentifierArrayDocument_dataable
			for _, wsID := range oldWorkspaceIDs.List() {
				wsIdent := models.NewWorkspacesIdentifierArrayDocument_data()
				wsIdent.SetId(ptr(wsID.(string)))
				wsIdent.SetTypeEscaped(&wsIdentType)
				rmData = append(rmData, wsIdent)
			}
			body := models.NewWorkspacesIdentifierArrayDocument()
			body.SetData(rmData)

			log.Printf("[DEBUG] Removing remote state consumers from workspace: %s", d.Id())
			if rmErr := api.Workspaces().ByWorkspace_id(d.Id()).Relationships().RemoteStateConsumers().Delete(ctx, body, nil); rmErr != nil {
				return fmt.Errorf("Error removing remote state consumers from workspace %s: %w", d.Id(), rmErr)
			}
		}
	}

	return resourceTFEWorkspaceRead(d, meta)
}

// errV2Conflict is the v2 sentinel for 409 Conflict, used in place of the
// v1 tfe.ErrWorkspaceStillProcessing sentinel when polling safe-delete.
var errV2Conflict = &tfev2.APIError{StatusCode: 409}

func safeWorkspaceDelete(ctx context.Context, config ConfiguredClient, id string) error {
	api := config.ClientV2.API
	return retry.RetryContext(ctx, time.Duration(5)*time.Minute, func() *retry.RetryError {
		err := api.Workspaces().ByWorkspace_id(id).Actions().SafeDelete().Post(ctx, nil)
		if err != nil {
			// Only the transient state-processing conflict should be retried.
			if errors.Is(err, errV2Conflict) && strings.Contains(strings.ToLower(v2ErrorDetails(err)), "being processed") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

func resourceTFEWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API
	id := d.Id()

	log.Printf("[DEBUG] Delete workspace %s", id)

	resp, err := api.Workspaces().ByWorkspace_id(id).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf(
			"Error reading workspace %s: %w", id, err)
	}

	wsData := resp.GetData()
	attrs := wsData.GetAttributes()
	forceDelete := d.Get("force_delete").(bool)

	// Permissions may be nil on older TFE versions that predate the
	// can-force-delete field; treat nil the same as v1 (CanForceDelete == nil).
	var canForceDelete *bool
	if perms := attrs.GetPermissions(); perms != nil {
		canForceDelete = perms.GetCanForceDelete()
	}

	resourceCount := int(valueOrZero(attrs.GetResourceCount()))

	if canForceDelete == nil {
		// Older TFE: no safe-delete support.
		if forceDelete {
			err = api.Workspaces().ByWorkspace_id(id).Delete(ctx, nil)
		} else {
			return fmt.Errorf(
				"Error deleting workspace %s: This version of Terraform Enterprise does not support workspace safe-delete. Workspaces must be force deleted by setting force_delete=true", id)
		}
	} else if *canForceDelete {
		if forceDelete {
			err = api.Workspaces().ByWorkspace_id(id).Delete(ctx, nil)
		} else {
			err = errWorkspaceResourceCountCheck(id, resourceCount)
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
		err = errWorkspaceResourceCountCheck(id, resourceCount)
		if err != nil {
			return err
		}
		err = safeWorkspaceDelete(ctx, config, id)
	}

	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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
	// since execution_mode and agent_pool_id are marked as Optional: true, and
	// Computed: true, unsetting the execution_mode/agent_pool_id in the config
	// after it's been set to a valid value is not detected by ResourceDiff so
	// we need to read the value from RawConfig instead
	configMap := d.GetRawConfig().AsValueMap()
	executionMode, executionModeReadOk := configMap["execution_mode"]
	agentPoolID, agentPoolIDReadOk := configMap["agent_pool_id"]
	executionModeSet := !executionMode.IsNull() && executionModeReadOk
	agentPoolIDSet := !agentPoolID.IsNull() && agentPoolIDReadOk
	if executionModeSet {
		executionModeIsAgent := executionMode.AsString() == "agent"
		if executionModeIsAgent && !agentPoolIDSet {
			return fmt.Errorf("agent_pool_id must be provided when execution_mode is 'agent'")
		} else if !executionModeIsAgent && agentPoolIDSet {
			return fmt.Errorf("execution_mode must be set to 'agent' to assign agent_pool_id")
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

func resourceTFEWorkspaceImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) >= 3 {
		return nil, fmt.Errorf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <WORKSPACE ID>)",
			d.Id(),
		)
	} else if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalIDV2(s[0]+"/"+s[1], config.ClientV2)
		if err != nil {
			return nil, fmt.Errorf(
				"error retrieving workspace with name %s from organization %s %w", s[1], s[0], err)
		}

		d.SetId(workspaceID)
	}

	identity, err := d.Identity()
	if err != nil {
		return nil, fmt.Errorf("error reading workspace identity: %w", err)
	}

	if externalID := identity.Get("id").(string); externalID != "" {
		// We are importing by identity
		// This only supported when using an import block, since import blocks
		// are the only way to specify an identity. Importing via TF CLI does
		// not support specifying an identity.
		d.SetId(externalID)
	}

	return []*schema.ResourceData{d}, nil
}

func errWorkspaceSafeDeleteWithPermission(workspaceID string, err error) error {
	if err != nil {
		// Check for v2 409 Conflict (workspace has managed resources or is locked),
		// and preserve backward compat with v1 "conflict" prefix error strings.
		if errors.Is(err, errV2Conflict) || strings.HasPrefix(err.Error(), "conflict") {
			return fmt.Errorf("error deleting workspace %s: %w%s\nThis workspace may either have managed resources in state or has a latest state that's still being processed. Add force_delete = true to the resource config to delete this workspace", workspaceID, err, v2ErrorDetails(err))
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

func customizeDiffAutoDestroyAt(_ context.Context, d *schema.ResourceDiff) error {
	config := d.GetRawConfig()

	// check if auto_destroy_activity_duration is set in config
	if !config.GetAttr("auto_destroy_activity_duration").IsNull() {
		return nil
	}

	inheritsProjectAutoDestroy, ok := d.GetOk("inherits_project_auto_destroy")
	if ok && inheritsProjectAutoDestroy.(bool) {
		return nil
	}

	// if config auto_destroy_at is unset but it exists in state, clear it out
	// required because auto_destroy_at is computed and we want to set it to null
	if _, ok := d.GetOk("auto_destroy_at"); ok && config.GetAttr("auto_destroy_at").IsNull() {
		return d.SetNew("auto_destroy_at", nil)
	}

	return nil
}

func customizeDiffAutoDestroyActivityDuration(_ context.Context, d *schema.ResourceDiff) error {
	inheritsProjectAutoDestroy, ok := d.GetOk("inherits_project_auto_destroy")
	if ok && inheritsProjectAutoDestroy.(bool) {
		return nil
	}

	if _, ok := d.GetOk("auto_destroy_activity_duration"); ok && d.GetRawConfig().GetAttr("auto_destroy_activity_duration").IsNull() {
		return d.SetNew("auto_destroy_activity_duration", nil)
	}

	return nil
}

// flattenAutoDestroyAt is retained for compatibility; the v2 CRUD handlers
// handle auto_destroy_at inline using *time.Time.
func flattenAutoDestroyAt(a jsonapi.NullableAttr[time.Time]) (*string, error) {
	if !a.IsSpecified() {
		return nil, nil
	}

	autoDestroyTime, err := a.Get()
	if err != nil {
		return nil, err
	}

	autoDestroyAt := autoDestroyTime.Format(time.RFC3339)
	return &autoDestroyAt, nil
}

func hasAutoDestroyAtChange(d *schema.ResourceData) bool {
	state := d.GetRawState()
	if state.IsNull() {
		return d.HasChange("auto_destroy_at")
	}

	config := d.GetRawConfig()
	autoDestroyAt := config.GetAttr("auto_destroy_at")
	if !autoDestroyAt.IsNull() {
		return d.HasChange("auto_destroy_at")
	}

	return config.GetAttr("auto_destroy_at") != state.GetAttr("auto_destroy_at")
}

// assignSSHKeyV2 assigns an SSH key to a workspace using the go-tfe v2 client.
func assignSSHKeyV2(ctx context.Context, client *tfev2.Client, workspaceID, sshKeyID string) error {
	sshType := models.SSHKEYS_SSHKEYSNULLABLEIDENTIFIERDOCUMENT_DATA_TYPE
	sshData := models.NewSshKeysNullableIdentifierDocument_data()
	sshData.SetId(ptr(sshKeyID))
	sshData.SetTypeEscaped(&sshType)
	sshBody := models.NewSshKeysNullableIdentifierDocument()
	sshBody.SetData(sshData)
	_, err := client.API.Workspaces().ByWorkspace_id(workspaceID).Relationships().SshKey().Patch(ctx, sshBody, nil)
	return err
}

// unassignSSHKeyV2 unassigns the SSH key from a workspace using the go-tfe v2 client.
func unassignSSHKeyV2(ctx context.Context, client *tfev2.Client, workspaceID string) error {
	sshBody := models.NewSshKeysNullableIdentifierDocument()
	sshBody.GetAdditionalData()["data"] = serialization.NewUntypedNull()
	_, err := client.API.Workspaces().ByWorkspace_id(workspaceID).Relationships().SshKey().Patch(ctx, sshBody, nil)
	return err
}

// removeTagNamesByNameV2 removes tag_names from a workspace by resolving their
// names to IDs via a GET, then issuing a DELETE. If the GET fails, it logs and
// returns gracefully, matching the v1 behavior of ignoring missing tag IDs.
func removeTagNamesByNameV2(ctx context.Context, client *tfev2.Client, workspaceID string, tagNames []interface{}) error {
	api := client.API

	tagsResp, err := api.Workspaces().ByWorkspace_id(workspaceID).Relationships().Tags().Get(ctx, nil)
	if err != nil {
		log.Printf("[DEBUG] Could not list tags for workspace %s to resolve tag IDs: %v", workspaceID, err)
		return err
	}

	// Build a name→ID map from the GET response.
	nameToID := map[string]string{}
	for _, t := range tagsResp.GetData() {
		if tAttrs := t.GetAttributes(); tAttrs != nil {
			name := valueOrZero(tAttrs.GetName())
			id := valueOrZero(t.GetId())
			if name != "" && id != "" {
				nameToID[name] = id
			}
		}
	}

	var rmData []models.TagsRemoveArrayDocument_dataable
	for _, tagName := range tagNames {
		name := tagName.(string)
		tagID, ok := nameToID[name]
		if !ok {
			log.Printf("[DEBUG] Tag %q not found on workspace %s, skipping removal", name, workspaceID)
			continue
		}
		rmItem := models.NewTagsRemoveArrayDocument_data()
		rmItem.SetId(ptr(tagID))
		rmType := models.TAGS_TAGSREMOVEARRAYDOCUMENT_DATA_TYPE
		rmItem.SetTypeEscaped(&rmType)
		rmData = append(rmData, rmItem)
	}

	if len(rmData) == 0 {
		return nil
	}

	log.Printf("[DEBUG] Removing tags from workspace: %s", workspaceID)
	removeBody := models.NewTagsRemoveArrayDocument()
	removeBody.SetData(rmData)
	return api.Workspaces().ByWorkspace_id(workspaceID).Relationships().Tags().Delete(ctx, removeBody, nil)
}

// v2EffectiveTagBindings converts a slice of go-tfe v2 EffectiveTagBindingsable
// items into the []*tfe.EffectiveTagBinding slice expected by helpers.NewTagInfo.
// An item whose InheritedFrom relationship is non-nil is treated as inherited
// (matching the v1 behavior where binding.Links["inherited-from"] was non-nil).
func v2EffectiveTagBindings(items []models.EffectiveTagBindingsable) []*tfe.EffectiveTagBinding {
	result := make([]*tfe.EffectiveTagBinding, 0, len(items))
	for _, item := range items {
		etb := &tfe.EffectiveTagBinding{
			ID: valueOrZero(item.GetId()),
		}
		if attrs := item.GetAttributes(); attrs != nil {
			etb.Key = valueOrZero(attrs.GetKey())
			etb.Value = valueOrZero(attrs.GetValue())
		}
		if rels := item.GetRelationships(); rels != nil && rels.GetInheritedFrom() != nil {
			etb.Links = map[string]interface{}{"inherited-from": "set"}
		}
		result = append(result, etb)
	}
	return result
}
