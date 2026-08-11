// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
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

// modelFromTFEStackV2 builds a modelTFEStack struct from a v2 Stacksable
// value. trigger_patterns, creation_source, and vcs_repo.github_app_installation_id
// are left at their zero value: the pinned go-tfe/v2 generated client has no
// getter for any of the three (go-tfe/v2 gap), so callers must backfill them
// via fillStackV1OnlyFields.
func modelFromTFEStackV2(v models.Stacksable) modelTFEStack {
	result := modelTFEStack{
		ID:               types.StringValue(valueOrZero(v.GetId())),
		AgentPoolID:      types.StringNull(),
		Migration:        types.BoolNull(),
		CreationSource:   types.StringNull(),
		Description:      types.StringNull(),
		WorkingDirectory: types.StringNull(),
		TriggerPatterns:  types.ListNull(types.StringType),
	}

	if attrs := v.GetAttributes(); attrs != nil {
		applyStackAttributesV2(&result, attrs)
	}

	if rel := v.GetRelationships(); rel != nil {
		applyStackRelationshipsV2(&result, rel)
	}

	return result
}

// applyStackAttributesV2 populates result's plain-attribute fields from a v2
// Stacks_attributesable value.
func applyStackAttributesV2(result *modelTFEStack, attrs models.Stacks_attributesable) {
	result.Name = types.StringValue(valueOrZero(attrs.GetName()))
	result.SpeculativeEnabled = types.BoolValue(valueOrZero(attrs.GetSpeculativeEnabled()))

	if createdAt := attrs.GetCreatedAt(); createdAt != nil {
		result.CreatedAt = types.StringValue(createdAt.Format(time.RFC3339))
	}
	if updatedAt := attrs.GetUpdatedAt(); updatedAt != nil {
		result.UpdatedAt = types.StringValue(updatedAt.Format(time.RFC3339))
	}
	if desc := valueOrZero(attrs.GetDescription()); desc != "" {
		result.Description = types.StringValue(desc)
	}
	if wd := valueOrZero(attrs.GetWorkingDirectory()); wd != "" {
		result.WorkingDirectory = types.StringValue(wd)
	}

	if vcs := attrs.GetVcsRepo(); vcs != nil {
		result.VCSRepo = &modelTFEStackVCSRepo{
			Identifier:        types.StringValue(valueOrZero(vcs.GetIdentifier())),
			Branch:            types.StringNull(),
			GHAInstallationID: types.StringNull(),
			OAuthTokenID:      types.StringNull(),
		}
		if branch := valueOrZero(vcs.GetBranch()); branch != "" {
			result.VCSRepo.Branch = types.StringValue(branch)
		}
		if oauthTokenID := valueOrZero(vcs.GetOauthTokenId()); oauthTokenID != "" {
			result.VCSRepo.OAuthTokenID = types.StringValue(oauthTokenID)
		}
	}
}

// applyStackRelationshipsV2 populates result's project_id and agent_pool_id
// fields from a v2 Stacks_relationshipsable value.
func applyStackRelationshipsV2(result *modelTFEStack, rel models.Stacks_relationshipsable) {
	if proj := rel.GetProject(); proj != nil && proj.GetData() != nil {
		result.ProjectID = types.StringValue(valueOrZero(proj.GetData().GetId()))
	}
	if ap := rel.GetAgentPool(); ap != nil && ap.GetData() != nil {
		result.AgentPoolID = types.StringValue(valueOrZero(ap.GetData().GetId()))
	}
}

// fillStackV1OnlyFields backfills trigger_patterns, creation_source, and
// vcs_repo.github_app_installation_id onto result from a v1 tfe.Stack read of
// the same stack. See modelFromTFEStackV2's doc comment for why this is
// necessary.
func fillStackV1OnlyFields(result *modelTFEStack, v *tfe.Stack) {
	result.TriggerPatterns = triggerPatternsToList(v.TriggerPatterns)

	if v.CreationSource != "" {
		result.CreationSource = types.StringValue(v.CreationSource)
	}

	if result.VCSRepo != nil && v.VCSRepo != nil && v.VCSRepo.GHAInstallationID != "" {
		result.VCSRepo.GHAInstallationID = types.StringValue(v.VCSRepo.GHAInstallationID)
	}
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
