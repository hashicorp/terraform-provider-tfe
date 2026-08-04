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
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	v2orgs "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceVariableSet() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes a workspace from a variable set's scope." +
			"\n\n~> **Note:** `tfe_variable_set` has a deprecated argument `workspace_ids` that should not be used alongside this resource. They manage the same attachments and are mutually exclusive.",

		Create: resourceTFEWorkspaceVariableSetCreate,
		Read:   resourceTFEWorkspaceVariableSetRead,
		Delete: resourceTFEWorkspaceVariableSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceVariableSetImporter,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the variable set attachment. ID format: `<workspace-id>_<variable-set-id>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"variable_set_id": {
				Description: "The variable set ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"workspace_id": {
				Description: "Workspace ID to add the variable set to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEWorkspaceVariableSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	vSID := d.Get("variable_set_id").(string)
	wID := d.Get("workspace_id").(string)

	wsType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	wsData := models.NewWorkspacesIdentifierArrayDocument_data()
	wsData.SetId(ptr(wID))
	wsData.SetTypeEscaped(&wsType)

	body := models.NewWorkspacesIdentifierArrayDocument()
	body.SetData([]models.WorkspacesIdentifierArrayDocument_dataable{wsData})

	err := api.Varsets().ByVarset_id(vSID).Relationships().Workspaces().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"Error applying variable set id %s to workspace %s: %w", vSID, wID, err)
	}

	id := encodeVariableSetWorkspaceAttachment(wID, vSID)
	d.SetId(id)

	return resourceTFEWorkspaceVariableSetRead(d, meta)
}

func resourceTFEWorkspaceVariableSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	wID := d.Get("workspace_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace variable set: %s", d.Id())
	resp, err := api.Varsets().ByVarset_id(vSID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Variable set %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of variable set %s: %w", d.Id(), err)
	}

	// Verify workspace listed in variable set
	check := false
	if data := resp.GetData(); data != nil {
		if rels := data.GetRelationships(); rels != nil {
			if wsRel := rels.GetWorkspaces(); wsRel != nil {
				for _, ws := range wsRel.GetData() {
					if valueOrZero(ws.GetId()) == wID {
						check = true
						d.Set("workspace_id", wID)
						break
					}
				}
			}
		}
	}
	if !check {
		log.Printf("[DEBUG] Workspace %s not attached to variable set %s. Removing from state.", wID, vSID)
		d.SetId("")
		return nil
	}

	d.Set("variable_set_id", vSID)
	return nil
}

func resourceTFEWorkspaceVariableSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	wID := d.Get("workspace_id").(string)
	vSID := d.Get("variable_set_id").(string)

	log.Printf("[DEBUG] Delete workspace (%s) from variable set (%s)", wID, vSID)

	wsType := models.WORKSPACES_WORKSPACESIDENTIFIERARRAYDOCUMENT_DATA_TYPE
	wsData := models.NewWorkspacesIdentifierArrayDocument_data()
	wsData.SetId(ptr(wID))
	wsData.SetTypeEscaped(&wsType)

	body := models.NewWorkspacesIdentifierArrayDocument()
	body.SetData([]models.WorkspacesIdentifierArrayDocument_dataable{wsData})

	err := api.Varsets().ByVarset_id(vSID).Relationships().Workspaces().Delete(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"Error removing workspace %s from variable set %s: %w", wID, vSID, err)
	}

	return nil
}

func encodeVariableSetWorkspaceAttachment(wID, vSID string) string {
	return fmt.Sprintf("%s_%s", wID, vSID)
}

func resourceTFEWorkspaceVariableSetImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/WORKSPACE NAME/VARSET NAME> but be aware
	// that variable set names can contain forward slash characters but organization/workspace
	// names cannot. Therefore, we split the import ID into at most 3 substrings.
	organization, wsName, vSName, err := destructureImportID(strings.SplitN(d.Id(), "/", 3))
	if err != nil {
		return nil, err
	}

	config := meta.(ConfiguredClient)
	api := config.ClientV2.API

	// Resolve the workspace name to its external ID.
	wsResp, err := api.Organizations().ByOrganization_name(organization).Workspaces().ByWorkspace_name(wsName).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of workspace %s in organization %s: %w", wsName, organization, err)
	}
	wsID := ""
	if wsResp.GetData() != nil {
		wsID = valueOrZero(wsResp.GetData().GetId())
	}
	if wsID == "" {
		return nil, fmt.Errorf("workspace %s in organization %s returned empty ID", wsName, organization)
	}

	// List varsets in the org, filtering by name, and find the one attached to this workspace.
	pageSize := int32(100)
	queryParams := &v2orgs.ItemVarsetsRequestBuilderGetQueryParameters{
		Q:        &vSName,
		Pagesize: &pageSize,
	}

	for {
		list, err := api.Organizations().ByOrganization_name(organization).Varsets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("Error retrieving variable sets: %w", err)
		}
		for _, vs := range list.GetData() {
			vsAttrs := vs.GetAttributes()
			if vsAttrs == nil || valueOrZero(vsAttrs.GetName()) != vSName {
				continue
			}

			vsID := valueOrZero(vs.GetId())

			// Fetch the full varset to get workspace relationship data.
			vsResp, err := api.Varsets().ByVarset_id(vsID).Get(ctx, nil)
			if err != nil {
				return nil, fmt.Errorf("Error reading variable set %s: %w", vsID, err)
			}
			if vsResp.GetData() == nil {
				continue
			}

			wsRel := vsResp.GetData().GetRelationships().GetWorkspaces()
			if wsRel == nil {
				continue
			}
			for _, ws := range wsRel.GetData() {
				if valueOrZero(ws.GetId()) != wsID {
					continue
				}

				d.Set("workspace_id", wsID)
				d.Set("variable_set_id", vsID)
				d.SetId(encodeVariableSetWorkspaceAttachment(wsID, vsID))

				return []*schema.ResourceData{d}, nil
			}
		}

		// Exit the loop when we've seen all pages.
		nextPage := nextPageNumber(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return nil, fmt.Errorf("workspace %s has not been assigned to variable set %s", wsName, vSName)
}

func destructureImportID(splitID []string) (string, string, string, error) {
	if len(splitID) != 3 {
		return "", "", "", fmt.Errorf(
			"invalid workspace variable set input format: %s (expected <ORGANIZATION><WORKSPACE NAME>/<VARIABLE SET NAME>)",
			splitID,
		)
	}

	return splitID[0], splitID[1], splitID[2], nil
}
