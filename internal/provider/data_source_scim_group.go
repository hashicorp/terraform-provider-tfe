// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// scimGroupNonWhitespaceRegex rejects whitespace-only inputs that slip past LengthAtLeast(1).
var scimGroupNonWhitespaceRegex = regexp.MustCompile(`\S`)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFESCIMGroup{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESCIMGroup{}
)

// NewSCIMGroupDataSource is a helper function to simplify the provider implementation.
func NewSCIMGroupDataSource() datasource.DataSource {
	return &dataSourceTFESCIMGroup{}
}

// dataSourceTFESCIMGroup is the data source implementation.
type dataSourceTFESCIMGroup struct {
	client *tfe.Client
}

// modelDataTFESCIMGroup maps the data source schema data.
type modelDataTFESCIMGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESCIMGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_group"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESCIMGroup) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads SCIM groups synchronized from the configured Identity Provider into Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the SCIM group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The exact name of the SCIM group to retrieve (case-insensitive).",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(
						scimGroupNonWhitespaceRegex,
						"must contain at least one non-whitespace character",
					),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESCIMGroup) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	d.client = client.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFESCIMGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFESCIMGroup
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If `name` is still unknown, defer to apply by returning without
	// writing state; the framework will mark all computed attributes as
	// unknown for downstream consumers.
	if data.Name.IsUnknown() {
		return
	}
	name := data.Name.ValueString()

	options := &tfe.AdminSCIMGroupListOptions{
		// ?q= is a fuzzy substring match used here only as a server-side
		// prefilter; we still narrow to an exact, case-insensitive match below.
		Query: name,
	}

	tflog.Debug(ctx, "Listing SCIM groups", map[string]any{
		"name": name,
	})

	// Keep the case-insensitive exact match (at most one) and stop
	// paginating as soon as we find it.
	var match *tfe.AdminSCIMGroup
	for {
		list, err := d.client.Admin.Settings.SCIM.Groups.List(ctx, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list SCIM groups", err.Error())
			return
		}

		if matched := filterExactSCIMGroups(list.Items, name); len(matched) > 0 {
			match = matched[0]
			break
		}
		if list.Pagination == nil || list.CurrentPage >= list.TotalPages {
			break
		}
		options.PageNumber = list.NextPage
	}

	if match == nil {
		resp.Diagnostics.AddError(
			"SCIM group not found",
			fmt.Sprintf("No SCIM group found with name %q.", name),
		)
		return
	}

	data.ID = types.StringValue(match.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
