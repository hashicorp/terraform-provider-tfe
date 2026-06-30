// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource = &ProviderSetListResource{}
)

type ProviderSetListResource struct {
	config ConfiguredClient
}

type ProviderSetListResourceModel struct {
	OrganizationName types.String `tfsdk:"organization_name"`
}

func NewProviderSetListResource() list.ListResource {
	return &ProviderSetListResource{}
}

func (r *ProviderSetListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"organization_name": listschema.StringAttribute{
				Description: "Name of the organization to list things in.",
				Required:    true,
			},
		},
	}
}

func (r *ProviderSetListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_set"
}

func (r *ProviderSetListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	r.config = client
}

func (r *ProviderSetListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ProviderSetListResourceModel

	// Read list config data into the model
	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	options := &tfe.ProviderSetListOptions{}

	// Not checking includeResource, because there is not API support for the difference
	pl, err := r.config.Client.ProviderSets.List(ctx, data.OrganizationName.ValueString(), options)
	if err != nil {
		var errorDiags diag.Diagnostics
		errorDiags.AddError(
			"Error Retrieving Provider Sets",
			fmt.Sprintf("Could not list provider sets for organization %s: %s",
				data.OrganizationName.ValueString(), err.Error()),
		)
		stream.Results = list.ListResultsStreamDiagnostics(errorDiags)
		return
	}

	// Define the function that will push results into the stream
	stream.Results = func(push func(list.ListResult) bool) {
		for _, providerSet := range pl.Items {
			// Initialize a new result object for each thing
			result := req.NewListResult(ctx)

			// Set the user-friendly name of this thing
			result.DisplayName = providerSet.Name

			// Set resource identity data on the result
			identity := TFEProviderSetIdentityModel{
				ID: types.StringValue(providerSet.ID),
			}
			result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

			// Only set full resource data when Terraform requested it.
			if req.IncludeResource {
				resourceModel, diags := modelFromTFEProviderSet(ctx, *providerSet, types.Int64Null())
				if diags.HasError() {
					stream.Results = list.ListResultsStreamDiagnostics(diags)
					return
				}
				result.Diagnostics.Append(result.Resource.Set(ctx, resourceModel)...)
			}

			// Send the result to the stream.
			if !push(result) {
				return
			}
		}
	}
}
