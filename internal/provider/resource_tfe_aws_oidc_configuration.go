// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	roleARN := plan.RoleARN.ValueString()
	attrs := models.NewAwsOidcConfigurations_attributes()
	attrs.SetRoleArn(&roleARN)

	awsData := models.NewAwsOidcConfigurations()
	awsData.SetAttributes(attrs)
	awsType := models.AWSOIDCCONFIGURATIONS_AWSOIDCCONFIGURATIONS_TYPE
	awsData.SetTypeEscaped(&awsType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetAwsOidcConfigurations(awsData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE AWS OIDC Configuration for organization %s", orgName))
	oidcEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).OidcConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE AWS OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractAWSOIDCData(oidcEnvelope)
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
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
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

	oidc, err := extractAWSOIDCData(oidcEnvelope)
	if err != nil {
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

	oidcID := state.ID.ValueString()

	roleARN := plan.RoleARN.ValueString()
	attrs := models.NewAwsOidcConfigurations_attributes()
	attrs.SetRoleArn(&roleARN)

	awsData := models.NewAwsOidcConfigurations()
	awsData.SetAttributes(attrs)
	awsType := models.AWSOIDCCONFIGURATIONS_AWSOIDCCONFIGURATIONS_TYPE
	awsData.SetTypeEscaped(&awsType)

	inner := models.NewOidcConfigurationEnvelope_OidcConfigurationEnvelope_data()
	inner.SetAwsOidcConfigurations(awsData)

	envelope := models.NewOidcConfigurationEnvelope()
	envelope.SetData(inner)

	tflog.Debug(ctx, fmt.Sprintf("Update TFE AWS OIDC Configuration %s", oidcID))
	oidcEnvelope, err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE AWS OIDC Configuration", err.Error())
		return
	}

	oidc, err := extractAWSOIDCData(oidcEnvelope)
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
	err := r.config.ClientV2.API.OidcConfigurations().ByOidc_configuration_id(oidcID).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE AWS OIDC configuration %s no longer exists", oidcID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE AWS OIDC Configuration", err.Error())
		return
	}
}

// extractAWSOIDCData pulls the AwsOidcConfigurations typed value out of a composed-type envelope.
func extractAWSOIDCData(envelope models.OidcConfigurationEnvelopeable) (models.AwsOidcConfigurationsable, error) {
	if envelope == nil || envelope.GetData() == nil {
		return nil, fmt.Errorf("no data returned by API")
	}
	data := envelope.GetData().GetAwsOidcConfigurations()
	if data == nil {
		return nil, fmt.Errorf("unexpected OIDC configuration type in API response")
	}
	return data, nil
}

func modelFromTFEAWSOIDCConfiguration(p models.AwsOidcConfigurationsable) modelTFEAWSOIDCConfiguration {
	m := modelTFEAWSOIDCConfiguration{
		ID: types.StringValue(valueOrZero(p.GetId())),
	}
	if attrs := p.GetAttributes(); attrs != nil {
		m.RoleARN = types.StringValue(valueOrZero(attrs.GetRoleArn()))
	}
	if rel := p.GetRelationships(); rel != nil {
		if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
			m.Organization = types.StringValue(valueOrZero(org.GetData().GetId()))
		}
	}
	return m
}
