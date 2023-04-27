package tfe

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceTFEVariable struct {
	config ConfiguredClient
}

// Configure implements resource.ResourceWithConfigure. TODO: dry this out for other rscs
func (r *resourceTFEVariable) Configure(c context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		res.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.config = client
}

// Metadata implements resource.Resource
func (r *resourceTFEVariable) Metadata(_ context.Context, _ resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = "tfe_variable"
}

// Schema implements resource.Resource
func (r *resourceTFEVariable) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Name of the variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, res *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var stateSensitive types.Bool
							diags := req.State.GetAttribute(ctx, path.Root("sensitive"), &stateSensitive)
							if diags.HasError() {
								res.Diagnostics.Append(diags...)
								return
							}
							if stateSensitive.ValueBool() && req.PlanValue.ValueString() != req.StateValue.ValueString() {
								res.RequiresReplace = true
							}
						},
						"Force replacement if key changed and sensitive is true",
						"Force replacement if key changed and sensitive is true",
					),
				},
			},
			"value": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Sensitive:   true,
				Description: "Value of the variable",
				// TODO: do descriptions cause a schema upgrade? how bout the rest of the stuff I'm doing here?
			},
			"category": schema.StringAttribute{
				Required:    true,
				Description: `Whether this is a Terraform or environment variable. Valid values are "terraform" or "env".`,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.CategoryEnv),
						string(tfe.CategoryTerraform),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Default:  stringdefault.StaticString(""),
			},
			"hcl": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"sensitive": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, res *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							if req.StateValue.ValueBool() && !req.ConfigValue.ValueBool() {
								res.RequiresReplace = true
							}
						},
						"Force replacement if sensitive argument changed from true to false.",
						"Force replacement if sensitive argument changed from true to false.",
					),
				},
			},
			"workspace_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("variable_set_id"),
						// TODO: double-check behavior and ensure it includes current attr in that list
					),
					stringvalidator.RegexMatches(
						workspaceIDRegexp,
						"must be a valid workspace ID (ws-<RANDOM STRING>)",
					),
				},
			},
			"variable_set_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("workspace_id"),
					),
					stringvalidator.RegexMatches(
						variableSetIDRegexp,
						"must be a valid variable set ID (varset-<RANDOM STRING>)",
					),
				},
			},
		},
		Description:         "",
		MarkdownDescription: "",
		DeprecationMessage:  "",
		Version:             0,
	}
}

// Create implements resource.Resource
func (r *resourceTFEVariable) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

// Delete implements resource.Resource
func (r *resourceTFEVariable) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}

// Read implements resource.Resource
func (r *resourceTFEVariable) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
	panic("unimplemented")
}

// Update implements resource.Resource
func (r *resourceTFEVariable) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}

// Compile-time interface check
var _ resource.ResourceWithConfigure = &resourceTFEVariable{}

// NewResourceVariable is a resource function for the framework provider.
func NewResourceVariable() resource.Resource {
	return &resourceTFEVariable{}
}
