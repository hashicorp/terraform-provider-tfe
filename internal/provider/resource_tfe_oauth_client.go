// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// Compile-time proof of interface implementation.
	_ resource.Resource                   = &resourceTFEOAuthClient{}
	_ resource.ResourceWithConfigure      = &resourceTFEOAuthClient{}
	_ resource.ResourceWithValidateConfig = &resourceTFEOAuthClient{}
)

func NewOAuthClient() resource.Resource {
	return &resourceTFEOAuthClient{}
}

type resourceTFEOAuthClient struct {
	config ConfiguredClient
}

type modelTFEOAuthClient struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Organization       types.String `tfsdk:"organization"`
	APIURL             types.String `tfsdk:"api_url"`
	HTTPURL            types.String `tfsdk:"http_url"`
	Key                types.String `tfsdk:"key"`
	OAuthToken         types.String `tfsdk:"oauth_token"`
	PrivateKey         types.String `tfsdk:"private_key"`
	Secret             types.String `tfsdk:"secret"`
	RSAPublicKey       types.String `tfsdk:"rsa_public_key"`
	ServiceProvider    types.String `tfsdk:"service_provider"`
	OAuthTokenID       types.String `tfsdk:"oauth_token_id"`
	AgentPoolID        types.String `tfsdk:"agent_pool_id"`
	OrganizationScoped types.Bool   `tfsdk:"organization_scoped"`
}

func modelFromTFEOAuthClient(c *tfe.OAuthClient, lastValues map[string]types.String) (*modelTFEOAuthClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := modelTFEOAuthClient{
		ID:                 types.StringValue(c.ID),
		Name:               types.StringPointerValue(c.Name),
		Organization:       types.StringValue(c.Organization.Name),
		APIURL:             types.StringValue(c.APIURL),
		HTTPURL:            types.StringValue(c.HTTPURL),
		ServiceProvider:    types.StringValue(string(c.ServiceProvider)),
		OrganizationScoped: types.BoolPointerValue(c.OrganizationScoped),
	}

	if oauthToken, ok := lastValues["oauth_token"]; ok {
		m.OAuthToken = oauthToken
	}

	if privateKey, ok := lastValues["private_key"]; ok {
		m.PrivateKey = privateKey
	}

	if key, ok := lastValues["key"]; ok {
		m.Key = key
	}

	if secret, ok := lastValues["secret"]; ok {
		m.Secret = secret
	}

	if c.AgentPool != nil {
		m.AgentPoolID = types.StringValue(c.AgentPool.ID)
	}

	if len(c.RSAPublicKey) > 0 {
		m.RSAPublicKey = types.StringValue(c.RSAPublicKey)
	}

	if len(c.RSAPublicKey) > 0 {
		m.RSAPublicKey = types.StringValue(c.RSAPublicKey)
	}

	switch len(c.OAuthTokens) {
	case 0:
		m.OAuthTokenID = types.StringValue("")
	case 1:
		m.OAuthTokenID = types.StringValue(c.OAuthTokens[0].ID)
	default:
		diags.AddError("Error parsing API result", fmt.Sprintf("unexpected number of OAuth tokens: %d", len(c.OAuthTokens)))
	}

	return &m, diags
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEOAuthClient) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
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

// Metadata implements resource.Resource
func (r *resourceTFEOAuthClient) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_client"
}

// Schema implements resource.Resource
func (r *resourceTFEOAuthClient) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-generated identifier for the variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Display name for the OAuth Client",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"organization": schema.StringAttribute{
				Description: "Name of the organization",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"api_url": schema.StringAttribute{
				Description: "The base URL of the VCS provider's API",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"http_url": schema.StringAttribute{
				Description: "The homepage of the VCS provider",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"key": schema.StringAttribute{
				Description: "The OAuth Client key can refer to a Consumer Key, Application Key, or another type of client key for the VCS provider",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"oauth_token": schema.StringAttribute{
				Description: "The OAuth token string for the VCS provider",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"private_key": schema.StringAttribute{
				Description: "The text of the private key associated with a Azure DevOps Server account",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"secret": schema.StringAttribute{
				Description: "The text of the SSH private key associated with a Bitbucket Data Center Application Link",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"rsa_public_key": schema.StringAttribute{
				Description: "The text of the SSH public key associated with a Bitbucket Data Center Application Link",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"service_provider": schema.StringAttribute{
				Description: "The VCS provider being connected with",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.ServiceProviderGithub),
						string(tfe.ServiceProviderGithubEE),
						string(tfe.ServiceProviderGitlab),
						string(tfe.ServiceProviderGitlabCE),
						string(tfe.ServiceProviderGitlabEE),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderBitbucketServer),
						string(tfe.ServiceProviderBitbucketServerLegacy),
						string(tfe.ServiceProviderBitbucketDataCenter),
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
					),
				},
			},

			"oauth_token_id": schema.StringAttribute{
				Description: "OAuth Token ID for the OAuth Client",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"agent_pool_id": schema.StringAttribute{
				Description: "An existing agent pool ID within the organization that has Private VCS support enabled",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"organization_scoped": schema.BoolAttribute{
				Description: "Whether or not the oauth client is scoped to all projects and workspaces in the organization",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEOAuthClient) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Load the plan into the model
	var plan modelTFEOAuthClient
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new options struct.
	// The tfe.OAuthClientCreateOptions has omitempty for these values, so if it
	// is empty, then it will be ignored in the create request
	options := tfe.OAuthClientCreateOptions{
		Name:               plan.Name.ValueStringPointer(),
		APIURL:             plan.APIURL.ValueStringPointer(),
		HTTPURL:            plan.HTTPURL.ValueStringPointer(),
		OAuthToken:         plan.OAuthToken.ValueStringPointer(),
		Key:                plan.Key.ValueStringPointer(),
		ServiceProvider:    tfe.ServiceProvider(tfe.ServiceProviderType(plan.ServiceProvider.ValueString())),
		OrganizationScoped: plan.OrganizationScoped.ValueBoolPointer(),
	}

	serviceProviderType := tfe.ServiceProviderType(plan.ServiceProvider.ValueString())

	if serviceProviderType == tfe.ServiceProviderAzureDevOpsServer {
		options.PrivateKey = plan.PrivateKey.ValueStringPointer()
	}

	if serviceProviderType == tfe.ServiceProviderBitbucketServer || serviceProviderType == tfe.ServiceProviderBitbucketDataCenter {
		options.RSAPublicKey = plan.RSAPublicKey.ValueStringPointer()
		options.Secret = plan.Secret.ValueStringPointer()
	}

	if serviceProviderType == tfe.ServiceProviderBitbucket {
		options.Secret = plan.Secret.ValueStringPointer()
	}

	if !plan.AgentPoolID.IsNull() {
		options.AgentPool = &tfe.AgentPool{ID: plan.AgentPoolID.ValueString()}
	}

	tflog.Debug(ctx, fmt.Sprintf("Create an OAuth client for organization: %s", organization))
	oc, err := r.config.Client.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating OAuth client", err.Error())
		return
	}

	lastValues := map[string]types.String{
		"oauth_token":    plan.OAuthToken,
		"rsa_public_key": plan.RSAPublicKey,
		"key":            plan.Key,
		"secret":         plan.Secret,
	}

	// Load the result into the model
	result, diags := modelFromTFEOAuthClient(oc, lastValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read implements resource.Resource
func (r *resourceTFEOAuthClient) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the state into the model
	var state modelTFEOAuthClient
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// Read the OAuth client
	tflog.Debug(ctx, fmt.Sprintf("Read OAuth client: %s", id))
	oc, err := r.config.Client.OAuthClients.Read(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("OAuth client %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OAuth client", err.Error())
		return
	}

	lastValues := map[string]types.String{
		"oauth_token":    state.OAuthToken,
		"rsa_public_key": state.RSAPublicKey,
		"key":            state.Key,
		"secret":         state.Secret,
	}

	// Load the result into the model
	result, diags := modelFromTFEOAuthClient(oc, lastValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Update implements resource.Resource
func (r *resourceTFEOAuthClient) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Load the plan and state into the models
	var plan, state modelTFEOAuthClient
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new options struct.
	options := tfe.OAuthClientUpdateOptions{
		OrganizationScoped: plan.OrganizationScoped.ValueBoolPointer(),
		OAuthToken:         plan.OAuthToken.ValueStringPointer(),
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Update OAuth client: %s", id))
	oc, err := r.config.Client.OAuthClients.Update(ctx, id, options)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("OAuth client %s no longer exists", id))
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error updating OAuth client", err.Error())
		return
	}

	lastValues := map[string]types.String{
		"oauth_token":    plan.OAuthToken,
		"rsa_public_key": plan.RSAPublicKey,
		"key":            plan.Key,
		"secret":         plan.Secret,
	}

	// Load the result into the model
	result, diags := modelFromTFEOAuthClient(oc, lastValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete implements resource.Resource
func (r *resourceTFEOAuthClient) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load the state into the model
	var state modelTFEOAuthClient
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Delete OAuth client: %s", id))
	err := r.config.Client.OAuthClients.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("OAuth client %s no longer exists", id))
			// The resource will implicitly be removed from state on return
			return
		}

		resp.Diagnostics.AddError("Error deleting OAuth client", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceTFEOAuthClient) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Load the config into the model
	var config modelTFEOAuthClient
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceProviderType := tfe.ServiceProviderType(config.ServiceProvider.ValueString())
	if serviceProviderType == tfe.ServiceProviderAzureDevOpsServer &&
		config.PrivateKey.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key"),
			"Invalid configuration",
			fmt.Sprintf("private_key is required for service_provider %s", serviceProviderType))
		return
	}
}
