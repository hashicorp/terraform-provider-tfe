// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var (
	_ resource.Resource                = &resourceTFENotificationConfiguration{}
	_ resource.ResourceWithConfigure   = &resourceTFENotificationConfiguration{}
	_ resource.ResourceWithImportState = &resourceTFENotificationConfiguration{}
	_ resource.ResourceWithModifyPlan  = &resourceTFENotificationConfiguration{}
)

// NewNotificationConfigurationResource
func NewNotificationConfigurationResource() resource.Resource {
	return &resourceTFENotificationConfiguration{}
}

type resourceTFENotificationConfiguration struct {
	config ConfiguredClient
}

// modelTFENotificationConfiguration maps the resource schema data to a struct.
type modelTFENotificationConfiguration struct {
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
	WorkspaceID     types.String `tfsdk:"workspace_id"`
}

// modelFromTFENotificationConfiguration builds a modelTFENotificationConfiguration struct from a tfe.NotificationConfiguration value.
func modelFromTFENotificationConfiguration(v *tfe.NotificationConfiguration, tokenWOVersion, urlWOVersion types.Int64) (modelTFENotificationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := modelTFENotificationConfiguration{
		ID:              types.StringValue(v.ID),
		Name:            types.StringValue(v.Name),
		DestinationType: types.StringValue(string(v.DestinationType)),
		Enabled:         types.BoolValue(v.Enabled),
		WorkspaceID:     types.StringValue(v.SubscribableChoice.Workspace.ID),
		TokenWOVersion:  tokenWOVersion,
		URLWOVersion:    urlWOVersion,
	}

	if len(v.EmailAddresses) == 0 {
		result.EmailAddresses = types.SetNull(types.StringType)
	} else {
		emailAddresses, diags := types.SetValueFrom(ctx, types.StringType, v.EmailAddresses)
		if diags != nil && diags.HasError() {
			return result, diags
		}
		result.EmailAddresses = emailAddresses
	}

	if len(v.Triggers) == 0 {
		result.Triggers = types.SetNull(types.StringType)
	} else {
		triggers, diags := types.SetValueFrom(ctx, types.StringType, v.Triggers)
		if diags != nil && diags.HasError() {
			return result, diags
		}

		result.Triggers = triggers
	}

	if len(v.EmailUsers) == 0 {
		result.EmailUserIDs = types.SetNull(types.StringType)
	} else {
		emailUserIDs := make([]attr.Value, len(v.EmailUsers))
		for i, emailUser := range v.EmailUsers {
			emailUserIDs[i] = types.StringValue(emailUser.ID)
		}

		result.EmailUserIDs = types.SetValueMust(types.StringType, emailUserIDs)
	}

	if v.Token != "" {
		result.Token = types.StringValue(v.Token)
	}
	isWriteOnlyValue := !tokenWOVersion.IsNull()
	// Don't retrieve values if write-only is being used. Unset the value and readable_value fields before updating the state.
	if isWriteOnlyValue {
		result.Token = types.StringNull()
	}

	isURLWriteOnly := !urlWOVersion.IsNull()
	if isURLWriteOnly {
		result.URL = types.StringNull()
	} else if v.URL != "" {
		result.URL = types.StringValue(v.URL)
	}

	return result, diags
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFENotificationConfiguration) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata implements resource.Resource
func (r *resourceTFENotificationConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_configuration"
}

// Schema implements resource.Resource
func (r *resourceTFENotificationConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a notification configuration resource.",
		Version:     0,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the notification configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Name of the notification configuration.",
				Required:    true,
			},

			"destination_type": schema.StringAttribute{
				Description: "The type of notification configuration payload to send.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.NotificationDestinationTypeEmail),
						string(tfe.NotificationDestinationTypeGeneric),
						string(tfe.NotificationDestinationTypeSlack),
						string(tfe.NotificationDestinationTypeMicrosoftTeams),
					),
				},
			},

			"email_addresses": schema.SetAttribute{
				Description: "A list of email addresses. This value must not be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
					),
				},
			},

			"email_user_ids": schema.SetAttribute{
				Description: "A list of user IDs. This value must not be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
					),
				},
			},

			"enabled": schema.BoolAttribute{
				Description: "Whether the notification configuration should be enabled or not. Disabled configurations will not send any notifications. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"token": schema.StringAttribute{
				Description: "A write-only secure token for the notification configuration, which can be used by the receiving server to verify request authenticity when configured for notification configurations with a destination type of `generic`. Defaults to `null`. This value _must not_ be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email", "microsoft-teams", "slack"},
					),
					stringvalidator.ConflictsWith(path.MatchRoot("token_wo")),
				},
			},
			"token_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Value of the token in write-only mode",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("token")),
					stringvalidator.AlsoRequires(path.MatchRoot("token_wo_version")),
				},
			},
			"token_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only token to trigger updates",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("token")),
					int64validator.AlsoRequires(path.MatchRoot("token_wo")),
				},
			},
			"triggers": schema.SetAttribute{
				Description: "The array of triggers for which this notification configuration will send notifications. If omitted, no notification triggers are configured.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							string(tfe.NotificationTriggerCreated),
							string(tfe.NotificationTriggerPlanning),
							string(tfe.NotificationTriggerNeedsAttention),
							string(tfe.NotificationTriggerApplying),
							string(tfe.NotificationTriggerCompleted),
							string(tfe.NotificationTriggerErrored),
							string(tfe.NotificationTriggerAssessmentCheckFailed),
							string(tfe.NotificationTriggerAssessmentDrifted),
							string(tfe.NotificationTriggerAssessmentFailed),
							string(tfe.NotificationTriggerWorkspaceAutoDestroyReminder),
							string(tfe.NotificationTriggerWorkspaceAutoDestroyRunResults),
						),
					),
				},
			},

			"url": schema.StringAttribute{
				Description: "The HTTP or HTTPS URL where notification requests will be made. This value must not be provided if `email_addresses` or `email_user_ids` is present, or if `destination_type` is `email`. Use `url_wo` instead to prevent the URL from being stored in state.",
				Optional:    true,
				Sensitive:   true,
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

			"workspace_id": schema.StringAttribute{
				Description: "The ID of the workspace that owns the notification configuration.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// computeURLWOHash returns a hex-encoded SHA-256 hash of the given URL.
func computeURLWOHash(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:])
}

// ModifyPlan implements resource.ResourceWithModifyPlan. It auto-manages url_wo_version
// by hashing the url_wo value and incrementing the version when the hash changes,
// unless url_wo_version is explicitly set in config (manual mode).
func (r *resourceTFENotificationConfiguration) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	// If url_wo_version is explicitly set in config, use manual mode — skip auto-detection
	var configURLWOVersion types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("url_wo_version"), &configURLWOVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !configURLWOVersion.IsNull() {
		return
	}

	// Get url_wo from config
	var urlWO types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("url_wo"), &urlWO)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if urlWO.IsNull() || urlWO.IsUnknown() {
		// url_wo not set — clear the version
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("url_wo_version"), types.Int64Null())...)
		return
	}

	newHash := computeURLWOHash(urlWO.ValueString())

	// On create (no prior state), set initial version to 1
	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("url_wo_version"), types.Int64Value(1))...)
		return
	}

	// On update: compare new hash against stored hash in private state
	storedHashBytes, diags := req.Private.GetKey(ctx, "url_wo_hash")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storedHash string
	if storedHashBytes != nil {
		if err := json.Unmarshal(storedHashBytes, &storedHash); err != nil {
			resp.Diagnostics.AddError("Failed to decode url_wo hash", err.Error())
			return
		}
	}

	if !bytes.Equal([]byte(newHash), []byte(storedHash)) {
		// Hash changed — increment version
		var stateVersion types.Int64
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("url_wo_version"), &stateVersion)...)
		if resp.Diagnostics.HasError() {
			return
		}
		currentVersion := int64(0)
		if !stateVersion.IsNull() && !stateVersion.IsUnknown() {
			currentVersion = stateVersion.ValueInt64()
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("url_wo_version"), types.Int64Value(currentVersion+1))...)
	}
}

// Create implements resource.Resource
func (r *resourceTFENotificationConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config modelTFENotificationConfiguration
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()

	// Create a new options struct
	options := tfe.NotificationConfigurationCreateOptions{
		DestinationType: tfe.NotificationDestination(tfe.NotificationDestinationType(plan.DestinationType.ValueString())),
		Enabled:         plan.Enabled.ValueBoolPointer(),
		Name:            plan.Name.ValueStringPointer(),
		URL:             plan.URL.ValueStringPointer(),
		SubscribableChoice: &tfe.NotificationConfigurationSubscribableChoice{
			Workspace: &tfe.Workspace{ID: workspaceID},
		},
	}

	// Set Value from `token_wo` if set, otherwise use the normal value
	if !config.TokenWO.IsNull() {
		options.Token = config.TokenWO.ValueStringPointer()
	} else {
		options.Token = plan.Token.ValueStringPointer()
	}

	// Set URL from `url_wo` if set, otherwise use the normal value
	if !config.URLWO.IsNull() {
		options.URL = config.URLWO.ValueStringPointer()
	}

	// Add triggers set to the options struct
	var triggers []types.String
	resp.Diagnostics.Append(plan.Triggers.ElementsAs(ctx, &triggers, true)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options.Triggers = []tfe.NotificationTriggerType{}
	for _, trigger := range triggers {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.ValueString()))
	}

	// Add email_addresses set to the options struct
	emailAddresses := make([]types.String, len(plan.EmailAddresses.Elements()))
	resp.Diagnostics.Append(plan.EmailAddresses.ElementsAs(ctx, &emailAddresses, true)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options.EmailAddresses = []string{}
	for _, emailAddress := range emailAddresses {
		options.EmailAddresses = append(options.EmailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDs := make([]types.String, len(plan.EmailUserIDs.Elements()))
	resp.Diagnostics.Append(plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDs, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options.EmailUsers = []*tfe.User{}
	for _, emailUserID := range emailUserIDs {
		options.EmailUsers = append(options.EmailUsers, &tfe.User{ID: emailUserID.ValueString()})
	}

	tflog.Debug(ctx, "Creating notification configuration")

	nc, err := r.config.Client.NotificationConfigurations.Create(ctx, workspaceID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create notification configuration", err.Error())
		return
	} else if len(nc.EmailUsers) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	// Restore token from plan because it is write only
	if !plan.Token.IsNull() {
		nc.Token = plan.Token.ValueString()
	}

	// We got a notification, so set state to new values
	result, diags := modelFromTFENotificationConfiguration(nc, config.TokenWOVersion, plan.URLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store hash of url_wo in private state for auto change detection
	if !config.URLWO.IsNull() {
		hash := computeURLWOHash(config.URLWO.ValueString())
		hashJSON, err := json.Marshal(hash)
		if err != nil {
			resp.Diagnostics.AddError("Failed to encode url_wo hash", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, "url_wo_hash", hashJSON)...)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Read implements resource.Resource
func (r *resourceTFENotificationConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFENotificationConfiguration
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading notification configuration %q", state.ID.ValueString()))
	nc, err := r.config.Client.NotificationConfigurations.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("`Notification configuration %s no longer exists", state.ID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading notification configuration", "Could not read notification configuration, unexpected error: "+err.Error())
		}
		return
	}

	// Restore token from state because it is write only
	if !state.Token.IsNull() {
		nc.Token = state.Token.ValueString()
	}

	result, diags := modelFromTFENotificationConfiguration(nc, state.TokenWOVersion, state.URLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFENotificationConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config modelTFENotificationConfiguration
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new options struct
	options := tfe.NotificationConfigurationUpdateOptions{
		Enabled: plan.Enabled.ValueBoolPointer(),
		Name:    plan.Name.ValueStringPointer(),
		Token:   plan.Token.ValueStringPointer(),
		URL:     plan.URL.ValueStringPointer(),
	}

	// Add triggers set to the options struct
	triggers := make([]types.String, len(plan.Triggers.Elements()))
	resp.Diagnostics.Append(plan.Triggers.ElementsAs(ctx, &triggers, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options.Triggers = []tfe.NotificationTriggerType{}
	for _, trigger := range triggers {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.ValueString()))
	}

	// Add email_addresses set to the options struct
	emailAddresses := make([]types.String, len(plan.EmailAddresses.Elements()))
	resp.Diagnostics.Append(plan.EmailAddresses.ElementsAs(ctx, &emailAddresses, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options.EmailAddresses = []string{}
	for _, emailAddress := range emailAddresses {
		options.EmailAddresses = append(options.EmailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDs := make([]types.String, len(plan.EmailUserIDs.Elements()))
	resp.Diagnostics.Append(plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDs, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options.EmailUsers = []*tfe.User{}
	for _, emailUserID := range emailUserIDs {
		options.EmailUsers = append(options.EmailUsers, &tfe.User{ID: emailUserID.ValueString()})
	}

	// check is needed to prevent accidentally unsetting the token when no changes to token or token_wo were made
	// this is important when an update is triggered by changes in other attributes
	if tkn := r.determineTokenForUpdate(plan, state, config); tkn != nil {
		options.Token = tkn
	}

	// check is needed to prevent accidentally unsetting the URL when no changes to url or url_wo were made
	if u := r.determineURLForUpdate(plan, state, config); u != nil {
		options.URL = u
	}

	tflog.Debug(ctx, "Updating notification configuration")
	nc, err := r.config.Client.NotificationConfigurations.Update(ctx, state.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update notification configuration", err.Error())
		return
	} else if len(nc.EmailUsers) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	// Restore token from plan because it is write only
	if !plan.Token.IsNull() {
		nc.Token = plan.Token.ValueString()
	}

	result, diags := modelFromTFENotificationConfiguration(nc, config.TokenWOVersion, plan.URLWOVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update hash of url_wo in private state for auto change detection
	if !config.URLWO.IsNull() {
		hash := computeURLWOHash(config.URLWO.ValueString())
		hashJSON, err := json.Marshal(hash)
		if err != nil {
			resp.Diagnostics.AddError("Failed to encode url_wo hash", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, "url_wo_hash", hashJSON)...)
	} else if !state.URLWOVersion.IsNull() && plan.URLWOVersion.IsNull() {
		// Switching from url_wo to url — clear the stored hash
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, "url_wo_hash", nil)...)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Delete implements resource.Resource
func (r *resourceTFENotificationConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFENotificationConfiguration
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting notification configuration")
	err := r.config.Client.NotificationConfigurations.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete notification configuration", err.Error())
		return
	}
}

func (r *resourceTFENotificationConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// determineURLForUpdate is invoked only after terraform determines that an attribute update is needed.
// It prevents accidentally unsetting the URL when changes to other attributes trigger an update.
// Returns nil if no URL update is needed.
func (r *resourceTFENotificationConfiguration) determineURLForUpdate(plan, state, config modelTFENotificationConfiguration) *string {
	usingWriteOnlyInPlan := !plan.URLWOVersion.IsNull()
	usingWriteOnlyInState := !state.URLWOVersion.IsNull()

	// Case 1: Switching FROM url TO url_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.URLWO.IsNull() {
		return config.URLWO.ValueStringPointer()
	}
	// Case 2: Switching FROM url_wo TO url
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.URL.IsNull() {
		return plan.URL.ValueStringPointer()
	}
	// Case 3: url_wo version changed in plan (auto-detected hash change or manual increment)
	if usingWriteOnlyInPlan && plan.URLWOVersion.ValueInt64() != state.URLWOVersion.ValueInt64() && !config.URLWO.IsNull() {
		return config.URLWO.ValueStringPointer()
	}
	// Case 4: Regular url changed
	if state.URL.ValueString() != plan.URL.ValueString() {
		return plan.URL.ValueStringPointer()
	}
	return nil
}

// determineTokenForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the token/token_wo attributes.
// this function compares the TokenWOVersion vs Token to ensure that during api update call, token is not mistakenly unset.
// Returns nil if no token update is needed.
func (r *resourceTFENotificationConfiguration) determineTokenForUpdate(plan, state, config modelTFENotificationConfiguration) *string {
	// Determine if we're using write-only token in plan vs state
	usingWriteOnlyInPlan := !plan.TokenWOVersion.IsNull()
	usingWriteOnlyInState := !state.TokenWOVersion.IsNull()

	// Case 1: Switching FROM token TO token_wo
	if !usingWriteOnlyInState && usingWriteOnlyInPlan && !config.TokenWO.IsNull() {
		return config.TokenWO.ValueStringPointer()
	}
	// Case 2: Switching FROM token_wo TO token
	if usingWriteOnlyInState && !usingWriteOnlyInPlan && !plan.Token.IsNull() {
		return plan.Token.ValueStringPointer()
	}
	// Case 3: token_wo version changed in plan
	if usingWriteOnlyInPlan && plan.TokenWOVersion.ValueInt64() != state.TokenWOVersion.ValueInt64() && !config.TokenWO.IsNull() {
		return config.TokenWO.ValueStringPointer()
	}
	// Case 4: Regular token changed. Only set Token if our planned value would be a CHANGE from
	// the prior state. This prevents accidentally resetting the token on unrelated changes.
	if state.Token.ValueString() != plan.Token.ValueString() {
		return plan.Token.ValueStringPointer()
	}
	return nil
}
