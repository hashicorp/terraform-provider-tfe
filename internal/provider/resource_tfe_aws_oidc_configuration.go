// // Copyright IBM Corp. 2018, 2025
// // SPDX-License-Identifier: MPL-2.0

// NOTE: This resource is intentionally NOT migrated to go-tfe v2.
//
// All CRUD operations for OIDC configurations (AWS, GCP, Azure, Vault) share a
// single generated v2 endpoint, /oidc-configurations/{id}, whose response is a
// JSON:API composed type (OidcConfigurationEnvelope's data is one of
// AwsOidcConfigurations/AzureOidcConfigurations/GcpOidcConfigurations/
// VaultOidcConfigurations). The generated discriminator function for that
// composed type calls parseNode.GetChildNode("") with an empty discriminator
// key (the OpenAPI spec never declares one for this union), and
// kiota-serialization-json-go's GetChildNode unconditionally returns
// errors.New("index is empty") for an empty key. This means every Get/Post/
// Patch to this endpoint fails outright with that error -- confirmed
// empirically against a real go-tfe v2 client via an httptest server
// returning a well-formed aws-oidc-configurations payload -- not just a case
// of fields silently coming back nil. The endpoint also has no generated
// Delete method at all. Until upstream adds a discriminator mapping for this
// union (the same class of fix already applied to the organization-
// memberships "included" array in hashicorp/go-tfe@cae29a5b) and generates a
// Delete operation, this resource must stay on go-tfe v1 in its entirety.

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure   = &resourceTFEAWSOIDCConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFEAWSOIDCConfiguration{}
)

func NewAWSOIDCConfigurationResource() resource.Resource {
	return &resourceTFEAWSOIDCConfiguration{}
}

type resourceTFEAWSOIDCConfiguration struct {
	config ConfiguredClient
}

type modelTFEAWSOIDCConfiguration struct {
	ID           types.String `tfsdk:"id"`
	RoleARN      types.String `tfsdk:"role_arn"`
	Organization types.String `tfsdk:"organization"`
}

func (r *resourceTFEAWSOIDCConfiguration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
}

func (r *resourceTFEAWSOIDCConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_oidc_configuration"
}

func (r *resourceTFEAWSOIDCConfiguration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an AWS OIDC configuration.\n\n~> **Note:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the AWS OIDC configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_arn": schema.StringAttribute{
				Description: "The AWS ARN of your role.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceTFEAWSOIDCConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceTFEAWSOIDCConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan into the model
	var plan modelTFEAWSOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the organization name from resource or provider config
	var orgName string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &orgName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.AWSOIDCConfigurationCreateOptions{
		RoleARN: plan.RoleARN.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Create TFE AWS OIDC Configuration for organization %s", orgName))
	oidc, err := r.config.Client.AWSOIDCConfigurations.Create(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE AWS OIDC Configuration", err.Error())
		return
	}
	result := modelFromTFEAWSOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAWSOIDCConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state into the model
	var state modelTFEAWSOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read AWS OIDC configuration: %s", oidcID))
	oidc, err := r.config.Client.AWSOIDCConfigurations.Read(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("AWS OIDC configuration %s no longer exists", oidcID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading AWS OIDC configuration %s", oidcID),
			err.Error(),
		)
		return
	}
	result := modelFromTFEAWSOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAWSOIDCConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEAWSOIDCConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state modelTFEAWSOIDCConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.AWSOIDCConfigurationUpdateOptions{
		RoleARN: plan.RoleARN.ValueString(),
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE AWS OIDC Configuration %s", oidcID))
	oidc, err := r.config.Client.AWSOIDCConfigurations.Update(ctx, oidcID, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE AWS OIDC Configuration", err.Error())
		return
	}

	result := modelFromTFEAWSOIDCConfiguration(oidc)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEAWSOIDCConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEAWSOIDCConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE AWS OIDC configuration: %s", oidcID))
	err := r.config.Client.AWSOIDCConfigurations.Delete(ctx, oidcID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE AWS OIDC configuration %s no longer exists", oidcID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE AWS OIDC Configuration", err.Error())
		return
	}
}

func modelFromTFEAWSOIDCConfiguration(p *tfe.AWSOIDCConfiguration) modelTFEAWSOIDCConfiguration {
	return modelTFEAWSOIDCConfiguration{
		ID:           types.StringValue(p.ID),
		RoleARN:      types.StringValue(p.RoleARN),
		Organization: types.StringValue(p.Organization.Name),
	}
}
