// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfev2models "github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	customValidators "github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

type resourceTFEIPAllowlist struct {
	config ConfiguredClient
}

var (
	_ resource.Resource                = &resourceTFEIPAllowlist{}
	_ resource.ResourceWithConfigure   = &resourceTFEIPAllowlist{}
	_ resource.ResourceWithImportState = &resourceTFEIPAllowlist{}
	_ resource.ResourceWithModifyPlan  = &resourceTFEIPAllowlist{}
)

// NewIPAllowlistResource returns a new IP allowlist (CIDR range list) resource.
func NewIPAllowlistResource() resource.Resource {
	return &resourceTFEIPAllowlist{}
}

// modelTFEIPAllowlist is the resource model for an IP allowlist.
type modelTFEIPAllowlist struct {
	ID               types.String `tfsdk:"id"`
	Organization     types.String `tfsdk:"organization"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	EnforcementScope types.String `tfsdk:"enforcement_scope"`
	AgentPoolIDs     types.Set    `tfsdk:"agent_pool_ids"`
	CIDRRanges       types.Set    `tfsdk:"cidr_range"`
}

func (r *resourceTFEIPAllowlist) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_allowlist"
}

func (r *resourceTFEIPAllowlist) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IP allowlists (CIDR range lists) in an organization. IP allowlists restrict which client IP addresses may access HCP Terraform for the organization or its agent pools.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the IP allowlist.",
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
			"name": schema.StringAttribute{
				Description: "Name of the IP allowlist. Must be unique within the organization.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the IP allowlist.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enforcement_scope": schema.StringAttribute{
				Description: "Where the IP allowlist is enforced. Must be one of `organization`, `all_agent_pools`, or `selected_agent_pools`. Only one `organization`-scoped allowlist may exist per organization.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						ipAllowlistScopeOrganization,
						ipAllowlistScopeAllAgentPools,
						ipAllowlistScopeSelectedAgentPools,
					),
				},
			},
			"agent_pool_ids": schema.SetAttribute{
				Description: "The IDs of the agent pools the IP allowlist applies to. Only valid when `enforcement_scope` is `selected_agent_pools`.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"cidr_range": schema.SetNestedAttribute{
				Description: "The set of CIDR ranges that belong to the IP allowlist. At least one range is required.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"range": schema.StringAttribute{
							Description: "An IPv4 CIDR range, e.g. `10.0.0.0/24`.",
							Required:    true,
							Validators: []validator.String{
								customValidators.IsIPv4CIDR(),
							},
						},
						"description": schema.StringAttribute{
							Description: "A description for the CIDR range.",
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the CIDR range is enforced. Defaults to `true`.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
					},
				},
			},
		},
	}
}

func (r *resourceTFEIPAllowlist) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// When the allowlist uses the provider's default organization, a change to
	// that default should trigger a replacement.
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)

	// Skip further validation on destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan modelTFEIPAllowlist
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// agent_pool_ids is only valid for the selected_agent_pools scope.
	if plan.EnforcementScope.ValueString() != ipAllowlistScopeSelectedAgentPools &&
		!plan.AgentPoolIDs.IsNull() && !plan.AgentPoolIDs.IsUnknown() && len(plan.AgentPoolIDs.Elements()) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("agent_pool_ids"),
			"Invalid Attribute Combination",
			"agent_pool_ids can only be set when enforcement_scope is \"selected_agent_pools\".",
		)
	}
}

func (r *resourceTFEIPAllowlist) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.config = client
}

func (r *resourceTFEIPAllowlist) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEIPAllowlist
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Config, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scope, err := enforcementScopeToV2(plan.EnforcementScope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid enforcement_scope", err.Error())
		return
	}

	// Build the list attributes.
	attrs := tfev2models.NewCidrRangeLists_attributes()
	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	attrs.SetName(&name)
	attrs.SetDescription(&desc)
	attrs.SetEnforcementScope(&scope)

	listType := tfev2models.CIDRRANGELISTS_CIDRRANGELISTS_TYPE
	data := tfev2models.NewCidrRangeLists()
	data.SetTypeEscaped(&listType)
	data.SetAttributes(attrs)

	// Seed the CIDR ranges at creation time. The create endpoint accepts nested
	// ranges and creates them atomically with the list.
	var planRanges []modelTFECIDRRange
	resp.Diagnostics.Append(plan.CIDRRanges.ElementsAs(ctx, &planRanges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(planRanges) > 0 {
		nested := make([]tfev2models.NestedCidrRangeable, 0, len(planRanges))
		for _, m := range planRanges {
			nested = append(nested, nestedCidrRangeData(m))
		}
		cidrRanges := tfev2models.NewCidrRangeLists_relationships_cidrRanges()
		cidrRanges.SetData(nested)
		relationships := tfev2models.NewCidrRangeLists_relationships()
		relationships.SetCidrRanges(cidrRanges)
		data.SetRelationships(relationships)
	}

	body := tfev2models.NewCidrRangeListWithRangesEnvelope()
	body.SetData(data)

	tflog.Debug(ctx, fmt.Sprintf("Creating IP allowlist %q for organization %q", name, organization))
	created, err := r.config.ClientV2.API.
		Organizations().
		ByOrganization_name(organization).
		CidrRangeLists().
		Post(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP allowlist", err.Error())
		return
	}
	if created == nil || created.GetData() == nil || created.GetData().GetId() == nil {
		resp.Diagnostics.AddError("Error creating IP allowlist", "The API did not return an ID for the created IP allowlist.")
		return
	}

	listID := *created.GetData().GetId()

	// Assign agent pools for the selected_agent_pools scope.
	if plan.EnforcementScope.ValueString() == ipAllowlistScopeSelectedAgentPools {
		desired := setToStringSlice(ctx, plan.AgentPoolIDs, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.reconcileAgentPools(ctx, listID, desired); err != nil {
			resp.Diagnostics.AddError("Error assigning agent pools to IP allowlist", err.Error())
			return
		}
	}

	// Build state from the plan rather than reading back, because reads on HCP
	// Terraform are served by a read replica that may lag behind the primary
	// immediately after a write. The create request is atomic and fully
	// determines the result, so the plan (plus the server-assigned ID) is the
	// authoritative post-create state.
	result := plan
	result.ID = types.StringValue(listID)
	result.Organization = types.StringValue(organization)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEIPAllowlist) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEIPAllowlist
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID := state.ID.ValueString()
	result, diags, err := r.fetchIPAllowlist(ctx, listID)
	if errors.Is(err, errIPAllowlistNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("IP allowlist %s no longer exists", listID))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP allowlist", err.Error())
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEIPAllowlist) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEIPAllowlist
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID := plan.ID.ValueString()

	scope, err := enforcementScopeToV2(plan.EnforcementScope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid enforcement_scope", err.Error())
		return
	}

	// Update the list attributes. The list PATCH endpoint treats nested ranges
	// as append-only, so ranges are intentionally omitted here and reconciled
	// individually below.
	attrs := tfev2models.NewCidrRangeLists_attributes()
	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	attrs.SetName(&name)
	attrs.SetDescription(&desc)
	attrs.SetEnforcementScope(&scope)

	listType := tfev2models.CIDRRANGELISTS_CIDRRANGELISTS_TYPE
	data := tfev2models.NewCidrRangeLists()
	data.SetTypeEscaped(&listType)
	data.SetAttributes(attrs)

	body := tfev2models.NewCidrRangeListEnvelope()
	body.SetData(data)

	tflog.Debug(ctx, fmt.Sprintf("Updating IP allowlist %s", listID))
	if _, err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Patch(ctx, body, nil); err != nil {
		resp.Diagnostics.AddError("Error updating IP allowlist", err.Error())
		return
	}

	// Reconcile CIDR ranges.
	var planRanges []modelTFECIDRRange
	resp.Diagnostics.Append(plan.CIDRRanges.ElementsAs(ctx, &planRanges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconcileRanges(ctx, listID, planRanges); err != nil {
		resp.Diagnostics.AddError("Error updating IP allowlist CIDR ranges", err.Error())
		return
	}

	// Reconcile agent pool assignments. For non-selected scopes the API clears
	// assignments automatically when the scope changes, so only reconcile when
	// the scope is selected_agent_pools.
	if plan.EnforcementScope.ValueString() == ipAllowlistScopeSelectedAgentPools {
		desired := setToStringSlice(ctx, plan.AgentPoolIDs, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.reconcileAgentPools(ctx, listID, desired); err != nil {
			resp.Diagnostics.AddError("Error updating IP allowlist agent pools", err.Error())
			return
		}
	}

	// As with Create, build state from the plan instead of reading back to
	// avoid read-replica lag immediately after the write. The reconcile steps
	// above bring the server to the planned state.
	result := plan
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEIPAllowlist) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEIPAllowlist
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID := state.ID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Deleting IP allowlist %s", listID))
	err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Delete(ctx, nil)
	if err != nil && !isV2ResourceNotFound(err) {
		resp.Diagnostics.AddError(
			"Error deleting IP allowlist",
			fmt.Sprintf("Couldn't delete IP allowlist %s: %s", listID, err.Error()),
		)
	}
}

func (r *resourceTFEIPAllowlist) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// errIPAllowlistNotFound is returned by fetchIPAllowlist when the IP allowlist
// (or one of its sub-resources) responds with an HTTP 404.
var errIPAllowlistNotFound = errors.New("IP allowlist not found")

// fetchIPAllowlist fetches the IP allowlist and its CIDR ranges and builds the
// resource model. It returns errIPAllowlistNotFound when the allowlist responds
// with a 404 so callers can distinguish deletion from other errors.
func (r *resourceTFEIPAllowlist) fetchIPAllowlist(ctx context.Context, listID string) (modelTFEIPAllowlist, diag.Diagnostics, error) {
	var diags diag.Diagnostics
	var model modelTFEIPAllowlist

	envelope, err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Get(ctx, nil)
	if err != nil {
		if isV2ResourceNotFound(err) {
			return model, diags, errIPAllowlistNotFound
		}
		return model, diags, fmt.Errorf("couldn't read IP allowlist %s: %w", listID, err)
	}
	if envelope == nil || envelope.GetData() == nil {
		return model, diags, errIPAllowlistNotFound
	}

	list := envelope.GetData()

	model.ID = types.StringValue(listID)
	if list.GetAttributes() != nil {
		if list.GetAttributes().GetName() != nil {
			model.Name = types.StringValue(*list.GetAttributes().GetName())
		}
		desc := ""
		if list.GetAttributes().GetDescription() != nil {
			desc = *list.GetAttributes().GetDescription()
		}
		model.Description = types.StringValue(desc)
		model.EnforcementScope = types.StringValue(enforcementScopeFromV2(list.GetAttributes().GetEnforcementScope()))
	}

	// Organization from the relationship, when present.
	model.Organization = types.StringNull()
	if list.GetRelationships() != nil && list.GetRelationships().GetOrganization() != nil {
		if id := list.GetRelationships().GetOrganization().GetData(); id != nil && id.GetId() != nil {
			model.Organization = types.StringValue(*id.GetId())
		}
	}

	// Agent pool IDs.
	agentPoolIDs := currentAgentPoolIDs(list)
	if len(agentPoolIDs) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, agentPoolIDs)
		diags.Append(d...)
		model.AgentPoolIDs = set
	} else {
		model.AgentPoolIDs = types.SetNull(types.StringType)
	}

	// CIDR ranges.
	apiRanges, err := readIPAllowlistRanges(ctx, r.config.ClientV2, listID)
	if err != nil {
		if isV2ResourceNotFound(err) {
			return model, diags, errIPAllowlistNotFound
		}
		return model, diags, fmt.Errorf("couldn't read CIDR ranges for IP allowlist %s: %w", listID, err)
	}
	set, d := cidrRangeSetFromAPI(ctx, apiRanges)
	diags.Append(d...)
	model.CIDRRanges = set

	return model, diags, nil
}

// reconcileAgentPools ensures the assigned agent pools match the desired set.
func (r *resourceTFEIPAllowlist) reconcileAgentPools(ctx context.Context, listID string, desired []string) error {
	envelope, err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Get(ctx, nil)
	if err != nil {
		return err
	}

	var current []string
	if envelope != nil {
		current = currentAgentPoolIDs(envelope.GetData())
	}

	toAdd := stringSliceDifference(desired, current)
	toRemove := stringSliceDifference(current, desired)

	if len(toAdd) > 0 {
		if err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Relationships().AgentPools().Post(ctx, agentPoolIDsBody(toAdd), nil); err != nil {
			return err
		}
	}
	if len(toRemove) > 0 {
		if err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Relationships().AgentPools().Delete(ctx, agentPoolIDsBody(toRemove), nil); err != nil {
			return err
		}
	}
	return nil
}

// reconcileRanges ensures the CIDR ranges belonging to the list match the
// desired set, creating, updating, and deleting individual ranges as needed.
func (r *resourceTFEIPAllowlist) reconcileRanges(ctx context.Context, listID string, desired []modelTFECIDRRange) error {
	current, err := readIPAllowlistRanges(ctx, r.config.ClientV2, listID)
	if err != nil {
		return err
	}

	currentByRange := make(map[string]tfev2models.CidrRangesable, len(current))
	for _, c := range current {
		if c.GetAttributes() != nil && c.GetAttributes().GetRangeEscaped() != nil {
			currentByRange[*c.GetAttributes().GetRangeEscaped()] = c
		}
	}

	desiredByRange := make(map[string]modelTFECIDRRange, len(desired))
	for _, d := range desired {
		desiredByRange[d.Range.ValueString()] = d
	}

	// Create or update desired ranges.
	for rng, d := range desiredByRange {
		existing, ok := currentByRange[rng]
		if !ok {
			// Create a new range.
			if _, err := r.config.ClientV2.API.CidrRangeLists().ByCidr_range_list_id(listID).Relationships().CidrRanges().Post(ctx, cidrRangeEnvelope(d), nil); err != nil {
				return err
			}
			continue
		}

		// Update the range if its mutable attributes changed.
		attrs := existing.GetAttributes()
		curDesc := ""
		if attrs != nil && attrs.GetDescription() != nil {
			curDesc = *attrs.GetDescription()
		}
		curEnabled := true
		if attrs != nil && attrs.GetEnabled() != nil {
			curEnabled = *attrs.GetEnabled()
		}
		if curDesc != d.Description.ValueString() || curEnabled != d.Enabled.ValueBool() {
			if existing.GetId() == nil {
				continue
			}
			if _, err := r.config.ClientV2.API.CidrRanges().ByCidr_range_id(*existing.GetId()).Patch(ctx, cidrRangeEnvelope(d), nil); err != nil {
				return err
			}
		}
	}

	// Delete ranges that are no longer desired.
	for rng, existing := range currentByRange {
		if _, ok := desiredByRange[rng]; ok {
			continue
		}
		if existing.GetId() == nil {
			continue
		}
		if err := r.config.ClientV2.API.CidrRanges().ByCidr_range_id(*existing.GetId()).Delete(ctx, nil); err != nil {
			return err
		}
	}

	return nil
}
