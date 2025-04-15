// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dataSourceTFERegistryModule{}
	_ datasource.DataSourceWithConfigure = &dataSourceTFERegistryModule{}
)

// NewModuleDataSource is a helper function to simplify the implementation.
func NewRegistryModuleDataSource() datasource.DataSource {
	return &dataSourceTFERegistryModule{}
}

// dataSourceTFEModule is the data source implementation.
type dataSourceTFERegistryModule struct {
	config ConfiguredClient
}

// modelModule maps the data source schema data.
type modelRegistryModule struct {
	ID                  types.String                         `tfsdk:"id"`
	Organization        types.String                         `tfsdk:"organization"`
	Namespace           types.String                         `tfsdk:"namespace"`
	Name                types.String                         `tfsdk:"name"`
	RegistryName        types.String                         `tfsdk:"registry_name"`
	ModuleProvider      types.String                         `tfsdk:"module_provider"`
	NoCodeModuleID      types.String                         `tfsdk:"no_code_module_id"`
	NoCodeModuleSource  types.String                         `tfsdk:"no_code_module_source"`
	NoCode              types.Bool                           `tfsdk:"no_code"`
	Permissions         []modelRegistryModulePermissions     `tfsdk:"permissions"`
	PublishingMechanism types.String                         `tfsdk:"publishing_mechanism"`
	Status              types.String                         `tfsdk:"status"`
	TestConfig          []modelTestConfig                    `tfsdk:"test_config"`
	VCSRepo             []modelTFEVCSRepo                    `tfsdk:"vcs_repo"`
	VersionStatuses     []modelRegistryModuleVersionStatuses `tfsdk:"version_statuses"`
	CreatedAt           types.String                         `tfsdk:"created_at"`
	UpdatedAt           types.String                         `tfsdk:"updated_at"`
}

type modelRegistryModuleVersionStatuses struct {
	Version types.String `tfsdk:"version"`
	Status  types.String `tfsdk:"status"`
	Error   types.String `tfsdk:"error"`
}

func (m modelRegistryModuleVersionStatuses) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"version": types.StringType,
		"status":  types.StringType,
		"error":   types.StringType,
	}
}

func modelFromTFERegistryModuleVersionStatuses(v *tfe.RegistryModuleVersionStatuses) modelRegistryModuleVersionStatuses {
	return modelRegistryModuleVersionStatuses{
		Version: types.StringValue(v.Version),
		Status:  types.StringValue(string(v.Status)),
		Error:   types.StringValue(v.Error),
	}
}

type modelTestConfig struct {
	TestsEnabled types.Bool `tfsdk:"tests_enabled"`
}

func (m modelTestConfig) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tests_enabled": types.BoolType,
	}
}

type modelTFEVCSRepo struct {
	Branch            types.String `tfsdk:"branch"`
	DisplayIdentifier types.String `tfsdk:"display_identifier"`
	Identifier        types.String `tfsdk:"identifier"`
	IngressSubmodules types.Bool   `tfsdk:"ingress_submodules"`
	OAuthTokenID      types.String `tfsdk:"oauth_token_id"`
	GHAInstallationID types.String `tfsdk:"github_app_installation_id"`
	RepositoryHTTPURL types.String `tfsdk:"repository_http_url"`
	ServiceProvider   types.String `tfsdk:"service_provider"`
	Tags              types.Bool   `tfsdk:"tags"`
	TagsRegex         types.String `tfsdk:"tags_regex"`
	WebhookURL        types.String `tfsdk:"webhook_url"`
}

func (m modelTFEVCSRepo) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"branch":                     types.StringType,
		"display_identifier":         types.StringType,
		"identifier":                 types.StringType,
		"ingress_submodules":         types.BoolType,
		"oauth_token_id":             types.StringType,
		"github_app_installation_id": types.StringType,
		"repository_http_url":        types.StringType,
		"service_provider":           types.StringType,
		"tags":                       types.BoolType,
		"tags_regex":                 types.StringType,
		"webhook_url":                types.StringType,
	}
}

func modelFromTFEVCSRepo(v *tfe.VCSRepo) modelTFEVCSRepo {
	return modelTFEVCSRepo{
		Branch:            types.StringValue(v.Branch),
		DisplayIdentifier: types.StringValue(v.DisplayIdentifier),
		Identifier:        types.StringValue(v.Identifier),
		IngressSubmodules: types.BoolValue(v.IngressSubmodules),
		OAuthTokenID:      types.StringValue(v.OAuthTokenID),
		GHAInstallationID: types.StringValue(v.GHAInstallationID),
		RepositoryHTTPURL: types.StringValue(v.RepositoryHTTPURL),
		ServiceProvider:   types.StringValue(v.ServiceProvider),
		Tags:              types.BoolValue(v.Tags),
		TagsRegex:         types.StringValue(v.TagsRegex),
		WebhookURL:        types.StringValue(v.WebhookURL),
	}
}

func modelFromTFETestConfig(v *tfe.TestConfig) modelTestConfig {
	return modelTestConfig{
		TestsEnabled: types.BoolValue(v.TestsEnabled),
	}
}

type modelRegistryModulePermissions struct {
	CanDelete types.Bool `tfsdk:"can_delete"`
	CanResync types.Bool `tfsdk:"can_resync"`
	CanRetry  types.Bool `tfsdk:"can_retry"`
}

func (m modelRegistryModulePermissions) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"can_delete": types.BoolType,
		"can_resync": types.BoolType,
		"can_retry":  types.BoolType,
	}
}

func modelFromTFERegistryModulePermission(v *tfe.RegistryModulePermissions) modelRegistryModulePermissions {
	return modelRegistryModulePermissions{
		CanDelete: types.BoolValue(v.CanDelete),
		CanResync: types.BoolValue(v.CanResync),
		CanRetry:  types.BoolValue(v.CanRetry),
	}
}

// Metadata returns the data source type name.
func (d *dataSourceTFERegistryModule) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_module"
}

// Schema defines the schema for the data source.
func (d *dataSourceTFERegistryModule) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used to retrieve a public or private no-code module.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the no-code module.",
				Computed:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the module.",
				Required:    true,
			},
			"registry_name": schema.StringAttribute{
				Description: "Name of the registry. Valid options: \"public\", \"private\". Defaults to \"private\".",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.PrivateRegistry),
						string(tfe.PublicRegistry),
					),
				},
			},
			"module_provider": schema.StringAttribute{
				Description: "Name of the module provider.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the no-code module. Uses organization name if not provided.",
				Optional:    true,
				Computed:    true,
			},
			"no_code_module_id": schema.StringAttribute{
				Description: "ID of the no-code module.",
				Computed:    true,
			},
			"no_code_module_source": schema.StringAttribute{
				Description: "Source value of the no-code module.",
				Computed:    true,
			},
			"no_code": schema.BoolAttribute{
				Description: "Whether or not this is a no-code module.",
				Computed:    true,
			},
			"publishing_mechanism": schema.StringAttribute{
				Description: "The publishing mechanism of the module.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the module.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The time when the modules was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The time when the modules was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.ListNestedBlock{
				Description: "The permissions for this module.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"can_delete": schema.BoolAttribute{
							Description: "Whether or not this is a delete permission.",
							Computed:    true,
						},
						"can_resync": schema.BoolAttribute{
							Description: "Whether or not this is a resync permission.",
							Computed:    true,
						},
						"can_retry": schema.BoolAttribute{
							Description: "Whether or not this is a retry permission.",
							Computed:    true,
						},
					},
				},
			},
			"test_config": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tests_enabled": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"version_statuses": schema.ListNestedBlock{
				Description: "The status of each version of this module.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							Description: "The version of this module.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"error": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"vcs_repo": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"branch": schema.StringAttribute{
							Computed: true,
						},
						"display_identifier": schema.StringAttribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"ingress_submodules": schema.BoolAttribute{
							Computed: true,
						},
						"oauth_token_id": schema.StringAttribute{
							Computed: true,
						},
						"github_app_installation_id": schema.StringAttribute{
							Computed: true,
						},
						"repository_http_url": schema.StringAttribute{
							Computed: true,
						},
						"service_provider": schema.StringAttribute{
							Computed: true,
						},
						"tags": schema.BoolAttribute{
							Computed: true,
						},
						"tags_regex": schema.StringAttribute{
							Computed: true,
						},
						"webhook_url": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataSourceTFERegistryModule) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe modules, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	d.config = client
}

// Read refreshes the Terraform state with the latest data.
func (d *dataSourceTFERegistryModule) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data modelRegistryModule

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the namespace is not provided, use the organization name
	var namespace types.String
	if data.Namespace.IsNull() {
		namespace = data.Organization
	} else {
		namespace = data.Namespace
	}

	// defaults to private registry
	var registryName types.String
	if data.RegistryName.IsNull() {
		registryName = types.StringValue(string(tfe.PrivateRegistry))
	} else {
		registryName = data.RegistryName
	}

	rmID := tfe.RegistryModuleID{
		Organization: data.Organization.ValueString(),
		Name:         data.Name.ValueString(),
		Provider:     data.ModuleProvider.ValueString(),
		Namespace:    namespace.ValueString(),
		RegistryName: tfe.RegistryName(registryName.ValueString()),
	}

	tflog.Debug(ctx, "Reading module")
	module, err := d.config.Client.RegistryModules.Read(ctx, rmID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read module", err.Error())
		return
	}

	data = modelFromTFERegistryModule(module)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func modelFromTFERegistryModule(v *tfe.RegistryModule) modelRegistryModule {
	m := modelRegistryModule{
		ID:                  types.StringValue(v.ID),
		Organization:        types.StringValue(v.Organization.Name),
		Name:                types.StringValue(v.Name),
		Namespace:           types.StringValue(v.Namespace),
		RegistryName:        types.StringValue(string(v.RegistryName)),
		ModuleProvider:      types.StringValue(v.Provider),
		NoCode:              types.BoolValue(v.NoCode),
		PublishingMechanism: types.StringValue(string(v.PublishingMechanism)),
		Status:              types.StringValue(string(v.Status)),
		CreatedAt:           types.StringValue(v.CreatedAt),
		UpdatedAt:           types.StringValue(v.UpdatedAt),
	}
	for _, s := range v.VersionStatuses {
		m.VersionStatuses = append(m.VersionStatuses, modelFromTFERegistryModuleVersionStatuses(&s))
	}
	if v.TestConfig != nil {
		m.TestConfig = append(m.TestConfig, modelFromTFETestConfig(v.TestConfig))
	}
	if v.Permissions != nil {
		m.Permissions = append(m.Permissions, modelFromTFERegistryModulePermission(v.Permissions))
	}
	if v.VCSRepo != nil {
		m.VCSRepo = append(m.VCSRepo, modelFromTFEVCSRepo(v.VCSRepo))
	}
	// Only valid options are no RegistryNoCodeModule or a single entry
	if v.RegistryNoCodeModule != nil {
		m.NoCodeModuleID = types.StringValue(v.RegistryNoCodeModule[0].ID)
		m.NoCodeModuleSource = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", v.RegistryName, v.Namespace, v.Name, v.Provider))
	}
	return m
}
