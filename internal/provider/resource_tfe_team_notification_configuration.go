// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/planmodifiers"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/validators"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &resourceTFETeamNotificationConfiguration{}
var _ resource.ResourceWithConfigure = &resourceTFETeamNotificationConfiguration{}
var _ resource.ResourceWithImportState = &resourceTFETeamNotificationConfiguration{}

func NewTeamNotificationConfigurationResource() resource.Resource {
	return &resourceTFETeamNotificationConfiguration{}
}

// resourceTFETeamNotificationConfiguration implements the tfe_team_notification_configuration resource type
type resourceTFETeamNotificationConfiguration struct {
	config ConfiguredClient
}

func (r *resourceTFETeamNotificationConfiguration) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_notification_configuration"
}

type modelTFETeamNotificationConfiguration struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DestinationType types.String `tfsdk:"destination_type"`
	EmailAddresses  types.Set    `tfsdk:"email_addresses"`
	EmailUserIDs    types.Set    `tfsdk:"email_user_ids"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Token           types.String `tfsdk:"token"`
	Triggers        types.Set    `tfsdk:"triggers"`
	URL             types.String `tfsdk:"url"`
	TeamID          types.String `tfsdk:"team_id"`
	TokenWO         types.String `tfsdk:"token_wo"`
}

// modelFromTFETeamNotificationConfiguration builds a modelTFETeamNotificationConfiguration
// struct from a tfe.TeamNotificationConfiguration value.
func modelFromTFETeamNotificationConfiguration(v *tfe.NotificationConfiguration, isWriteOnly bool, lastValue types.String) (*modelTFETeamNotificationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := modelTFETeamNotificationConfiguration{
		ID:              types.StringValue(v.ID),
		Name:            types.StringValue(v.Name),
		DestinationType: types.StringValue(string(v.DestinationType)),
		Enabled:         types.BoolValue(v.Enabled),
		TeamID:          types.StringValue(v.SubscribableChoice.Team.ID),
		Token:           types.StringValue(""),
	}

	if len(v.EmailAddresses) == 0 {
		result.EmailAddresses = types.SetNull(types.StringType)
	} else {
		emailAddresses, diags := types.SetValueFrom(ctx, types.StringType, v.EmailAddresses)
		if diags != nil && diags.HasError() {
			return nil, diags
		}
		result.EmailAddresses = emailAddresses
	}

	if len(v.Triggers) == 0 {
		result.Triggers = types.SetNull(types.StringType)
	} else {
		triggers, diags := types.SetValueFrom(ctx, types.StringType, v.Triggers)
		if diags != nil && diags.HasError() {
			return nil, diags
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

	if lastValue.String() != "" {
		result.Token = lastValue
	}

	if isWriteOnly {
		result.Token = types.StringNull()
	}

	if v.URL != "" {
		result.URL = types.StringValue(v.URL)
	}

	return &result, diags
}

func (r *resourceTFETeamNotificationConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a team notification configuration resource.",
		Version:     0,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the team notification configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Description: "Name of the team notification configuration.",
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
				Description: "Whether the team notification configuration should be enabled or not. Disabled configurations will not send any notifications. Defaults to `false`.",
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
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("token_wo")),
				},
			},

			"token_wo": schema.StringAttribute{
				Description: "A write-only secure token for the notification configuration, guaranteed not to be written to plan or state artifacts.",
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Validators: []validator.String{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email", "microsoft-teams", "slack"},
					),
					stringvalidator.ConflictsWith(path.MatchRoot("token")),
				},
				PlanModifiers: []planmodifier.String{
					planmodifiers.NewReplaceForWriteOnlyStringValue("token_wo"),
				},
			},

			"triggers": schema.SetAttribute{
				Description: "The array of triggers for which this team notification configuration will send notifications. If omitted, no notification triggers are configured.",
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

			"team_id": schema.StringAttribute{
				Description: "The ID of the team that owns the notification configuration.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFETeamNotificationConfiguration) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *resourceTFETeamNotificationConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config modelTFETeamNotificationConfiguration

	// Read Terraform plan and config data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get team
	teamID := plan.TeamID.ValueString()

	// Create a new options struct
	options := tfe.NotificationConfigurationCreateOptions{
		DestinationType: tfe.NotificationDestination(tfe.NotificationDestinationType(plan.DestinationType.ValueString())),
		Enabled:         plan.Enabled.ValueBoolPointer(),
		Name:            plan.Name.ValueStringPointer(),
		URL:             plan.URL.ValueStringPointer(),
		SubscribableChoice: &tfe.NotificationConfigurationSubscribableChoice{
			Team: &tfe.Team{ID: teamID},
		},
	}

	// Set Token from `token_wo` if set, otherwise use the normal value
	isWriteOnly := !config.TokenWO.IsNull()
	if isWriteOnly {
		options.Token = config.TokenWO.ValueStringPointer()
	} else {
		options.Token = plan.Token.ValueStringPointer()
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

	tflog.Debug(ctx, "Creating team notification configuration")
	tnc, err := r.config.Client.NotificationConfigurations.Create(ctx, teamID, options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create team notification configuration", err.Error())
		return
	} else if len(tnc.EmailUsers) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	// Restore token from plan because it is write only
	if !plan.Token.IsNull() {
		tnc.Token = plan.Token.ValueString()
	}

	result, diags := modelFromTFETeamNotificationConfiguration(tnc, isWriteOnly, plan.Token)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Write the hashed private token to the state if it was provided
	store := r.writeOnlyValueStore(resp.Private)
	resp.Diagnostics.Append(store.SetPriorValue(ctx, config.TokenWO)...)

	if diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFETeamNotificationConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFETeamNotificationConfiguration

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading team notification configuration %q", state.ID.ValueString()))
	tnc, err := r.config.Client.NotificationConfigurations.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("`Notification configuration %s no longer exists", state.ID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading notification configuration", "Could not read notification configuration, unexpected error: "+err.Error())
		}
		return
	}

	// Check if the parameter is write-only
	isWriteOnly, diags := r.writeOnlyValueStore(resp.Private).PriorValueExists(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	result, diags := modelFromTFETeamNotificationConfiguration(tnc, isWriteOnly, state.Token)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFETeamNotificationConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFETeamNotificationConfiguration
	var state modelTFETeamNotificationConfiguration
	var config modelTFETeamNotificationConfiguration

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

	tflog.Debug(ctx, "Updating team notification configuration")
	tnc, err := r.config.Client.NotificationConfigurations.Update(ctx, state.ID.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update team notification configuration", err.Error())
		return
	} else if len(tnc.EmailUsers) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	result, diags := modelFromTFETeamNotificationConfiguration(tnc, !config.TokenWO.IsNull(), plan.Token)
	if diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
	}

	// Write the hashed private key to the state if it was provided
	store := r.writeOnlyValueStore(resp.Private)
	resp.Diagnostics.Append(store.SetPriorValue(ctx, config.TokenWO)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *resourceTFETeamNotificationConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFETeamNotificationConfiguration

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting team notification configuration")
	err := r.config.Client.NotificationConfigurations.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete team notification configuration", err.Error())
		return
	}
}

func (r *resourceTFETeamNotificationConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *resourceTFETeamNotificationConfiguration) writeOnlyValueStore(private helpers.PrivateState) *helpers.WriteOnlyValueStore {
	return helpers.NewWriteOnlyValueStore(private, "token_wo")
}
