// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	tfev2 "github.com/hashicorp/go-tfe/v2"
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
		Description: "This ephemeral resource can be used to retrieve state outputs for a given workspace. It enables output values in one Terraform configuration to be used in another. The retrieved output values are guaranteed not to be written to state." +
			"\n\n~> **Warning:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral)." +
			"\n\n~> **Note:** Regardless of sensitivity of the output values as set in HCP Terraform, this ephemeral resource treats both `values` and `nonsensitive_values` as sensitive.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "System-generated unique identifier for the resource. Do not rely on this value.",
				Computed:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Name of the organization. If omitted, the organization must be defined in the provider config.",
				Optional:            true,
				Computed:            true,
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Name of the workspace.",
				Required:            true,
			},
			"values": schema.DynamicAttribute{
				MarkdownDescription: "The current output values for the specified workspace.",
				Computed:            true,
			},
			"nonsensitive_values": schema.DynamicAttribute{
				MarkdownDescription: "The current non-sensitive output values for the specified workspace, this is a subset of all output values.",
				Computed:            true,
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

	wsName := config.Workspace.ValueString()
	api := e.config.ClientV2.API

	log.Printf("[DEBUG] Reading the workspace %s in organization %s", wsName, config.Organization.ValueString())

	// Resolve workspace name to external ID.
	wsResp, err := api.Organizations().ByOrganization_name(orgName).Workspaces().ByWorkspace_name(wsName).Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workspace", err.Error())
		return
	}
	wsID := valueOrZero(wsResp.GetData().GetId())

	// Fetch current state version outputs for the workspace.
	outputsResp, err := api.Workspaces().ByWorkspace_id(wsID).CurrentStateVersionOutputs().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workspace outputs", err.Error())
		return
	}

	sensitiveTypes := map[string]attr.Type{}
	sensitiveValues := map[string]attr.Value{}
	nonSensitiveTypes := map[string]attr.Type{}
	nonSensitiveValues := map[string]attr.Value{}

	for _, op := range outputsResp.GetData() {
		attrs := op.GetAttributes()
		if attrs == nil {
			continue
		}
		opName := valueOrZero(attrs.GetName())
		opSensitive := valueOrZero(attrs.GetSensitive())
		opID := valueOrZero(op.GetId())

		// The value field is not modeled in the spec and lives in additionalData.
		var rawValue interface{}
		if ad := attrs.GetAdditionalData(); ad != nil {
			rawValue = ad["value"]
		}

		if opSensitive {
			// An additional API call is required to read sensitive output values.
			svResp, svErr := api.StateVersionOutputs().ByState_version_output_id(opID).Get(ctx, nil)
			if svErr != nil && errors.Is(svErr, tfev2.ErrNotFound) {
				continue
			}
			if svErr != nil {
				resp.Diagnostics.AddError("Unable to read resource", svErr.Error())
				return
			}
			if svData := svResp.GetData(); svData != nil && svData.GetAttributes() != nil {
				rawValue = svData.GetAttributes().GetAdditionalData()["value"]
			}
		}

		attrType, err := inferAttrType(rawValue)
		if err != nil {
			resp.Diagnostics.AddError("Error inferring attribute type", err.Error())
			return
		}

		attrValue, diags := convertToAttrValue(rawValue, attrType)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sensitiveTypes[opName] = attrType
		sensitiveValues[opName] = attrValue

		if !opSensitive {
			nonSensitiveTypes[opName] = attrType
			nonSensitiveValues[opName] = attrValue
		}
	}

	// Create dynamic attribute value for `values`
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

	diags.Append(resp.Result.Set(ctx, modelFromOutputs(orgName, wsName, sensitiveOutputs, nonSensitiveOutputs))...)
}
