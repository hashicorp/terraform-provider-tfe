// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithConfigure = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithImportState = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithModifyPlan = &resourceTFEOrgMaxTokenTTLPolicy{}

func NewOrgMaxTokenTTLPolicyResource() resource.Resource {
	return &resourceTFEOrgMaxTokenTTLPolicy{}
}

type resourceTFEOrgMaxTokenTTLPolicy struct {
	config ConfiguredClient
}

type modelTFEOrgMaxTokenTTLPolicy struct {
	ID                    types.String `tfsdk:"id"`
	Organization          types.String `tfsdk:"organization"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	OrgTokenMaxTTL        types.String `tfsdk:"org_token_max_ttl"`
	TeamTokenMaxTTL       types.String `tfsdk:"team_token_max_ttl"`
	AuditTrailTokenMaxTTL types.String `tfsdk:"audit_trail_token_max_ttl"`
	UserTokenMaxTTL       types.String `tfsdk:"user_token_max_ttl"`

	// Hidden computed attributes (int64 milliseconds from API)
	OrgTokenMaxTTLMs        types.Int64 `tfsdk:"org_token_max_ttl_ms"`
	TeamTokenMaxTTLMs       types.Int64 `tfsdk:"team_token_max_ttl_ms"`
	AuditTrailTokenMaxTTLMs types.Int64 `tfsdk:"audit_trail_token_max_ttl_ms"`
	UserTokenMaxTTLMs       types.Int64 `tfsdk:"user_token_max_ttl_ms"`
}

// validTTLPattern is a regex pattern for validating TTL duration strings.
var validTTLPattern = `^[0-9]+(\.[0-9]+)?(h|d|w|mo|y)$`

// defaultTokenTTL is the default maximum TTL for all token types when policy is disabled
const defaultTokenTTL = "2y"
const defaultTokenTTLMs = int64(63072000000) // 2 years in milliseconds

func (r *resourceTFEOrgMaxTokenTTLPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_max_token_ttl_policy"
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the maximum time-to-live (TTL) policy for API tokens in an organization. " +
			"When enabled, this policy enforces maximum lifespans for organization, team, audit trail, " +
			"and user tokens, revoking any tokens that exceed the configured limits.",
		Version: 0,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the token TTL policy (same as the organization name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Denotes whether the maximum TTL token policy is enabled (true) or disabled (false) for the organization.",
				Required:    true,
			},
			"org_token_max_ttl": schema.StringAttribute{
				Description: "Maximum lifespan allowed for organization tokens to access the organization's resources. " +
					"Defaults to two years (2y). " +
					"Format: <number><unit> where unit is h (hours), d (days), w (weeks), mo (months), or y (years). " +
					"Decimals are supported (e.g., 0.5h for 30 minutes). Examples: 1h, 2.5d, 3w, 1mo, 2y.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(defaultTokenTTL),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(validTTLPattern),
						"must be a valid duration string (e.g., 1h, 2.5d, 3w, 1mo, 2y)",
					),
				},
			},
			"team_token_max_ttl": schema.StringAttribute{
				Description: "Maximum lifespan allowed for team tokens to access the organization's resources. " +
					"Defaults to two years (2y). " +
					"Format: <number><unit> where unit is h (hours), d (days), w (weeks), mo (months), or y (years). " +
					"Decimals are supported (e.g., 0.5h for 30 minutes). Examples: 1h, 2.5d, 3w, 1mo, 2y.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(defaultTokenTTL),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(validTTLPattern),
						"must be a valid duration string (e.g., 1h, 2.5d, 3w, 1mo, 2y)",
					),
				},
			},
			"audit_trail_token_max_ttl": schema.StringAttribute{
				Description: "Maximum lifespan allowed for audit trail tokens to access the organization's resources. " +
					"Defaults to two years (2y). " +
					"Format: <number><unit> where unit is h (hours), d (days), w (weeks), mo (months), or y (years). " +
					"Decimals are supported (e.g., 0.5h for 30 minutes). Examples: 1h, 2.5d, 3w, 1mo, 2y.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(defaultTokenTTL),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(validTTLPattern),
						"must be a valid duration string (e.g., 1h, 2.5d, 3w, 1mo, 2y)",
					),
				},
			},
			"user_token_max_ttl": schema.StringAttribute{
				Description: "Maximum lifespan allowed for user tokens to access the organization's resources. " +
					"Defaults to two years (2y). " +
					"Format: <number><unit> where unit is h (hours), d (days), w (weeks), mo (months), or y (years). " +
					"Decimals are supported (e.g., 0.5h for 30 minutes). Examples: 1h, 2.5d, 3w, 1mo, 2y.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(defaultTokenTTL),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(validTTLPattern),
						"must be a valid duration string (e.g., 1h, 2.5d, 3w, 1mo, 2y)",
					),
				},
			},
			"org_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Internal: Exact milliseconds for organization token TTL as returned by the API.",
				Computed:    true,
			},
			"team_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Internal: Exact milliseconds for team token TTL as returned by the API.",
				Computed:    true,
			},
			"audit_trail_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Internal: Exact milliseconds for audit trail token TTL as returned by the API.",
				Computed:    true,
			},
			"user_token_max_ttl_ms": schema.Int64Attribute{
				Description: "Internal: Exact milliseconds for user token TTL as returned by the API.",
				Computed:    true,
			},
		},
	}
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceTFEOrgMaxTokenTTLPolicy) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)

	if req.Plan.Raw.IsNull() {
		return
	}

	// When enabled=false, override user-provided TTL values with defaults
	var plan modelTFEOrgMaxTokenTTLPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.ValueBool() {
		plan.OrgTokenMaxTTL = types.StringValue(defaultTokenTTL)
		plan.TeamTokenMaxTTL = types.StringValue(defaultTokenTTL)
		plan.AuditTrailTokenMaxTTL = types.StringValue(defaultTokenTTL)
		plan.UserTokenMaxTTL = types.StringValue(defaultTokenTTL)

		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

func newDisabledPolicyModel(organization string) modelTFEOrgMaxTokenTTLPolicy {
	return modelTFEOrgMaxTokenTTLPolicy{
		ID:                      types.StringValue(organization),
		Organization:            types.StringValue(organization),
		Enabled:                 types.BoolValue(false),
		OrgTokenMaxTTL:          types.StringValue(defaultTokenTTL),
		TeamTokenMaxTTL:         types.StringValue(defaultTokenTTL),
		AuditTrailTokenMaxTTL:   types.StringValue(defaultTokenTTL),
		UserTokenMaxTTL:         types.StringValue(defaultTokenTTL),
		OrgTokenMaxTTLMs:        types.Int64Value(defaultTokenTTLMs),
		TeamTokenMaxTTLMs:       types.Int64Value(defaultTokenTTLMs),
		AuditTrailTokenMaxTTLMs: types.Int64Value(defaultTokenTTLMs),
		UserTokenMaxTTLMs:       types.Int64Value(defaultTokenTTLMs),
	}
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If disabled, just set state without creating policies
	if !plan.Enabled.ValueBool() {
		tflog.Debug(ctx, "Token TTL policy is disabled, skipping creation", map[string]any{
			"organization": organization,
		})
		result := newDisabledPolicyModel(organization)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	policies, diagErr := r.buildPolicyUpdateItems(plan)
	if diagErr != nil {
		resp.Diagnostics.AddError("Invalid TTL values", diagErr.Error())
		return
	}

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: policies,
	}

	tflog.Debug(ctx, "Creating token TTL policies", map[string]any{
		"organization": organization,
	})
	updatedPolicies, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPolicies(organization, true, updatedPolicies, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading token TTL policies", map[string]any{
		"organization": organization,
	})
	policyList, err := r.config.Client.OrganizationTokenTTLPolicies.List(ctx, organization, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization token TTL policies", err.Error())
		return
	}
	enabled := len(policyList.Items) > 0
	result := modelFromTokenTTLPolicies(organization, enabled, policyList.Items, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Enabled.ValueBool() {
		tflog.Debug(ctx, "Disabling token TTL policy", map[string]any{
			"organization": organization,
		})

		options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
			Policies: []tfe.OrganizationTokenTTLPolicyUpdateItem{
				{TokenType: tfe.TokenTypeOrganization, MaxTTLMs: defaultTokenTTLMs},
				{TokenType: tfe.TokenTypeTeam, MaxTTLMs: defaultTokenTTLMs},
				{TokenType: tfe.TokenTypeUser, MaxTTLMs: defaultTokenTTLMs},
				{TokenType: tfe.TokenTypeAuditTrails, MaxTTLMs: defaultTokenTTLMs},
			},
		}

		_, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to disable organization token TTL policy", err.Error())
			return
		}

		result := newDisabledPolicyModel(organization)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	policies, diagErr := r.buildPolicyUpdateItems(plan)
	if diagErr != nil {
		resp.Diagnostics.AddError("Invalid TTL values", diagErr.Error())
		return
	}

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: policies,
	}

	tflog.Debug(ctx, "Updating token TTL policies", map[string]any{
		"organization": organization,
	})
	updatedPolicies, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPolicies(organization, true, updatedPolicies, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: []tfe.OrganizationTokenTTLPolicyUpdateItem{
			{TokenType: tfe.TokenTypeOrganization, MaxTTLMs: defaultTokenTTLMs},
			{TokenType: tfe.TokenTypeTeam, MaxTTLMs: defaultTokenTTLMs},
			{TokenType: tfe.TokenTypeUser, MaxTTLMs: defaultTokenTTLMs},
			{TokenType: tfe.TokenTypeAuditTrails, MaxTTLMs: defaultTokenTTLMs},
		},
	}

	tflog.Debug(ctx, "Deleting token TTL policy", map[string]any{
		"organization": organization,
	})
	_, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete organization token TTL policy", err.Error())
		return
	}
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization := req.ID

	tflog.Debug(ctx, "Importing token TTL policies", map[string]any{
		"organization": organization,
	})
	policyList, err := r.config.Client.OrganizationTokenTTLPolicies.List(ctx, organization, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization token TTL policies", err.Error())
		return
	}

	enabled := len(policyList.Items) > 0
	result := modelFromTokenTTLPolicies(organization, enabled, policyList.Items, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Converts plan model to go-tfe update items with milliseconds
func (r *resourceTFEOrgMaxTokenTTLPolicy) buildPolicyUpdateItems(plan modelTFEOrgMaxTokenTTLPolicy) ([]tfe.OrganizationTokenTTLPolicyUpdateItem, error) {
	var policies []tfe.OrganizationTokenTTLPolicyUpdateItem

	// Define token type and field value
	tokenConfigs := []struct {
		tokenType string
		ttlValue  types.String
	}{
		{tfe.TokenTypeOrganization, plan.OrgTokenMaxTTL},
		{tfe.TokenTypeTeam, plan.TeamTokenMaxTTL},
		{tfe.TokenTypeAuditTrails, plan.AuditTrailTokenMaxTTL},
		{tfe.TokenTypeUser, plan.UserTokenMaxTTL},
	}

	for _, config := range tokenConfigs {
		if err := r.addPolicyIfSet(config.tokenType, config.ttlValue, &policies); err != nil {
			return nil, err
		}
	}

	return policies, nil
}

// Adds a policy to the list if the TTL value is set
func (r *resourceTFEOrgMaxTokenTTLPolicy) addPolicyIfSet(tokenType string, ttlValue types.String, policies *[]tfe.OrganizationTokenTTLPolicyUpdateItem) error {
	if ttlValue.IsNull() || ttlValue.IsUnknown() {
		return nil
	}

	ms, err := durationStringToMilliseconds(ttlValue.ValueString())
	if err != nil {
		return fmt.Errorf("invalid %s token TTL: %w", tokenType, err)
	}

	*policies = append(*policies, tfe.OrganizationTokenTTLPolicyUpdateItem{
		TokenType: tokenType,
		MaxTTLMs:  ms,
	})

	return nil
}

// Converts duration strings like "1y", "30d", "24h" to milliseconds
func durationStringToMilliseconds(duration string) (int64, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	re := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(h|d|w|mo|y)$`)
	matches := re.FindStringSubmatch(duration)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	var milliseconds int64

	switch unit {
	case "h": // hours
		milliseconds = int64(value * 60 * 60 * 1000)
	case "d": // days
		milliseconds = int64(value * 24 * 60 * 60 * 1000)
	case "w": // weeks
		milliseconds = int64(value * 7 * 24 * 60 * 60 * 1000)
	case "mo": // months (30 days)
		milliseconds = int64(value * 30 * 24 * 60 * 60 * 1000)
	case "y": // years (365 days)
		milliseconds = int64(value * 365 * 24 * 60 * 60 * 1000)
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return milliseconds, nil
}

func modelFromTokenTTLPolicies(organization string, enabled bool, policies []*tfe.OrganizationTokenTTLPolicy, plan *modelTFEOrgMaxTokenTTLPolicy) modelTFEOrgMaxTokenTTLPolicy {
	result := modelTFEOrgMaxTokenTTLPolicy{
		ID:           types.StringValue(organization),
		Organization: types.StringValue(organization),
		Enabled:      types.BoolValue(enabled),
	}

	// If plan is provided (Create/Update), preserve user input (includes schema defaults)
	if plan != nil {
		result.OrgTokenMaxTTL = plan.OrgTokenMaxTTL
		result.TeamTokenMaxTTL = plan.TeamTokenMaxTTL
		result.AuditTrailTokenMaxTTL = plan.AuditTrailTokenMaxTTL
		result.UserTokenMaxTTL = plan.UserTokenMaxTTL
	} else {
		// For Read/ImportState without plan, set defaults
		result.OrgTokenMaxTTL = types.StringValue(defaultTokenTTL)
		result.TeamTokenMaxTTL = types.StringValue(defaultTokenTTL)
		result.AuditTrailTokenMaxTTL = types.StringValue(defaultTokenTTL)
		result.UserTokenMaxTTL = types.StringValue(defaultTokenTTL)
	}

	// Store API milliseconds in _ms fields (schema defaults handle missing values)
	for _, policy := range policies {
		switch policy.TokenType {
		case tfe.TokenTypeOrganization:
			result.OrgTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeTeam:
			result.TeamTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeAuditTrails:
			result.AuditTrailTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		case tfe.TokenTypeUser:
			result.UserTokenMaxTTLMs = types.Int64Value(policy.MaxTTLMs)
		}
	}

	return result
}
