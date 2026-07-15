// // Copyright IBM Corp. 2018, 2025
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	tfe "github.com/hashicorp/go-tfe/v2"
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
	_ resource.ResourceWithConfigure   = &resourceTFEHYOKConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFEHYOKConfiguration{}
)

func NewHYOKConfigurationResource() resource.Resource {
	return &resourceTFEHYOKConfiguration{}
}

type resourceTFEHYOKConfiguration struct {
	config ConfiguredClient
}

type modelTFEHYOKConfiguration struct {
	ID                    types.String        `tfsdk:"id"`
	Name                  types.String        `tfsdk:"name"`
	KEKID                 types.String        `tfsdk:"kek_id"`
	KMSOptions            *modelTFEKMSOptions `tfsdk:"kms_options"`
	OIDCConfigurationID   types.String        `tfsdk:"oidc_configuration_id"`
	OIDCConfigurationType types.String        `tfsdk:"oidc_configuration_type"`
	AgentPoolID           types.String        `tfsdk:"agent_pool_id"`
	Organization          types.String        `tfsdk:"organization"`
}

type modelTFEKMSOptions struct {
	KeyRegion   types.String `tfsdk:"key_region"`
	KeyLocation types.String `tfsdk:"key_location"`
	KeyRingID   types.String `tfsdk:"key_ring_id"`
}

// List all available OIDC configuration types.
const (
	OIDCConfigurationTypeAWS   string = "aws"
	OIDCConfigurationTypeGCP   string = "gcp"
	OIDCConfigurationTypeVault string = "vault"
	OIDCConfigurationTypeAzure string = "azure"
)

var idDataTypeToOidcConfigurationType = map[models.OidcConfigurationsId_data_type]string{
	models.AWSOIDCCONFIGURATIONS_OIDCCONFIGURATIONSID_DATA_TYPE:   OIDCConfigurationTypeAWS,
	models.GCPOIDCCONFIGURATIONS_OIDCCONFIGURATIONSID_DATA_TYPE:   OIDCConfigurationTypeGCP,
	models.VAULTOIDCCONFIGURATIONS_OIDCCONFIGURATIONSID_DATA_TYPE: OIDCConfigurationTypeVault,
	models.AZUREOIDCCONFIGURATIONS_OIDCCONFIGURATIONSID_DATA_TYPE: OIDCConfigurationTypeAzure,
}

func (r *resourceTFEHYOKConfiguration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEHYOKConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hyok_configuration"
}

func (r *resourceTFEHYOKConfiguration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the HYOK configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Label for the HYOK configuration to be used within HCP Terraform.",
				Required:    true,
			},
			"kek_id": schema.StringAttribute{
				Description: "Refers to the name of your key encryption key stored in your key management service.",
				Required:    true,
			},
			"oidc_configuration_id": schema.StringAttribute{
				Description: "The ID of the TFE OIDC configuration.",
				Required:    true,
			},
			"oidc_configuration_type": schema.StringAttribute{
				Description: "The type of the TFE OIDC configuration.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						OIDCConfigurationTypeAWS,
						OIDCConfigurationTypeGCP,
						OIDCConfigurationTypeVault,
						OIDCConfigurationTypeAzure,
					),
				},
			},
			"agent_pool_id": schema.StringAttribute{
				Description: "The ID of the agent-pool to associate with the HYOK configuration.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization to which the TFE HYOK configuration belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kms_options": schema.SingleNestedBlock{
				Description: "Optional object used to specify additional fields for some key management services.",
				Attributes: map[string]schema.Attribute{
					"key_region": schema.StringAttribute{
						Description: "The AWS region where your key is located.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
					"key_location": schema.StringAttribute{
						Description: "The location in which the GCP key ring exists.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
					"key_ring_id": schema.StringAttribute{
						Description: "The root resource for Google Cloud KMS keys and key versions.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
				},
			},
		},
		Description: "Generates a new TFE HYOK Configuration.",
	}
}

func (r *resourceTFEHYOKConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceTFEHYOKConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan into the model
	var plan modelTFEHYOKConfiguration
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

	options := hyokConfigurationEnvelopeFromModel(plan)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE HYOK Configuration for organization %s", orgName))
	envelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).HyokConfigurations().Post(ctx, options, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE HYOK Configuration", err.Error())
		return
	}
	result := modelFromTFEHYOKConfiguration(envelope)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEHYOKConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state into the model
	var state modelTFEHYOKConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read HYOK configuration: %s", id))
	envelope, err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(state.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("HYOK configuration %s no longer exists", id))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error reading HYOK configuration %s", id),
				err.Error(),
			)
		}
		return
	}
	result := modelFromTFEHYOKConfiguration(envelope)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEHYOKConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEHYOKConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state modelTFEHYOKConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := hyokConfigurationEnvelopeFromModel(plan)

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE HYOK Configuration %s", id))
	envelope, err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(id).Patch(ctx, options, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE HYOK Configuration", err.Error())
		return
	}

	result := modelFromTFEHYOKConfiguration(envelope)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEHYOKConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEHYOKConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE HYOK configuration: %s", id))
	err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(id).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE HYOK configuration %s no longer exists", id))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE HYOK Configuration", err.Error())
		return
	}
}

func hyokConfigurationEnvelopeFromModel(m modelTFEHYOKConfiguration) models.HyokConfigurationsEnvelopeable {
	// Attributes
	attributes := models.NewHyokConfigurations_attributes()
	attributes.SetName(m.Name.ValueStringPointer())
	attributes.SetKekId(m.KEKID.ValueStringPointer())
	if m.KMSOptions != nil {
		attributes.SetKmsOptions(v2KMSOptions(m.KMSOptions))
	}

	// Relationships
	relationships := models.NewHyokConfigurations_relationships()
	relationships.SetAgentPool(v2AgentPoolRelationship(m.AgentPoolID.ValueString()))
	relationships.SetOidcConfiguration(v2OIDCConfigurationRelationship(m.OIDCConfigurationID.ValueString()))

	hyokConfiguration := models.NewHyokConfigurations()
	hyokConfiguration.SetAttributes(attributes)
	hyokConfiguration.SetRelationships(relationships)

	envelope := models.NewHyokConfigurationsEnvelope()
	envelope.SetData(hyokConfiguration)
	return envelope
}

func v2KMSOptions(m *modelTFEKMSOptions) *models.HyokConfigurations_attributes_kmsOptions {
	kmsOptionsAttributes := models.NewHyokConfigurations_attributes_kmsOptions()

	if m == nil {
		return nil
	}

	if v := m.KeyRegion.ValueString(); v != "" {
		kmsOptionsAttributes.SetKeyRegion(&v)
	}
	if v := m.KeyLocation.ValueString(); v != "" {
		kmsOptionsAttributes.SetKeyLocation(&v)
	}
	if v := m.KeyRingID.ValueString(); v != "" {
		kmsOptionsAttributes.SetKeyRingId(&v)
	}

	return kmsOptionsAttributes
}

func v2AgentPoolRelationship(id string) models.AgentPoolsIdable {
	agentPoolsIdData := models.NewAgentPoolsId_data()
	agentPoolsIdData.SetId(&id)

	agentPoolsId := models.NewAgentPoolsId()
	agentPoolsId.SetData(agentPoolsIdData)

	return agentPoolsId
}

func v2OIDCConfigurationRelationship(id string) models.OidcConfigurationsIdable {
	oidcIdData := models.NewOidcConfigurationsId_data()
	oidcIdData.SetId(&id)

	oidcId := models.NewOidcConfigurationsId()
	oidcId.SetData(oidcIdData)

	return oidcId
}

func modelFromTFEHYOKConfiguration(p models.HyokConfigurationsEnvelopeable) modelTFEHYOKConfiguration {
	model := modelTFEHYOKConfiguration{}

	data := p.GetData()
	if data == nil {
		return model
	}

	model.ID = types.StringValue(*data.GetId())

	attributes := data.GetAttributes()
	if attributes != nil {
		model.Name = types.StringValue(*attributes.GetName())
		model.KEKID = types.StringValue(*attributes.GetKekId())
		model.KMSOptions = tfeKMSOptions(attributes.GetKmsOptions())
	}

	relationships := data.GetRelationships()
	if relationships == nil {
		return model
	}

	organization := relationships.GetOrganization()
	if organization != nil && organization.GetData() != nil {
		model.Organization = types.StringValue(*organization.GetData().GetId())
	}

	agentPool := relationships.GetAgentPool()
	if agentPool != nil && agentPool.GetData() != nil {
		model.AgentPoolID = types.StringValue(*agentPool.GetData().GetId())
	}

	oidc := relationships.GetOidcConfiguration()
	if oidc != nil && oidc.GetData() != nil {
		model.OIDCConfigurationID = types.StringValue(*oidc.GetData().GetId())

		oidcType := oidc.GetData().GetTypeEscaped()
		if oidcType != nil {
			model.OIDCConfigurationType = types.StringValue(idDataTypeToOidcConfigurationType[*oidcType])
		}
	}

	return model
}

func tfeKMSOptions(m models.HyokConfigurations_attributes_kmsOptionsable) *modelTFEKMSOptions {
	if m == nil {
		return nil
	}

	kmsOptions := &modelTFEKMSOptions{
		KeyRegion:   types.StringValue(""),
		KeyLocation: types.StringValue(""),
		KeyRingID:   types.StringValue(""),
	}

	if v := m.GetKeyRegion(); v != nil {
		kmsOptions.KeyRegion = types.StringValue(*v)
	}
	if v := m.GetKeyLocation(); v != nil {
		kmsOptions.KeyLocation = types.StringValue(*v)
	}
	if v := m.GetKeyRingId(); v != nil {
		kmsOptions.KeyRingID = types.StringValue(*v)
	}

	return kmsOptions
}
