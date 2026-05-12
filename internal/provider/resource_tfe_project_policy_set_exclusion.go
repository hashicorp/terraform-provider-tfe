// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFEProjectPolicySetExclusionParameter{}
	_ resource.ResourceWithConfigure   = &resourceTFEProjectPolicySetExclusionParameter{}
	_ resource.ResourceWithImportState = &resourceTFEProjectPolicySetExclusionParameter{}
)

type resourceTFEProjectPolicySetExclusionParameter struct {
	config ConfiguredClient
}

func NewProjectPolicySetExclusionResource() resource.Resource {
	return &resourceTFEProjectPolicySetExclusionParameter{}
}

type modelProjectPolicySetExclusionParameter struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	PolicySetID types.String `tfsdk:"policy_set_id"`
}

func (r *resourceTFEProjectPolicySetExclusionParameter) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	r.config = client
}

// Metadata implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_policy_set_exclusion"
}

// Schema implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource which manages the exclusion of a project from a policy set.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the project exclusion. This is a synthetic ID in the format <policy_set_id>_<project_id>.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "The ID of the policy set that will have an exclusion for the project",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^polset-[a-zA-Z0-9]{16}$`),
						"must be a valid policy set ID (e.g. polset-<RANDOM_STRING>)",
					),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project that will be excluded from the policy set",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelProjectPolicySetExclusionParameter
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating project exclusion for policy set %s", plan.PolicySetID.ValueString()))
	if ok, err := r.checkProjectExists(ctx, r.config.Client, plan.ProjectID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Checking if Project Exists",
			fmt.Sprintf("An error was encountered when checking if project %q exists: %s", plan.ProjectID.ValueString(), err),
		)
		return
	} else if !ok {
		resp.Diagnostics.AddError(
			"Project Not Found During Creation",
			fmt.Sprintf("The project with ID %q was not found. Please verify that the project ID is correct and that the project exists.", plan.ProjectID.ValueString()),
		)
		return
	}

	err := r.config.Client.PolicySets.AddProjectExclusions(ctx, plan.PolicySetID.ValueString(), tfe.PolicySetAddProjectExclusionsOptions{
		ProjectExclusions: []*tfe.Project{
			{
				ID: plan.ProjectID.ValueString(),
			},
		},
	})
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", plan.PolicySetID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Project Exclusion to Policy Set",
			fmt.Sprintf("An error was encountered when adding project exclusion %q to policy set %q: %s", plan.ProjectID.ValueString(), plan.PolicySetID.ValueString(), err),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.ProjectID.ValueString(), plan.PolicySetID.ValueString()))

	tflog.Debug(ctx, "Creation of project exclusion for policy set is complete", map[string]interface{}{
		"ID":            plan.ID.ValueString(),
		"policy_set_id": plan.PolicySetID.ValueString(),
		"project_id":    plan.ProjectID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Debug(ctx, "Failed to set state for project exclusion of policy set after creation", map[string]interface{}{
			"ID":            plan.ID.ValueString(),
			"policy_set_id": plan.PolicySetID.ValueString(),
			"project_id":    plan.ProjectID.ValueString(),
			"diagnostics":   resp.Diagnostics,
		},
		)
	}
}

// Read implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelProjectPolicySetExclusionParameter
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySet, err := r.config.Client.PolicySets.ReadWithOptions(ctx, state.PolicySetID.ValueString(), &tfe.PolicySetReadOptions{
		Include: []tfe.PolicySetIncludeOpt{
			tfe.PolicySetProjectExclusions,
		},
	})
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", state.PolicySetID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		tflog.Debug(ctx, "Policy set exclusion is not supported")
		resp.Diagnostics.AddError(
			"Policy Set Exclusion Not Supported",
			"The API operation to manage project exclusions on policy sets is not supported in the current version of Terraform Enterprise. Please upgrade to a newer version of Terraform Enterprise that supports this feature.",
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Policy Set",
			fmt.Sprintf("An error was encountered when reading policy set %q: %s", state.PolicySetID.ValueString(), err),
		)
		return
	}

	for _, excludedProject := range policySet.ProjectExclusions {
		if excludedProject.ID == state.ProjectID.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Project Exclusion Not Attached",
		fmt.Sprintf("The project exclusion with project ID %q was not found on policy set %q. It may have been removed outside of Terraform.", state.ProjectID.ValueString(), state.PolicySetID.ValueString()),
	)
	resp.State.RemoveResource(ctx)
}

// Update implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This method is a no-op but required by the framework
	var plan modelProjectPolicySetExclusionParameter

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements [resource.Resource].
func (r *resourceTFEProjectPolicySetExclusionParameter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelProjectPolicySetExclusionParameter
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Removing project (%s) from exclusion list of policy set (%s)", state.ProjectID.ValueString(), state.PolicySetID.ValueString()))
	err := r.config.Client.PolicySets.RemoveProjectExclusions(ctx, state.PolicySetID.ValueString(), tfe.PolicySetRemoveProjectExclusionsOptions{
		ProjectExclusions: []*tfe.Project{
			{
				ID: state.ProjectID.ValueString(),
			},
		},
	})

	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", state.PolicySetID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Project Exclusion from Policy Set",
			fmt.Sprintf("An error was encountered when removing project exclusion %q from policy set %q: %s", state.ProjectID.ValueString(), state.PolicySetID.ValueString(), err),
		)
		return
	}
}

// ImportState implements [resource.ResourceWithImportState].
func (r *resourceTFEProjectPolicySetExclusionParameter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID Format",
			fmt.Sprintf("The import ID must be in the format <project_id>/<policy_id>. Got: %q", id),
		)
		return
	}

	policySetID := parts[1]
	projectID := parts[0]

	tflog.Debug(ctx, fmt.Sprintf("Importing project exclusion for policy set with import ID %s", id))
	if ok, err := r.checkProjectExists(ctx, r.config.Client, projectID); err != nil {
		resp.Diagnostics.AddError(
			"Error Checking if Project Exists",
			fmt.Sprintf("An error was encountered when checking if project %q exists: %s", projectID, err),
		)
		return
	} else if !ok {
		resp.Diagnostics.AddError(
			"Project Not Found During Import",
			fmt.Sprintf("The project with ID %q was not found. Please verify that the project ID is correct and that the project exists. ID: %q", projectID, id),
		)
		return
	}

	ps, err := r.config.Client.PolicySets.ReadWithOptions(ctx, policySetID, &tfe.PolicySetReadOptions{
		Include: []tfe.PolicySetIncludeOpt{
			tfe.PolicySetProjectExclusions,
		},
	})

	if err != nil && errors.Is(err, tfe.ErrInvalidIncludeValue) {
		tflog.Debug(ctx, "Policy set exclusion is not supported")
		resp.Diagnostics.AddError(
			"Policy Set Exclusion Not Supported",
			"The API operation to manage project exclusions on policy sets is not supported in the current version of Terraform Enterprise. Please upgrade to a newer version of Terraform Enterprise that supports this feature.",
		)
		return
	}

	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", policySetID))
		resp.State.RemoveResource(ctx)
		return
	}

	for _, excludedProject := range ps.ProjectExclusions {
		if excludedProject.ID == projectID {
			state := modelProjectPolicySetExclusionParameter{
				ID:          types.StringValue(fmt.Sprintf("%s/%s", projectID, policySetID)),
				PolicySetID: types.StringValue(policySetID),
				ProjectID:   types.StringValue(projectID),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Project Exclusion Not Attached",
		fmt.Sprintf("The project exclusion with project ID %q was not found on policy set %q. It may have been removed outside of Terraform.", projectID, policySetID),
	)
	resp.State.RemoveResource(ctx)
}

func (r *resourceTFEProjectPolicySetExclusionParameter) checkProjectExists(ctx context.Context, client *tfe.Client, projectId string) (bool, error) {
	tflog.Debug(ctx, fmt.Sprintf("Checking if project %s exists", projectId))
	_, err := client.Projects.Read(ctx, projectId)
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Project %s does not exist.", projectId))
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
