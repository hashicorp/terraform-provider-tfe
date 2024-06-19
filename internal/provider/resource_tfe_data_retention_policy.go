package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/numbervalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"log"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFEDataRetentionPolicy{}
var _ resource.ResourceWithConfigure = &resourceTFEDataRetentionPolicy{}
var _ resource.ResourceWithImportState = &resourceTFEDataRetentionPolicy{}

func NewDataRetentionPolicyResource() resource.Resource {
	return &resourceTFEDataRetentionPolicy{}
}

// resourceTFEDataRetentionPolicy implements the tfe_data_retention_policy resource type
type resourceTFEDataRetentionPolicy struct {
	config ConfiguredClient
}

func (r *resourceTFEDataRetentionPolicy) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_retention_policy"
}

func (r *resourceTFEDataRetentionPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the data retention policies for a specific workspace or an the entire organization.",
		Version:     1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Data Retention Policy.",
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
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Description: "ID of the workspace that the data retention policy should apply to. If omitted, the data retention policy will apply to the entire organization.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				//Validators: []validator.String{
				//	stringvalidator.ExactlyOneOf(
				//		path.MatchRelative().AtParent().AtName("organization"),
				//	),
				//},
			},
		},
		Blocks: map[string]schema.Block{
			"delete_older_than": schema.SingleNestedBlock{
				Description: "Sets the maximum number of days, months, years data is allowed to exist before it is scheduled for deletion. Cannot be configured if the dont_delete attribute is also configured.",
				Attributes: map[string]schema.Attribute{
					"days": schema.NumberAttribute{
						Description: "Number of days",
						Optional:    true,
						PlanModifiers: []planmodifier.Number{
							numberplanmodifier.RequiresReplace(),
						},
						Validators: []validator.Number{
							numbervalidator.ExactlyOneOf(
								path.MatchRelative().AtParent().AtParent().AtName("dont_delete"),
							),
						},
					},
				},
			},
			"dont_delete": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("delete_older_than"),
					),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEDataRetentionPolicy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEDataRetentionPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEDataRetentionPolicy

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.ensureOrganizationIsSet(ctx, &plan, req.Plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeleteOlderThan.IsNull() {
		r.createDeleteOlderThanRetentionPolicy(ctx, plan, resp)
		return
	}

	if !plan.DontDelete.IsNull() {
		r.createDontDeleteRetentionPolicy(ctx, plan, resp)
		return
	}
}

func (r *resourceTFEDataRetentionPolicy) ensureOrganizationIsSet(ctx context.Context, model *modelTFEDataRetentionPolicy, data AttrGettable, diags *diag.Diagnostics) {
	if !model.Organization.IsUnknown() || model.Organization.ValueString() != "" {
		// skip this method if the organization has already been set
		return
	}

	if model.WorkspaceID.IsNull() {
		var organization string
		diags.Append(r.config.dataOrDefaultOrganization(ctx, data, &organization)...)
		model.Organization = types.StringValue(organization)
	}
}

func (r *resourceTFEDataRetentionPolicy) createDeleteOlderThanRetentionPolicy(ctx context.Context, plan modelTFEDataRetentionPolicy, resp *resource.CreateResponse) {
	deleteOlderThan := &modelTFEDeleteOlderThan{}

	diags := plan.DeleteOlderThan.As(ctx, &deleteOlderThan, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	deleteOlderThanDays, _ := deleteOlderThan.Days.ValueBigFloat().Int64()
	options := tfe.DataRetentionPolicyDeleteOlderSetOptions{
		DeleteOlderThanNDays: int(deleteOlderThanDays),
	}

	tflog.Debug(ctx, "Creating data retention policy")
	var dataRetentionPolicy *tfe.DataRetentionPolicyDeleteOlder
	var err error
	if plan.WorkspaceID.IsNull() {
		dataRetentionPolicy, err = r.config.Client.Organizations.SetDataRetentionPolicyDeleteOlder(ctx, plan.Organization.ValueString(), options)
	} else {
		dataRetentionPolicy, err = r.config.Client.Workspaces.SetDataRetentionPolicyDeleteOlder(ctx, plan.WorkspaceID.ValueString(), options)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", err.Error())
		return
	}

	result, diags := modelFromTFEDataRetentionPolicyDeleteOlder(ctx, plan, dataRetentionPolicy)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// set organization if it is still not known after creating the data retention policy
	r.ensureOrganizationSetAfterApply(&result, &resp.Diagnostics)

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceTFEDataRetentionPolicy) createDontDeleteRetentionPolicy(ctx context.Context, plan modelTFEDataRetentionPolicy, resp *resource.CreateResponse) {
	deleteOlderThan := &modelTFEDeleteOlderThan{}

	diags := plan.DeleteOlderThan.As(ctx, &deleteOlderThan, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	options := tfe.DataRetentionPolicyDontDeleteSetOptions{}

	tflog.Debug(ctx, "Creating data retention policy")
	var dataRetentionPolicy *tfe.DataRetentionPolicyDontDelete
	var err error
	if plan.WorkspaceID.IsNull() {
		dataRetentionPolicy, err = r.config.Client.Organizations.SetDataRetentionPolicyDontDelete(ctx, plan.Organization.ValueString(), options)
	} else {
		dataRetentionPolicy, err = r.config.Client.Workspaces.SetDataRetentionPolicyDontDelete(ctx, plan.WorkspaceID.ValueString(), options)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", err.Error())
		return
	}

	result := modelFromTFEDataRetentionPolicyDontDelete(plan, dataRetentionPolicy)

	// set organization if it is still not known after creating the data retention policy
	r.ensureOrganizationSetAfterApply(&result, &resp.Diagnostics)

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceTFEDataRetentionPolicy) ensureOrganizationSetAfterApply(policy *modelTFEDataRetentionPolicy, diags *diag.Diagnostics) {
	if policy.Organization.IsUnknown() {
		workspace, err := r.config.Client.Workspaces.ReadByID(ctx, policy.WorkspaceID.ValueString())
		if err != nil {
			diags.AddError("Unable to create data retention policy", err.Error())
			return
		}
		policy.Organization = types.StringValue(workspace.Organization.Name)
	}
}

func (r *resourceTFEDataRetentionPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEDataRetentionPolicy

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var policy *tfe.DataRetentionPolicyChoice
	var err error
	if state.WorkspaceID.IsNull() {
		policy, err = r.config.Client.Organizations.ReadDataRetentionPolicyChoice(ctx, state.Organization.ValueString())
	} else {
		policy, err = r.config.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, state.WorkspaceID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read data retention policy", err.Error())
		return
	}
	// remove the policy from state if it no longer exists or has been replaced by another policy
	if policy == nil || r.getPolicyID(policy) != state.ID.ValueString() {
		log.Printf("[DEBUG] Data retention policy %s no longer exists", state.ID)
		resp.State.RemoveResource(ctx)
		return
	}
	result, diags := modelFromTFEDataRetentionPolicyChoice(ctx, state, policy)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEDataRetentionPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// If the resource does not support modification and should always be recreated on
	// configuration value updates, the Update logic can be left empty and ensure all
	// configurable schema attributes implement the resource.RequiresReplace()
	// attribute plan modifier.
	resp.Diagnostics.AddError("Update not supported", "The update operation is not supported on this resource. This is a bug in the provider.")
}

func (r *resourceTFEDataRetentionPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEDataRetentionPolicy

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.WorkspaceID.IsNull() {
		tflog.Debug(ctx, fmt.Sprintf("Deleting data retention policy for organization: %s", state.Organization))
		err := r.config.Client.Organizations.DeleteDataRetentionPolicy(ctx, state.Organization.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Deleting data retention policy for organization: %s", state.Organization), err.Error())
			return
		}
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Deleting data retention policy for workspace: %s", state.WorkspaceID))
		err := r.config.Client.Workspaces.DeleteDataRetentionPolicy(ctx, state.WorkspaceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Deleting data retention policy for workspace: %s", state.WorkspaceID), err.Error())
			return
		}
	}
}

func (r *resourceTFEDataRetentionPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := strings.Split(req.ID, "/")
	if len(s) >= 3 || len(s) == 0 {
		resp.Diagnostics.AddError("Error importing workspace settings", fmt.Sprintf(
			"invalid workspace input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME> or <ORGANIZATION>)",
			req.ID,
		))
		return
	}

	if len(s) == 2 {
		workspaceID, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], r.config.Client)
		if err != nil {
			resp.Diagnostics.AddError("Error importing data retention policy", fmt.Sprintf(
				"error retrieving workspace with name %s from organization %s: %s", s[1], s[0], err.Error(),
			))
			return
		}

		policy, err := r.config.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, workspaceID)
		if err != nil {
			resp.Diagnostics.AddError("Error importing data retention policy", fmt.Sprintf(
				"error retrieving data policy for workspace %s from organization %s: %s", s[1], s[0], err.Error(),
			))
			return
		}

		req.ID = r.getPolicyID(policy)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), r.getPolicyID(policy))...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), s[0])...)
		return
	}

	policy, err := r.config.Client.Organizations.ReadDataRetentionPolicyChoice(ctx, s[0])
	if err != nil {
		resp.Diagnostics.AddError("Error importing data retention policy", fmt.Sprintf(
			"error retrieving data policy for organization %s: %s", s[0], err.Error(),
		))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), r.getPolicyID(policy))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), s[0])...)
}

func (r *resourceTFEDataRetentionPolicy) getPolicyID(policy *tfe.DataRetentionPolicyChoice) string {
	if policy.DataRetentionPolicyDeleteOlder != nil {
		return policy.DataRetentionPolicyDeleteOlder.ID
	}

	if policy.DataRetentionPolicyDontDelete != nil {
		return policy.DataRetentionPolicyDontDelete.ID
	}

	return policy.ConvertToLegacyStruct().ID
}
