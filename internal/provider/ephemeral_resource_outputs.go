// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &outputsEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &outputsEphemeralResource{}
)

func NewOutputsEphemeralResource() ephemeral.EphemeralResource {
	return &outputsEphemeralResource{}
}

type outputsEphemeralResource struct {
	config ConfiguredClient
}

func (e *outputsEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve a workspace's state outputs without saving them in state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: `System-generated unique identifier for the resource.`,
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: `Name of the organization.`,
				Required:    true,
			},
			"workspace": schema.StringAttribute{
				Description: `Name of the workspace.`,
				Required:    true,
			},
			"values": schema.DynamicAttribute{
				Description: `Values of the workspace outputs.`,
				Computed:    true,
			},
			"nonsensitive_values": schema.DynamicAttribute{
				Description: `Non-sensitive values of the workspace outputs.`,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (e *outputsEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}

	e.config = client
}

func (e *outputsEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outputs"
}

func (e *outputsEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	config := outputsModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org name or default
	var orgName string
	resp.Diagnostics.Append(e.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Reading the workspace %s in organization %s", config.Workspace.ValueString(), config.Organization.ValueString())
	opts := &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSOutputs},
	}

	ws, err := e.config.Client.Workspaces.ReadWithOptions(ctx, config.Organization.ValueString(), config.Workspace.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}

	sensitiveTypes := map[string]attr.Type{}
	sensitiveValues := map[string]attr.Value{}
	nonSensitiveTypes := map[string]attr.Type{}
	nonSensitiveValues := map[string]attr.Value{}

	for _, op := range ws.Outputs {
		if op.Sensitive {
			// An additional API call is required to read sensitive output values.
			result, err := e.config.Client.StateVersionOutputs.Read(ctx, op.ID)
			if err != nil {
				resp.Diagnostics.AddError("Unable to read resource", err.Error())
				return
			}

			op.Value = result.Value
		}

		attrType, err := inferAttrType(op.Value)
		if err != nil {
			resp.Diagnostics.AddError("Error inferring attribute type", err.Error())
			return
		}

		attrValue, diags := convertToAttrValue(op.Value, attrType)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sensitiveTypes[op.Name] = attrType
		sensitiveValues[op.Name] = attrValue

		if !op.Sensitive {
			nonSensitiveTypes[op.Name] = attrType
			nonSensitiveValues[op.Name] = attrValue
		}
	}

	// Create dynamic attribute value for `sensitive_values`
	obj, diags := types.ObjectValue(sensitiveTypes, sensitiveValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sensitiveOutputs := types.DynamicValue(obj)

	// Create dynamic attribute value for `nonsensitive_values`
	obj, diags = types.ObjectValue(nonSensitiveTypes, nonSensitiveValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nonSensitiveOutputs := types.DynamicValue(obj)

	diags.Append(resp.Result.Set(ctx, modelFromOutputs(ws, sensitiveOutputs, nonSensitiveOutputs))...)
}
