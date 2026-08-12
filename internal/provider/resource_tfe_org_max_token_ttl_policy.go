// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/hashicorp/go-tfe/v2/api/models"
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

const minTFEVersionOrgMaxTokenTTLPolicy = "2.0.1"

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
		Description: "Manages the maximum time-to-live (TTL) policy for API tokens in an organization. When enabled, this policy enforces maximum lifespans for organization, team, audit trail, and user tokens. Any tokens that exceed the configured limits will cease to work." +
			"\n\n-> **Note:** To enable or disable the maximum TTL policy feature for an organization, use the `max_ttl_enabled` attribute on the `tfe_organization` resource." +
			"\n\n~> **Warning:** Maximum TTL policies are enforced immediately upon creation or update. Existing tokens that exceed newly configured limits will stop working and return an unauthorized error. Ensure all active tokens comply with the new limits before applying changes." +
			"\n\n~> **Note:** This resource requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v2.0.1." +
			"\n\nAll TTL attributes accept duration strings in the format `<number><unit>`, where unit is one of: `h` (hours), `d` (days), `w` (weeks), `mo` (months), or `y` (years). Decimal values are supported (e.g., `0.5h` = 30 minutes, `2.5d` = 2 days 12 hours).",
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
			"org_token_max_ttl": schema.StringAttribute{
				Description: "Maximum lifespan allowed for organization tokens to access the organization's resources. " +
					"Defaults to two years (`2y`). ",
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
					"Defaults to two years (`2y`). ",
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
					"Defaults to two years (`2y`). ",
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
					"Defaults to two years (`2y`). ",
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
				Description: "The computed maximum time-to-live for organization tokens, in milliseconds, as returned by the API.",
				Computed:    true,
			},
			"team_token_max_ttl_ms": schema.Int64Attribute{
				Description: "The computed maximum time-to-live for team tokens, in milliseconds, as returned by the API.",
				Computed:    true,
			},
			"audit_trail_token_max_ttl_ms": schema.Int64Attribute{
				Description: "The computed maximum time-to-live for audit trail tokens, in milliseconds, as returned by the API.",
				Computed:    true,
			},
			"user_token_max_ttl_ms": schema.Int64Attribute{
				Description: "The computed maximum time-to-live for user tokens, in milliseconds, as returned by the API.",
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
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) checkMaxTokenTTLPolicySupport() error {
	meetsMinVersionRequirement, err := r.config.MeetsMinRemoteTFEVersion(minTFEVersionOrgMaxTokenTTLPolicy)
	if err != nil {
		return fmt.Errorf("could not determine if Terraform Enterprise version %s meets minimum required version %s: %w",
			r.config.RemoteTFEVersion(), minTFEVersionOrgMaxTokenTTLPolicy, err)
	}
	if !meetsMinVersionRequirement {
		return fmt.Errorf("organization max token TTL policy requires Terraform Enterprise version %s or later. Current version: %s",
			minTFEVersionOrgMaxTokenTTLPolicy, r.config.RemoteTFEVersion())
	}
	return nil
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if TFE version supports max token TTL policy
	if err := r.checkMaxTokenTTLPolicySupport(); err != nil {
		resp.Diagnostics.AddError("Feature not supported", err.Error())
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

	response, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TokenTtlPolicies().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPoliciesV2(organization, response.GetData(), &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if TFE version supports max token TTL policy
	if err := r.checkMaxTokenTTLPolicySupport(); err != nil {
		resp.Diagnostics.AddError("Feature not supported", err.Error())
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.updateTokenTTLPolicies(ctx, organization, plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization token TTL policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if TFE version supports max token TTL policy
	if err := r.checkMaxTokenTTLPolicySupport(); err != nil {
		resp.Diagnostics.AddError("Feature not supported", err.Error())
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.Plan, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.updateTokenTTLPolicies(ctx, organization, plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update organization token TTL policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) updateTokenTTLPolicies(ctx context.Context, organization string, plan modelTFEOrgMaxTokenTTLPolicy) (modelTFEOrgMaxTokenTTLPolicy, error) {
	// Build policy update options from user's plan values
	entries, diagErr := r.buildPolicyEntries(plan)
	if diagErr != nil {
		return modelTFEOrgMaxTokenTTLPolicy{}, fmt.Errorf("invalid TTL values: %w", diagErr)
	}

	// Build the v2 request body.
	body := models.NewTokenTtlPoliciesEnvelope()
	policies := models.NewTokenTtlPolicies()
	policiesType := models.ORGANIZATIONTOKENTTLPOLICIES_TOKENTTLPOLICIES_TYPE
	policies.SetTypeEscaped(&policiesType)
	attrs := models.NewTokenTtlPolicies_attributes()
	attrs.SetTokenTtlPolicies(entries)
	policies.SetAttributes(attrs)
	body.SetData(policies)

	tflog.Debug(ctx, "Updating token TTL policies", map[string]any{
		"organization": organization,
	})

	response, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TokenTtlPolicies().Patch(ctx, body, nil)
	if err != nil {
		return modelTFEOrgMaxTokenTTLPolicy{}, fmt.Errorf("unable to update organization token TTL policies: %w", err)
	}

	return modelFromTokenTTLPoliciesV2(organization, response.GetData(), &plan), nil
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEOrgMaxTokenTTLPolicy

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if TFE version supports max token TTL policy
	if err := r.checkMaxTokenTTLPolicySupport(); err != nil {
		resp.Diagnostics.AddError("Feature not supported", err.Error())
		return
	}

	var organization string
	resp.Diagnostics.Append(r.config.dataOrDefaultOrganization(ctx, req.State, &organization)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting token TTL policy", map[string]any{
		"organization": organization,
	})

	// Reset all token types to their 2-year defaults.
	defaultTokenTypes := []models.TokenTtlPolicyEntry_tokenType{
		models.ORGANIZATION_TOKENTTLPOLICYENTRY_TOKENTYPE,
		models.TEAM_TOKENTTLPOLICYENTRY_TOKENTYPE,
		models.USER_TOKENTTLPOLICYENTRY_TOKENTYPE,
		models.AUDIT_TRAILS_TOKENTTLPOLICYENTRY_TOKENTYPE,
	}
	defaultMs := defaultTokenTTLMs
	var entries []models.TokenTtlPolicyEntryable
	for _, tt := range defaultTokenTypes {
		entry := models.NewTokenTtlPolicyEntry()
		entry.SetTokenType(&tt)
		entry.SetMaxTtlMs(&defaultMs)
		entries = append(entries, entry)
	}

	body := models.NewTokenTtlPoliciesEnvelope()
	policies := models.NewTokenTtlPolicies()
	policiesType := models.ORGANIZATIONTOKENTTLPOLICIES_TOKENTTLPOLICIES_TYPE
	policies.SetTypeEscaped(&policiesType)
	attrs := models.NewTokenTtlPolicies_attributes()
	attrs.SetTokenTtlPolicies(entries)
	policies.SetAttributes(attrs)
	body.SetData(policies)

	_, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TokenTtlPolicies().Patch(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete organization token TTL policy", err.Error())
		return
	}
}

func (r *resourceTFEOrgMaxTokenTTLPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization := req.ID

	// Check if TFE version supports max token TTL policy
	if err := r.checkMaxTokenTTLPolicySupport(); err != nil {
		resp.Diagnostics.AddError("Feature not supported", err.Error())
		return
	}

	tflog.Debug(ctx, "Importing token TTL policies", map[string]any{
		"organization": organization,
	})

	response, err := r.config.ClientV2.API.Organizations().ByOrganization_name(organization).TokenTtlPolicies().Get(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization token TTL policies", err.Error())
		return
	}

	result := modelFromTokenTTLPoliciesV2(organization, response.GetData(), nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// buildPolicyEntries converts a plan model to a slice of v2 TokenTtlPolicyEntry objects.
func (r *resourceTFEOrgMaxTokenTTLPolicy) buildPolicyEntries(plan modelTFEOrgMaxTokenTTLPolicy) ([]models.TokenTtlPolicyEntryable, error) {
	tokenConfigs := []struct {
		tokenType models.TokenTtlPolicyEntry_tokenType
		ttlValue  types.String
	}{
		{models.ORGANIZATION_TOKENTTLPOLICYENTRY_TOKENTYPE, plan.OrgTokenMaxTTL},
		{models.TEAM_TOKENTTLPOLICYENTRY_TOKENTYPE, plan.TeamTokenMaxTTL},
		{models.AUDIT_TRAILS_TOKENTTLPOLICYENTRY_TOKENTYPE, plan.AuditTrailTokenMaxTTL},
		{models.USER_TOKENTTLPOLICYENTRY_TOKENTYPE, plan.UserTokenMaxTTL},
	}

	var entries []models.TokenTtlPolicyEntryable
	for _, cfg := range tokenConfigs {
		if cfg.ttlValue.IsNull() || cfg.ttlValue.IsUnknown() {
			continue
		}
		ms, err := durationStringToMilliseconds(cfg.ttlValue.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid %s token TTL: %w", cfg.tokenType.String(), err)
		}
		tt := cfg.tokenType
		entry := models.NewTokenTtlPolicyEntry()
		entry.SetTokenType(&tt)
		entry.SetMaxTtlMs(&ms)
		entries = append(entries, entry)
	}

	return entries, nil
}

// modelFromTokenTTLPoliciesV2 builds a modelTFEOrgMaxTokenTTLPolicy from a v2
// list of TokenTtlPolicyable items.
//
//   - For Create/Update: pass plan to preserve user's exact input format.
//   - For Read: pass state to enable smart conversion.
//   - For ImportState: pass nil to convert all values to readable format.
func modelFromTokenTTLPoliciesV2(organization string, policies []models.TokenTtlPolicyable, stateOrPlan *modelTFEOrgMaxTokenTTLPolicy) modelTFEOrgMaxTokenTTLPolicy {
	result := modelTFEOrgMaxTokenTTLPolicy{
		ID:           types.StringValue(organization),
		Organization: types.StringValue(organization),
	}

	// Initialize with defaults for ImportState case
	result.OrgTokenMaxTTL = types.StringValue(defaultTokenTTL)
	result.TeamTokenMaxTTL = types.StringValue(defaultTokenTTL)
	result.AuditTrailTokenMaxTTL = types.StringValue(defaultTokenTTL)
	result.UserTokenMaxTTL = types.StringValue(defaultTokenTTL)

	for _, policy := range policies {
		if policy == nil {
			continue
		}
		attrs := policy.GetAttributes()
		if attrs == nil {
			continue
		}
		tokenType := attrs.GetTokenType()
		maxTTLMs := attrs.GetMaxTtlMs()
		if tokenType == nil || maxTTLMs == nil {
			continue
		}

		switch *tokenType {
		case models.ORGANIZATION_TOKENTTLPOLICY_ATTRIBUTES_TOKENTYPE:
			result.OrgTokenMaxTTLMs = types.Int64Value(*maxTTLMs)
			if stateOrPlan != nil {
				result.OrgTokenMaxTTL = types.StringValue(durationConversion(*maxTTLMs, stateOrPlan.OrgTokenMaxTTL.ValueString()))
			} else {
				result.OrgTokenMaxTTL = types.StringValue(millisecondsToDurationString(*maxTTLMs))
			}
		case models.TEAM_TOKENTTLPOLICY_ATTRIBUTES_TOKENTYPE:
			result.TeamTokenMaxTTLMs = types.Int64Value(*maxTTLMs)
			if stateOrPlan != nil {
				result.TeamTokenMaxTTL = types.StringValue(durationConversion(*maxTTLMs, stateOrPlan.TeamTokenMaxTTL.ValueString()))
			} else {
				result.TeamTokenMaxTTL = types.StringValue(millisecondsToDurationString(*maxTTLMs))
			}
		case models.AUDIT_TRAILS_TOKENTTLPOLICY_ATTRIBUTES_TOKENTYPE:
			result.AuditTrailTokenMaxTTLMs = types.Int64Value(*maxTTLMs)
			if stateOrPlan != nil {
				result.AuditTrailTokenMaxTTL = types.StringValue(durationConversion(*maxTTLMs, stateOrPlan.AuditTrailTokenMaxTTL.ValueString()))
			} else {
				result.AuditTrailTokenMaxTTL = types.StringValue(millisecondsToDurationString(*maxTTLMs))
			}
		case models.USER_TOKENTTLPOLICY_ATTRIBUTES_TOKENTYPE:
			result.UserTokenMaxTTLMs = types.Int64Value(*maxTTLMs)
			if stateOrPlan != nil {
				result.UserTokenMaxTTL = types.StringValue(durationConversion(*maxTTLMs, stateOrPlan.UserTokenMaxTTL.ValueString()))
			} else {
				result.UserTokenMaxTTL = types.StringValue(millisecondsToDurationString(*maxTTLMs))
			}
		}
	}

	return result
}

// Converts duration strings like "1y", "30d", "24h" to milliseconds
func durationStringToMilliseconds(duration string) (int64, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)(h|d|w|mo|y)$`)
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
		milliseconds = int64(math.Round(value * 60 * 60 * 1000))
	case "d": // days
		milliseconds = int64(math.Round(value * 24 * 60 * 60 * 1000))
	case "w": // weeks
		milliseconds = int64(math.Round(value * 7 * 24 * 60 * 60 * 1000))
	case "mo": // months (30 days)
		milliseconds = int64(math.Round(value * 30 * 24 * 60 * 60 * 1000))
	case "y": // years (365 days)
		milliseconds = int64(math.Round(value * 365 * 24 * 60 * 60 * 1000))
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return milliseconds, nil
}

// formatDuration formats a value with its unit, supporting fractional values
// Returns empty string if value < 1 or not cleanly representable
func formatDuration(value float64, unit string) string {
	if value < 1 {
		return ""
	}

	// Check for whole number
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d%s", int64(value), unit)
	}

	// Check for clean 1 decimal place (e.g., 5.5)
	if value*10 == float64(int64(value*10)) {
		return fmt.Sprintf("%.1f%s", value, unit)
	}

	// Check for clean 2 decimal places (e.g., 2.25)
	if value*100 == float64(int64(value*100)) {
		return fmt.Sprintf("%.2f%s", value, unit)
	}

	return ""
}

// Supports both whole and fractional values (e.g., 5y, 5.5y, 0.5h)
func millisecondsToDurationString(ms int64) string {
	hours := float64(ms) / (60 * 60 * 1000)

	// Try each unit in descending order of size
	if result := formatDuration(hours/8760, "y"); result != "" {
		return result
	}
	if result := formatDuration(hours/720, "mo"); result != "" {
		return result
	}
	if result := formatDuration(hours/168, "w"); result != "" {
		return result
	}
	if result := formatDuration(hours/24, "d"); result != "" {
		return result
	}
	if result := formatDuration(hours, "h"); result != "" {
		return result
	}

	// Fallback: format hours with 2 decimal places
	return fmt.Sprintf("%.2fh", hours)
}

func durationConversion(apiMs int64, userDuration string) string {
	// Parse the user's duration string to milliseconds
	userMs, err := durationStringToMilliseconds(userDuration)
	if err != nil {
		// If we can't parse the user's duration, just convert the API value
		return millisecondsToDurationString(apiMs)
	}

	// If the API value matches the user's value, preserve the user's format
	if apiMs == userMs {
		return userDuration
	}

	// Otherwise, convert the API value to a readable format (drift detected)
	return millisecondsToDurationString(apiMs)
}
