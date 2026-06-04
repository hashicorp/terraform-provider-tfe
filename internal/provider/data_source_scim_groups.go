// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

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
	_ datasource.DataSource              = &dataSourceTFESCIMGroups{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFESCIMGroups{}
)

// NewSCIMGroupsDataSource is a helper function to simplify the provider implementation.
func NewSCIMGroupsDataSource() datasource.DataSource {
	return &dataSourceTFESCIMGroups{}
}

// dataSourceTFESCIMGroups is the data source implementation.
type dataSourceTFESCIMGroups struct {
	client *tfe.Client
}

// modelDataTFESCIMGroup represents a single SCIM group entry in the groups list.
type modelDataTFESCIMGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// modelDataTFESCIMGroups maps the data source schema data.
type modelDataTFESCIMGroups struct {
	ID        types.String            `tfsdk:"id"`
	Name      types.String            `tfsdk:"name"`
	GroupID   types.String            `tfsdk:"group_id"`
	GroupName types.String            `tfsdk:"group_name"`
	Groups    []modelDataTFESCIMGroup `tfsdk:"groups"`
}

// Metadata returns the data source type name.
func (d *dataSourceTFESCIMGroups) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_groups"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFESCIMGroups) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads SCIM groups synchronized from the configured Identity Provider into Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal ID of the data source, formatted as `name/<value>`, where the `<value>` portion is URL-path-escaped before being stored in state (for example, spaces become `%20` and `/` becomes `%2F`; e.g., `name/admin-team` or `name/Platform%20Ops`).",
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
			"group_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the SCIM group. Only populated when exactly one matching group is found.",
			},
			"group_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the SCIM group. Only populated when exactly one matching group is found.",
			},
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of all matching SCIM groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the SCIM group.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the SCIM group.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFESCIMGroups) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dataSourceTFESCIMGroups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelDataTFESCIMGroups
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
	value := strings.TrimSpace(data.Name.ValueString())

	options := &tfe.AdminSCIMGroupListOptions{
		// ?q= is a fuzzy substring match used here only as a server-side
		// prefilter; we still narrow to an exact, case-insensitive match below.
		Query: value,
	}

	tflog.Debug(ctx, "Listing SCIM groups", map[string]any{
		"name": value,
	})

	// Keep only case-insensitive exact matches (at most one) and stop
	// paginating as soon as we find it.
	matched := make([]*tfe.AdminSCIMGroup, 0, 1)
	for {
		list, err := d.client.Admin.Settings.SCIM.Groups.List(ctx, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list SCIM groups", err.Error())
			return
		}

		matched = append(matched, filterExactSCIMGroups(list.Items, value)...)

		if len(matched) > 0 {
			break
		}
		if list.Pagination == nil || list.CurrentPage >= list.TotalPages {
			break
		}
		options.PageNumber = list.NextPage
	}

	// PathEscape so `/` or spaces in value don't break the id format.
	data.ID = types.StringValue(fmt.Sprintf("name/%s", url.PathEscape(value)))

	groups := make([]modelDataTFESCIMGroup, 0, len(matched))
	for _, g := range matched {
		groups = append(groups, modelDataTFESCIMGroup{
			ID:   types.StringValue(g.ID),
			Name: types.StringValue(g.Name),
		})
	}
	data.Groups = groups

	if len(matched) == 1 {
		data.GroupID = types.StringValue(matched[0].ID)
		data.GroupName = types.StringValue(matched[0].Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
