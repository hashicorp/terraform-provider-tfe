// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	tfe "github.com/hashicorp/go-tfe"
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
	ID         types.String        `tfsdk:"id"`
	Name       types.String        `tfsdk:"name"`
	KEKID      types.String        `tfsdk:"kek_id"`
	KMSOptions *modelTFEKMSOptions `tfsdk:"kms_options"`

	AWSOIDCConfigurationID   types.String `tfsdk:"aws_oidc_configuration_id"`
	GCPOIDCConfigurationID   types.String `tfsdk:"gcp_oidc_configuration_id"`
	VaultOIDCConfigurationID types.String `tfsdk:"vault_oidc_configuration_id"`
	AzureOIDCConfigurationID types.String `tfsdk:"azure_oidc_configuration_id"`

	AgentPoolID  types.String `tfsdk:"agent_pool_id"`
	Organization types.String `tfsdk:"organization"`
}

type modelTFEKMSOptions struct {
	KeyRegion   types.String `tfsdk:"key_region"`
	KeyLocation types.String `tfsdk:"key_location"`
	KeyRingID   types.String `tfsdk:"key_ring_id"`
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
			"aws_oidc_configuration_id": schema.StringAttribute{
				Description: "The ID of the TFE AWS OIDC configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validateSingleOIDCConfigurationChoice(),
				},
			},
			"gcp_oidc_configuration_id": schema.StringAttribute{
				Description: "The ID of the TFE HYOK configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validateSingleOIDCConfigurationChoice(),
				},
			},
			"vault_oidc_configuration_id": schema.StringAttribute{
				Description: "The ID of the TFE Vault OIDC configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validateSingleOIDCConfigurationChoice(),
				},
			},
			"azure_oidc_configuration_id": schema.StringAttribute{
				Description: "The ID of the TFE Azure OIDC configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validateSingleOIDCConfigurationChoice(),
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

func validateSingleOIDCConfigurationChoice() validator.String {
	return stringvalidator.ExactlyOneOf(
		path.MatchRoot("aws_oidc_configuration_id"),
		path.MatchRoot("gcp_oidc_configuration_id"),
		path.MatchRoot("azure_oidc_configuration_id"),
		path.MatchRoot("vault_oidc_configuration_id"),
	)
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

	var awsOIDCConfig *tfe.AWSOIDCConfiguration
	if plan.AWSOIDCConfigurationID.ValueString() != "" {
		awsOIDCConfig = &tfe.AWSOIDCConfiguration{ID: plan.AWSOIDCConfigurationID.ValueString()}
	}

	var gcpOIDCConfig *tfe.GCPOIDCConfiguration
	if plan.GCPOIDCConfigurationID.ValueString() != "" {
		gcpOIDCConfig = &tfe.GCPOIDCConfiguration{ID: plan.GCPOIDCConfigurationID.ValueString()}
	}

	var vaultOIDCConfig *tfe.VaultOIDCConfiguration
	if plan.VaultOIDCConfigurationID.ValueString() != "" {
		vaultOIDCConfig = &tfe.VaultOIDCConfiguration{ID: plan.VaultOIDCConfigurationID.ValueString()}
	}

	var azureOIDCConfig *tfe.AzureOIDCConfiguration
	if plan.AzureOIDCConfigurationID.ValueString() != "" {
		azureOIDCConfig = &tfe.AzureOIDCConfiguration{ID: plan.AzureOIDCConfigurationID.ValueString()}
	}

	var kmsOptions *tfe.KMSOptions
	if plan.KMSOptions != nil {
		kmsOptions = &tfe.KMSOptions{
			KeyRegion:   plan.KMSOptions.KeyRegion.ValueString(),
			KeyLocation: plan.KMSOptions.KeyLocation.ValueString(),
			KeyRingID:   plan.KMSOptions.KeyRingID.ValueString(),
		}
	}

	options := tfe.HYOKConfigurationsCreateOptions{
		KEKID:      plan.KEKID.ValueString(),
		Name:       plan.Name.ValueString(),
		KMSOptions: kmsOptions,
		OIDCConfiguration: &tfe.OIDCConfigurationTypeChoice{
			AWSOIDCConfiguration:   awsOIDCConfig,
			GCPOIDCConfiguration:   gcpOIDCConfig,
			VaultOIDCConfiguration: vaultOIDCConfig,
			AzureOIDCConfiguration: azureOIDCConfig,
		},
		AgentPool: &tfe.AgentPool{ID: plan.AgentPoolID.ValueString()},
	}

	tflog.Debug(ctx, fmt.Sprintf("Create TFE HYOK Configuration for organization %s", orgName))
	hyok, err := r.config.Client.HYOKConfigurations.Create(ctx, orgName, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TFE HYOK Configuration", err.Error())
		return
	}
	result := modelFromTFEHYOKConfiguration(hyok)
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
	opts := tfe.HYOKConfigurationsReadOptions{
		Include: []tfe.HYOKConfigurationsIncludeOpt{
			tfe.HYOKConfigurationsIncludeOIDCConfiguration,
		},
	}
	tflog.Debug(ctx, fmt.Sprintf("Read HYOK configuration: %s", hyokID))
	hyok, err := r.config.Client.HYOKConfigurations.Read(ctx, state.ID.ValueString(), &opts)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
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
	result := modelFromTFEHYOKConfiguration(hyok)
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

	var kmsOptions *tfe.KMSOptions
	if plan.KMSOptions != nil {
		kmsOptions = &tfe.KMSOptions{
			KeyRegion:   plan.KMSOptions.KeyRegion.ValueString(),
			KeyLocation: plan.KMSOptions.KeyLocation.ValueString(),
			KeyRingID:   plan.KMSOptions.KeyRingID.ValueString(),
		}
	}

	options := tfe.HYOKConfigurationsUpdateOptions{
		Name:       plan.Name.ValueStringPointer(),
		KEKID:      plan.KEKID.ValueStringPointer(),
		KMSOptions: kmsOptions,
		AgentPool:  &tfe.AgentPool{ID: plan.AgentPoolID.ValueString()},
	}

	hyokID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Update TFE HYOK Configuration %s", hyokID))
	hyok, err := r.config.Client.HYOKConfigurations.Update(ctx, hyokID, options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TFE HYOK Configuration", err.Error())
		return
	}

	result := modelFromTFEHYOKConfiguration(hyok)
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
	err := r.config.Client.HYOKConfigurations.Delete(ctx, hyokID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("TFE HYOK configuration %s no longer exists", hyokID))
		}

		resp.Diagnostics.AddError("Error deleting TFE HYOK Configuration", err.Error())
		return
	}
}

func modelFromTFEHYOKConfiguration(p *tfe.HYOKConfiguration) modelTFEHYOKConfiguration {
	var kmsOptions *modelTFEKMSOptions
	if p.KMSOptions != nil {
		kmsOptions = &modelTFEKMSOptions{
			KeyRegion:   types.StringValue(p.KMSOptions.KeyRegion),
			KeyLocation: types.StringValue(p.KMSOptions.KeyLocation),
			KeyRingID:   types.StringValue(p.KMSOptions.KeyRingID),
		}
	}

	model := modelTFEHYOKConfiguration{
		ID:           types.StringValue(p.ID),
		Name:         types.StringValue(p.Name),
		KEKID:        types.StringValue(p.KEKID),
		Organization: types.StringValue(p.Organization.Name),
		AgentPoolID:  types.StringValue(p.AgentPool.ID),
		KMSOptions:   kmsOptions,
	}

	if p.OIDCConfiguration.AWSOIDCConfiguration != nil {
		model.AWSOIDCConfigurationID = types.StringValue(p.OIDCConfiguration.AWSOIDCConfiguration.ID)
	} else if p.OIDCConfiguration.GCPOIDCConfiguration != nil {
		model.GCPOIDCConfigurationID = types.StringValue(p.OIDCConfiguration.GCPOIDCConfiguration.ID)
	} else if p.OIDCConfiguration.AzureOIDCConfiguration != nil {
		model.AzureOIDCConfigurationID = types.StringValue(p.OIDCConfiguration.AzureOIDCConfiguration.ID)
	} else if p.OIDCConfiguration.VaultOIDCConfiguration != nil {
		model.VaultOIDCConfigurationID = types.StringValue(p.OIDCConfiguration.VaultOIDCConfiguration.ID)
	}

	return model
}
