package tfe

import (
	"context"
	"fmt"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	signatureMethodSHA1                     string = "SHA1"
	signatureMethodSHA256                   string = "SHA256"
	defaultSSOAPITokenSessionTimeoutSeconds int64  = 1209600 // 14 days
)

type resourceTFESAMLSettings struct {
	//config ConfiguredClient
	client *tfe.Client
}

// modelTFESAMLSettings maps the resource schema data to a struct.
type modelTFESAMLSettings struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Debug                     types.Bool   `tfsdk:"debug"`
	AuthnRequestsSigned       types.Bool   `tfsdk:"authn_requests_signed"`
	WantAssertionsSigned      types.Bool   `tfsdk:"want_assertions_signed"`
	TeamManagementEnabled     types.Bool   `tfsdk:"team_management_enabled"`
	OldIDPCert                types.String `tfsdk:"old_idp_cert"`
	IDPCert                   types.String `tfsdk:"idp_cert"`
	SLOEndpointURL            types.String `tfsdk:"slo_endpoint_url"`
	SSOEndpointURL            types.String `tfsdk:"sso_endpoint_url"`
	AttrUsername              types.String `tfsdk:"attr_username"`
	AttrGroups                types.String `tfsdk:"attr_groups"`
	AttrSiteAdmin             types.String `tfsdk:"attr_site_admin"`
	SiteAdminRole             types.String `tfsdk:"site_admin_role"`
	SSOAPITokenSessionTimeout types.Int64  `tfsdk:"sso_api_token_session_timeout"`
	ACSConsumerURL            types.String `tfsdk:"acs_consumer_url"`
	MetadataURL               types.String `tfsdk:"metadata_url"`
	Certificate               types.String `tfsdk:"certificate"`
	PrivateKey                types.String `tfsdk:"private_key"`
	SignatureSigningMethod    types.String `tfsdk:"signature_signing_method"`
	SignatureDigestMethod     types.String `tfsdk:"signature_digest_method"`
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESAMLSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client.Client
}

// Metadata implements resource.Resource
func (r *resourceTFESAMLSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_settings"
}

// Schema implements resource.Resource
func (r *resourceTFESAMLSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether or not to enable SAML single sign-on",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"debug": schema.BoolAttribute{
				Description: "When sign-on fails and this is enabled, the SAMLResponse XML will be displayed on the login page",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"authn_requests_signed": schema.BoolAttribute{
				Description: "Ensure that <samlp:AuthnRequest> messages are signed",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"want_assertions_signed": schema.BoolAttribute{
				Description: "Ensure that <saml:Assertion> elements are signed",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"team_management_enabled": schema.BoolAttribute{
				Description: "Set it to false if you would rather use Terraform Enterprise to manage team membership",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"old_idp_cert": schema.StringAttribute{
				Computed: true,
			},
			"idp_cert": schema.StringAttribute{
				Description: "Identity Provider Certificate specifies the PEM encoded X.509 Certificate as provided by the IdP configuration",
				Optional:    true,
				Computed:    true,
			},
			"slo_endpoint_url": schema.StringAttribute{
				Description: "Single Log Out URL specifies the HTTPS endpoint on your IdP for single logout requests. This value is provided by the IdP configuration",
				Optional:    true,
				Computed:    true,
			},
			"sso_endpoint_url": schema.StringAttribute{
				Description: "Single Sign On URL specifies the HTTPS endpoint on your IdP for single sign-on requests. This value is provided by the IdP configuration",
				Optional:    true,
				Computed:    true,
			},
			"attr_username": schema.StringAttribute{
				Description: "Username Attribute Name specifies the name of the SAML attribute that determines the user's username",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Username"),
			},
			"attr_site_admin": schema.StringAttribute{
				Description: "Specifies the role for site admin access. Overrides the \"Site Admin Role\" method",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SiteAdmin"),
			},
			"attr_groups": schema.StringAttribute{
				Description: "Team Attribute Name specifies the name of the SAML attribute that determines team membership",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("MemberOf"),
			},
			"site_admin_role": schema.StringAttribute{
				Description: "Specifies the role for site admin access, provided in the list of roles sent in the Team Attribute Name attribute",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("site-admins"),
			},
			"sso_api_token_session_timeout": schema.Int64Attribute{
				Description: "Specifies the Single Sign On session timeout in seconds. Defaults to 14 days",
				Optional:    true,
				Computed:    true,
				Default:     staticInt64(defaultSSOAPITokenSessionTimeoutSeconds),
			},
			"acs_consumer_url": schema.StringAttribute{
				Description: "ACS Consumer (Recipient) URL",
				Computed:    true,
			},
			"metadata_url": schema.StringAttribute{
				Description: "Metadata (Audience) URL",
				Computed:    true,
			},
			"certificate": schema.StringAttribute{
				Description: "The certificate used for request and assertion signing",
				Optional:    true,
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key used for request and assertion signing",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"signature_signing_method": schema.StringAttribute{
				Description: fmt.Sprintf("Signature Signing Method. Must be either `%s` or `%s`. Defaults to `%s`", signatureMethodSHA1, signatureMethodSHA256, signatureMethodSHA256),
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(signatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						signatureMethodSHA1,
						signatureMethodSHA256,
					),
				},
			},
			"signature_digest_method": schema.StringAttribute{
				Description: fmt.Sprintf("Signature Digest Method. Must be either `%s` or `%s`. Defaults to `%s`", signatureMethodSHA1, signatureMethodSHA256, signatureMethodSHA256),
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(signatureMethodSHA256),
				Validators: []validator.String{
					stringvalidator.OneOf(
						signatureMethodSHA1,
						signatureMethodSHA256,
					),
				},
			},
		},
		Version: 1,
	}
}

func (r *resourceTFESAMLSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceTFESAMLSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *resourceTFESAMLSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r resourceTFESAMLSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

var (
	_ resource.Resource              = &resourceTFESAMLSettings{}
	_ resource.ResourceWithConfigure = &resourceTFESAMLSettings{}
)

// NewSAMLSettingsResource is a resource function for the framework provider.
func NewSAMLSettingsResource() resource.Resource {
	return &resourceTFESAMLSettings{}
}
