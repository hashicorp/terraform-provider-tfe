// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                     = &resourceTFEAdminBanner{}
	_ resource.ResourceWithConfigure        = &resourceTFEAdminBanner{}
	_ resource.ResourceWithImportState      = &resourceTFEAdminBanner{}
	_ resource.ResourceWithConfigValidators = &resourceTFEAdminBanner{}
	_ resource.ResourceWithValidateConfig   = &resourceTFEAdminBanner{}
)

// minTFEVersionAdminBanner is the minimum TFE release that ships the
// /api/v2/admin/banners endpoint. The commit message expresses this as
// "TFE v2.1.0+".
const minTFEVersionAdminBanner = "2.1.0"

// bannerScheduleAttributes are the JSON:API attribute names that make up the
// banner schedule. They must be set or cleared together.
var bannerScheduleAttributes = []string{"scheduled-publish-at", "scheduled-expire-at", "timezone"}

// NewAdminBannerResource returns the tfe_admin_banner resource.
func NewAdminBannerResource() resource.Resource {
	return &resourceTFEAdminBanner{}
}

// resourceTFEAdminBanner implements the tfe_admin_banner resource type
type resourceTFEAdminBanner struct {
	config ConfiguredClient
}

type modelTFEAdminBanner struct {
	ID                 types.String      `tfsdk:"id"`
	Title              types.String      `tfsdk:"title"`
	Body               types.String      `tfsdk:"body"`
	Style              types.String      `tfsdk:"style"`
	Audience           types.String      `tfsdk:"audience"`
	ScheduledPublishAt timetypes.RFC3339 `tfsdk:"scheduled_publish_at"`
	ScheduledExpireAt  timetypes.RFC3339 `tfsdk:"scheduled_expire_at"`
	Timezone           types.String      `tfsdk:"timezone"`
	PublishedAt        timetypes.RFC3339 `tfsdk:"published_at"`
}

func timeValueFromPtr(t *time.Time) timetypes.RFC3339 {
	if t == nil || t.IsZero() {
		return timetypes.NewRFC3339Null()
	}
	return timetypes.NewRFC3339TimeValue(*t)
}

func modelFromAdminBanner(b models.AdminBannersable) modelTFEAdminBanner {
	m := modelTFEAdminBanner{
		ID:                 types.StringValue(valueOrZero(b.GetId())),
		Title:              types.StringNull(),
		Body:               types.StringNull(),
		Style:              types.StringNull(),
		Audience:           types.StringNull(),
		Timezone:           types.StringNull(),
		ScheduledPublishAt: timetypes.NewRFC3339Null(),
		ScheduledExpireAt:  timetypes.NewRFC3339Null(),
		PublishedAt:        timetypes.NewRFC3339Null(),
	}

	attrs := b.GetAttributes()
	if attrs == nil {
		return m
	}

	if v := attrs.GetTitle(); v != nil {
		m.Title = types.StringValue(*v)
	}
	if v := attrs.GetBody(); v != nil {
		m.Body = types.StringValue(*v)
	}
	if v := enumStringOrEmpty(attrs.GetStyle()); v != "" {
		m.Style = types.StringValue(v)
	}
	if v := enumStringOrEmpty(attrs.GetAudience()); v != "" {
		m.Audience = types.StringValue(v)
	}
	if v := attrs.GetTimezone(); v != nil && *v != "" {
		m.Timezone = types.StringValue(*v)
	}
	m.ScheduledPublishAt = timeValueFromPtr(attrs.GetScheduledPublishAt())
	m.ScheduledExpireAt = timeValueFromPtr(attrs.GetScheduledExpireAt())
	m.PublishedAt = timeValueFromPtr(attrs.GetPublishedAt())

	return m
}

// bannerEnvelopeFromModel builds the JSON:API request body for a create or
// update. When clearSchedule is true, the schedule attributes are explicitly
// sent as null so that a previously configured schedule is removed; the
// generated model otherwise omits nil values, which a partial PATCH would
// interpret as "leave unchanged".
func bannerEnvelopeFromModel(m modelTFEAdminBanner, clearSchedule bool) (models.AdminBannersEnvelopeable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := models.NewAdminBanners_attributes()
	attrs.SetTitle(m.Title.ValueStringPointer())
	attrs.SetBody(m.Body.ValueStringPointer())

	if !m.Style.IsNull() && !m.Style.IsUnknown() {
		style := parseEnumPtr[models.AdminBanners_attributes_style](models.ParseAdminBanners_attributes_style, m.Style.ValueString())
		if style == nil {
			diags.AddError("Invalid style", fmt.Sprintf("%q is not a valid banner style", m.Style.ValueString()))
			return nil, diags
		}
		attrs.SetStyle(style)
	}
	if !m.Audience.IsNull() && !m.Audience.IsUnknown() {
		audience := parseEnumPtr[models.AdminBanners_attributes_audience](models.ParseAdminBanners_attributes_audience, m.Audience.ValueString())
		if audience == nil {
			diags.AddError("Invalid audience", fmt.Sprintf("%q is not a valid banner audience", m.Audience.ValueString()))
			return nil, diags
		}
		attrs.SetAudience(audience)
	}

	if !m.ScheduledPublishAt.IsNull() {
		publishAt, d := m.ScheduledPublishAt.ValueRFC3339Time()
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		attrs.SetScheduledPublishAt(&publishAt)
	}
	if !m.ScheduledExpireAt.IsNull() {
		expireAt, d := m.ScheduledExpireAt.ValueRFC3339Time()
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		attrs.SetScheduledExpireAt(&expireAt)
	}
	attrs.SetTimezone(m.Timezone.ValueStringPointer())

	if clearSchedule {
		additional := make(map[string]any, len(bannerScheduleAttributes))
		for _, name := range bannerScheduleAttributes {
			additional[name] = nil
		}
		attrs.SetAdditionalData(additional)
	}

	data := models.NewAdminBanners()
	data.SetTypeEscaped(ptr(models.ADMINBANNERS_ADMINBANNERS_TYPE))
	data.SetAttributes(attrs)

	envelope := models.NewAdminBannersEnvelope()
	envelope.SetData(data)
	return envelope, diags
}

// preserveEquivalentTime returns prior when prior and server denote the same
// instant, and server otherwise. TFE normalizes timestamps to UTC, but
// timetypes.RFC3339 (v0.5.0) only treats "Z" and "+00:00" as equal, not other
// offsets of the same instant. Keeping the user's representation avoids a
// spurious "inconsistent result after apply" and a perpetual plan diff, while a
// genuinely different instant still surfaces the server value.
func preserveEquivalentTime(prior, server timetypes.RFC3339) timetypes.RFC3339 {
	if prior.IsNull() || prior.IsUnknown() || server.IsNull() || server.IsUnknown() {
		return server
	}
	priorTime, diags := prior.ValueRFC3339Time()
	if diags.HasError() {
		return server
	}
	serverTime, diags := server.ValueRFC3339Time()
	if diags.HasError() {
		return server
	}
	if priorTime.Equal(serverTime) {
		return prior
	}
	return server
}

// unsupportedReason returns a diagnostic summary and detail when the connected
// instance does not support admin banners, or empty strings when it does.
// Banners are TFE-only, so HCP Terraform is rejected, as are TFE instances older
// than minTFEVersionAdminBanner.
func (r *resourceTFEAdminBanner) unsupportedReason() (summary, detail string) {
	meets, err := r.config.MeetsMinRemoteTFEOnlyVersion(minTFEVersionAdminBanner)
	if err != nil {
		return "Error checking minimum Terraform Enterprise version",
			fmt.Sprintf("Could not determine whether Terraform Enterprise version %s meets the minimum required version %s: %v",
				r.config.RemoteTFEVersion(), minTFEVersionAdminBanner, err)
	}
	if meets {
		return "", ""
	}
	if r.config.Client.IsCloud() {
		return "Unsupported platform",
			"tfe_admin_banner is only supported on Terraform Enterprise and cannot be used with HCP Terraform."
	}
	return "Unsupported Terraform Enterprise version",
		fmt.Sprintf("tfe_admin_banner requires Terraform Enterprise version %s or later, but the connected instance is running version %s.",
			minTFEVersionAdminBanner, r.config.RemoteTFEVersion())
}

// checkBannerSupport records an error diagnostic and returns false when the
// connected instance does not support admin banners.
func (r *resourceTFEAdminBanner) checkBannerSupport(diags *diag.Diagnostics) bool {
	summary, detail := r.unsupportedReason()
	if summary == "" {
		return true
	}
	diags.AddError(summary, detail)
	return false
}

// stateAfterWrite builds the post-apply state for Create and Update. Required
// and Optional attributes that the server may echo back differently are set to
// the planned value so Terraform's post-apply consistency check passes:
//
//   - body: the server sanitizes HTML. A warning is raised when it did, and the
//     next Read refreshes the stored value so the change shows up as a plan diff.
//   - scheduled_*: the server normalizes timestamps to UTC; the planned
//     representation is kept when it denotes the same instant.
func stateAfterWrite(plan modelTFEAdminBanner, server models.AdminBannersable, diags *diag.Diagnostics) modelTFEAdminBanner {
	result := modelFromAdminBanner(server)

	if !result.Body.IsNull() && !result.Body.Equal(plan.Body) {
		diags.AddWarning(
			"Terraform Enterprise modified the banner body",
			"The server sanitized `body` and stored a different value than configured. "+
				"The next `terraform plan` will show a diff until the configuration is updated to match the stored value.",
		)
	}
	result.Body = plan.Body
	result.ScheduledPublishAt = preserveEquivalentTime(plan.ScheduledPublishAt, result.ScheduledPublishAt)
	result.ScheduledExpireAt = preserveEquivalentTime(plan.ScheduledExpireAt, result.ScheduledExpireAt)
	return result
}

// Metadata implements resource.Resource
func (r *resourceTFEAdminBanner) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_banner"
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEAdminBanner) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.config = client
}

// ConfigValidators implements resource.ResourceWithConfigValidators
func (r *resourceTFEAdminBanner) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("scheduled_publish_at"),
			path.MatchRoot("scheduled_expire_at"),
			path.MatchRoot("timezone"),
		),
	}
}

// ValidateConfig implements resource.ResourceWithValidateConfig
func (r *resourceTFEAdminBanner) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config modelTFEAdminBanner
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Values that are null or not yet known cannot be compared.
	if config.ScheduledPublishAt.IsNull() || config.ScheduledPublishAt.IsUnknown() ||
		config.ScheduledExpireAt.IsNull() || config.ScheduledExpireAt.IsUnknown() {
		return
	}

	publishAt, diags := config.ScheduledPublishAt.ValueRFC3339Time()
	resp.Diagnostics.Append(diags...)
	expireAt, diags := config.ScheduledExpireAt.ValueRFC3339Time()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !expireAt.After(publishAt) {
		resp.Diagnostics.AddAttributeError(
			path.Root("scheduled_expire_at"),
			"Invalid scheduled_expire_at",
			"scheduled_expire_at must be after scheduled_publish_at.",
		)
	}
}

// Schema implements resource.Resource
func (r *resourceTFEAdminBanner) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a site-wide announcement banner on a Terraform Enterprise instance. " +
			"This resource is not supported on HCP Terraform.\n\n" +
			"Declaring this resource publishes the banner; destroying it deactivates the banner.\n\n" +
			"~> **NOTE:** Only one non-deactivated banner can exist on a Terraform Enterprise instance at a time. " +
			"Applying a second `tfe_admin_banner` resource fails with a `409 Conflict` error.\n\n" +
			"~> **NOTE:** Terraform Enterprise removes any HTML tags or attributes in `body` that it does not allow when the banner is saved. " +
			"If `body` contains disallowed markup, the stored value differs from the configuration and `terraform plan` reports a perpetual diff.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the banner.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The short heading of the banner. Maximum 100 characters.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"body": schema.StringAttribute{
				Description: "The body text of the banner. Basic HTML formatting, such as `<strong>`, `<em>`, and `<a href>` links, is supported. " +
					"Terraform Enterprise removes any HTML tags or attributes it does not allow when the banner is saved. " +
					"The maximum length of 500 characters is enforced by Terraform Enterprise on the saved value, after those tags and attributes are removed. " +
					"If the configured value contains markup that is removed, the next `terraform plan` will show a diff. " +
					"Update the configuration to match the stored value to resolve the diff.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"style": schema.StringAttribute{
				Description: "The visual style of the banner. Valid values are `info`, `warning`, and `critical`. Defaults to `info`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("info"),
				Validators: []validator.String{
					stringvalidator.OneOf("info", "warning", "critical"),
				},
			},
			"audience": schema.StringAttribute{
				Description: "Which users see the banner. Valid values are `all_users` and `authenticated_only`. Defaults to `authenticated_only`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("authenticated_only"),
				Validators: []validator.String{
					stringvalidator.OneOf("all_users", "authenticated_only"),
				},
			},
			"scheduled_publish_at": schema.StringAttribute{
				Description: "The date and time, in RFC3339 format (e.g., `2026-12-31T22:00:00Z`), when the banner becomes visible. " +
					"If set, `scheduled_expire_at` and `timezone` must also be set. If omitted, the banner is published immediately.",
				Optional:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"scheduled_expire_at": schema.StringAttribute{
				Description: "The date and time, in RFC3339 format, when the banner stops being visible. Must be after `scheduled_publish_at`. " +
					"If set, `scheduled_publish_at` and `timezone` must also be set. An expired banner is not destroyed automatically; " +
					"remove the resource from the configuration to deactivate it.",
				Optional:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"timezone": schema.StringAttribute{
				Description: "The IANA timezone identifier (e.g., `America/New_York`) used to display the schedule in the UI. " +
					"If set, `scheduled_publish_at` and `scheduled_expire_at` must also be set.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"published_at": schema.StringAttribute{
				Description: "The date and time, in RFC3339 format, when the banner was published.",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEAdminBanner) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFEAdminBanner
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.checkBannerSupport(&resp.Diagnostics) {
		return
	}

	body, diags := bannerEnvelopeFromModel(plan, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating admin banner")
	created, err := r.config.ClientV2.API.Admin().Banners().Post(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create admin banner", apiErrorDetail(err))
		return
	}
	if created == nil || created.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create admin banner", "no data was returned by the API")
		return
	}

	result := stateAfterWrite(plan, created.GetData(), &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFEAdminBanner) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEAdminBanner
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An unsupported version (e.g. TFE downgraded below minTFEVersionAdminBanner)
	// must not block refresh, or a plain `terraform destroy` could never reach
	// Delete. Keep the existing state untouched and warn instead; Create and
	// Update still fail with an error.
	if summary, detail := r.unsupportedReason(); summary != "" {
		resp.Diagnostics.AddWarning(summary, detail+" The resource state was not refreshed.")
		return
	}

	tflog.Debug(ctx, "Reading admin banner", map[string]any{"id": state.ID.ValueString()})
	banner, err := r.findBanner(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading admin banner", apiErrorDetail(err))
		return
	}
	if banner == nil {
		// The banner no longer exists (or was deactivated out of band).
		resp.State.RemoveResource(ctx)
		return
	}

	// Keep the configured representation of a timestamp when the server returns
	// the same instant in another offset, to avoid a perpetual plan diff.
	result := modelFromAdminBanner(banner)
	result.ScheduledPublishAt = preserveEquivalentTime(state.ScheduledPublishAt, result.ScheduledPublishAt)
	result.ScheduledExpireAt = preserveEquivalentTime(state.ScheduledExpireAt, result.ScheduledExpireAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEAdminBanner) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state modelTFEAdminBanner
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A schedule that was set in state but is no longer configured must be
	// cleared explicitly, since the PATCH only modifies fields it contains.
	clearSchedule := plan.ScheduledPublishAt.IsNull() && !state.ScheduledPublishAt.IsNull()

	if !r.checkBannerSupport(&resp.Diagnostics) {
		return
	}

	id := state.ID.ValueString()
	body, diags := bannerEnvelopeFromModel(plan, clearSchedule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating admin banner", map[string]any{"id": id})
	updated, err := r.config.ClientV2.API.Admin().Banners().ById(id).Patch(ctx, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update admin banner", apiErrorDetail(err))
		return
	}
	if updated == nil || updated.GetData() == nil {
		resp.Diagnostics.AddError("Unable to update admin banner", "no data was returned by the API")
		return
	}

	result := stateAfterWrite(plan, updated.GetData(), &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource. Deleting a banner deactivates it.
// Note: the version guard is intentionally omitted here so that destroy always
// succeeds regardless of the current TFE version (e.g. after a downgrade).
func (r *resourceTFEAdminBanner) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEAdminBanner
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Deactivating admin banner", map[string]any{"id": id})
	err := r.config.ClientV2.API.Admin().Banners().ById(id).Delete(ctx, nil)
	// Ignore 404s for delete
	if err != nil && !errors.Is(err, tfev2.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting admin banner", fmt.Sprintf("Couldn't delete admin banner %s: %s", id, apiErrorDetail(err)))
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFEAdminBanner) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !r.checkBannerSupport(&resp.Diagnostics) {
		return
	}

	banner, err := r.findBanner(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing admin banner", apiErrorDetail(err))
		return
	}
	if banner == nil {
		resp.Diagnostics.AddError(
			"Error importing admin banner",
			fmt.Sprintf("Admin banner %q does not exist or has been deactivated.", req.ID),
		)
		return
	}

	result := modelFromAdminBanner(banner)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// findBanner returns the non-deactivated banner with the given ID, or nil if
// there is none. The admin API only exposes a collection endpoint for reads.
func (r *resourceTFEAdminBanner) findBanner(ctx context.Context, id string) (models.AdminBannersable, error) {
	banners, err := r.config.ClientV2.API.Admin().Banners().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	if banners == nil {
		return nil, nil
	}
	for _, b := range banners.GetData() {
		if b != nil && valueOrZero(b.GetId()) == id {
			return b, nil
		}
	}
	return nil, nil
}
