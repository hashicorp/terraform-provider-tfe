package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Resource schema definition
type organizationTTLPolicyResource struct{}

type organizationTTLPolicyModel struct {
	Organization      types.String `tfsdk:"organization"`
	UserToken         types.String `tfsdk:"user_token"`
	OrganizationToken types.String `tfsdk:"organization_token"`
	TeamToken         types.String `tfsdk:"team_token"`
	AuditTrailToken   types.String `tfsdk:"audit_trail_token"`
}

func NewOrganizationTTLPolicyResource() resource.Resource {
	return &organizationTTLPolicyResource{}
}

func (r *organizationTTLPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_ttl_policy"
}

func (r *organizationTTLPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "The name of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_token_max_ttl": schema.StringAttribute{
				Description: "The maximum TTL allowed for usage of user tokens in the organization",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("2y"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_token_max_ttl": schema.StringAttribute{
				Description: "The maximum TTL allowed for usage of organization tokens in the organization",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("2y"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_token_max_ttl": schema.StringAttribute{
				Description: "The maximum TTL allowed for usage of team tokens in the organization",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("2y"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"audit_trail_token_max_ttl": schema.StringAttribute{
				Description: "The maximum TTL allowed for usage of audit trail tokens in the organization",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("2y"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create operation
func (r *organizationTTLPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationTTLPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// TODO: Call API to create the policy
	//plan.ID = types.StringValue(fmt.Sprintf("org-ttl-policy-%s", plan.UserToken.ValueString()))
	//resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read operation
func (r *organizationTTLPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationTTLPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// TODO: Call API to read the policy
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update operation
func (r *organizationTTLPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationTTLPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// TODO: Call API to update the policy
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete operation
func (r *organizationTTLPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationTTLPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// TODO: Call API to delete the policy
	// resp.Diagnostics.Append(resp.State.RemoveResource(ctx)...)
}
