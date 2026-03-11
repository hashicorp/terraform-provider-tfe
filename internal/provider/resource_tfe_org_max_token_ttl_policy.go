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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithConfigure = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithImportState = &resourceTFEOrgMaxTokenTTLPolicy{}
var _ resource.ResourceWithModifyPlan = &resourceTFEOrgMaxTokenTTLPolicy{}

// NewOrgMaxTokenTTLPolicyResource returns a new instance of the resource.
func NewOrgMaxTokenTTLPolicyResource() resource.Resource {
	return &resourceTFEOrgMaxTokenTTLPolicy{}
}

// resourceTFEOrgMaxTokenTTLPolicy implements the tfe_org_max_token_ttl_policy resource type.
type resourceTFEOrgMaxTokenTTLPolicy struct {
	config ConfiguredClient
}

// modelTFEOrgMaxTokenTTLPolicy is the Terraform state model for this resource.
type modelTFEOrgMaxTokenTTLPolicy struct {
	ID                    types.String `tfsdk:"id"`
	Organization          types.String `tfsdk:"organization"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	OrgTokenMaxTTL        types.String `tfsdk:"org_token_max_ttl"`
	TeamTokenMaxTTL       types.String `tfsdk:"team_token_max_ttl"`
	AuditTrailTokenMaxTTL types.String `tfsdk:"audit_trail_token_max_ttl"`
	UserTokenMaxTTL       types.String `tfsdk:"user_token_max_ttl"`
}

// validTTLPattern is a regex pattern for validating TTL duration strings.
// Accepts formats like: 1h, 2.5d, 0.5h, 3w, 1mo, 2y
// Units: h (hours), d (days), w (weeks), mo (months), y (years)
var validTTLPattern = `^[0-9]+(\.[0-9]+)?(h|d|w|mo|y)$`

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
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(validTTLPattern),
						"must be a valid duration string (e.g., 1h, 2.5d, 3w, 1mo, 2y)",
					),
				},
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure.
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

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *resourceTFEOrgMaxTokenTTLPolicy) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanForDefaultOrganizationChange(ctx, r.config.Organization, req.State, req.Config, req.Plan, resp)
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
		tflog.Debug(ctx, fmt.Sprintf("Token TTL policy is disabled for organization: %s, skipping creation", organization))
		result := modelTFEOrgMaxTokenTTLPolicy{
			ID:                    types.StringValue(organization),
			Organization:          types.StringValue(organization),
			Enabled:               types.BoolValue(false),
			OrgTokenMaxTTL:        types.StringValue("2y"),
			TeamTokenMaxTTL:       types.StringValue("2y"),
			AuditTrailTokenMaxTTL: types.StringValue("2y"),
			UserTokenMaxTTL:       types.StringValue("2y"),
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	// Build array of policies with milliseconds
	policies, diagErr := r.buildPolicyUpdateItems(plan)
	if diagErr != nil {
		resp.Diagnostics.AddError("Invalid TTL values", diagErr.Error())
		return
	}

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: policies,
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating token TTL policies for organization: %s", organization))
	updatedPolicies, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPolicies(organization, true, updatedPolicies)
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

	tflog.Debug(ctx, fmt.Sprintf("Reading token TTL policies for organization: %s", organization))
	policyList, err := r.config.Client.OrganizationTokenTTLPolicies.List(ctx, organization, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization token TTL policies", err.Error())
		return
	}

	// If no policies exist, treat as disabled
	enabled := len(policyList.Items) > 0

	result := modelFromTokenTTLPolicies(organization, enabled, policyList.Items)
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

	// If disabled, set very high TTLs (effectively disabled)
	if !plan.Enabled.ValueBool() {
		tflog.Debug(ctx, fmt.Sprintf("Disabling token TTL policy for organization: %s", organization))
		maxTTL := int64(63072000000000) // 2 years in milliseconds

		options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
			Policies: []tfe.OrganizationTokenTTLPolicyUpdateItem{
				{TokenType: tfe.TokenTypeOrganization, MaxTTLMs: maxTTL},
				{TokenType: tfe.TokenTypeTeam, MaxTTLMs: maxTTL},
				{TokenType: tfe.TokenTypeUser, MaxTTLMs: maxTTL},
				{TokenType: tfe.TokenTypeAuditTrails, MaxTTLMs: maxTTL},
			},
		}

		_, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
		if err != nil {
			resp.Diagnostics.AddError("Unable to disable organization token TTL policy", err.Error())
			return
		}

		result := modelTFEOrgMaxTokenTTLPolicy{
			ID:                    types.StringValue(organization),
			Organization:          types.StringValue(organization),
			Enabled:               types.BoolValue(false),
			OrgTokenMaxTTL:        types.StringValue("2y"),
			TeamTokenMaxTTL:       types.StringValue("2y"),
			AuditTrailTokenMaxTTL: types.StringValue("2y"),
			UserTokenMaxTTL:       types.StringValue("2y"),
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	// Build array of policies with milliseconds
	policies, diagErr := r.buildPolicyUpdateItems(plan)
	if diagErr != nil {
		resp.Diagnostics.AddError("Invalid TTL values", diagErr.Error())
		return
	}

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: policies,
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating token TTL policies for organization: %s", organization))
	updatedPolicies, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPolicies(organization, true, updatedPolicies)
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

	// Set very high TTLs (effectively disabled)
	maxTTL := int64(63072000000000) // 2 years in milliseconds

	options := tfe.OrganizationTokenTTLPolicyUpdateOptions{
		Policies: []tfe.OrganizationTokenTTLPolicyUpdateItem{
			{TokenType: tfe.TokenTypeOrganization, MaxTTLMs: maxTTL},
			{TokenType: tfe.TokenTypeTeam, MaxTTLMs: maxTTL},
			{TokenType: tfe.TokenTypeUser, MaxTTLMs: maxTTL},
			{TokenType: tfe.TokenTypeAuditTrails, MaxTTLMs: maxTTL},
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting token TTL policy for organization: %s", organization))
	_, err := r.config.Client.OrganizationTokenTTLPolicies.Update(ctx, organization, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete organization token TTL policy", err.Error())
		return
	}
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization := req.ID

	tflog.Debug(ctx, fmt.Sprintf("Importing token TTL policies for organization: %s", organization))
	policyList, err := r.config.Client.OrganizationTokenTTLPolicies.List(ctx, organization, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization token TTL policies", err.Error())
		return
	}

	enabled := len(policyList.Items) > 0
	result := modelFromTokenTTLPolicies(organization, enabled, policyList.Items)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// buildPolicyUpdateItems converts plan model to go-tfe update items with milliseconds
func (r *resourceTFEOrgMaxTokenTTLPolicy) buildPolicyUpdateItems(plan modelTFEOrgMaxTokenTTLPolicy) ([]tfe.OrganizationTokenTTLPolicyUpdateItem, error) {
	var policies []tfe.OrganizationTokenTTLPolicyUpdateItem

	if !plan.OrgTokenMaxTTL.IsNull() && !plan.OrgTokenMaxTTL.IsUnknown() {
		ms, err := durationStringToMilliseconds(plan.OrgTokenMaxTTL.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid org_token_max_ttl: %w", err)
		}
		policies = append(policies, tfe.OrganizationTokenTTLPolicyUpdateItem{
			TokenType: tfe.TokenTypeOrganization,
			MaxTTLMs:  ms,
		})
	}

	if !plan.TeamTokenMaxTTL.IsNull() && !plan.TeamTokenMaxTTL.IsUnknown() {
		ms, err := durationStringToMilliseconds(plan.TeamTokenMaxTTL.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid team_token_max_ttl: %w", err)
		}
		policies = append(policies, tfe.OrganizationTokenTTLPolicyUpdateItem{
			TokenType: tfe.TokenTypeTeam,
			MaxTTLMs:  ms,
		})
	}

	if !plan.AuditTrailTokenMaxTTL.IsNull() && !plan.AuditTrailTokenMaxTTL.IsUnknown() {
		ms, err := durationStringToMilliseconds(plan.AuditTrailTokenMaxTTL.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid audit_trail_token_max_ttl: %w", err)
		}
		policies = append(policies, tfe.OrganizationTokenTTLPolicyUpdateItem{
			TokenType: tfe.TokenTypeAuditTrails,
			MaxTTLMs:  ms,
		})
	}

	if !plan.UserTokenMaxTTL.IsNull() && !plan.UserTokenMaxTTL.IsUnknown() {
		ms, err := durationStringToMilliseconds(plan.UserTokenMaxTTL.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid user_token_max_ttl: %w", err)
		}
		policies = append(policies, tfe.OrganizationTokenTTLPolicyUpdateItem{
			TokenType: tfe.TokenTypeUser,
			MaxTTLMs:  ms,
		})
	}

	return policies, nil
}

// durationStringToMilliseconds converts duration strings like "1y", "30d", "24h" to milliseconds
func durationStringToMilliseconds(duration string) (int64, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	// Parse the duration string
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

func millisecondsToHumanReadable(ms int64) string {
	// Convert to most appropriate unit
	seconds := ms / 1000
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	if days >= 365 && days%365 == 0 {
		return fmt.Sprintf("%dy", days/365)
	} else if days >= 30 && days%30 == 0 {
		return fmt.Sprintf("%dmo", days/30)
	} else if days >= 7 && days%7 == 0 {
		return fmt.Sprintf("%dw", days/7)
	} else if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return "0h"
}

// modelFromTokenTTLPolicies converts API response array to Terraform state model
func modelFromTokenTTLPolicies(organization string, enabled bool, policies []*tfe.OrganizationTokenTTLPolicy) modelTFEOrgMaxTokenTTLPolicy {
	result := modelTFEOrgMaxTokenTTLPolicy{
		ID:           types.StringValue(organization),
		Organization: types.StringValue(organization),
		Enabled:      types.BoolValue(enabled),
	}

	// Convert each policy from milliseconds to duration string
	for _, policy := range policies {
		durationStr := millisecondsToHumanReadable(policy.MaxTTLMs)

		switch policy.TokenType {
		case tfe.TokenTypeOrganization:
			result.OrgTokenMaxTTL = types.StringValue(durationStr)
		case tfe.TokenTypeTeam:
			result.TeamTokenMaxTTL = types.StringValue(durationStr)
		case tfe.TokenTypeAuditTrails:
			result.AuditTrailTokenMaxTTL = types.StringValue(durationStr)
		case tfe.TokenTypeUser:
			result.UserTokenMaxTTL = types.StringValue(durationStr)
		}
	}

	// Set defaults for missing policies
	if result.OrgTokenMaxTTL.IsNull() {
		result.OrgTokenMaxTTL = types.StringValue("2y")
	}
	if result.TeamTokenMaxTTL.IsNull() {
		result.TeamTokenMaxTTL = types.StringValue("2y")
	}
	if result.AuditTrailTokenMaxTTL.IsNull() {
		result.AuditTrailTokenMaxTTL = types.StringValue("2y")
	}
	if result.UserTokenMaxTTL.IsNull() {
		result.UserTokenMaxTTL = types.StringValue("2y")
	}

	return result
}

// Made with Bob
