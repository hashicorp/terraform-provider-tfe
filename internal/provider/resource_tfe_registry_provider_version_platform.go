// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFERegistryProviderVersionPlatform{}
var _ resource.ResourceWithConfigure = &resourceTFERegistryProviderVersionPlatform{}
var _ resource.ResourceWithImportState = &resourceTFERegistryProviderVersionPlatform{}
var _ resource.ResourceWithModifyPlan = &resourceTFERegistryProviderVersionPlatform{}

func NewRegistryProviderVersionPlatformResource() resource.Resource {
	return &resourceTFERegistryProviderVersionPlatform{}
}

// resourceTFERegistryProviderVersionPlatform implements the tfe_registry_provider_version_platform resource type
type resourceTFERegistryProviderVersionPlatform struct {
	config ConfiguredClient
}

type modelTFERegistryProviderVersionPlatform struct {
	ID                     types.String `tfsdk:"id"`
	RegistryProviderID     types.String `tfsdk:"registry_provider_id"`
	Organization           types.String `tfsdk:"organization"`
	RegistryName           types.String `tfsdk:"registry_name"`
	Namespace              types.String `tfsdk:"namespace"`
	ProviderName           types.String `tfsdk:"name"`
	Version                types.String `tfsdk:"version"`
	OSArch                 types.String `tfsdk:"os_arch"`
	Filename               types.String `tfsdk:"filename"`
	Shasum                 types.String `tfsdk:"shasum"`
	ProviderBinaryUploaded types.Bool   `tfsdk:"provider_binary_uploaded"`
}

func modelFromTFERegistryProviderVersionPlatform(p *tfe.RegistryProviderPlatform, organization, registryName, namespace, providerName, version string) modelTFERegistryProviderVersionPlatform {
	osArch := fmt.Sprintf("%s_%s", p.OS, p.Arch)

	var registryProviderID string
	if p.RegistryProviderVersion != nil {
		registryProviderID = p.RegistryProviderVersion.ID
	}

	return modelTFERegistryProviderVersionPlatform{
		ID:                     types.StringValue(p.ID),
		RegistryProviderID:     types.StringValue(registryProviderID),
		Organization:           types.StringValue(organization),
		RegistryName:           types.StringValue(registryName),
		Namespace:              types.StringValue(namespace),
		ProviderName:           types.StringValue(providerName),
		Version:                types.StringValue(version),
		OSArch:                 types.StringValue(osArch),
		Filename:               types.StringValue(p.Filename),
		Shasum:                 types.StringValue(p.Shasum),
		ProviderBinaryUploaded: types.BoolValue(p.ProviderBinaryUploaded),
	}
}

func (r *resourceTFERegistryProviderVersionPlatform) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_provider_version_platform"
}

func (r *resourceTFERegistryProviderVersionPlatform) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a platform release binary of a Provider Version in the private registry.",
		Version:     1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the provider version platform.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registry_provider_id": schema.StringAttribute{
				Description: "ID of the registry provider version this platform belongs to.",
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
				Description: "Version of the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"os_arch": schema.StringAttribute{
				Description: "A valid operating system string and architecture string, separated by an underscore. Example: `linux_amd64`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filename": schema.StringAttribute{
				Description: "The path to the provider binary file (local path or HTTP/HTTPS URL). The file will be uploaded to the registry.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"shasum": schema.StringAttribute{
				Description: "The SHA256 checksum of the provider binary.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_binary_uploaded": schema.BoolAttribute{
				Description: "Indicates whether the provider binary has been uploaded.",
				Computed:    true,
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFERegistryProviderVersionPlatform) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFERegistryProviderVersionPlatform) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
}

// downloadURL fetches content from an HTTP/HTTPS URL and returns the data and filename.
func downloadURL(ctx context.Context, source string) ([]byte, string, error) {
	tflog.Debug(ctx, "Downloading file from URL", map[string]interface{}{"url": source})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	filename := filepath.Base(source)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return data, filename, nil
}

// readFileContent reads content from a local file path or remote URL
func readFileContent(ctx context.Context, source string) ([]byte, string, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return downloadURL(ctx, source)
	}

	// Otherwise treat as local file path
	tflog.Debug(ctx, "Reading local file", map[string]interface{}{"path": source})

	data, err := os.ReadFile(source)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read local file: %w", err)
	}

	return data, filepath.Base(source), nil
}

// calculateSHA256 calculates the SHA256 checksum of the given data
func calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// uploadProviderBinary uploads the provider binary to the given URL
func uploadProviderBinary(ctx context.Context, uploadURL string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(data))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (r *resourceTFERegistryProviderVersionPlatform) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFERegistryProviderVersionPlatform

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

	// Parse os_arch into OS and Arch
	osArch := plan.OSArch.ValueString()
	parts := strings.Split(osArch, "_")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid os_arch format",
			fmt.Sprintf("os_arch must be in the format 'os_arch' (e.g., 'linux_amd64'), got: %s", osArch),
		)
		return
	}
	osName := parts[0]
	arch := parts[1]

	// Read the file content (from local path or URL)
	fileSource := plan.Filename.ValueString()
	fileContent, filename, err := readFileContent(ctx, fileSource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read provider binary file",
			fmt.Sprintf("Could not read file from %s: %s", fileSource, err.Error()),
		)
		return
	}

	// Calculate SHA256 checksum
	shasum := calculateSHA256(fileContent)

	versionID := tfe.RegistryProviderVersionID{
		RegistryProviderID: tfe.RegistryProviderID{
			OrganizationName: organization,
			RegistryName:     tfe.RegistryName(registryName),
			Namespace:        namespace,
			Name:             plan.ProviderName.ValueString(),
		},
		Version: plan.Version.ValueString(),
	}

	options := tfe.RegistryProviderPlatformCreateOptions{
		OS:       osName,
		Arch:     arch,
		Shasum:   shasum,
		Filename: filename,
	}

	tflog.Debug(ctx, "Creating registry provider version platform")
	platform, err := r.config.Client.RegistryProviderPlatforms.Create(ctx, versionID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create registry provider version platform", err.Error())
		return
	}

	// Upload the provider binary if there's an upload URL in the links
	if uploadURL, ok := platform.Links["provider-binary-upload"].(string); ok && uploadURL != "" {
		tflog.Debug(ctx, "Uploading provider binary", map[string]interface{}{
			"upload_url": uploadURL,
			"filename":   filename,
			"size":       len(fileContent),
		})

		err = uploadProviderBinary(ctx, uploadURL, fileContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to upload provider binary",
				fmt.Sprintf("Failed to upload binary file: %s", err.Error()),
			)
			return
		}

		// Re-read the platform to get updated status
		platformID := tfe.RegistryProviderPlatformID{
			RegistryProviderVersionID: versionID,
			OS:                        osName,
			Arch:                      arch,
		}
		platform, err = r.config.Client.RegistryProviderPlatforms.Read(ctx, platformID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read registry provider version platform after upload", err.Error())
			return
		}
	}

	result := modelFromTFERegistryProviderVersionPlatform(platform, organization, registryName, namespace, plan.ProviderName.ValueString(), plan.Version.ValueString())

	// Preserve the original filename from the plan (could be a URL or local path)
	// The API only returns the extracted filename, but we want to keep the original source
	result.Filename = plan.Filename

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryProviderVersionPlatform) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFERegistryProviderVersionPlatform

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

	// Parse os_arch into OS and Arch
	osArch := state.OSArch.ValueString()
	parts := strings.Split(osArch, "_")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid os_arch format in state",
			fmt.Sprintf("os_arch must be in the format 'os_arch' (e.g., 'linux_amd64'), got: %s", osArch),
		)
		return
	}
	osName := parts[0]
	arch := parts[1]

	platformID := tfe.RegistryProviderPlatformID{
		RegistryProviderVersionID: tfe.RegistryProviderVersionID{
			RegistryProviderID: tfe.RegistryProviderID{
				OrganizationName: organization,
				RegistryName:     tfe.RegistryName(registryName),
				Namespace:        namespace,
				Name:             state.ProviderName.ValueString(),
			},
			Version: state.Version.ValueString(),
		},
		OS:   osName,
		Arch: arch,
	}

	tflog.Debug(ctx, "Reading registry provider version platform")
	platform, err := r.config.Client.RegistryProviderPlatforms.Read(ctx, platformID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read registry provider version platform", err.Error())
		return
	}

	result := modelFromTFERegistryProviderVersionPlatform(platform, organization, registryName, namespace, state.ProviderName.ValueString(), state.Version.ValueString())

	// Preserve the original filename from state (could be a URL or local path)
	// The API only returns the extracted filename, but we want to keep the original source
	result.Filename = state.Filename

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFERegistryProviderVersionPlatform) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// If the resource does not support modification and should always be recreated on
	// configuration value updates, the Update logic can be left empty and ensure all
	// configurable schema attributes implement the resource.RequiresReplace()
	// attribute plan modifier.
	resp.Diagnostics.AddError("Update not supported", "The update operation is not supported on this resource. This is a bug in the provider.")
}

func (r *resourceTFERegistryProviderVersionPlatform) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFERegistryProviderVersionPlatform

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registryName := state.RegistryName.ValueString()
	namespace := state.Namespace.ValueString()

	// Parse os_arch into OS and Arch
	osArch := state.OSArch.ValueString()
	parts := strings.Split(osArch, "_")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid os_arch format in state",
			fmt.Sprintf("os_arch must be in the format 'os_arch' (e.g., 'linux_amd64'), got: %s", osArch),
		)
		return
	}
	osName := parts[0]
	arch := parts[1]

	platformID := tfe.RegistryProviderPlatformID{
		RegistryProviderVersionID: tfe.RegistryProviderVersionID{
			RegistryProviderID: tfe.RegistryProviderID{
				OrganizationName: state.Organization.ValueString(),
				RegistryName:     tfe.RegistryName(registryName),
				Namespace:        namespace,
				Name:             state.ProviderName.ValueString(),
			},
			Version: state.Version.ValueString(),
		},
		OS:   osName,
		Arch: arch,
	}

	tflog.Debug(ctx, "Deleting registry provider version platform")
	err := r.config.Client.RegistryProviderPlatforms.Delete(ctx, platformID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete registry provider version platform", err.Error())
		return
	}
}

func (r *resourceTFERegistryProviderVersionPlatform) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>/<VERSION>/<OS>/<ARCH>
	s := strings.SplitN(req.ID, "/", 7)
	if len(s) != 7 {
		resp.Diagnostics.AddError(
			"Error importing registry provider version platform",
			fmt.Sprintf("Invalid import format: %s (expected <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>/<VERSION>/<OS>/<ARCH>)", req.ID),
		)
		return
	}

	organization := s[0]
	registryName := s[1]
	namespace := s[2]
	providerName := s[3]
	version := s[4]
	osName := s[5]
	arch := s[6]
	osArch := fmt.Sprintf("%s_%s", osName, arch)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), organization)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("registry_name"), registryName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), providerName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("os_arch"), osArch)...)
}
