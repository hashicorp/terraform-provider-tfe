// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type modelTFEWorkspaceRunTaskV0 struct {
	ID               types.String `tfsdk:"id"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
	TaskID           types.String `tfsdk:"task_id"`
	EnforcementLevel types.String `tfsdk:"enforcement_level"`
	Stage            types.String `tfsdk:"stage"`
}

var resourceWorkspaceRunTaskSchemaV0 = schema.Schema{
	Version: 0,
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Service-generated identifier for the workspace task",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"workspace_id": schema.StringAttribute{
			Description: "The id of the workspace to associate the Run task to.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"task_id": schema.StringAttribute{
			Description: "The id of the Run task to associate to the Workspace.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"enforcement_level": schema.StringAttribute{
			Description: fmt.Sprintf("The enforcement level of the task. Valid values are %s.", sentenceList(
				workspaceRunTaskEnforcementLevels(),
				"`",
				"`",
				"and",
			)),
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(workspaceRunTaskEnforcementLevels()...),
			},
		},
		"stage": schema.StringAttribute{
			Description: fmt.Sprintf("The stage to run the task in. Valid values are %s.", sentenceList(
				workspaceRunTaskStages(),
				"`",
				"`",
				"and",
			)),
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(string(tfe.PostPlan)),
			Validators: []validator.String{
				stringvalidator.OneOf(workspaceRunTaskStages()...),
			},
		},
	},
}

type modelTFEWorkspaceRunTaskV1 struct {
	ID               types.String `tfsdk:"id"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
	TaskID           types.String `tfsdk:"task_id"`
	EnforcementLevel types.String `tfsdk:"enforcement_level"`
	Stage            types.String `tfsdk:"stage"`
	Stages           types.List   `tfsdk:"stages"`
}

var resourceWorkspaceRunTaskSchemaV1 = schema.Schema{
	Version: 1,
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Service-generated identifier for the workspace task",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"workspace_id": schema.StringAttribute{
			Description: "The id of the workspace to associate the Run task to.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"task_id": schema.StringAttribute{
			Description: "The id of the Run task to associate to the Workspace.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"enforcement_level": schema.StringAttribute{
			Description: fmt.Sprintf("The enforcement level of the task. Valid values are %s.", sentenceList(
				workspaceRunTaskEnforcementLevels(),
				"`",
				"`",
				"and",
			)),
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(workspaceRunTaskEnforcementLevels()...),
			},
		},
		"stage": schema.StringAttribute{
			DeprecationMessage: "stage is deprecated, please use stages instead",
			Description: fmt.Sprintf("The stage to run the task in. Valid values are %s.", sentenceList(
				workspaceRunTaskStages(),
				"`",
				"`",
				"and",
			)),
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(string(tfe.PostPlan)),
			Validators: []validator.String{
				stringvalidator.OneOf(workspaceRunTaskStages()...),
			},
		},
		"stages": schema.ListAttribute{
			ElementType: types.StringType,
			Description: fmt.Sprintf("The stages to run the task in. Valid values are %s.", sentenceList(
				workspaceRunTaskStages(),
				"`",
				"`",
				"and",
			)),
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.UniqueValues(),
				listvalidator.ValueStringsAre(
					stringvalidator.OneOf(workspaceRunTaskStages()...),
				),
			},
			Optional: true,
			Computed: true,
		},
	},
}
