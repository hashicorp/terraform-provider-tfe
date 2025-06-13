// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
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

// modelTFEStack maps the resource or data source schema data to a struct.
type modelTFEStack struct {
	ID              types.String          `tfsdk:"id"`
	ProjectID       types.String          `tfsdk:"project_id"`
	Name            types.String          `tfsdk:"name"`
	Description     types.String          `tfsdk:"description"`
	DeploymentNames types.Set             `tfsdk:"deployment_names"`
	VCSRepo         *modelTFEStackVCSRepo `tfsdk:"vcs_repo"`
	CreatedAt       types.String          `tfsdk:"created_at"`
	UpdatedAt       types.String          `tfsdk:"updated_at"`
}

// modelFromTFEStack builds a modelTFEStack struct from a tfe.Stack value.
func modelFromTFEStack(v *tfe.Stack) modelTFEStack {
	names := make([]attr.Value, len(v.DeploymentNames))
	for i, name := range v.DeploymentNames {
		names[i] = types.StringValue(name)
	}

	result := modelTFEStack{
		ID:              types.StringValue(v.ID),
		ProjectID:       types.StringValue(v.Project.ID),
		Name:            types.StringValue(v.Name),
		Description:     types.StringNull(),
		DeploymentNames: types.SetValueMust(types.StringType, names),
		CreatedAt:       types.StringValue(v.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:       types.StringValue(v.UpdatedAt.Format(time.RFC3339)),
	}

	if v.Description != "" {
		result.Description = types.StringValue(v.Description)
	}

	if v.VCSRepo != nil {
		result.VCSRepo = &modelTFEStackVCSRepo{
			Identifier: types.StringValue(v.VCSRepo.Identifier),
		}

		if v.VCSRepo.GHAInstallationID != "" {
			result.VCSRepo.GHAInstallationID = types.StringValue(v.VCSRepo.GHAInstallationID)
		} else {
			result.VCSRepo.GHAInstallationID = types.StringNull()
		}

		if v.VCSRepo.OAuthTokenID != "" {
			result.VCSRepo.OAuthTokenID = types.StringValue(v.VCSRepo.OAuthTokenID)
		} else {
			result.VCSRepo.OAuthTokenID = types.StringNull()
		}

		if v.VCSRepo.Branch != "" {
			result.VCSRepo.Branch = types.StringValue(v.VCSRepo.Branch)
		} else {
			result.VCSRepo.Branch = types.StringNull()
		}
	}

	return result
}
