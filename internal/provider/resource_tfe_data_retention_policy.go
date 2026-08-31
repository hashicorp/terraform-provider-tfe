// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const minTFEVersionGranularDRP = "v2.1.0"

var _ resource.Resource = &resourceTFEDataRetentionPolicy{}
var _ resource.ResourceWithConfigure = &resourceTFEDataRetentionPolicy{}
var _ resource.ResourceWithImportState = &resourceTFEDataRetentionPolicy{}

func NewDataRetentionPolicyResource() resource.Resource {
	return &resourceTFEDataRetentionPolicy{}
}

type resourceTFEDataRetentionPolicy struct {
	config ConfiguredClient
}

func (r *resourceTFEDataRetentionPolicy) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_retention_policy"
}

func (r *resourceTFEDataRetentionPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "(Only for Terraform Enterprise) Manages a data retention policy attached to either an organization or workspace.",
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
				Description: "The name of the organization the policy will apply to. Must not be set if `workspace_id` is set.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Description: "The ID of the workspace that the data retention policy should apply to. If omitted, the data retention policy will apply to the entire organization. Must not be set if `organization` is set.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("organization"),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"delete_older_than": schema.SingleNestedBlock{
				Description: "Sets the maximum number of days data is allowed to exist before it is scheduled for deletion. Cannot be configured if the `dont_delete` attribute is also configured.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = req.StateValue.IsNull() != req.PlanValue.IsNull()
						},
						"Requires replace only when the delete_older_than block is added or removed, not when its attributes change.",
						"Requires replace only when the `delete_older_than` block is added or removed, not when its attributes change.",
					),
				},
				Attributes: map[string]schema.Attribute{
					"days": schema.NumberAttribute{
						Description:        "Number of days old data must be before it is scheduled for deletion. Used as the global window when per-artifact-type windows are not set. Deprecated for TFE v2.1.0+: use per-artifact-type fields instead.",
						DeprecationMessage: "The days field is deprecated for TFE v2.1.0+. Use state_versions_delete_after_n_days, configuration_versions_delete_after_n_days, and run_data_and_logs_delete_after_n_days instead.",
						Optional:           true,
					},
					"delete_state_versions": schema.BoolAttribute{
						Description: "When true, state versions are eligible for deletion under this policy.",
						Optional:    true,
					},
					"delete_configuration_versions": schema.BoolAttribute{
						Description: "When true, configuration versions are eligible for deletion under this policy.",
						Optional:    true,
					},
					"delete_run_data_and_logs": schema.BoolAttribute{
						Description: "When true, run data and logs (plans, applies, assessments) are eligible for deletion under this policy.",
						Optional:    true,
					},
					"state_versions_delete_after_n_days": schema.NumberAttribute{
						Description: "Number of days after which state versions are eligible for deletion. Requires `delete_state_versions` to be true.",
						Optional:    true,
					},
					"configuration_versions_delete_after_n_days": schema.NumberAttribute{
						Description: "Number of days after which configuration versions are eligible for deletion. Requires `delete_configuration_versions` to be true.",
						Optional:    true,
					},
					"run_data_and_logs_delete_after_n_days": schema.NumberAttribute{
						Description: "Number of days after which run data and logs are eligible for deletion. Requires `delete_run_data_and_logs` to be true.",
						Optional:    true,
					},
					"state_versions_keep_latest_count": schema.NumberAttribute{
						Description: "Minimum number of state versions to keep per workspace, regardless of age. Requires `delete_state_versions` to be true.",
						Optional:    true,
					},
					"configuration_versions_keep_latest_count": schema.NumberAttribute{
						Description: "Minimum number of configuration versions to keep per workspace, regardless of age. Requires `delete_configuration_versions` to be true.",
						Optional:    true,
					},
					"run_data_keep_latest_count": schema.NumberAttribute{
						Description: "Minimum number of runs (and associated plan/apply data) to keep per workspace, regardless of age. Requires `delete_run_data_and_logs` to be true.",
						Optional:    true,
					},
				},
				Validators: []validator.Object{
					ValidateDeleteOlderThan(),
				},
			},
			"dont_delete": schema.SingleNestedBlock{
				Description: "If this block is set, the created policy will prevent other policies from deleting data from this workspace or organization. Must not be set if `delete_older_than` is set. Note that the empty nested block schema is not an error.",
				Attributes:  map[string]schema.Attribute{},
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

func (r *resourceTFEDataRetentionPolicy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	if !model.Organization.IsUnknown() && model.Organization.ValueString() != "" {
		return
	}

	if model.WorkspaceID.IsNull() {
		var organization string
		diags.Append(r.config.dataOrDefaultOrganization(ctx, data, &organization)...)
		model.Organization = types.StringValue(organization)
	}
}

// patchPolicy sends a PATCH request and returns the resulting policy envelope.
// For org-level policies the endpoint returns no body, so a follow-up GET is performed.
func (r *resourceTFEDataRetentionPolicy) patchPolicy(ctx context.Context, plan modelTFEDataRetentionPolicy, requestEnvelope models.DataRetentionPolicyEnvelopeable) (models.DataRetentionPolicyEnvelopeable, error) {
	if plan.WorkspaceID.IsNull() {
		if err := r.config.ClientV2.API.Organizations().ByOrganization_name(plan.Organization.ValueString()).Relationships().DataRetentionPolicy().Patch(ctx, requestEnvelope, nil); err != nil {
			return nil, err
		}
		return r.config.ClientV2.API.Organizations().ByOrganization_name(plan.Organization.ValueString()).Relationships().DataRetentionPolicy().Get(ctx, nil)
	}
	return r.config.ClientV2.API.Workspaces().ByWorkspace_id(plan.WorkspaceID.ValueString()).Relationships().DataRetentionPolicy().Patch(ctx, requestEnvelope, nil)
}

func (r *resourceTFEDataRetentionPolicy) createDeleteOlderThanRetentionPolicy(ctx context.Context, plan modelTFEDataRetentionPolicy, resp *resource.CreateResponse) {
	deleteOlderThan := &modelTFEDeleteOlderThan{}
	resp.Diagnostics.Append(plan.DeleteOlderThan.As(ctx, deleteOlderThan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.warnIfDaysDeprecated(deleteOlderThan, &resp.Diagnostics)

	requestEnvelope, err := newDeleteOlderEnvelope(*deleteOlderThan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build data retention policy request", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating data retention policy")
	responseEnvelope, err := r.patchPolicy(ctx, plan, requestEnvelope)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", err.Error())
		return
	}
	if responseEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", "API returned empty response")
		return
	}

	result, diags := deleteOlderThanFromAPIResponse(ctx, plan.Organization, plan.WorkspaceID, plan.DeleteOlderThan, responseEnvelope.GetData())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEDataRetentionPolicy) createDontDeleteRetentionPolicy(ctx context.Context, plan modelTFEDataRetentionPolicy, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating data retention policy")
	responseEnvelope, err := r.patchPolicy(ctx, plan, newDontDeleteEnvelope())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", err.Error())
		return
	}
	if responseEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create data retention policy", "API returned empty response")
		return
	}
	result := dontDeleteFromAPIResponse(plan.Organization, plan.WorkspaceID, responseEnvelope.GetData())
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEDataRetentionPolicy) warnIfDaysDeprecated(dot *modelTFEDeleteOlderThan, diags *diag.Diagnostics) {
	// hasGranularFields is checked against the plan value (not prior state). This is safe because
	// the validator enforces that `days` and granular fields are mutually exclusive at plan time,
	// so if any granular fields are set in the plan, `days` cannot be set. The two paths can
	// never coexist in a valid plan, making the plan value a reliable signal here.
	if dot.Days.IsNull() || hasGranularFields(*dot) {
		return
	}
	meets, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionGranularDRP)
	if err != nil {
		log.Printf("[DEBUG] could not determine if TFE version meets minimum required version %s: %v", minTFEVersionGranularDRP, err)
		return
	}
	if meets {
		diags.AddWarning(
			"days field is deprecated for TFE v2.1.0+",
			fmt.Sprintf(
				"The days field is supported for backwards compatibility with TFE versions prior to %s. "+
					"Your TFE version (%s) supports per-artifact-type retention fields. "+
					"Consider migrating to state_versions_delete_after_n_days, configuration_versions_delete_after_n_days, and run_data_and_logs_delete_after_n_days.",
				minTFEVersionGranularDRP, r.config.RemoteTFEVersion(),
			),
		)
	}
}

func (r *resourceTFEDataRetentionPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEDataRetentionPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var responseEnvelope models.DataRetentionPolicyEnvelopeable
	var err error
	if state.WorkspaceID.IsNull() {
		responseEnvelope, err = r.config.ClientV2.API.Organizations().ByOrganization_name(state.Organization.ValueString()).Relationships().DataRetentionPolicy().Get(ctx, nil)
	} else {
		responseEnvelope, err = r.config.ClientV2.API.Workspaces().ByWorkspace_id(state.WorkspaceID.ValueString()).Relationships().DataRetentionPolicy().Get(ctx, nil)
	}
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Data retention policy no longer exists")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read data retention policy", err.Error())
		return
	}

	policy := responseEnvelope.GetData()
	if policy == nil {
		log.Printf("[DEBUG] Data retention policy %s no longer exists", state.ID)
		resp.State.RemoveResource(ctx)
		return
	}

	policyID, err := getPolicyIDFromV2(policy)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read data retention policy", err.Error())
		return
	}
	if policyID != state.ID.ValueString() {
		log.Printf("[DEBUG] Data retention policy %s has been replaced (new ID: %s)", state.ID, policyID)
		resp.State.RemoveResource(ctx)
		return
	}

	result, diags := dataRetentionPolicyFromAPIResponse(ctx, state, policy)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEDataRetentionPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEDataRetentionPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ensureOrganizationIsSet(ctx, &plan, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeleteOlderThan.IsNull() {
		deleteOlderThan := &modelTFEDeleteOlderThan{}
		resp.Diagnostics.Append(plan.DeleteOlderThan.As(ctx, deleteOlderThan, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		r.warnIfDaysDeprecated(deleteOlderThan, &resp.Diagnostics)

		requestEnvelope, err := newDeleteOlderEnvelope(*deleteOlderThan)
		if err != nil {
			resp.Diagnostics.AddError("Unable to build data retention policy request", err.Error())
			return
		}

		tflog.Debug(ctx, "Updating data retention policy")
		responseEnvelope, err := r.patchPolicy(ctx, plan, requestEnvelope)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update data retention policy", err.Error())
			return
		}
		if responseEnvelope.GetData() == nil {
			resp.Diagnostics.AddError("Unable to update data retention policy", "API returned empty response")
			return
		}

		result, diags := deleteOlderThanFromAPIResponse(ctx, plan.Organization, plan.WorkspaceID, plan.DeleteOlderThan, responseEnvelope.GetData())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	if !plan.DontDelete.IsNull() {
		tflog.Debug(ctx, "Updating data retention policy")
		responseEnvelope, err := r.patchPolicy(ctx, plan, newDontDeleteEnvelope())
		if err != nil {
			resp.Diagnostics.AddError("Unable to update data retention policy", err.Error())
			return
		}
		if responseEnvelope.GetData() == nil {
			resp.Diagnostics.AddError("Unable to update data retention policy", "API returned empty response")
			return
		}
		result := dontDeleteFromAPIResponse(plan.Organization, plan.WorkspaceID, responseEnvelope.GetData())
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}

func (r *resourceTFEDataRetentionPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEDataRetentionPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.WorkspaceID.IsNull() {
		tflog.Debug(ctx, fmt.Sprintf("Deleting data retention policy for organization: %s", state.Organization))
		err := r.config.ClientV2.API.Organizations().ByOrganization_name(state.Organization.ValueString()).Relationships().DataRetentionPolicy().Delete(ctx, nil)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Deleting data retention policy for organization: %s", state.Organization), err.Error())
			return
		}
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Deleting data retention policy for workspace: %s", state.WorkspaceID))
		err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(state.WorkspaceID.ValueString()).Relationships().DataRetentionPolicy().Delete(ctx, nil)
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

		wsEnvelope, err := r.config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).Relationships().DataRetentionPolicy().Get(ctx, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error importing data retention policy", fmt.Sprintf(
				"error retrieving data policy for workspace %s from organization %s: %s", s[1], s[0], err.Error(),
			))
			return
		}

		policyID, err := getPolicyIDFromV2(wsEnvelope.GetData())
		if err != nil {
			resp.Diagnostics.AddError("Error importing data retention policy", err.Error())
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
		return
	}

	orgEnvelope, err := r.config.ClientV2.API.Organizations().ByOrganization_name(s[0]).Relationships().DataRetentionPolicy().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing data retention policy", fmt.Sprintf(
			"error retrieving data policy for organization %s: %s", s[0], err.Error(),
		))
		return
	}

	policyID, err := getPolicyIDFromV2(orgEnvelope.GetData())
	if err != nil {
		resp.Diagnostics.AddError("Error importing data retention policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), s[0])...)
}
