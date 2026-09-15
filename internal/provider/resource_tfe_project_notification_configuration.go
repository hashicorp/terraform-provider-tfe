// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFEProjectNotificationConfiguration{}
var _ resource.ResourceWithConfigure = &resourceTFEProjectNotificationConfiguration{}
var _ resource.ResourceWithImportState = &resourceTFEProjectNotificationConfiguration{}
var _ resource.ResourceWithModifyPlan = &resourceTFEProjectNotificationConfiguration{}

func NewProjectNotificationConfigurationResource() resource.Resource {
	return &resourceTFEProjectNotificationConfiguration{}
}

// resourceTFEProjectNotificationConfiguration implements the tfe_project_notification_configuration resource type
type resourceTFEProjectNotificationConfiguration struct {
	config ConfiguredClient
}

func (r *resourceTFEProjectNotificationConfiguration) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_notification_configuration"
}

type modelTFEProjectNotificationConfiguration struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DestinationType types.String `tfsdk:"destination_type"`
	EmailAddresses  types.Set    `tfsdk:"email_addresses"`
	EmailUserIDs    types.Set    `tfsdk:"email_user_ids"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Token           types.String `tfsdk:"token"`
	TokenWO         types.String `tfsdk:"token_wo"`
	TokenWOVersion  types.Int64  `tfsdk:"token_wo_version"`
	Triggers        types.Set    `tfsdk:"triggers"`
	URL             types.String `tfsdk:"url"`
	URLWO           types.String `tfsdk:"url_wo"`
	URLWOVersion    types.Int64  `tfsdk:"url_wo_version"`
	ProjectID       types.String `tfsdk:"project_id"`
}

// modelFromTFEProjectNotificationConfiguration builds a modelTFEProjectNotificationConfiguration
// struct from a v2 notification configuration resource.
func modelFromTFEProjectNotificationConfiguration(v models.NotificationConfigurationsable, tokenWOVersion types.Int64, urlWOVersion types.Int64, lastValue types.String, priorTriggers types.Set) (*modelTFEProjectNotificationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := modelTFEProjectNotificationConfiguration{
		ID:             types.StringValue(valueOrZero(v.GetId())),
		ProjectID:      types.StringValue(notificationConfigurationSubscribableID(v.GetRelationships())),
		TokenWOVersion: tokenWOVersion,
		URLWOVersion:   urlWOVersion,
	}

	var emailAddresses, triggers []string
	var url string
	if attrs := v.GetAttributes(); attrs != nil {
		result.Name = types.StringValue(valueOrZero(attrs.GetName()))
		result.Enabled = types.BoolValue(valueOrZero(attrs.GetEnabled()))
		if dt := attrs.GetDestinationType(); dt != nil {
			result.DestinationType = types.StringValue(dt.String())
		}
		emailAddresses = attrs.GetEmailAddresses()
		triggers = attrs.GetTriggers()
		url = valueOrZero(attrs.GetUrl())
	}

	if len(emailAddresses) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, emailAddresses)
		if d != nil && d.HasError() {
			return nil, d
		}
		result.EmailAddresses = set
	}

	if len(triggers) == 0 {
		// triggers is optional and not computed, so preserve the configured
		// intent (an explicit empty set vs. null) to avoid an inconsistent
		// result after apply.
		result.Triggers = priorTriggers
	} else {
		set, d := types.SetValueFrom(ctx, types.StringType, triggers)
		if d != nil && d.HasError() {
			return nil, d
		}
		result.Triggers = set
	}

	emailUserIDs := notificationConfigurationUserIDs(v.GetRelationships())
	if len(emailUserIDs) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, emailUserIDs)
		if d != nil && d.HasError() {
			return nil, d
		}
		result.EmailUserIDs = set
	}

	if lastValue.ValueString() != "" {
		result.Token = lastValue
	}

	if !tokenWOVersion.IsNull() {
		result.Token = types.StringNull()
	}

	isURLWriteOnly := !urlWOVersion.IsNull()
	if isURLWriteOnly {
		result.URL = types.StringNull()
	} else if url != "" {
		result.URL = types.StringValue(url)
	}

	return &result, diags
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFEProjectNotificationConfiguration) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is unconfigured (i.e. we're only validating config or something)
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

// Schema implements resource.Resource
func (r *resourceTFEProjectNotificationConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a project notification configuration." +
			"\n\nHCP Terraform can be configured to send notifications for run state transitions within a project. Project notification configurations allow you to specify a URL, destination type, and what events will trigger the notification. Each project can have up to 20 notification configurations, and they apply to all runs for all workspaces within that project.",
		Version: 0,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the project notification configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Name of the project notification configuration.",
				Required:    true,
			},

			"destination_type": schema.StringAttribute{
				Description: "The type of notification configuration payload to send. Valid values are `generic`, `email` (available in HCP Terraform or Terraform Enterprise v202206-1 or later), `slack`, and `microsoft-teams` (available in HCP Terraform or Terraform Enterprise v202206-1 or later).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"email",
						"generic",
						"slack",
						"microsoft-teams",
					),
				},
			},

			"email_addresses": schema.SetAttribute{
				MarkdownDescription: "(Terraform Enterprise Only) A list of email addresses. This value **must not** be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
					),
				},
			},

			"email_user_ids": schema.SetAttribute{
				MarkdownDescription: "A list of user IDs. This value **must not** be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
					),
				},
			},

			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the project notification configuration should be enabled or not. Disabled configurations will not send any notifications. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "A write-only secure token for the notification configuration, which can be used by the receiving server to verify request authenticity when configured for notification configurations with a destination type of `generic`. Defaults to `null`. This value **must not** be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email", "microsoft-teams", "slack"},
					),
					stringvalidator.ConflictsWith(path.MatchRoot("token_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("token_wo")),
				},
			},

			"token_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				Sensitive:           true,
				MarkdownDescription: "Write-only alternative to `token`. Never stored in Terraform state. Cannot be used with `token`. This value _must not_ be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`. The provider automatically detects changes by storing a SHA-256 hash of the value in [private state](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state) and incrementing `token_wo_version` when it changes. No additional configuration is required.\n\nFor maximum privacy — to prevent even the hash from being stored — omit `token_wo` from your config and set `token_wo_version` manually instead, incrementing it whenever you need to push a new token value.",
				Validators: []validator.String{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email", "microsoft-teams", "slack"},
					),
					stringvalidator.ConflictsWith(path.MatchRoot("token")),
				},
			},
			"token_wo_version": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Tracks the version of `token_wo`. In **auto-managed mode** (the default when `token_wo_version` is not set in config), the provider computes this value automatically: it is set to `1` on resource creation and incremented whenever the value of `token_wo` changes. In **manual mode** (when you explicitly set `token_wo_version` in config), auto-detection is disabled and you control updates by incrementing this value yourself — no hash is stored in private state. Cannot be used with `token`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("token")),
					int64validator.AlsoRequires(path.MatchRoot("token_wo")),
				},
			},

			"triggers": schema.SetAttribute{
				Description: "The array of triggers for which this project notification configuration will send notifications. If omitted, no notification triggers are configured. Valid values are `run:created`, `run:planning`, `run:needs_attention`, `run:applying`, `run:completed`, `run:errored`, `assessment:check_failure`, `assessment:drifted`, `assessment:failed`, `workspace:auto_destroy_reminder`, or `workspace:auto_destroy_run_results`.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"run:created",
							"run:planning",
							"run:needs_attention",
							"run:applying",
							"run:completed",
							"run:errored",
							"assessment:check_failure",
							"assessment:drifted",
							"assessment:failed",
							"workspace:auto_destroy_reminder",
							"workspace:auto_destroy_run_results",
						),
					),
				},
			},

			"url": schema.StringAttribute{
				MarkdownDescription: "The HTTP or HTTPS URL where notification requests will be made. This value must not be provided if `email_addresses` or `email_user_ids` is present, or if `destination_type` is `email`. Use `url_wo` instead to prevent the URL from being stored in state.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					validators.AttributeRequiredIfValueStringUnlessOtherSet(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
						"url_wo",
					),
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email"},
					),
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("email_addresses"),
						path.MatchRelative().AtParent().AtName("email_user_ids"),
						path.MatchRelative().AtParent().AtName("url_wo"),
					),
				},
			},

			"url_wo": schema.StringAttribute{
				Description: "Write-only alternative to `url`. The HTTP or HTTPS URL where notification requests will be made. Use this instead of `url` to prevent the URL from being stored in state. Changes are detected automatically via a hash stored in private state; increment `url_wo_version` manually to force an update without changing the value.",
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Validators: []validator.String{
					validators.AttributeRequiredIfValueStringUnlessOtherSet(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
						"url",
					),
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email"},
					),
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("email_addresses"),
						path.MatchRelative().AtParent().AtName("email_user_ids"),
						path.MatchRelative().AtParent().AtName("url"),
					),
				},
			},

			"url_wo_version": schema.Int64Attribute{
				Description: "Tracks the version of the write-only URL. When `url_wo` is set and this attribute is not explicitly configured, the provider automatically detects URL changes via a hash stored in private state and increments this value. Set this manually to force a URL update without changing the value, or for maximum privacy (disables hash storage).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("url")),
					int64validator.AlsoRequires(path.MatchRoot("url_wo")),
				},
			},

			"project_id": schema.StringAttribute{
				Description: "The ID of the project that owns the notification configuration.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFEProjectNotificationConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config modelTFEProjectNotificationConfiguration

	// Read Terraform plan and config data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()

	url := plan.URL.ValueStringPointer()
	// Set URL from `url_wo` if set, otherwise use the normal value
	if !config.URLWO.IsNull() {
		url = config.URLWO.ValueStringPointer()
	}

	lastTokenValue := types.StringValue("")
	token := plan.Token.ValueStringPointer()
	// Set Token from `token_wo` if set, otherwise use the normal value
	if !config.TokenWO.IsNull() {
		token = config.TokenWO.ValueStringPointer()
	} else {
		lastTokenValue = plan.Token
	}

	// Add triggers set to the options struct
	var triggerVals []types.String
	if diags := plan.Triggers.ElementsAs(ctx, &triggerVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	triggers := make([]string, 0, len(triggerVals))
	for _, trigger := range triggerVals {
		triggers = append(triggers, trigger.ValueString())
	}

	// Add email_addresses set to the options struct
	emailAddressVals := make([]types.String, len(plan.EmailAddresses.Elements()))
	if diags := plan.EmailAddresses.ElementsAs(ctx, &emailAddressVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	emailAddresses := make([]string, 0, len(emailAddressVals))
	for _, emailAddress := range emailAddressVals {
		emailAddresses = append(emailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDVals := make([]types.String, len(plan.EmailUserIDs.Elements()))
	if diags := plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	emailUserIDs := make([]string, 0, len(emailUserIDVals))
	for _, emailUserID := range emailUserIDVals {
		emailUserIDs = append(emailUserIDs, emailUserID.ValueString())
	}

	envelope, err := newNotificationConfigurationCreateEnvelope(projectID, models.PROJECTS_SUBSCRIBABLEIDENTIFIER_TYPE, plan.Name.ValueString(), plan.Enabled.ValueBool(), plan.DestinationType.ValueString(), token, url, triggers, emailAddresses, emailUserIDs)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create project notification configuration", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating project notification configuration")
	pncEnvelope, err := r.config.ClientV2.API.Projects().ByProject_id(projectID).NotificationConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create project notification configuration", err.Error())
		return
	}
	if pncEnvelope == nil || pncEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create project notification configuration", "No notification configuration data was returned by the API")
		return
	}
	pnc := pncEnvelope.GetData()

	if got := notificationConfigurationUserIDs(pnc.GetRelationships()); len(got) != len(emailUserIDs) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	result, diags := modelFromTFEProjectNotificationConfiguration(pnc, plan.TokenWOVersion, plan.URLWOVersion, lastTokenValue, plan.Triggers)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Store hashes in private state for auto change detection
	storeWOHash(ctx, resp.Private, "token_wo_hash", config.TokenWO, &resp.Diagnostics)
	storeWOHash(ctx, resp.Private, "url_wo_hash", config.URLWO, &resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFEProjectNotificationConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFEProjectNotificationConfiguration
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading project notification configuration %q", state.ID.ValueString()))
	pncEnvelope, err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Project notification configuration %s no longer exists", state.ID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading project notification configuration", "Could not read project notification configuration, unexpected error: "+err.Error())
		}
		return
	}
	if pncEnvelope == nil || pncEnvelope.GetData() == nil {
		tflog.Debug(ctx, fmt.Sprintf("Project notification configuration %s no longer exists", state.ID))
		resp.State.RemoveResource(ctx)
		return
	}

	result, diags := modelFromTFEProjectNotificationConfiguration(pncEnvelope.GetData(), state.TokenWOVersion, state.URLWOVersion, state.Token, state.Triggers)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFEProjectNotificationConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFEProjectNotificationConfiguration
	var state modelTFEProjectNotificationConfiguration
	var config modelTFEProjectNotificationConfiguration

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Read Terraform configuration data into the model so write-only values are available.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := plan.Token.ValueStringPointer()
	url := plan.URL.ValueStringPointer()

	// Preserve the previously known token unless this update explicitly changes token or token_wo.
	// The API never returns token values, so we must carry it forward in state to avoid sensitive value drift
	// when updates are triggered by unrelated attributes.
	lastTokenValue := state.Token

	tkn, isWOVal := r.determineTokenForUpdate(plan, state, config)
	// check is needed to prevent accidentally unsetting the token when no changes to token or token_wo were made
	// this is important when an update is triggered by changes in other attributes
	if tkn != nil {
		token = tkn

		if !isWOVal {
			lastTokenValue = types.StringValue(*tkn)
		}
	}

	// check is needed to prevent accidentally unsetting the URL when no changes to url or url_wo were made
	if u := r.determineURLForUpdate(plan, state, config); u != nil {
		url = u
	}

	// Add triggers set to the options struct
	triggerVals := make([]types.String, len(plan.Triggers.Elements()))
	if diags := plan.Triggers.ElementsAs(ctx, &triggerVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	triggers := make([]string, 0, len(triggerVals))
	for _, trigger := range triggerVals {
		triggers = append(triggers, trigger.ValueString())
	}

	// Add email_addresses set to the options struct
	emailAddressVals := make([]types.String, len(plan.EmailAddresses.Elements()))
	if diags := plan.EmailAddresses.ElementsAs(ctx, &emailAddressVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	emailAddresses := make([]string, 0, len(emailAddressVals))
	for _, emailAddress := range emailAddressVals {
		emailAddresses = append(emailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDVals := make([]types.String, len(plan.EmailUserIDs.Elements()))
	if diags := plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDVals, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	emailUserIDs := make([]string, 0, len(emailUserIDVals))
	for _, emailUserID := range emailUserIDVals {
		emailUserIDs = append(emailUserIDs, emailUserID.ValueString())
	}

	envelope, err := newNotificationConfigurationUpdateEnvelope(state.ID.ValueString(), plan.Name.ValueString(), plan.Enabled.ValueBool(), plan.DestinationType.ValueString(), token, url, triggers, emailAddresses, emailUserIDs)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update project notification configuration", err.Error())
		return
	}

	tflog.Debug(ctx, "Updating project notification configuration")
	pncEnvelope, err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update project notification configuration", err.Error())
		return
	}
	if pncEnvelope == nil || pncEnvelope.GetData() == nil {
		resp.Diagnostics.AddError("Unable to update project notification configuration", "No notification configuration data was returned by the API")
		return
	}
	pnc := pncEnvelope.GetData()

	if got := notificationConfigurationUserIDs(pnc.GetRelationships()); len(got) != len(emailUserIDs) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	result, diags := modelFromTFEProjectNotificationConfiguration(pnc, plan.TokenWOVersion, plan.URLWOVersion, lastTokenValue, plan.Triggers)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update hashes in private state for auto change detection
	storeWOHash(ctx, resp.Private, "token_wo_hash", config.TokenWO, &resp.Diagnostics)
	storeWOHash(ctx, resp.Private, "url_wo_hash", config.URLWO, &resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFEProjectNotificationConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFEProjectNotificationConfiguration
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting project notification configuration")
	err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Delete(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete project notification configuration", err.Error())
		return
	}
}

func (r *resourceTFEProjectNotificationConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// ModifyPlan implements resource.ResourceWithModifyPlan. It auto-manages token_wo_version
// and url_wo_version by hashing the write-only values and incrementing the version when
// the hash changes, unless the version is explicitly set in config (manual mode).
// It also blocks switching from a write-only attribute to its plaintext equivalent, which
// would expose a previously secret value in state.
func (r *resourceTFEProjectNotificationConfiguration) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	// Block write-only → plaintext transitions on existing resources
	if !req.State.Raw.IsNull() {
		blockWOToPlaintextTransition(ctx, req, resp, "token_wo_version", "token")
		if resp.Diagnostics.HasError() {
			return
		}
		blockWOToPlaintextTransition(ctx, req, resp, "url_wo_version", "url")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	r.modifyPlanWOVersion(ctx, req, resp, "token_wo", "token_wo_version", "token_wo_hash")
	if resp.Diagnostics.HasError() {
		return
	}
	r.modifyPlanWOVersion(ctx, req, resp, "url_wo", "url_wo_version", "url_wo_hash")
}

// modifyPlanWOVersion manages the auto-detection version for a write-only attribute.
// If the version attribute is explicitly set in config (manual mode), no auto-detection is performed.
func (r *resourceTFEProjectNotificationConfiguration) modifyPlanWOVersion(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	woAttr, versionAttr, hashKey string,
) {
	// If version is explicitly set in config, use manual mode — skip auto-detection
	var configVersion types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(versionAttr), &configVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !configVersion.IsNull() {
		return
	}

	// Get write-only value from config
	var woValue types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(woAttr), &woValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if woValue.IsNull() || woValue.IsUnknown() {
		// Write-only value not set — clear the version
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Null())...)
		return
	}

	newHash := computeWOHash(woValue.ValueString())

	// On create (no prior state), set initial version to 1
	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Value(1))...)
		return
	}

	// On update: compare new hash against stored hash in private state
	storedHashBytes, diags := req.Private.GetKey(ctx, hashKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storedHash string
	if storedHashBytes != nil {
		if err := json.Unmarshal(storedHashBytes, &storedHash); err != nil {
			resp.Diagnostics.AddError("Failed to decode "+woAttr+" hash", err.Error())
			return
		}
	}

	if !bytes.Equal([]byte(newHash), []byte(storedHash)) {
		// Hash changed — increment version
		var stateVersion types.Int64
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(versionAttr), &stateVersion)...)
		if resp.Diagnostics.HasError() {
			return
		}
		currentVersion := int64(0)
		if !stateVersion.IsNull() && !stateVersion.IsUnknown() {
			currentVersion = stateVersion.ValueInt64()
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Value(currentVersion+1))...)
	}
}

// determineURLForUpdate is invoked only after terraform determines that an attribute update is needed.
// It prevents accidentally unsetting the URL when changes to other attributes trigger an update.
// Returns nil if no URL update is needed.
func (r *resourceTFEProjectNotificationConfiguration) determineURLForUpdate(plan, state, config modelTFEProjectNotificationConfiguration) *string {
	usingWriteOnlyInPlan := !plan.URLWOVersion.IsNull()
	usingWriteOnlyInState := !state.URLWOVersion.IsNull()

	// Case 1: Switching FROM url TO url_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.URLWO.IsNull() {
		return config.URLWO.ValueStringPointer()
	}
	// Case 2: url_wo version changed in plan (auto-detected hash change or manual increment)
	if usingWriteOnlyInPlan && plan.URLWOVersion.ValueInt64() != state.URLWOVersion.ValueInt64() && !config.URLWO.IsNull() {
		return config.URLWO.ValueStringPointer()
	}
	// Case 3: Regular url changed
	if state.URL.ValueString() != plan.URL.ValueString() {
		return plan.URL.ValueStringPointer()
	}
	return nil
}

// determineTokenForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the token/token_wo attributes.
// this function compares the TokenWOVersion vs Token to ensure that during api update call, token is not mistakenly unset.
// Returns nil if no token update is needed.
func (r *resourceTFEProjectNotificationConfiguration) determineTokenForUpdate(plan, state, config modelTFEProjectNotificationConfiguration) (updateToken *string, isWOVal bool) {
	// Determine if we're using write-only token in plan vs state
	usingWriteOnlyInPlan := !plan.TokenWOVersion.IsNull()
	usingWriteOnlyInState := !state.TokenWOVersion.IsNull()

	// Case 1: Switching FROM token TO token_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.TokenWO.IsNull() {
		return config.TokenWO.ValueStringPointer(), true
	}
	// Case 2: Switching FROM token_wo TO token
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.Token.IsNull() {
		return plan.Token.ValueStringPointer(), false
	}
	// Case 3: token_wo version changed in plan
	if usingWriteOnlyInPlan && plan.TokenWOVersion.ValueInt64() != state.TokenWOVersion.ValueInt64() && !config.TokenWO.IsNull() {
		return config.TokenWO.ValueStringPointer(), true
	}
	// Case 4: Regular token changed. Only set Token if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the token on unrelated changes.
	if state.Token.ValueString() != plan.Token.ValueString() {
		return plan.Token.ValueStringPointer(), false
	}
	return nil, false
}
