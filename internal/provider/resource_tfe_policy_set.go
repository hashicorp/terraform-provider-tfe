// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/policysets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func resourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Description: "Manages policy sets." +
			"\n\nPolicies are rules enforced on Terraform runs. Two policy-as-code frameworks are integrated with Terraform Enterprise: Sentinel and Open Policy Agent (OPA)." +
			"\n\nPolicy sets are groups of policies that are applied together to related workspaces. By using policy sets, you can group your policies by attributes such as environment or region. Individual policies that are members of policy sets will only be checked for workspaces that the policy set is attached to." +
			"\n\n-> **Note:** When neither `vcs_repo` nor `policy_ids` is specified, the default behavior is to create an empty non-VCS policy set.",

		Create: resourceTFEPolicySetCreate,
		Read:   resourceTFEPolicySetRead,
		Update: resourceTFEPolicySetUpdate,
		Delete: resourceTFEPolicySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughWithIdentity("id"),
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

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
				Description: "The ID of the policy set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description:  "Name of the policy set.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\A[\w\_\-]+\z`), "can only include letters, numbers, -, and _."),
			},

			"description": {
				Description: "A description of the policy set's purpose.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"global": {
				Description:   "Whether or not policies in this set will apply to all workspaces. Defaults to `false`. Conflicts with `workspace_ids`.",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_ids"},
			},

			"kind": {
				Description: "The policy-as-code framework associated with the policy. Defaults to `sentinel` if not provided. Valid values are `sentinel` and `opa`. A policy set can only have policies that have the same underlying kind.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(tfe.Sentinel),
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.OPA),
						string(tfe.Sentinel),
					}, false),
			},

			"overridable": {
				Description: "Whether or not users can override this policy when it fails during a run. Defaults to `false`. Only valid for `opa` policies.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"agent_enabled": {
				Description: "Whether the policy set is executed in the HCP Terraform agent. `true` by default for `opa` policy sets.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"policy_tool_version": {
				Description: "The policy tool version to run the policy evaluation against. For both `sentinel` and `opa` leaving this argument unspecified results in selecting the latest available version at time of creation. For `opa` policy sets, `latest` will not be a valid input.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"policy_update_patterns": {
				Description: "A list of glob patterns specifying which file changes trigger policy set updates. Patterns are relative to the repository root, and you can specify a maximum of 100 patterns. This argument is only valid when you specify a VCS repository for the policy set.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},

			"policies_path": {
				Description:   "The sub-path within the attached VCS repository to ingress when using vcs_repo. All files and directories outside of this sub-path will be ignored. This option can only be supplied when `vcs_repo` is present. Forces a new resource if changed.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"policy_ids"},
			},

			"slug": {
				Description:   "A reference to the `tfe_slug` data source that contains the `source_path` to where the local policies are located. This is used when policies are located locally, and can only be used when there is no VCS repo or explicit policy IDs. Specifically requires the `tfe_slug` data source.",
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"policy_ids", "vcs_repo"},
			},

			"policy_ids": {
				Description:   "A list of Sentinel policy IDs. This value **must not** be provided if `vcs_repo` is provided.",
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vcs_repo", "policies_path"},
			},

			"vcs_repo": {
				Description:   "Settings for the policy sets VCS repository. Forces a new resource if changed. This value must not be provided if `policy_ids` are provided.",
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"policy_ids"},
				MinItems:      1,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Required:    true,
						},

						"branch": {
							Description: "The repository branch that Terraform will execute from. This defaults to the repository's default branch (e.g. main).",
							Type:        schema.TypeString,
							Optional:    true,
						},

						"ingress_submodules": {
							Description: "Whether submodules should be fetched when cloning the VCS repository. Defaults to `false`.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},

						"oauth_token_id": {
							Description:   "Token ID of the VCS Connection (OAuth Connection Token) to use. Conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.",
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.github_app_installation_id"},
						},

						"github_app_installation_id": {
							Description:   "The installation id of the GitHub App. Conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.",
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.oauth_token_id"},
							AtLeastOneOf:  []string{"vcs_repo.0.oauth_token_id", "vcs_repo.0.github_app_installation_id"},
						},
					},
				},
			},

			"workspace_ids": {
				Description:   "A list of workspace IDs. This value must not be provided if `global` is provided.",
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

	attrs := models.NewPolicySets_attributes()
	attrs.SetName(ptr(name))
	attrs.SetGlobal(ptr(d.Get("global").(bool)))

	if vKind, ok := d.GetOk("kind"); ok {
		parsedKind, _ := models.ParsePolicySets_attributes_kind(vKind.(string))
		if parsedKind != nil {
			k := parsedKind.(*models.PolicySets_attributes_kind)
			attrs.SetKind(k)
		}
	}

	if vOverridable, ok := d.GetOk("overridable"); ok {
		attrs.SetOverridable(ptr(vOverridable.(bool)))
	}

	if vAgentEnabled, ok := d.GetOk("agent_enabled"); ok {
		attrs.SetAgentEnabled(ptr(vAgentEnabled.(bool)))
	}

	if vPolicyToolVersion, ok := d.GetOk("policy_tool_version"); ok {
		attrs.SetPolicyToolVersion(ptr(vPolicyToolVersion.(string)))
	}

	if vPolicyUpdatePatterns, ok := d.GetOk("policy_update_patterns"); ok {
		var patterns []string
		for _, p := range vPolicyUpdatePatterns.([]interface{}) {
			patterns = append(patterns, p.(string))
		}
		attrs.SetPolicyUpdatePatterns(patterns)
	} else {
		attrs.SetPolicyUpdatePatterns([]string{})
	}
	if d.GetRawConfig().GetAttr("policy_update_patterns").IsNull() {
		attrs.SetPolicyUpdatePatterns(nil)
	}

	if desc, ok := d.GetOk("description"); ok {
		attrs.SetDescription(ptr(desc.(string)))
	}

	if policiesPath, ok := d.GetOk("policies_path"); ok {
		attrs.SetPoliciesPath(ptr(policiesPath.(string)))
	}

	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})
		vcsAttrs := models.NewPolicySets_attributes_vcsRepo()
		vcsAttrs.SetIdentifier(ptr(vcsRepo["identifier"].(string)))
		vcsAttrs.SetIngressSubmodules(ptr(vcsRepo["ingress_submodules"].(bool)))
		vcsAttrs.SetOauthTokenId(ptr(vcsRepo["oauth_token_id"].(string)))
		vcsAttrs.SetGithubAppInstallationId(ptr(vcsRepo["github_app_installation_id"].(string)))
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			vcsAttrs.SetBranch(ptr(branch))
		}
		attrs.SetVcsRepo(vcsAttrs)
	}

	rels := models.NewPolicySets_relationships()

	var policyData []models.PoliciesIdentifierable
	for _, policyID := range d.Get("policy_ids").(*schema.Set).List() {
		item := models.NewPoliciesIdentifier()
		item.SetId(ptr(policyID.(string)))
		policyData = append(policyData, item)
	}
	if len(policyData) > 0 {
		policiesRel := models.NewPoliciesHasMany()
		policiesRel.SetData(policyData)
		rels.SetPolicies(policiesRel)
	}

	var wsData []models.WorkspacesIdentifierable
	for _, workspaceID := range d.Get("workspace_ids").(*schema.Set).List() {
		item := models.NewWorkspacesIdentifier()
		item.SetId(ptr(workspaceID.(string)))
		wsData = append(wsData, item)
	}
	if len(wsData) > 0 {
		wsRel := models.NewWorkspacesHasMany()
		wsRel.SetData(wsData)
		rels.SetWorkspaces(wsRel)
	}

	body := models.NewPolicySets()
	body.SetAttributes(attrs)
	body.SetRelationships(rels)
	env := models.NewPolicySetsEnvelope()
	env.SetData(body)

	log.Printf("[DEBUG] Create policy set %s for organization: %s", name, organization)
	policySetEnv, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).PolicySets().Post(ctx, env, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating policy set %s for organization %s: %w", name, organization, err)
	}

	policySetID := valueOrZero(policySetEnv.GetData().GetId())

	_, hasVCSRepo := d.GetOk("vcs_repo")
	_, hasSlug := d.GetOk("slug")
	if hasSlug && !hasVCSRepo {
		err := resourceTFEPolicySetUploadVersion(config.Client, d, policySetID)
		if err != nil {
			return err
		}
	}

	d.SetId(policySetID)

	err = helpers.WriteTFEIdentity(d, policySetID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return resourceTFEPolicySetRead(d, meta)
}

func resourceTFEPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read policy set: %s", d.Id())
	policySetEnv, err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] Policy set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading policy set %s: %w", d.Id(), err)
	}

	policySet := policySetEnv.GetData()
	attrs := policySet.GetAttributes()
	rels := policySet.GetRelationships()

	d.Set("name", valueOrZero(attrs.GetName()))
	d.Set("description", valueOrZero(attrs.GetDescription()))
	d.Set("global", valueOrZero(attrs.GetGlobal()))
	d.Set("policies_path", valueOrZero(attrs.GetPoliciesPath()))
	d.Set("policy_update_patterns", attrs.GetPolicyUpdatePatterns())
	d.Set("agent_enabled", valueOrZero(attrs.GetAgentEnabled()))

	if rels != nil && rels.GetOrganization() != nil {
		orgData := rels.GetOrganization().GetData()
		if orgData != nil {
			d.Set("organization", valueOrZero(orgData.GetId()))
		}
	}

	if k := attrs.GetKind(); k != nil {
		d.Set("kind", k.String())
	}

	if ov := attrs.GetOverridable(); ov != nil {
		d.Set("overridable", *ov)
	}

	if ptv := attrs.GetPolicyToolVersion(); ptv != nil {
		d.Set("policy_tool_version", *ptv)
	}

	var vcsRepo []interface{}
	if vcs := attrs.GetVcsRepo(); vcs != nil {
		vcsConfig := map[string]interface{}{
			"identifier":                 valueOrZero(vcs.GetIdentifier()),
			"ingress_submodules":         valueOrZero(vcs.GetIngressSubmodules()),
			"oauth_token_id":             valueOrZero(vcs.GetOauthTokenId()),
			"github_app_installation_id": valueOrZero(vcs.GetGithubAppInstallationId()),
		}
		if v, ok := d.GetOk("vcs_repo"); ok {
			if vcsRepo, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
					vcsConfig["branch"] = valueOrZero(vcs.GetBranch())
				}
			}
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	d.Set("vcs_repo", vcsRepo)

	var policyIDs []interface{}
	if rels != nil && rels.GetPolicies() != nil {
		for _, p := range rels.GetPolicies().GetData() {
			policyIDs = append(policyIDs, valueOrZero(p.GetId()))
		}
	}
	d.Set("policy_ids", policyIDs)

	var workspaceIDs []interface{}
	if !valueOrZero(attrs.GetGlobal()) {
		if rels != nil && rels.GetWorkspaces() != nil {
			for _, ws := range rels.GetWorkspaces().GetData() {
				workspaceIDs = append(workspaceIDs, valueOrZero(ws.GetId()))
			}
		}
	}
	d.Set("workspace_ids", workspaceIDs)

	err = helpers.WriteTFEIdentity(d, d.Id(), config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFEPolicySetUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name := d.Get("name").(string)
	global := d.Get("global").(bool)

	if global && d.HasChange("global") {
		oldWorkspaceIDs, _ := d.GetChange("workspace_ids")
		if oldWorkspaceIDs.(*schema.Set).Len() > 0 {
			body := makeWorkspaceIdentifierArrayDocument(oldWorkspaceIDs.(*schema.Set).List())
			log.Printf("[DEBUG] Removing previous workspaces from now-global policy set: %s", d.Id())
			err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Relationships().Workspaces().Delete(ctx, body, nil)
			if err != nil {
				return fmt.Errorf("Error detaching policy set %s from workspaces: %w", d.Id(), err)
			}
		}
	}

	fields := []string{
		"name", "description", "global", "vcs_repo",
		"overridable", "agent_enabled", "policy_tool_version",
		"policy_update_patterns",
	}
	hasAnyChange := false
	for _, field := range fields {
		if d.HasChange(field) {
			hasAnyChange = true
			break
		}
	}

	if hasAnyChange {
		attrs := models.NewPolicySets_attributes()
		attrs.SetName(ptr(name))
		attrs.SetGlobal(ptr(global))

		if desc, ok := d.GetOk("description"); ok {
			attrs.SetDescription(ptr(desc.(string)))
		}

		if d.HasChange("overridable") {
			attrs.SetOverridable(ptr(d.Get("overridable").(bool)))
		}

		if d.HasChange("agent_enabled") {
			attrs.SetAgentEnabled(ptr(d.Get("agent_enabled").(bool)))
		}

		if policyToolVersion, ok := d.GetOk("policy_tool_version"); ok {
			attrs.SetPolicyToolVersion(ptr(policyToolVersion.(string)))
		}

		if d.HasChange("policy_update_patterns") {
			if vPolicyUpdatePatterns, ok := d.GetOk("policy_update_patterns"); ok {
				var patterns []string
				for _, p := range vPolicyUpdatePatterns.([]interface{}) {
					patterns = append(patterns, p.(string))
				}
				attrs.SetPolicyUpdatePatterns(patterns)
			} else {
				attrs.SetPolicyUpdatePatterns([]string{})
			}
			if d.GetRawConfig().GetAttr("policy_update_patterns").IsNull() {
				attrs.SetPolicyUpdatePatterns(nil)
			}
		}

		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})
			vcsAttrs := models.NewPolicySets_attributes_vcsRepo()
			vcsAttrs.SetIdentifier(ptr(vcsRepo["identifier"].(string)))
			vcsAttrs.SetBranch(ptr(vcsRepo["branch"].(string)))
			vcsAttrs.SetIngressSubmodules(ptr(vcsRepo["ingress_submodules"].(bool)))
			vcsAttrs.SetOauthTokenId(ptr(vcsRepo["oauth_token_id"].(string)))
			vcsAttrs.SetGithubAppInstallationId(ptr(vcsRepo["github_app_installation_id"].(string)))
			attrs.SetVcsRepo(vcsAttrs)
		}

		body := models.NewPolicySets()
		body.SetAttributes(attrs)
		env := models.NewPolicySetsEnvelope()
		env.SetData(body)

		log.Printf("[DEBUG] Update configuration for policy set: %s", d.Id())
		_, err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Patch(ctx, env, nil)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for policy set %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("policy_ids") {
		oldSet, newSet := d.GetChange("policy_ids")
		oldPolicyIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newPolicyIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		if newPolicyIDs.Len() > 0 {
			body := makePolicyIdentifierArrayDocument(newPolicyIDs.List())
			log.Printf("[DEBUG] Add policies to policy set: %s", d.Id())
			err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Relationships().Policies().Post(ctx, body, nil)
			if err != nil {
				return fmt.Errorf("Error adding policies to policy set %s: %w", d.Id(), err)
			}
		}

		if oldPolicyIDs.Len() > 0 {
			body := makePolicyIdentifierArrayDocument(oldPolicyIDs.List())
			log.Printf("[DEBUG] Remove policies from policy set: %s", d.Id())
			err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Relationships().Policies().Delete(ctx, body, nil)
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

		if newWorkspaceIDs.Len() > 0 {
			body := makeWorkspaceIdentifierArrayDocument(newWorkspaceIDs.List())
			log.Printf("[DEBUG] Attach policy set to workspaces: %s", d.Id())
			err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Relationships().Workspaces().Post(ctx, body, nil)
			if err != nil {
				return fmt.Errorf("Error attaching policy set %s to workspaces: %w", d.Id(), err)
			}
		}

		if oldWorkspaceIDs.Len() > 0 {
			body := makeWorkspaceIdentifierArrayDocument(oldWorkspaceIDs.List())
			log.Printf("[DEBUG] Detach policy set from workspaces: %s", d.Id())
			err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Relationships().Workspaces().Delete(ctx, body, nil)
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
	err := config.ClientV2.API.PolicySets().ByPolicy_set_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting policy set %s: %w", d.Id(), err)
	}

	return nil
}

func makeWorkspaceIdentifierArrayDocument(ids []interface{}) *models.WorkspacesIdentifierArrayDocument {
	body := models.NewWorkspacesIdentifierArrayDocument()
	var data []models.WorkspacesIdentifierArrayDocument_dataable
	for _, id := range ids {
		item := models.NewWorkspacesIdentifierArrayDocument_data()
		item.SetId(ptr(id.(string)))
		data = append(data, item)
	}
	body.SetData(data)
	return body
}

func makePolicyIdentifierArrayDocument(ids []interface{}) *models.PoliciesIdentifierArrayDocument {
	body := models.NewPoliciesIdentifierArrayDocument()
	var data []models.PoliciesIdentifierArrayDocument_dataable
	for _, id := range ids {
		item := models.NewPoliciesIdentifierArrayDocument_data()
		item.SetId(ptr(id.(string)))
		data = append(data, item)
	}
	body.SetData(data)
	return body
}

func makeProjectIdentifierArrayDocument(ids []interface{}) *models.ProjectsIdentifierArrayDocument {
	body := models.NewProjectsIdentifierArrayDocument()
	var data []models.ProjectsIdentifierArrayDocument_dataable
	for _, id := range ids {
		item := models.NewProjectsIdentifierArrayDocument_data()
		item.SetId(ptr(id.(string)))
		data = append(data, item)
	}
	body.SetData(data)
	return body
}

func makeTagSelectorPostBody(key string, value *string, isExclude bool) policysets.ItemTagSelectorsPostRequestBodyable {
	item := policysets.NewItemTagSelectorsPostRequestBody_data()
	item.SetTagKey(&key)
	item.SetTagValue(value)
	item.SetIsExclude(&isExclude)
	body := policysets.NewItemTagSelectorsPostRequestBody()
	body.SetData([]policysets.ItemTagSelectorsPostRequestBody_dataable{item})
	return body
}

func makeTagSelectorDeleteBody(key string, value *string, isExclude bool) policysets.ItemTagSelectorsDeleteRequestBodyable {
	item := policysets.NewItemTagSelectorsDeleteRequestBody_data()
	item.SetTagKey(&key)
	item.SetTagValue(value)
	item.SetIsExclude(&isExclude)
	body := policysets.NewItemTagSelectorsDeleteRequestBody()
	body.SetData([]policysets.ItemTagSelectorsDeleteRequestBody_dataable{item})
	return body
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
