// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	TokenWO         types.String `tfsdk:"token_wo"`
	TokenWOVersion  types.Int64  `tfsdk:"token_wo_version"`
	Triggers        types.Set    `tfsdk:"triggers"`
	URL             types.String `tfsdk:"url"`
	TeamID          types.String `tfsdk:"team_id"`
}

// setNotificationAttributeCollections populates the triggers and email_addresses
// attributes on attrs from plan, and returns the constructed users relationship
// (for email_user_ids). It is called identically from both Create and Update.
func setNotificationAttributeCollections(ctx context.Context, plan modelTFETeamNotificationConfiguration, attrs models.NotificationConfigurations_attributesable) (models.NotificationConfigurations_relationships_usersable, diag.Diagnostics) {
	var diags diag.Diagnostics

	var triggers []types.String
	if d := plan.Triggers.ElementsAs(ctx, &triggers, true); d != nil && d.HasError() {
		return nil, d
	}
	triggerValues := make([]string, 0, len(triggers))
	for _, t := range triggers {
		triggerValues = append(triggerValues, t.ValueString())
	}
	attrs.SetTriggers(triggerValues)

	emailAddresses := make([]types.String, len(plan.EmailAddresses.Elements()))
	if d := plan.EmailAddresses.ElementsAs(ctx, &emailAddresses, true); d != nil && d.HasError() {
		return nil, d
	}
	emailAddressValues := make([]string, 0, len(emailAddresses))
	for _, ea := range emailAddresses {
		emailAddressValues = append(emailAddressValues, ea.ValueString())
	}
	attrs.SetEmailAddresses(emailAddressValues)

	emailUserIDs := make([]types.String, len(plan.EmailUserIDs.Elements()))
	if d := plan.EmailUserIDs.ElementsAs(ctx, &emailUserIDs, true); d != nil && d.HasError() {
		return nil, d
	}
	emailUserData := make([]models.NotificationConfigurations_relationships_users_dataable, 0, len(emailUserIDs))
	for _, id := range emailUserIDs {
		userData := models.NewNotificationConfigurations_relationships_users_data()
		userData.SetId(id.ValueStringPointer())
		userData.SetTypeEscaped(ptr(models.USERS_NOTIFICATIONCONFIGURATIONS_RELATIONSHIPS_USERS_DATA_TYPE))
		emailUserData = append(emailUserData, userData)
	}
	users := models.NewNotificationConfigurations_relationships_users()
	users.SetData(emailUserData)

	return users, diags
}

// modelFromTFETeamNotificationConfiguration builds a modelTFETeamNotificationConfiguration
// struct from a go-tfe v2 NotificationConfigurations value.
func modelFromTFETeamNotificationConfiguration(ctx context.Context, v models.NotificationConfigurationsable, tokenWOVersion types.Int64, lastValue types.String) (*modelTFETeamNotificationConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := v.GetAttributes()

	var teamID string
	if relationships := v.GetRelationships(); relationships != nil && relationships.GetSubscribable() != nil && relationships.GetSubscribable().GetData() != nil {
		teamID = valueOrZero(relationships.GetSubscribable().GetData().GetId())
	}

	destinationType := ""
	if dt := attrs.GetDestinationType(); dt != nil {
		destinationType = dt.String()
	}

	result := modelTFETeamNotificationConfiguration{
		ID:              types.StringValue(valueOrZero(v.GetId())),
		Name:            types.StringValue(valueOrZero(attrs.GetName())),
		DestinationType: types.StringValue(destinationType),
		Enabled:         types.BoolValue(valueOrZero(attrs.GetEnabled())),
		TeamID:          types.StringValue(teamID),
		TokenWOVersion:  tokenWOVersion,
		Token:           types.StringValue(""),
	}

	emailAddressValues := attrs.GetEmailAddresses()
	if len(emailAddressValues) == 0 {
		result.EmailAddresses = types.SetNull(types.StringType)
	} else {
		emailAddresses, diags := types.SetValueFrom(ctx, types.StringType, emailAddressValues)
		if diags != nil && diags.HasError() {
			return nil, diags
		}
		result.EmailAddresses = emailAddresses
	}

	triggerValues := attrs.GetTriggers()
	if len(triggerValues) == 0 {
		result.Triggers = types.SetNull(types.StringType)
	} else {
		triggers, diags := types.SetValueFrom(ctx, types.StringType, triggerValues)
		if diags != nil && diags.HasError() {
			return nil, diags
		}

		result.Triggers = triggers
	}

	var emailUserData []models.NotificationConfigurations_relationships_users_dataable
	if relationships := v.GetRelationships(); relationships != nil && relationships.GetUsers() != nil {
		emailUserData = relationships.GetUsers().GetData()
	}
	if len(emailUserData) == 0 {
		result.EmailUserIDs = types.SetNull(types.StringType)
	} else {
		emailUserIDs := make([]attr.Value, len(emailUserData))
		for i, emailUser := range emailUserData {
			emailUserIDs[i] = types.StringValue(valueOrZero(emailUser.GetId()))
		}

		result.EmailUserIDs = types.SetValueMust(types.StringType, emailUserIDs)
	}

	if lastValue.String() != "" {
		result.Token = lastValue
	}

	// Don't retrieve values if write-only is being used. Unset the token field before updating the state.
	isWriteOnlyValue := !tokenWOVersion.IsNull()
	if isWriteOnlyValue {
		result.Token = types.StringNull()
	}

	if url := valueOrZero(attrs.GetUrl()); url != "" {
		result.URL = types.StringValue(url)
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
				MarkdownDescription: "(TFE Only) A list of email addresses. This value must not be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
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
				MarkdownDescription: "A list of user IDs. This value must not be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.",
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
				MarkdownDescription: "Whether the team notification configuration should be enabled or not. Disabled configurations will not send any notifications. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"token": schema.StringAttribute{
				MarkdownDescription: "A write-only secure token for the notification configuration, which can be used by the receiving server to verify request authenticity when configured for notification configurations with a destination type of `generic`. Defaults to `null`. This value _must not_ be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.",
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
			// since the token_wo write-only values are not saved to state, they will not trigger updates on their own.
			// Instead the token_wo_version responsibility is to trigger updates to the token_wo attribute when version number changes.
			"token_wo": schema.StringAttribute{
				Description: "Write-only secure token for the notification configuration, which can be used by the receiving server to verify request authenticity when configured for notification configurations with a destination type of `generic`. Either `token` or `token_wo` can be provided, but not both. Must be used with `token_wo_version`. This value must not be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.",
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Validators: []validator.String{
					validators.AttributeValueConflictValidator(
						"destination_type",
						[]string{"email", "microsoft-teams", "slack"},
					),
					stringvalidator.ConflictsWith(path.MatchRoot("token")),
					stringvalidator.AlsoRequires(path.MatchRoot("token_wo_version")),
				},
			},

			"token_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Version of the write-only token. This field is used to trigger updates when the write-only token changes. Must be used with `token_wo`. When `token_wo_version` changes, the write-only token will be updated.",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("token")),
					int64validator.AlsoRequires(path.MatchRoot("token_wo")),
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
				MarkdownDescription: "The HTTP or HTTPS URL where notification requests will be made. This value must not be provided if `email_addresses` or `email_user_ids` is present, or if `destination_type` is `email`.",
				Optional:            true,
				Sensitive:           true,
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

	destinationType, diags := models.ParseNotificationConfigurations_attributes_destinationType(plan.DestinationType.ValueString())
	if diags != nil {
		resp.Diagnostics.AddError("Unable to create team notification configuration", fmt.Sprintf("%v", diags))
		return
	}

	attributes := models.NewNotificationConfigurations_attributes()
	attributes.SetDestinationType(destinationType.(*models.NotificationConfigurations_attributes_destinationType))
	attributes.SetEnabled(plan.Enabled.ValueBoolPointer())
	attributes.SetName(plan.Name.ValueStringPointer())
	attributes.SetUrl(plan.URL.ValueStringPointer())

	lastTokenValue := types.StringValue("")
	// Set Token from `token_wo` if set, otherwise use the normal value
	if !config.TokenWO.IsNull() {
		// write-only value should not be persisted.
		attributes.SetToken(config.TokenWO.ValueStringPointer())
	} else {
		attributes.SetToken(plan.Token.ValueStringPointer())
		lastTokenValue = plan.Token
	}

	users, collDiags := setNotificationAttributeCollections(ctx, plan, attributes)
	if collDiags.HasError() {
		resp.Diagnostics.Append(collDiags...)
		return
	}

	subscribableData := models.NewNotificationConfigurations_relationships_subscribable_data()
	subscribableData.SetId(ptr(teamID))
	subscribableData.SetTypeEscaped(ptr(models.TEAMS_NOTIFICATIONCONFIGURATIONS_RELATIONSHIPS_SUBSCRIBABLE_DATA_TYPE))
	subscribable := models.NewNotificationConfigurations_relationships_subscribable()
	subscribable.SetData(subscribableData)

	relationships := models.NewNotificationConfigurations_relationships()
	relationships.SetSubscribable(subscribable)
	relationships.SetUsers(users)

	notificationConfig := models.NewNotificationConfigurations()
	notificationConfig.SetTypeEscaped(ptr(models.NOTIFICATIONCONFIGURATIONS_NOTIFICATIONCONFIGURATIONS_TYPE))
	notificationConfig.SetAttributes(attributes)
	notificationConfig.SetRelationships(relationships)

	envelope := models.NewNotificationConfigurationsEnvelope()
	envelope.SetData(notificationConfig)

	tflog.Debug(ctx, "Creating team notification configuration")
	result, err := r.config.ClientV2.API.Teams().ById(teamID).NotificationConfigurations().Post(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create team notification configuration", err.Error())
		return
	}
	if result == nil || result.GetData() == nil {
		resp.Diagnostics.AddError("Unable to create team notification configuration", "no data returned by the API")
		return
	}
	tnc := result.GetData()

	var createdEmailUserData []models.NotificationConfigurations_relationships_users_dataable
	if relationships := tnc.GetRelationships(); relationships != nil && relationships.GetUsers() != nil {
		createdEmailUserData = relationships.GetUsers().GetData()
	}
	if len(createdEmailUserData) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	modelResult, diags2 := modelFromTFETeamNotificationConfiguration(ctx, tnc, config.TokenWOVersion, lastTokenValue)
	if diags2.HasError() {
		resp.Diagnostics.Append(diags2...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &modelResult)...)
}

func (r *resourceTFETeamNotificationConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFETeamNotificationConfiguration

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading team notification configuration %q", state.ID.ValueString()))
	envelope, err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("`Notification configuration %s no longer exists", state.ID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error reading notification configuration", "Could not read notification configuration, unexpected error: "+err.Error())
		}
		return
	}
	if envelope == nil || envelope.GetData() == nil {
		tflog.Debug(ctx, fmt.Sprintf("`Notification configuration %s no longer exists", state.ID))
		resp.State.RemoveResource(ctx)
		return
	}

	result, diags := modelFromTFETeamNotificationConfiguration(ctx, envelope.GetData(), state.TokenWOVersion, state.Token)
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
	// Read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new attributes struct
	attributes := models.NewNotificationConfigurations_attributes()
	attributes.SetEnabled(plan.Enabled.ValueBoolPointer())
	attributes.SetName(plan.Name.ValueStringPointer())
	attributes.SetUrl(plan.URL.ValueStringPointer())

	// NOTE: while converting this resource to use write-only token-version, it was noted that the last token value should not be preserved since
	// the API will not return it. However, it seems like this was done to preserve token value consistency in the state after apply.
	// This is a todo pending discussions.

	// Preserve the previously known token unless this update explicitly sets a non-write-only token value.
	// The API never returns token values, so we must carry it forward in state to avoid sensitive value drift
	// when updates are triggered by unrelated attributes.
	lastTokenValue := state.Token

	tkn, isWOVal := r.determineTokenForUpdate(plan, state, config)
	// check is needed to prevent accidentally unsetting the token when no changes to token or token_wo were made
	// this is important when an update is triggered by changes in other attributes
	if tkn != nil {
		attributes.SetToken(tkn)

		if !isWOVal {
			lastTokenValue = types.StringValue(*tkn)
		}
	}

	users, collDiags := setNotificationAttributeCollections(ctx, plan, attributes)
	if collDiags.HasError() {
		resp.Diagnostics.Append(collDiags...)
		return
	}

	relationships := models.NewNotificationConfigurations_relationships()
	relationships.SetUsers(users)

	notificationConfig := models.NewNotificationConfigurations()
	notificationConfig.SetTypeEscaped(ptr(models.NOTIFICATIONCONFIGURATIONS_NOTIFICATIONCONFIGURATIONS_TYPE))
	notificationConfig.SetId(state.ID.ValueStringPointer())
	notificationConfig.SetAttributes(attributes)
	notificationConfig.SetRelationships(relationships)

	envelope := models.NewNotificationConfigurationsEnvelope()
	envelope.SetData(notificationConfig)

	tflog.Debug(ctx, "Updating team notification configuration")
	updated, err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Patch(ctx, envelope, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update team notification configuration", err.Error())
		return
	}
	if updated == nil || updated.GetData() == nil {
		resp.Diagnostics.AddError("Unable to update team notification configuration", "no data returned by the API")
		return
	}
	tnc := updated.GetData()

	var updatedEmailUserData []models.NotificationConfigurations_relationships_users_dataable
	if relationships := tnc.GetRelationships(); relationships != nil && relationships.GetUsers() != nil {
		updatedEmailUserData = relationships.GetUsers().GetData()
	}
	if len(updatedEmailUserData) != len(plan.EmailUserIDs.Elements()) {
		resp.Diagnostics.AddError("Email user IDs produced an inconsistent result", "API returned a different number of email user IDs than were provided in the plan.")
		return
	}

	result, diags := modelFromTFETeamNotificationConfiguration(ctx, tnc, config.TokenWOVersion, lastTokenValue)
	if diags.HasError() {
		resp.Diagnostics.Append((diags)...)
		return
	}

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
	err := r.config.ClientV2.API.NotificationConfigurations().ByNotification_configuration_id(state.ID.ValueString()).Delete(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete team notification configuration", err.Error())
		return
	}
}

func (r *resourceTFETeamNotificationConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// determineTokenForUpdate is invoked only after terraform determines that an attribute update is needed.
// note that the update can be triggered by other attributes outside of the token/token_wo attributes.
// this function compares the TokenWOVersion vs Token to ensure that during api update call, token is not mistakenly unset.
// Returns nil if no token update is needed.
func (r *resourceTFETeamNotificationConfiguration) determineTokenForUpdate(plan, state, config modelTFETeamNotificationConfiguration) (updateToken *string, isWOVal bool) {
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
