// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type modelTFESCIMGroupMapping struct {
	ID          types.String `tfsdk:"id"`
	TeamID      types.String `tfsdk:"team_id"`
	SCIMGroupID types.String `tfsdk:"scim_group_id"`
	Paused      types.Bool   `tfsdk:"paused"`
}

// resourceTFESCIMGroupMapping implements the tfe_scim_group_mapping resource type.
type resourceTFESCIMGroupMapping struct {
	client *tfe.Client
}

var (
	_ resource.Resource                = &resourceTFESCIMGroupMapping{}
	_ resource.ResourceWithConfigure   = &resourceTFESCIMGroupMapping{}
	_ resource.ResourceWithImportState = &resourceTFESCIMGroupMapping{}
)

// NewSCIMGroupMappingResource is a resource function for the framework provider.
func NewSCIMGroupMappingResource() resource.Resource {
	return &resourceTFESCIMGroupMapping{}
}

// Metadata implements resource.Resource
func (r *resourceTFESCIMGroupMapping) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_group_mapping"
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESCIMGroupMapping) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is not properly configured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}
	r.client = client.Client
}

// Schema implements resource.Resource
func (r *resourceTFESCIMGroupMapping) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Maps a SCIM group to a team in Terraform Enterprise. A team can be mapped to at most one SCIM group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SCIM group mapping. Since a team can only be mapped to one SCIM group, this is the same as `team_id`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "The ID of the team to map the SCIM group to. Changing this forces a new mapping to be created.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scim_group_id": schema.StringAttribute{
				Description: "The ID of the SCIM group to map to the team. Changing this forces a new mapping to be created, since the mapping API only supports updating the paused state.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paused": schema.BoolAttribute{
				Description: "Whether provisioning for this mapping is paused. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Read implements resource.Resource. The Teams API tells us if the team is
// SCIM-linked, its paused state and the group's name; we then look up the
// group's ID by name, since Teams doesn't return it.
func (r *resourceTFESCIMGroupMapping) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFESCIMGroupMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading SCIM group mapping for team %s", teamID))
	team, err := r.client.Teams.Read(ctx, teamID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Team %s no longer exists; removing SCIM group mapping from state", teamID))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading team %s for SCIM group mapping", teamID),
			"Could not read team, unexpected error: "+err.Error(),
		)
		return
	}

	// If the team is no longer SCIM-linked, the mapping was removed out-of-band.
	if team.SCIMLinked == nil || !*team.SCIMLinked {
		tflog.Debug(ctx, fmt.Sprintf("Team %s is no longer SCIM-linked; removing mapping from state", teamID))
		resp.State.RemoveResource(ctx)
		return
	}

	groupName := ""
	if team.SCIMGroupName != nil {
		groupName = *team.SCIMGroupName
	}

	scimGroupID, err := r.resolveSCIMGroupID(ctx, groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error resolving SCIM group for team %s", teamID),
			err.Error(),
		)
		return
	}

	paused := false
	if team.SCIMSyncPaused != nil {
		paused = *team.SCIMSyncPaused
	}

	result := modelTFESCIMGroupMapping{
		ID:          types.StringValue(teamID),
		TeamID:      types.StringValue(teamID),
		SCIMGroupID: types.StringValue(scimGroupID),
		Paused:      types.BoolValue(paused),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Create implements resource.Resource. Create can't set the paused state, so a
// paused mapping is created then paused in a follow-up update.
func (r *resourceTFESCIMGroupMapping) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFESCIMGroupMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := plan.TeamID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Creating SCIM group mapping for team %s", teamID))
	err := r.client.Admin.Settings.SCIM.SCIMGroupMappings.Create(ctx, teamID, &tfe.AdminSCIMGroupMappingCreateOptions{
		SCIMGroupID: plan.SCIMGroupID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SCIM group mapping",
			fmt.Sprintf("Could not create SCIM group mapping for team %s, unexpected error: %s", teamID, err.Error()),
		)
		return
	}

	// Save state now that the mapping exists, so a failed pause below doesn't
	// orphan it.
	paused := plan.Paused.ValueBool()
	result := modelTFESCIMGroupMapping{
		ID:          types.StringValue(teamID),
		TeamID:      types.StringValue(teamID),
		SCIMGroupID: plan.SCIMGroupID,
		Paused:      types.BoolValue(false),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create always starts unpaused, so pause it now if requested.
	if paused {
		tflog.Debug(ctx, fmt.Sprintf("Pausing SCIM group mapping for team %s", teamID))
		err = r.client.Admin.Settings.SCIM.SCIMGroupMappings.Update(ctx, teamID, &tfe.AdminSCIMGroupMappingUpdateOptions{
			SCIMSyncPaused: tfe.Bool(true),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error pausing SCIM group mapping",
				fmt.Sprintf("The mapping for team %s was created but could not be paused, unexpected error: %s", teamID, err.Error()),
			)
			return
		}
		result.Paused = types.BoolValue(true)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	}
}

// Update implements resource.Resource. Only the paused state can change in
// place; team_id and scim_group_id changes force a replacement.
func (r *resourceTFESCIMGroupMapping) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFESCIMGroupMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := plan.TeamID.ValueString()
	paused := plan.Paused.ValueBool()

	tflog.Debug(ctx, fmt.Sprintf("Updating SCIM group mapping for team %s", teamID))
	err := r.client.Admin.Settings.SCIM.SCIMGroupMappings.Update(ctx, teamID, &tfe.AdminSCIMGroupMappingUpdateOptions{
		SCIMSyncPaused: tfe.Bool(paused),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating SCIM group mapping",
			fmt.Sprintf("Could not update SCIM group mapping for team %s, unexpected error: %s", teamID, err.Error()),
		)
		return
	}

	result := modelTFESCIMGroupMapping{
		ID:          types.StringValue(teamID),
		TeamID:      types.StringValue(teamID),
		SCIMGroupID: plan.SCIMGroupID,
		Paused:      types.BoolValue(paused),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource.
func (r *resourceTFESCIMGroupMapping) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFESCIMGroupMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Deleting SCIM group mapping for team %s", teamID))
	err := r.client.Admin.Settings.SCIM.SCIMGroupMappings.Delete(ctx, teamID)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting SCIM group mapping",
			fmt.Sprintf("Could not delete SCIM group mapping for team %s, unexpected error: %s", teamID, err.Error()),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState. The import ID is the
// team ID; the remaining attributes are populated by the subsequent Read.
func (r *resourceTFESCIMGroupMapping) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// resolveSCIMGroupID looks up a SCIM group's ID by name, since the Teams API
// only gives us the group's name and not its ID.
func (r *resourceTFESCIMGroupMapping) resolveSCIMGroupID(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", errors.New("Team is SCIM-linked but the linked SCIM group name is empty; cannot resolve scim_group_id")
	}

	group, err := findSCIMGroupByName(ctx, r.client, name)
	if err != nil {
		return "", err
	}
	if group == nil {
		return "", fmt.Errorf("no SCIM group found with name %q", name)
	}

	return group.ID, nil
}
