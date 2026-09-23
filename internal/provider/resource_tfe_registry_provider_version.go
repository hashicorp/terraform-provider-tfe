// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFERegistryProviderVersion{}
var _ resource.ResourceWithConfigure = &resourceTFERegistryProviderVersion{}
var _ resource.ResourceWithImportState = &resourceTFERegistryProviderVersion{}
var _ resource.ResourceWithModifyPlan = &resourceTFERegistryProviderVersion{}

func NewRegistryProviderVersionResource() resource.Resource {
	return &resourceTFERegistryProviderVersion{}
}

// resourceTFERegistryProviderVersion implements the tfe_registry_provider_version resource type
type resourceTFERegistryProviderVersion struct {
	config ConfiguredClient
}

type modelTFERegistryProviderVersion struct {
	ID                 types.String `tfsdk:"id"`
	Organization       types.String `tfsdk:"organization"`
	RegistryName       types.String `tfsdk:"registry_name"`
	Namespace          types.String `tfsdk:"namespace"`
	ProviderName       types.String `tfsdk:"name"`
	Version            types.String `tfsdk:"version"`
	KeyID              types.String `tfsdk:"key_id"`
	Protocols          types.List   `tfsdk:"protocols"`
	ShasumsFile        types.String `tfsdk:"shasums_file"`
	ShasumsSigFile     types.String `tfsdk:"shasums_sig_file"`
	ShasumsUploaded    types.Bool   `tfsdk:"shasums_uploaded"`
	ShasumsSigUploaded types.Bool   `tfsdk:"shasums_sig_uploaded"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type modelTFERegistryProviderVersionIdentity struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	Organization types.String `tfsdk:"organization"`
	RegistryName types.String `tfsdk:"registry_name"`
	Namespace    types.String `tfsdk:"namespace"`
	ProviderName types.String `tfsdk:"name"`
	Version      types.String `tfsdk:"version"`
}

func modelFromTFERegistryProviderVersion(v *tfe.RegistryProviderVersion, organization, registryName, namespace, providerName string) modelTFERegistryProviderVersion {
	protocols, _ := types.ListValueFrom(context.Background(), types.StringType, v.Protocols)

	return modelTFERegistryProviderVersion{
		ID:                 types.StringValue(v.ID),
		Organization:       types.StringValue(organization),
		RegistryName:       types.StringValue(registryName),
		Namespace:          types.StringValue(namespace),
		ProviderName:       types.StringValue(providerName),
		Version:            types.StringValue(v.Version),
		KeyID:              types.StringValue(v.KeyID),
		Protocols:          protocols,
		ShasumsUploaded:    types.BoolValue(v.ShasumsUploaded),
		ShasumsSigUploaded: types.BoolValue(v.ShasumsSigUploaded),
		CreatedAt:          types.StringValue(v.CreatedAt),
		UpdatedAt:          types.StringValue(v.UpdatedAt),
	}
}

func (r *resourceTFERegistryProviderVersion) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider_version"
}

func (r *resourceTFERegistryProviderVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages provider versions in the private registry.",
		Version:     1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the provider version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the provider. For private providers this is the same as the organization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "A valid semver version string.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_id": schema.StringAttribute{
				Description: "The GPG key ID used to sign the provider version.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocols": schema.ListAttribute{
				Description: "An array of Terraform provider API versions that this version supports.",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"shasums_file": schema.StringAttribute{
				Description: "Path to the SHASUMS file (local file path or HTTP/HTTPS URL). Optional - if not provided, SHASUMS must be uploaded separately.",
				Optional:    true,
			},
			"shasums_sig_file": schema.StringAttribute{
				Description: "Path to the SHASUMS signature file (local file path or HTTP/HTTPS URL). Optional - if not provided, signature must be uploaded separately.",
				Optional:    true,
			},
			"shasums_uploaded": schema.BoolAttribute{
				Description: "Indicates whether the SHASUMS file has been uploaded.",
				Computed:    true,
			},
			"shasums_sig_uploaded": schema.BoolAttribute{
				Description: "Indicates whether the SHASUMS signature file has been uploaded.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the provider version was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the provider version was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *resourceTFERegistryProviderVersion) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"hostname": identityschema.StringAttribute{
				OptionalForImport: true,
			},
			"organization": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"registry_name": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"namespace": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"name": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"version": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFERegistryProviderVersion) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFERegistryProviderVersion) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

func (r *resourceTFERegistryProviderVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFERegistryProviderVersion

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine registry name and namespace
	var registryName string
	if plan.RegistryName.IsNull() || plan.RegistryName.IsUnknown() {
		registryName = string(tfe.PrivateRegistry)
	} else {
		registryName = plan.RegistryName.ValueString()
	}

	var namespace string
	if registryName == string(tfe.PrivateRegistry) {
		namespace = organization
	} else {
		namespace = plan.Namespace.ValueString()
	}

	// Extract protocols from list
	var protocols []string
	resp.Diagnostics.Append(plan.Protocols.ElementsAs(ctx, &protocols, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        namespace,
		Name:             plan.ProviderName.ValueString(),
	}

	options := tfe.RegistryProviderVersionCreateOptions{
		Version:   plan.Version.ValueString(),
		KeyID:     plan.KeyID.ValueString(),
		Protocols: protocols,
	}

	tflog.Debug(ctx, "Creating registry provider version")
	providerVersion, err := r.config.Client.RegistryProviderVersions.Create(ctx, providerID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create registry provider version", err.Error())
		return
	}

	// Upload SHASUMS file if provided
	if !plan.ShasumsFile.IsNull() && !plan.ShasumsFile.IsUnknown() {
		shasumsPath := plan.ShasumsFile.ValueString()
		tflog.Debug(ctx, "Uploading SHASUMS file", map[string]interface{}{"path": shasumsPath})

		if err := uploadShasumsFile(ctx, providerVersion, shasumsPath); err != nil {
			resp.Diagnostics.AddError("Unable to upload SHASUMS file", err.Error())
			return
		}
	}

	// Upload SHASUMS signature file if provided
	if !plan.ShasumsSigFile.IsNull() && !plan.ShasumsSigFile.IsUnknown() {
		shasumsSigPath := plan.ShasumsSigFile.ValueString()
		tflog.Debug(ctx, "Uploading SHASUMS signature file", map[string]interface{}{"path": shasumsSigPath})

		if err := uploadShasumsSigFile(ctx, providerVersion, shasumsSigPath); err != nil {
			resp.Diagnostics.AddError("Unable to upload SHASUMS signature file", err.Error())
			return
		}
	}

	// Re-read the provider version to get updated upload status
	if !plan.ShasumsFile.IsNull() || !plan.ShasumsSigFile.IsNull() {
		versionID := tfe.RegistryProviderVersionID{
			RegistryProviderID: providerID,
			Version:            plan.Version.ValueString(),
		}
		providerVersion, err = r.config.Client.RegistryProviderVersions.Read(ctx, versionID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read registry provider version after upload", err.Error())
			return
		}
	}

	result := modelFromTFERegistryProviderVersion(providerVersion, organization, registryName, namespace, plan.ProviderName.ValueString())

	// Preserve the file paths from plan since they're not returned by the API
	result.ShasumsFile = plan.ShasumsFile
	result.ShasumsSigFile = plan.ShasumsSigFile

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := modelTFERegistryProviderVersionIdentity{
		ID:           result.ID,
		Hostname:     types.StringValue(r.config.Client.BaseURL().Host),
		Organization: result.Organization,
		RegistryName: result.RegistryName,
		Namespace:    result.Namespace,
		ProviderName: result.ProviderName,
		Version:      result.Version,
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *resourceTFERegistryProviderVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFERegistryProviderVersion

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := state.RegistryName.ValueString()
	namespace := state.Namespace.ValueString()

	versionID := tfe.RegistryProviderVersionID{
		RegistryProviderID: tfe.RegistryProviderID{
			OrganizationName: organization,
			RegistryName:     tfe.RegistryName(registryName),
			Namespace:        namespace,
			Name:             state.ProviderName.ValueString(),
		},
		Version: state.Version.ValueString(),
	}

	tflog.Debug(ctx, "Reading registry provider version")
	providerVersion, err := r.config.Client.RegistryProviderVersions.Read(ctx, versionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read registry provider version", err.Error())
		return
	}

	result := modelFromTFERegistryProviderVersion(providerVersion, organization, registryName, namespace, state.ProviderName.ValueString())

	// Preserve the file paths from state since they're not returned by the API
	result.ShasumsFile = state.ShasumsFile
	result.ShasumsSigFile = state.ShasumsSigFile

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)

	identity := modelTFERegistryProviderVersionIdentity{
		ID:           result.ID,
		Hostname:     types.StringValue(r.config.Client.BaseURL().Host),
		Organization: result.Organization,
		RegistryName: result.RegistryName,
		Namespace:    result.Namespace,
		ProviderName: result.ProviderName,
		Version:      result.Version,
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *resourceTFERegistryProviderVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// If the resource does not support modification and should always be recreated on
	// configuration value updates, the Update logic can be left empty and ensure all
	// configurable schema attributes implement the resource.RequiresReplace()
	// attribute plan modifier.
	resp.Diagnostics.AddError("Update not supported", "The update operation is not supported on this resource. This is a bug in the provider.")
}

func (r *resourceTFERegistryProviderVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFERegistryProviderVersion

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := state.RegistryName.ValueString()
	namespace := state.Namespace.ValueString()

	versionID := tfe.RegistryProviderVersionID{
		RegistryProviderID: tfe.RegistryProviderID{
			OrganizationName: state.Organization.ValueString(),
			RegistryName:     tfe.RegistryName(registryName),
			Namespace:        namespace,
			Name:             state.ProviderName.ValueString(),
		},
		Version: state.Version.ValueString(),
	}

	tflog.Debug(ctx, "Deleting registry provider version")
	err := r.config.Client.RegistryProviderVersions.Delete(ctx, versionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete registry provider version", err.Error())
		return
	}
}

func (r *resourceTFERegistryProviderVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var organization string
	var registryName string
	var namespace string
	var providerName string
	var version string

	// We'll try to read the legacy import prefix
	if req.ID != "" {
		s := strings.SplitN(req.ID, "/", 5)
		if len(s) != 5 {
			resp.Diagnostics.AddError(
				"Error importing registry provider version",
				fmt.Sprintf("Invalid import format: %s (expected <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>/<VERSION>)", req.ID),
			)
			return
		}
		organization = s[0]
		registryName = s[1]
		namespace = s[2]
		providerName = s[3]
		version = s[4]
	} else {
		// If that is not present we'll read in the identity instead
		var identity modelTFERegistryProviderVersionIdentity
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}

		organization = identity.Organization.ValueString()
		registryName = identity.RegistryName.ValueString()
		namespace = identity.Namespace.ValueString()
		providerName = identity.ProviderName.ValueString()
		version = identity.Version.ValueString()
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), organization)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("registry_name"), registryName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), providerName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)
}

// uploadShasumsFile uploads the SHASUMS file to TFE/TFC
func uploadShasumsFile(ctx context.Context, version *tfe.RegistryProviderVersion, filePath string) error {
	// Get the upload URL
	uploadURL, err := version.ShasumsUploadURL()
	if err != nil {
		return fmt.Errorf("failed to get SHASUMS upload URL: %w", err)
	}

	// Read the file content
	data, err := readFileOrURL(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to read SHASUMS file: %w", err)
	}

	// Upload the file
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload SHASUMS file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// uploadShasumsSigFile uploads the SHASUMS signature file to TFE/TFC
func uploadShasumsSigFile(ctx context.Context, version *tfe.RegistryProviderVersion, filePath string) error {
	// Get the upload URL
	uploadURL, err := version.ShasumsSigUploadURL()
	if err != nil {
		return fmt.Errorf("failed to get SHASUMS signature upload URL: %w", err)
	}

	// Read the file content
	data, err := readFileOrURL(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to read SHASUMS signature file: %w", err)
	}

	// Upload the file
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload SHASUMS signature file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// readFileOrURL reads content from a local file or HTTP/HTTPS URL
func readFileOrURL(ctx context.Context, filePath string) ([]byte, error) {
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// Download from URL
		client := &http.Client{
			Timeout: 5 * time.Minute,
		}

		req, err := http.NewRequestWithContext(ctx, "GET", filePath, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		return data, nil
	}

	// Read from local file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
