// // Copyright IBM Corp. 2018, 2025
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	MultiRegion types.Bool   `tfsdk:"multi_region"`
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

// oidcTypeToV2 maps the provider's oidc_configuration_type string to the v2 OidcConfigurationsIdentifier_type enum.
func oidcTypeToV2(t string) *models.OidcConfigurationsIdentifier_type {
	var v models.OidcConfigurationsIdentifier_type
	switch t {
	case OIDCConfigurationTypeAWS:
		v = models.AWSOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE
	case OIDCConfigurationTypeAzure:
		v = models.AZUREOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE
	case OIDCConfigurationTypeGCP:
		v = models.GCPOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE
	case OIDCConfigurationTypeVault:
		v = models.VAULTOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE
	default:
		return nil
	}
	return &v
}

// oidcTypeFromV2 maps the v2 OidcConfigurationsIdentifier_type enum to the provider's oidc_configuration_type string.
func oidcTypeFromV2(t *models.OidcConfigurationsIdentifier_type) string {
	if t == nil {
		return ""
	}
	switch *t {
	case models.AWSOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE:
		return OIDCConfigurationTypeAWS
	case models.AZUREOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE:
		return OIDCConfigurationTypeAzure
	case models.GCPOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE:
		return OIDCConfigurationTypeGCP
	case models.VAULTOIDCCONFIGURATIONS_OIDCCONFIGURATIONSIDENTIFIER_TYPE:
		return OIDCConfigurationTypeVault
	}
	return ""
}

// hyokAgentPoolRelationship builds an AgentPoolsHasOne with the given ID.
func hyokAgentPoolRelationship(id string) models.AgentPoolsHasOneable {
	apType := models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE
	data := models.NewAgentPoolsHasOne_data()
	data.SetId(&id)
	data.SetTypeEscaped(&apType)
	rel := models.NewAgentPoolsHasOne()
	rel.SetData(data)
	return rel
}

// hyokOIDCRelationship builds an OidcConfigurationsHasOne with the given ID and type string.
func hyokOIDCRelationship(id string, oidcType string) models.OidcConfigurationsHasOneable {
	v2Type := oidcTypeToV2(oidcType)
	data := models.NewOidcConfigurationsHasOne_data()
	data.SetId(&id)
	if v2Type != nil {
		data.SetTypeEscaped(v2Type)
	}
	rel := models.NewOidcConfigurationsHasOne()
	rel.SetData(data)
	return rel
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
		MarkdownDescription: "Manages a HYOK configuration.\n\n~> **Note:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.",
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
				Description: "The ID of the TFE OIDC configuration. This is typically sourced from another OIDC configuration resource corresponding with the target cloud provider, such as `tfe_oidc_configuration_gcp`, `tfe_oidc_configuration_aws`, or `tfe_oidc_configuration_azure`.",
				Required:    true,
			},
			"oidc_configuration_type": schema.StringAttribute{
				Description: "The type of the TFE OIDC configuration. Valid values are `vault`, `aws`, `gcp`, and `azure`.",
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
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
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
					"multi_region": schema.BoolAttribute{
						Description: "Whether the AWS key is a multi-region key.",
						Optional:    true,
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
	}
}

func (r *resourceTFEHYOKConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildHYOKRequestBody constructs a HyokConfigurationsEnvelope for Create or Update.
// Pass oidcID/oidcType non-empty only for Create; omit them for Update.
func buildHYOKRequestBody(name, kekID string, kmsOpts *modelTFEKMSOptions, agentPoolID, oidcID, oidcType string) *models.HyokConfigurationsEnvelope {
	attrs := models.NewHyokConfigurations_attributes()
	attrs.SetName(&name)
	attrs.SetKekId(&kekID)

	if kmsOpts != nil {
		kms := models.NewHyokConfigurations_attributes_kmsOptions()
		kms.SetKeyRegion(valueOrNilString(kmsOpts.KeyRegion.ValueString()))
		kms.SetMultiRegion(kmsOpts.MultiRegion.ValueBoolPointer())
		kms.SetKeyLocation(valueOrNilString(kmsOpts.KeyLocation.ValueString()))
		kms.SetKeyRingId(valueOrNilString(kmsOpts.KeyRingID.ValueString()))
		attrs.SetKmsOptions(kms)
	}

	rels := models.NewHyokConfigurations_relationships()
	if agentPoolID != "" {
		rels.SetAgentPool(hyokAgentPoolRelationship(agentPoolID))
	}
	if oidcID != "" && oidcType != "" {
		rels.SetOidcConfiguration(hyokOIDCRelationship(oidcID, oidcType))
	}

	hyokType := models.HYOKCONFIGURATIONS_HYOKCONFIGURATIONS_TYPE
	data := models.NewHyokConfigurations()
	data.SetTypeEscaped(&hyokType)
	data.SetAttributes(attrs)
	data.SetRelationships(rels)

	envelope := models.NewHyokConfigurationsEnvelope()
	envelope.SetData(data)
	return envelope
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

	envelope := buildHYOKRequestBody(
		plan.Name.ValueString(),
		plan.KEKID.ValueString(),
		plan.KMSOptions,
		plan.AgentPoolID.ValueString(),
		plan.OIDCConfigurationID.ValueString(),
		plan.OIDCConfigurationType.ValueString(),
	)

	tflog.Debug(ctx, fmt.Sprintf("Create TFE HYOK Configuration for organization %s", orgName))
	hyokEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(orgName).HyokConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE HYOK Configuration", err.Error())
		return
	}
	if hyokEnvelope == nil || hyokEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Error creating TFE HYOK Configuration", "no data was returned by the API")
		return
	}

	result := modelFromTFEHYOKConfigurationV2(hyokEnvelope.GetData(), plan.OIDCConfigurationType.ValueString())
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

	hyokID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Read HYOK configuration: %s", hyokID))
	hyokEnvelope, err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(hyokID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("HYOK configuration %s no longer exists", hyokID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error reading HYOK configuration %s", hyokID),
				err.Error(),
			)
		}
		return
	}
	if hyokEnvelope == nil || hyokEnvelope.GetData() == nil {
		tflog.Debug(ctx, fmt.Sprintf("HYOK configuration %s no longer exists", hyokID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Recover oidc_configuration_type from the relationship type enum; fall back
	// to state if the relationship is absent (e.g. older TFE that omits it).
	oidcType := ""
	if rel := hyokEnvelope.GetData().GetRelationships(); rel != nil {
		if oidcRel := rel.GetOidcConfiguration(); oidcRel != nil && oidcRel.GetData() != nil {
			oidcType = oidcTypeFromV2(oidcRel.GetData().GetTypeEscaped())
		}
	}
	if oidcType == "" {
		oidcType = state.OIDCConfigurationType.ValueString()
	}

	result := modelFromTFEHYOKConfigurationV2(hyokEnvelope.GetData(), oidcType)
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

	// Update does not accept oidc_configuration (RequiresReplace); omit it.
	envelope := buildHYOKRequestBody(
		plan.Name.ValueString(),
		plan.KEKID.ValueString(),
		plan.KMSOptions,
		plan.AgentPoolID.ValueString(),
		"", // oidcID – not sent on update
		"", // oidcType – not sent on update
	)

	hyokID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE HYOK Configuration %s", hyokID))
	hyokEnvelope, err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(hyokID).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE HYOK Configuration", err.Error())
		return
	}
	if hyokEnvelope == nil || hyokEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Error updating TFE HYOK Configuration", "no data was returned by the API")
		return
	}

	result := modelFromTFEHYOKConfigurationV2(hyokEnvelope.GetData(), plan.OIDCConfigurationType.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *resourceTFEHYOKConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEHYOKConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hyokID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Delete TFE HYOK configuration: %s", hyokID))
	err := r.config.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(hyokID).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE HYOK configuration %s no longer exists", hyokID))
			return
		}

		resp.Diagnostics.AddError("Error deleting TFE HYOK Configuration", err.Error())
		return
	}
}

// modelFromTFEHYOKConfigurationV2 converts a v2 HyokConfigurationsable into the
// provider model. oidcTypeFallback is used when the relationship type is absent.
func modelFromTFEHYOKConfigurationV2(p models.HyokConfigurationsable, oidcTypeFallback string) modelTFEHYOKConfiguration {
	m := modelTFEHYOKConfiguration{
		ID: types.StringValue(valueOrZero(p.GetId())),
	}

	if attrs := p.GetAttributes(); attrs != nil {
		m.Name = types.StringValue(valueOrZero(attrs.GetName()))
		m.KEKID = types.StringValue(valueOrZero(attrs.GetKekId()))

		if kms := attrs.GetKmsOptions(); kms != nil {
			m.KMSOptions = &modelTFEKMSOptions{
				KeyRegion:   types.StringValue(valueOrZero(kms.GetKeyRegion())),
				MultiRegion: types.BoolNull(),
				KeyLocation: types.StringValue(valueOrZero(kms.GetKeyLocation())),
				KeyRingID:   types.StringValue(valueOrZero(kms.GetKeyRingId())),
			}

			// Explicitly populate MultiRegion bool when specified to avoid "inconsistent result after apply" error
			if multiRegion := kms.GetMultiRegion(); multiRegion != nil {
				m.KMSOptions.MultiRegion = types.BoolValue(valueOrZero(multiRegion))
			}
		}
	}

	populateHYOKRelationships(&m, p.GetRelationships(), oidcTypeFallback)
	return m
}

// populateHYOKRelationships fills organization, agent pool, and OIDC
// configuration fields from the v2 relationships block. When the OIDC type
// cannot be recovered from the relationship, oidcTypeFallback is used.
func populateHYOKRelationships(m *modelTFEHYOKConfiguration, rel models.HyokConfigurations_relationshipsable, oidcTypeFallback string) {
	if rel == nil {
		m.OIDCConfigurationType = types.StringValue(oidcTypeFallback)
		return
	}

	if org := rel.GetOrganization(); org != nil && org.GetData() != nil {
		m.Organization = types.StringValue(valueOrZero(org.GetData().GetId()))
	}
	if ap := rel.GetAgentPool(); ap != nil && ap.GetData() != nil {
		m.AgentPoolID = types.StringValue(valueOrZero(ap.GetData().GetId()))
	}

	oidcType := oidcTypeFallback
	if oidc := rel.GetOidcConfiguration(); oidc != nil && oidc.GetData() != nil {
		m.OIDCConfigurationID = types.StringValue(valueOrZero(oidc.GetData().GetId()))
		if t := oidcTypeFromV2(oidc.GetData().GetTypeEscaped()); t != "" {
			oidcType = t
		}
	}
	m.OIDCConfigurationType = types.StringValue(oidcType)
}

// valueOrNilString returns nil for an empty string, or a pointer to the string.
func valueOrNilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
