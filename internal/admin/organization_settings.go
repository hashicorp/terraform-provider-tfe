package admin

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OrganizationSettingsResource struct {
	client *tfe.Client
}

func NewOrganizationSettings() resource.Resource {
	return OrganizationSettingsResource{}
}

type OrganizationSettingsResourceModel struct {
	AccessBetaTools     types.Bool   `tfsdk:"access_beta_tools"`
	GlobalModuleSharing types.Bool   `tfsdk:"global_module_sharing"`
	NameID              types.String `tfsdk:"name_id"`
	IsDisabled          types.Bool   `tfsdk:"is_disabled"`
	SSOEnabled          types.Bool   `tfsdk:"sso_enabled"`
	WorkspaceLimit      types.Int64  `tfsdk:"workspace_limit"`
}

func (r OrganizationSettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tfe_admin_organization_settings"
}

func (r OrganizationSettingsResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Admin options for a Terraform Enterprise Organization. Requires that the provider be configured with an admin token.",
		Attributes: map[string]tfsdk.Attribute{
			"access_beta_tools": {
				MarkdownDescription: "Whether or not this organization has access to beta versions of Terraform",
				Optional:            true,
				Type:                types.BoolType,
			},
			"global_module_sharing": {
				MarkdownDescription: "If true, modules in the organization's private module repository will be available to all other organizations in this TFE instance. Mutually exclusive with any specifically configured module consumers.",
				Optional:            true,
				Type:                types.BoolType,
			},
			"is_disabled": {
				MarkdownDescription: "If true, removes all permissions from the organization and makes it inaccessible to users.",
				Optional:            true,
				Type:                types.BoolType,
			},
			"name_id": {
				MarkdownDescription: "The name (id) of the organization",
				Required:            true,
				Type:                types.StringType,
			},
			"sso_enabled": {
				MarkdownDescription: "Whether or not SSO is enabled",
				Optional:            true,
				Type:                types.BoolType,
			},
			"workspace_limit": {
				MarkdownDescription: "Maximum number of workspaces for this organization. If this number is set to a value lower than the number of workspaces the organization has, it will prevent additional workspaces from being created, but existing workspaces will not be affected. If set to 0, this limit will have no effect.",
				Optional:            true,
				Type:                types.Int64Type,
			},
		},
	}, nil
}

func (r *OrganizationSettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tfe.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider resource data", fmt.Sprintf("Expected *tfe.Client, got %T", req.ProviderData))
		return
	}

	r.client = client
}

func (r OrganizationSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *OrganizationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	remoteData, adminOrg, diags := r.readInternal(ctx, data.NameID.String())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.updateInternal(ctx, *adminOrg, &remoteData)
}

func (r OrganizationSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r OrganizationSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *OrganizationSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	remoteData, _, diags := r.readInternal(ctx, data.NameID.String())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remoteData)...)
}

func (r OrganizationSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *OrganizationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	remoteData, adminOrg, diags := r.readInternal(ctx, data.NameID.String())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateInternal(ctx, *adminOrg, &remoteData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remoteData)...)
}

func (r OrganizationSettingsResource) updateInternal(ctx context.Context, org tfe.AdminOrganization, data *OrganizationSettingsResourceModel) diag.Diagnostics {
	_, err := r.client.Admin.Organizations.Update(ctx, org.Name, tfe.AdminOrganizationUpdateOptions{
		AccessBetaTools:     tfe.Bool(data.AccessBetaTools.ValueBool()),
		GlobalModuleSharing: tfe.Bool(data.GlobalModuleSharing.ValueBool()),
		IsDisabled:          tfe.Bool(data.IsDisabled.ValueBool()),
		WorkspaceLimit:      tfe.Int(int(data.WorkspaceLimit.ValueInt64())),
	})

	var result diag.Diagnostics
	if err != nil {
		result.AddError("Failed to update admin organization settings", fmt.Sprintf("Unable to update resource, got error: %s", err))
	}
	return result
}

func (r *OrganizationSettingsResource) readInternal(ctx context.Context, nameID string) (OrganizationSettingsResourceModel, *tfe.AdminOrganization, diag.Diagnostics) {
	var data OrganizationSettingsResourceModel
	var diags diag.Diagnostics
	adminOrg, err := r.client.Admin.Organizations.Read(ctx, nameID)

	if err != nil {
		diags.AddError("Failed to fetch organization", fmt.Sprintf("Unable to read resource, got error: %s", err))
		return data, adminOrg, diags
	}

	data.AccessBetaTools = types.BoolValue(adminOrg.AccessBetaTools)
	if adminOrg.GlobalModuleSharing != nil {
		data.GlobalModuleSharing = types.BoolValue(*adminOrg.GlobalModuleSharing)
	}
	data.IsDisabled = types.BoolValue(adminOrg.IsDisabled)
	data.NameID = types.StringValue(adminOrg.Name)
	data.SSOEnabled = types.BoolValue(adminOrg.SsoEnabled)

	if adminOrg.WorkspaceLimit != nil {
		data.WorkspaceLimit = types.Int64Value(int64(*adminOrg.WorkspaceLimit))
	}

	return data, adminOrg, diags
}
