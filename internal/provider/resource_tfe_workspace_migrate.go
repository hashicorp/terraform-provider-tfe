// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type modelWorkspaceV0 struct {
	ID                  types.String `tfsdk:"id"`
	ExternalID          types.String `tfsdk:"external_id"`
	Name                types.String `tfsdk:"name"`
	Organization        types.String `tfsdk:"organization"`
	AssessmentsEnabled  types.Bool   `tfsdk:"assessments_enabled"`
	AutoApply           types.Bool   `tfsdk:"auto_apply"`
	FileTriggersEnabled types.Bool   `tfsdk:"file_triggers_enabled"`
	Operations          types.Bool   `tfsdk:"operations"`
	QueueAllRuns        types.Bool   `tfsdk:"queue_all_runs"`
	SSHKeyID            types.String `tfsdk:"ssh_key_id"`
	TerraformVersion    types.String `tfsdk:"terraform_version"`
	TriggerPrefixes     types.List   `tfsdk:"trigger_prefixes"`
	WorkingDirectory    types.String `tfsdk:"working_directory"`
	VCSRepo             types.List   `tfsdk:"vcs_repo"`
}

var resourceTFEWorkspaceSchemaV0 = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true},
		"external_id":           schema.StringAttribute{Computed: true},
		"name":                  schema.StringAttribute{Required: true},
		"organization":          schema.StringAttribute{Required: true},
		"assessments_enabled":   schema.BoolAttribute{Optional: true},
		"auto_apply":            schema.BoolAttribute{Optional: true, Computed: true},
		"file_triggers_enabled": schema.BoolAttribute{Optional: true, Computed: true},
		"operations":            schema.BoolAttribute{Optional: true, Computed: true},
		"queue_all_runs":        schema.BoolAttribute{Optional: true, Computed: true},
		"ssh_key_id":            schema.StringAttribute{Optional: true, Computed: true},
		"terraform_version":     schema.StringAttribute{Optional: true, Computed: true},
		"trigger_prefixes":      schema.ListAttribute{Optional: true, ElementType: types.StringType},
		"working_directory":     schema.StringAttribute{Optional: true, Computed: true},
	},
	Blocks: map[string]schema.Block{
		"vcs_repo": schema.ListNestedBlock{
			NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
				"identifier":                 schema.StringAttribute{Required: true},
				"branch":                     schema.StringAttribute{Optional: true},
				"ingress_submodules":         schema.BoolAttribute{Optional: true, Computed: true},
				"oauth_token_id":             schema.StringAttribute{Required: true},
				"github_app_installation_id": schema.StringAttribute{Computed: true},
			}},
		},
	},
}

func resourceTfeWorkspaceStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("raw state is nil")
	}

	rawState["id"] = rawState["external_id"]
	return rawState, nil
}
