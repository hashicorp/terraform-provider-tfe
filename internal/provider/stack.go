// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type modelTFEStackVCSRepo struct {
	Identifier        types.String `tfsdk:"identifier"`
	Branch            types.String `tfsdk:"branch"`
	GHAInstallationID types.String `tfsdk:"github_app_installation_id"`
	OAuthTokenID      types.String `tfsdk:"oauth_token_id"`
}

// modelTFEStack maps the resource or data source schema data to a
// struct.
type modelTFEStack struct {
	ID                 types.String          `tfsdk:"id"`
	ProjectID          types.String          `tfsdk:"project_id"`
	AgentPoolID        types.String          `tfsdk:"agent_pool_id"`
	Name               types.String          `tfsdk:"name"`
	Migration          types.Bool            `tfsdk:"migration"`
	SpeculativeEnabled types.Bool            `tfsdk:"speculative_enabled"`
	CreationSource     types.String          `tfsdk:"creation_source"`
	Description        types.String          `tfsdk:"description"`
	WorkingDirectory   types.String          `tfsdk:"working_directory"`
	TriggerPatterns    types.List            `tfsdk:"trigger_patterns"`
	VCSRepo            *modelTFEStackVCSRepo `tfsdk:"vcs_repo"`
	CreatedAt          types.String          `tfsdk:"created_at"`
	UpdatedAt          types.String          `tfsdk:"updated_at"`
}

type modelTFEStackIdentity struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
}

// modelFromTFEStack builds a modelTFEStack struct from a
// tfe.Stack value.“
func modelFromTFEStack(v *tfe.Stack) modelTFEStack {
	triggerPatterns := triggerPatternsToList(v.TriggerPatterns)

	result := modelTFEStack{
		ID:                 types.StringValue(v.ID),
		ProjectID:          types.StringValue(v.Project.ID),
		AgentPoolID:        types.StringNull(),
		Name:               types.StringValue(v.Name),
		Migration:          types.BoolNull(),
		SpeculativeEnabled: types.BoolValue(v.SpeculativeEnabled),
		CreationSource:     types.StringNull(),
		Description:        types.StringNull(),
		WorkingDirectory:   types.StringNull(),
		TriggerPatterns:    triggerPatterns,
		CreatedAt:          types.StringValue(v.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:          types.StringValue(v.UpdatedAt.Format(time.RFC3339)),
	}

	if v.VCSRepo != nil {
		result.VCSRepo = &modelTFEStackVCSRepo{
			Identifier:        types.StringValue(v.VCSRepo.Identifier),
			Branch:            types.StringNull(),
			GHAInstallationID: types.StringNull(),
			OAuthTokenID:      types.StringNull(),
		}
	}

	if v.AgentPool != nil {
		result.AgentPoolID = types.StringValue(v.AgentPool.ID)
	}

	if v.Description != "" {
		result.Description = types.StringValue(v.Description)
	}

	if v.WorkingDirectory != "" {
		result.WorkingDirectory = types.StringValue(v.WorkingDirectory)
	}

	if v.VCSRepo != nil {
		if v.VCSRepo.GHAInstallationID != "" {
			result.VCSRepo.GHAInstallationID = types.StringValue(v.VCSRepo.GHAInstallationID)
		}

		if v.VCSRepo.OAuthTokenID != "" {
			result.VCSRepo.OAuthTokenID = types.StringValue(v.VCSRepo.OAuthTokenID)
		}

		if v.VCSRepo.Branch != "" {
			result.VCSRepo.Branch = types.StringValue(v.VCSRepo.Branch)
		}
	}

	if v.CreationSource != "" {
		result.CreationSource = types.StringValue(v.CreationSource)
	}

	return result
}

func triggerPatternsToList(patterns []string) types.List {
	if len(patterns) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, len(patterns))
	for i, p := range patterns {
		elems[i] = types.StringValue(p)
	}
	list, _ := types.ListValue(types.StringType, elems)
	return list
}

func triggerPatternsFromList(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := make([]types.String, 0, len(list.Elements()))
	list.ElementsAs(ctx, &elems, false)
	result := make([]string, len(elems))
	for i, e := range elems {
		result[i] = e.ValueString()
	}
	return result
}
