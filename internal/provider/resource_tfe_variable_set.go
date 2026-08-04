// Copyright IBM Corp. 2018, 2026
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
	"regexp"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	tfev2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

var variableSetIDRegexp = regexp.MustCompile("varset-[a-zA-Z0-9]{16}$")

func resourceTFEVariableSet() *schema.Resource {
	return &schema.Resource{
		Description: "Creates, updates and destroys variable sets.",

		Create: resourceTFEVariableSetCreate,
		Read:   resourceTFEVariableSetRead,
		Update: resourceTFEVariableSetUpdate,
		Delete: resourceTFEVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughWithIdentity("id"),
		},

		CustomizeDiff: func(c context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if err := customizeDiffIfProviderDefaultOrganizationChanged(c, d, meta); err != nil {
				return err
			}

			if err := validateParentProjectID(d); err != nil {
				return err
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
				Description: "The ID of the variable set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the variable set.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"description": {
				Description: "Description of the variable set.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"global": {
				Description:   "Whether the variable set applies to all workspaces in the organization. Defaults to `false`. Conflicts with `workspace_ids`.",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_ids"},
			},

			"priority": {
				Description: "When true, the variables in this set take priority over workspace-level variables and cannot be overridden. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"workspace_ids": {
				Description: "IDs of the workspaces that use the variable set. Must not be set if `global` is set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Deprecated:  "Use the `tfe_workspace_variable_set` resource instead, which is the preferred method of associating a variable set to a workspace.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"stack_ids": {
				Description: "IDs of the stacks that use the variable set.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"parent_project_id": {
				Description: "ID of the project that should own the variable set. If set, the value of `global` must be `false`. To assign whether a variable set should be applied to a project, use the `tfe_project_variable_set` resource.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Build the varset attributes.
	attrs := models.NewVarsets_attributes()
	attrs.SetName(ptr(name))
	attrs.SetGlobal(ptr(d.Get("global").(bool)))
	attrs.SetPriority(ptr(d.Get("priority").(bool)))
	if description, ok := d.GetOk("description"); ok {
		attrs.SetDescription(ptr(description.(string)))
	}

	vsType := models.VARSETS_VARSETS_TYPE
	vsData := models.NewVarsets()
	vsData.SetTypeEscaped(&vsType)
	vsData.SetAttributes(attrs)

	// Set parent project relationship if specified.
	if parentProject, ok := d.GetOk("parent_project_id"); ok {
		parentType := models.PROJECTS_VARSETPARENTIDENTIFIER_TYPE
		parentData := models.NewVarsetParentHasOne_data()
		parentData.SetId(ptr(parentProject.(string)))
		parentData.SetTypeEscaped(&parentType)

		parentRel := models.NewVarsetParentHasOne()
		parentRel.SetData(parentData)

		rels := models.NewVarsets_relationships()
		rels.SetParent(parentRel)
		vsData.SetRelationships(rels)
	}

	envelope := models.NewVarsetsEnvelope()
	envelope.SetData(vsData)

	log.Printf("[DEBUG] Create variable set %s for organization: %s", name, organization)
	resp, err := api.Organizations().ByOrganization_name(organization).Varsets().Post(ctx, envelope, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating variable set %s, for organization: %s: %w", name, organization, err)
	}
	if resp.GetData() == nil {
		return fmt.Errorf("API returned empty response when creating variable set %s", name)
	}

	vsID := valueOrZero(resp.GetData().GetId())
	d.SetId(vsID)

	// Apply workspace_ids if configured (deprecated field).
	if workspaceIDs, workspacesSet := d.GetOk("workspace_ids"); !d.Get("global").(bool) && workspacesSet {
		log.Printf("[DEBUG] Apply variable set %s to workspaces %v", name, workspaceIDs)
		warnWorkspaceIdsDeprecation()

		if err := v2ApplyVarsetToWorkspaces(api, vsID, workspaceIDs.(*schema.Set).List()); err != nil {
			return fmt.Errorf("Error applying variable set %s (%s) to given workspaces: %w", name, vsID, err)
		}
	}

	// Apply stack_ids if configured.
	if stackIDs, stacksSet := d.GetOk("stack_ids"); !d.Get("global").(bool) && stacksSet {
		log.Printf("[DEBUG] Apply variable set %s to stacks %v", name, stackIDs)

		if err := v2ApplyVarsetToStacks(api, vsID, stackIDs.(*schema.Set).List()); err != nil {
			return fmt.Errorf("Error applying variable set %s (%s) to given stacks: %w", name, vsID, err)
		}
	}

	err = helpers.WriteTFEIdentity(d, vsID, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	log.Printf("[DEBUG] Read configuration of variable set: %s", d.Id())

	resp, err := api.Varsets().ByVarset_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	if resp.GetData() == nil {
		d.SetId("")
		return nil
	}

	vsData := resp.GetData()
	attrs := vsData.GetAttributes()
	rels := vsData.GetRelationships()

	if attrs != nil {
		d.Set("name", valueOrZero(attrs.GetName()))
		d.Set("description", valueOrZero(attrs.GetDescription()))
		d.Set("global", valueOrZero(attrs.GetGlobal()))
		d.Set("priority", valueOrZero(attrs.GetPriority()))
	}

	setVariableSetRelationships(d, rels)

	err = helpers.WriteTFEIdentity(d, d.Id(), config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFEVariableSetUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("global") || d.HasChange("priority") {
		attrs := models.NewVarsets_attributes()
		attrs.SetName(ptr(d.Get("name").(string)))
		attrs.SetDescription(ptr(d.Get("description").(string)))
		attrs.SetGlobal(ptr(d.Get("global").(bool)))
		attrs.SetPriority(ptr(d.Get("priority").(bool)))

		vsType := models.VARSETS_VARSETS_TYPE
		vsData := models.NewVarsets()
		vsData.SetTypeEscaped(&vsType)
		vsData.SetAttributes(attrs)

		envelope := models.NewVarsetsEnvelope()
		envelope.SetData(vsData)

		log.Printf("[DEBUG] Update variable set: %s", d.Id())
		_, err := api.Varsets().ByVarset_id(d.Id()).Patch(ctx, envelope, nil)
		if err != nil {
			return fmt.Errorf("Error updating variable %s: %w", d.Id(), err)
		}
	}

	if d.HasChanges("workspace_ids") {
		if err := updateVarsetWorkspaceAssociations(api, d); err != nil {
			return err
		}
	}

	if d.HasChanges("stack_ids") {
		if err := updateVarsetStackAssociations(api, d); err != nil {
			return err
		}
	}

	return resourceTFEVariableSetRead(d, meta)
}

func resourceTFEVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	log.Printf("[DEBUG] Delete variable set: %s", d.Id())
	err := api.Varsets().ByVarset_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting variable set %s: %w", d.Id(), err)
	}

	return nil
}

// v2ApplyVarsetToWorkspaces adds the given workspace IDs to a variable set via
// POST /varsets/{id}/relationships/workspaces.
func v2ApplyVarsetToWorkspaces(api *tfev2api.ApiClient, vsID string, workspaceIDs []interface{}) error {
	wsType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	var wsItems []models.WorkspacesIdentifierArrayDocument_dataable
	for _, id := range workspaceIDs {
		if val, ok := id.(string); ok && val != "" {
			item := models.NewWorkspacesIdentifierArrayDocument_data()
			item.SetId(ptr(val))
			item.SetTypeEscaped(&wsType)
			wsItems = append(wsItems, item)
		}
	}
	if len(wsItems) == 0 {
		return nil
	}

	body := models.NewWorkspacesIdentifierArrayDocument()
	body.SetData(wsItems)
	return api.Varsets().ByVarset_id(vsID).Relationships().Workspaces().Post(ctx, body, nil)
}

// v2RemoveVarsetFromWorkspaces removes the given workspace IDs from a variable set via
// DELETE /varsets/{id}/relationships/workspaces.
func v2RemoveVarsetFromWorkspaces(api *tfev2api.ApiClient, vsID string, workspaceIDs []interface{}) error {
	wsType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	var wsItems []models.WorkspacesIdentifierArrayDocument_dataable
	for _, id := range workspaceIDs {
		if val, ok := id.(string); ok && val != "" {
			item := models.NewWorkspacesIdentifierArrayDocument_data()
			item.SetId(ptr(val))
			item.SetTypeEscaped(&wsType)
			wsItems = append(wsItems, item)
		}
	}
	if len(wsItems) == 0 {
		return nil
	}

	body := models.NewWorkspacesIdentifierArrayDocument()
	body.SetData(wsItems)
	return api.Varsets().ByVarset_id(vsID).Relationships().Workspaces().Delete(ctx, body, nil)
}

// v2ApplyVarsetToStacks adds the given stack IDs to a variable set via
// POST /varsets/{id}/relationships/stacks.
func v2ApplyVarsetToStacks(api *tfev2api.ApiClient, vsID string, stackIDs []interface{}) error {
	var stackItems []models.JsonapiResourceIdentifierable
	for _, id := range stackIDs {
		if val, ok := id.(string); ok && val != "" {
			item := models.NewJsonapiResourceIdentifier()
			item.SetId(ptr(val))
			item.SetTypeEscaped(ptr("stacks"))
			stackItems = append(stackItems, item)
		}
	}
	if len(stackItems) == 0 {
		return nil
	}

	body := models.NewJsonapiIdentifierArrayDocument()
	body.SetData(stackItems)
	return api.Varsets().ByVarset_id(vsID).Relationships().Stacks().Post(ctx, body, nil)
}

// v2RemoveVarsetFromStacks removes the given stack IDs from a variable set via
// DELETE /varsets/{id}/relationships/stacks.
func v2RemoveVarsetFromStacks(api *tfev2api.ApiClient, vsID string, stackIDs []interface{}) error {
	var stackItems []models.JsonapiResourceIdentifierable
	for _, id := range stackIDs {
		if val, ok := id.(string); ok && val != "" {
			item := models.NewJsonapiResourceIdentifier()
			item.SetId(ptr(val))
			item.SetTypeEscaped(ptr("stacks"))
			stackItems = append(stackItems, item)
		}
	}
	if len(stackItems) == 0 {
		return nil
	}

	body := models.NewJsonapiIdentifierArrayDocument()
	body.SetData(stackItems)
	return api.Varsets().ByVarset_id(vsID).Relationships().Stacks().Delete(ctx, body, nil)
}

func warnWorkspaceIdsDeprecation() {
	log.Printf("[WARN] The workspace_ids field of tfe_variable_set is deprecated as of release 0.33.0 and may be removed in a future version. The preferred method of associating a variable set to a workspace is by using the tfe_workspace_variable_set resource.")
}

func validateParentProjectID(d *schema.ResourceDiff) error {
	_, ok := d.GetOk("parent_project_id")
	if !ok {
		return nil
	}

	// If parent_project_id is set, global must be false
	if global, ok := d.GetOk("global"); ok {
		if global.(bool) {
			return fmt.Errorf("global must be 'false' when setting parent_project_id")
		}
	}

	return nil
}

// setVariableSetRelationships writes the organization, workspace_ids,
// stack_ids, and parent_project_id fields to state from a varsets relationships
// object. It is a no-op when rels is nil.
func setVariableSetRelationships(d *schema.ResourceData, rels models.Varsets_relationshipsable) {
	if rels == nil {
		return
	}

	// Organization name (in TFE, organization ID == organization name).
	if orgRel := rels.GetOrganization(); orgRel != nil && orgRel.GetData() != nil {
		d.Set("organization", valueOrZero(orgRel.GetData().GetId()))
	}

	// Workspace IDs (deprecated workspace_ids field).
	var wids []interface{}
	if wsRel := rels.GetWorkspaces(); wsRel != nil {
		for _, ws := range wsRel.GetData() {
			wids = append(wids, valueOrZero(ws.GetId()))
		}
	}
	d.Set("workspace_ids", wids)

	// Stack IDs.
	var sids []interface{}
	if stkRel := rels.GetStacks(); stkRel != nil {
		for _, s := range stkRel.GetData() {
			sids = append(sids, valueOrZero(s.GetId()))
		}
	}
	d.Set("stack_ids", sids)

	// Parent project ID: present when parent relationship type is "projects".
	if parent := rels.GetParent(); parent != nil && parent.GetData() != nil {
		parentData := parent.GetData()
		if pt := parentData.GetTypeEscaped(); pt != nil && pt.String() == "projects" {
			d.Set("parent_project_id", valueOrZero(parentData.GetId()))
		}
	}
}

// updateVarsetWorkspaceAssociations applies/removes workspace associations for
// the workspace_ids change set. Separated to reduce nestif complexity.
func updateVarsetWorkspaceAssociations(api *tfev2api.ApiClient, d *schema.ResourceData) error {
	oldVal, newVal := d.GetChange("workspace_ids")
	oldSet := oldVal.(*schema.Set)
	newSet := newVal.(*schema.Set)

	toAdd := newSet.Difference(oldSet).List()
	toRemove := oldSet.Difference(newSet).List()

	warnWorkspaceIdsDeprecation()
	log.Printf("[DEBUG] Apply variable set %s to workspaces %v", d.Id(), newSet.List())

	if len(toAdd) > 0 {
		if err := v2ApplyVarsetToWorkspaces(api, d.Id(), toAdd); err != nil {
			return fmt.Errorf("Error applying variable set %s to given workspaces: %w", d.Id(), err)
		}
	}
	if len(toRemove) > 0 {
		if err := v2RemoveVarsetFromWorkspaces(api, d.Id(), toRemove); err != nil {
			return fmt.Errorf("Error removing variable set %s from workspaces: %w", d.Id(), err)
		}
	}
	return nil
}

// updateVarsetStackAssociations applies/removes stack associations for the
// stack_ids change set. Separated to reduce nestif complexity.
func updateVarsetStackAssociations(api *tfev2api.ApiClient, d *schema.ResourceData) error {
	oldVal, newVal := d.GetChange("stack_ids")
	oldSet := oldVal.(*schema.Set)
	newSet := newVal.(*schema.Set)

	toAdd := newSet.Difference(oldSet).List()
	toRemove := oldSet.Difference(newSet).List()

	log.Printf("[DEBUG] Apply variable set %s to stacks %v", d.Id(), newSet.List())
	if len(toAdd) > 0 {
		if err := v2ApplyVarsetToStacks(api, d.Id(), toAdd); err != nil {
			return fmt.Errorf("Error applying variable set %s to given stacks: %w", d.Id(), err)
		}
	}
	if len(toRemove) > 0 {
		if err := v2RemoveVarsetFromStacks(api, d.Id(), toRemove); err != nil {
			return fmt.Errorf("Error removing variable set %s from stacks: %w", d.Id(), err)
		}
	}
	return nil
}
