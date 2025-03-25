// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
)

// NewOrganizationRunTaskResource
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
	Triggers        types.Set    `tfsdk:"triggers"`
	URL             types.String `tfsdk:"url"`
	WorkspaceID     types.String `tfsdk:"workspace_id"`
}

// modelFromTFENotificationConfiguration builds a modelTFENotificationConfiguration struct from a tfe.NotificationConfiguration value.
func modelFromTFENotificationConfiguration(v *tfe.NotificationConfiguration) (modelTFENotificationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := modelTFENotificationConfiguration{
		ID:              types.StringValue(v.ID),
		Name:            types.StringValue(v.Name),
		DestinationType: types.StringValue(string(v.DestinationType)),
		Enabled:         types.BoolValue(v.Enabled),
		WorkspaceID:     types.StringValue(v.SubscribableChoice.Workspace.ID),
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

	if v.URL != "" {
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
				},
			},

			"triggers": schema.SetAttribute{
				Description: "The array of triggers for which this notification configuration will send notifications. If omitted, no notification triggers are configured.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							string(tfe.NotificationTriggerChangeRequestCreated),
						),
					),
				},
			},

			"url": schema.StringAttribute{
				Description: "The HTTP or HTTPS URL where notification requests will be made. This value must not be provided if `email_addresses` or `email_user_ids` is present, or if `destination_type` is `email`.",
				Optional:    true,
				Validators: []validator.String{
					validators.AttributeRequiredIfValueString(
						"destination_type",
						[]string{"generic", "microsoft-teams", "slack"},
					),
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email"},
					),
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("email_addresses"),
						path.MatchRelative().AtParent().AtName("email_user_ids"),
					),
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

// Create implements resource.Resource
func (r *resourceTFENotificationConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFENotificationConfiguration

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()

	// Create a new options struct
	options := tfe.NotificationConfigurationCreateOptions{
		DestinationType: tfe.NotificationDestination(tfe.NotificationDestinationType(plan.DestinationType.ValueString())),
		Enabled:         plan.Enabled.ValueBoolPointer(),
		Name:            plan.Name.ValueStringPointer(),
		Token:           plan.Token.ValueStringPointer(),
		URL:             plan.URL.ValueStringPointer(),
		SubscribableChoice: &tfe.NotificationConfigurationSubscribableChoice{
			Workspace: &tfe.Workspace{ID: workspaceID},
		},
	}

	// Add triggers set to the options struct
	var triggers []types.String
	if diags := plan.Triggers.ElementsAs(ctx, &triggers, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	options.Triggers = []tfe.NotificationTriggerType{}
	for _, trigger := range triggers {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.ValueString()))
	}

	// Add email_addresses set to the options struct
	emailAddresses := make([]types.String, len(plan.EmailAddresses.Elements()))
	if diags := plan.EmailAddresses.ElementsAs(ctx, &emailAddresses, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	options.EmailAddresses = []string{}
	for _, emailAddress := range emailAddresses {
		options.EmailAddresses = append(options.EmailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDs := make([]types.String, len(plan.EmailUserIDs.Elements()))
	if diags := plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDs, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
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

	result, diags := modelFromTFENotificationConfiguration(nc)
	if diags != nil && diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
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
		resp.Diagnostics.AddError("Unable to read notification configuration", err.Error())
		return
	}

	// Restore token from state because it is write only
	if !state.Token.IsNull() {
		nc.Token = state.Token.ValueString()
	}

	result, diags := modelFromTFENotificationConfiguration(nc)
	if diags != nil && diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFENotificationConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFENotificationConfiguration
	var state modelTFENotificationConfiguration

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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
	if diags := plan.Triggers.ElementsAs(ctx, &triggers, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	options.Triggers = []tfe.NotificationTriggerType{}
	for _, trigger := range triggers {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.ValueString()))
	}

	// Add email_addresses set to the options struct
	emailAddresses := make([]types.String, len(plan.EmailAddresses.Elements()))
	if diags := plan.EmailAddresses.ElementsAs(ctx, &emailAddresses, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	options.EmailAddresses = []string{}
	for _, emailAddress := range emailAddresses {
		options.EmailAddresses = append(options.EmailAddresses, emailAddress.ValueString())
	}

	// Add email_user_ids set to the options struct
	emailUserIDs := make([]types.String, len(plan.EmailUserIDs.Elements()))
	if diags := plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDs, true); diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	options.EmailUsers = []*tfe.User{}
	for _, emailUserID := range emailUserIDs {
		options.EmailUsers = append(options.EmailUsers, &tfe.User{ID: emailUserID.ValueString()})
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

	result, diags := modelFromTFENotificationConfiguration(nc)
	if diags != nil && diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
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
