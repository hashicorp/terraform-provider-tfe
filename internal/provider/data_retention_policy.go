// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type modelTFEDataRetentionPolicy struct {
	ID              types.String `tfsdk:"id"`
	Organization    types.String `tfsdk:"organization"`
	WorkspaceID     types.String `tfsdk:"workspace_id"`
	DeleteOlderThan types.Object `tfsdk:"delete_older_than"`
	DontDelete      types.Object `tfsdk:"dont_delete"`
}

type modelTFEDeleteOlderThan struct {
	// Days is a legacy global retention window applied to all artifact types. It exists for backwards
	// compatibility with TFE versions prior to v2.1.0, which do not support per-artifact-type
	// retention fields.
	// On TFE v2.1.0+, this field is deprecated and will not be sent to the
	// server if any of the per-artifact-type fields below are set.
	Days                                  types.Number `tfsdk:"days"`
	DeleteStateVersions                   types.Bool   `tfsdk:"delete_state_versions"`
	DeleteConfigurationVersions           types.Bool   `tfsdk:"delete_configuration_versions"`
	DeleteRunDataAndLogs                  types.Bool   `tfsdk:"delete_run_data_and_logs"`
	StateVersionsDeleteAfterNDays         types.Number `tfsdk:"state_versions_delete_after_n_days"`
	ConfigurationVersionsDeleteAfterNDays types.Number `tfsdk:"configuration_versions_delete_after_n_days"`
	RunDataAndLogsDeleteAfterNDays        types.Number `tfsdk:"run_data_and_logs_delete_after_n_days"`
	StateVersionsKeepLatestCount          types.Number `tfsdk:"state_versions_keep_latest_count"`
	ConfigurationVersionsKeepLatestCount  types.Number `tfsdk:"configuration_versions_keep_latest_count"`
	RunDataKeepLatestCount                types.Number `tfsdk:"run_data_keep_latest_count"`
}

func (m modelTFEDeleteOlderThan) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"days":                                       types.NumberType,
		"delete_state_versions":                      types.BoolType,
		"delete_configuration_versions":              types.BoolType,
		"delete_run_data_and_logs":                   types.BoolType,
		"state_versions_delete_after_n_days":         types.NumberType,
		"configuration_versions_delete_after_n_days": types.NumberType,
		"run_data_and_logs_delete_after_n_days":      types.NumberType,
		"state_versions_keep_latest_count":           types.NumberType,
		"configuration_versions_keep_latest_count":   types.NumberType,
		"run_data_keep_latest_count":                 types.NumberType,
	}
}

func DontDeleteEmptyObject() basetypes.ObjectValue {
	// ObjectValueMust is safe here: the attribute types are a static empty map
	// and this can never fail.
	return types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})
}

func int32PtrToNumber(v *int32) types.Number {
	if v == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(float64(*v)))
}

func numberToInt32Ptr(n types.Number) *int32 {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	v, _ := n.ValueBigFloat().Int64()
	i := int32(v)
	return &i
}

func boolPtrToTypes(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

// deleteFlagToTypes maps a server-returned delete flag to a types.Bool.
// The server may backfill false for artifact types not explicitly configured.
// We treat false and nil identically (both map to null) to avoid drift when
// the user omits a delete flag from their config.
func deleteFlagToTypes(v *bool) types.Bool {
	if v == nil || !*v {
		return types.BoolNull()
	}
	return types.BoolValue(true)
}

// mergeDeleteFlag returns the configured value if it was explicitly set by the user,
// otherwise falls back to the server value (normalizing server false → null).
// This allows explicit false in config to be preserved while preventing drift from
// server-backfilled false on unset flags.
func mergeDeleteFlag(configured types.Bool, serverVal *bool) types.Bool {
	if !configured.IsNull() && !configured.IsUnknown() {
		return configured
	}
	return deleteFlagToTypes(serverVal)
}

func typesBoolToPtr(b types.Bool) *bool {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	v := b.ValueBool()
	return &v
}

// deleteOlderThanFromAPIResponse converts an API policy response into a new local state model.
// organization and workspaceID are passed through directly into the returned state since they
// are not present in the API response.
// configuredDeleteOlderThan is the user's configured delete_older_than block — either from the
// plan (on Create) or from the prior state (on Read). It is used to detect whether the user
// configured days-only, which is needed to ignore granular fields that TFE v2.1.0+ backfills
// on the server when only days is sent, preventing spurious plan diffs.
func deleteOlderThanFromAPIResponse(ctx context.Context, organization types.String, workspaceID types.String, configuredDeleteOlderThan types.Object, policy models.DataRetentionPolicyable) (modelTFEDataRetentionPolicy, diag.Diagnostics) {
	attrs := policy.GetAttributes()
	if attrs == nil {
		var d diag.Diagnostics
		d.AddError("Invalid API response", "data retention policy response is missing attributes")
		return modelTFEDataRetentionPolicy{}, d
	}

	var configuredDot modelTFEDeleteOlderThan
	userSetDaysOnly := false
	hasConfigured := !configuredDeleteOlderThan.IsNull() && !configuredDeleteOlderThan.IsUnknown()

	// determine if the user has configured a legacy `days` attribute, or if we should acknowledge the
	// policy's granular fields
	if hasConfigured {
		diags := configuredDeleteOlderThan.As(ctx, &configuredDot, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return modelTFEDataRetentionPolicy{}, diags
		}
		// User set "days" only if days was non-null and all granular fields were null in the plan.
		userSetDaysOnly = !configuredDot.Days.IsNull() &&
			configuredDot.DeleteStateVersions.IsNull() &&
			configuredDot.DeleteConfigurationVersions.IsNull() &&
			configuredDot.DeleteRunDataAndLogs.IsNull() &&
			configuredDot.StateVersionsDeleteAfterNDays.IsNull() &&
			configuredDot.ConfigurationVersionsDeleteAfterNDays.IsNull() &&
			configuredDot.RunDataAndLogsDeleteAfterNDays.IsNull() &&
			configuredDot.StateVersionsKeepLatestCount.IsNull() &&
			configuredDot.ConfigurationVersionsKeepLatestCount.IsNull() &&
			configuredDot.RunDataKeepLatestCount.IsNull()
	}

	// When reading back the API response:
	// 1. If no prior state exists (first read), read everything from the API as-is.
	// 2. If user originally set "days" only, ignore any granular fields the server backfilled
	// 3. If user set granular fields, ignore the coalesced delete-older-than-n-days value.

	// Check if API response has artifact specific retention settings

	days := types.NumberNull()
	deleteStateVersions := types.BoolNull()
	deleteConfigurationVersions := types.BoolNull()
	deleteRunDataAndLogs := types.BoolNull()
	stateVersionsDeleteAfterNDays := types.NumberNull()
	configurationVersionsDeleteAfterNDays := types.NumberNull()
	runDataAndLogsDeleteAfterNDays := types.NumberNull()
	stateVersionsKeepLatestCount := types.NumberNull()
	configurationVersionsKeepLatestCount := types.NumberNull()
	runDataKeepLatestCount := types.NumberNull()

	if !hasConfigured {
		// First read: use whatever the API returns.
		apiHasGranular := attrs.GetDeleteStateVersions() != nil ||
			attrs.GetDeleteConfigurationVersions() != nil ||
			attrs.GetDeleteRunDataAndLogs() != nil

		if apiHasGranular {
			// Server has granular fields — ignore coalesced days
			deleteStateVersions = deleteFlagToTypes(attrs.GetDeleteStateVersions())
			deleteConfigurationVersions = deleteFlagToTypes(attrs.GetDeleteConfigurationVersions())
			deleteRunDataAndLogs = deleteFlagToTypes(attrs.GetDeleteRunDataAndLogs())
			stateVersionsDeleteAfterNDays = int32PtrToNumber(attrs.GetStateVersionsDeleteAfterNDays())
			configurationVersionsDeleteAfterNDays = int32PtrToNumber(attrs.GetConfigurationVersionsDeleteAfterNDays())
			runDataAndLogsDeleteAfterNDays = int32PtrToNumber(attrs.GetRunDataAndLogsDeleteAfterNDays())
			stateVersionsKeepLatestCount = int32PtrToNumber(attrs.GetStateVersionsKeepLatestCount())
			configurationVersionsKeepLatestCount = int32PtrToNumber(attrs.GetConfigurationVersionsKeepLatestCount())
			runDataKeepLatestCount = int32PtrToNumber(attrs.GetRunDataKeepLatestCount())
		} else {
			// Server has only days — use it
			days = int32PtrToNumber(attrs.GetDeleteOlderThanNDays())
		}
	} else if userSetDaysOnly {
		// User sent only "days" in their configuration or prior state
		// Even if we are on TFE >= 2.1.0, ignote the server-backfilled granular fields, preserve original "days", to avoid drift
		days = int32PtrToNumber(attrs.GetDeleteOlderThanNDays())
	} else {
		// User sent granular fields — use them, ignore coalesced "days" from the server.
		// For delete flags: prefer the configured value (which may be explicit false) over
		// the server value. The server may backfill false for flags the user omitted, which
		// would cause drift since the plan has null for those flags.
		deleteStateVersions = mergeDeleteFlag(configuredDot.DeleteStateVersions, attrs.GetDeleteStateVersions())
		deleteConfigurationVersions = mergeDeleteFlag(configuredDot.DeleteConfigurationVersions, attrs.GetDeleteConfigurationVersions())
		deleteRunDataAndLogs = mergeDeleteFlag(configuredDot.DeleteRunDataAndLogs, attrs.GetDeleteRunDataAndLogs())
		stateVersionsDeleteAfterNDays = int32PtrToNumber(attrs.GetStateVersionsDeleteAfterNDays())
		configurationVersionsDeleteAfterNDays = int32PtrToNumber(attrs.GetConfigurationVersionsDeleteAfterNDays())
		runDataAndLogsDeleteAfterNDays = int32PtrToNumber(attrs.GetRunDataAndLogsDeleteAfterNDays())
		stateVersionsKeepLatestCount = int32PtrToNumber(attrs.GetStateVersionsKeepLatestCount())
		configurationVersionsKeepLatestCount = int32PtrToNumber(attrs.GetConfigurationVersionsKeepLatestCount())
		runDataKeepLatestCount = int32PtrToNumber(attrs.GetRunDataKeepLatestCount())
	}

	deleteOlderThan := modelTFEDeleteOlderThan{
		Days:                                  days,
		DeleteStateVersions:                   deleteStateVersions,
		DeleteConfigurationVersions:           deleteConfigurationVersions,
		DeleteRunDataAndLogs:                  deleteRunDataAndLogs,
		StateVersionsDeleteAfterNDays:         stateVersionsDeleteAfterNDays,
		ConfigurationVersionsDeleteAfterNDays: configurationVersionsDeleteAfterNDays,
		RunDataAndLogsDeleteAfterNDays:        runDataAndLogsDeleteAfterNDays,
		StateVersionsKeepLatestCount:          stateVersionsKeepLatestCount,
		ConfigurationVersionsKeepLatestCount:  configurationVersionsKeepLatestCount,
		RunDataKeepLatestCount:                runDataKeepLatestCount,
	}
	deleteOlderThanObject, diags := types.ObjectValueFrom(ctx, deleteOlderThan.AttributeTypes(), deleteOlderThan)

	org := types.StringNull()
	if workspaceID.IsNull() {
		org = organization
	}

	id := ""
	if policy.GetId() != nil {
		id = *policy.GetId()
	}

	newState := modelTFEDataRetentionPolicy{
		ID:              types.StringValue(id),
		Organization:    org,
		WorkspaceID:     workspaceID,
		DeleteOlderThan: deleteOlderThanObject,
		DontDelete:      types.ObjectNull(map[string]attr.Type{}),
	}
	return newState, diags
}

func dontDeleteFromAPIResponse(organization types.String, workspaceID types.String, policy models.DataRetentionPolicyable) modelTFEDataRetentionPolicy {
	org := types.StringNull()
	if workspaceID.IsNull() {
		org = organization
	}

	id := ""
	if policy.GetId() != nil {
		id = *policy.GetId()
	}

	return modelTFEDataRetentionPolicy{
		ID:              types.StringValue(id),
		Organization:    org,
		WorkspaceID:     workspaceID,
		DeleteOlderThan: types.ObjectNull(modelTFEDeleteOlderThan{}.AttributeTypes()),
		DontDelete:      DontDeleteEmptyObject(),
	}
}

func dataRetentionPolicyFromAPIResponse(ctx context.Context, priorState modelTFEDataRetentionPolicy, policy models.DataRetentionPolicyable) (modelTFEDataRetentionPolicy, diag.Diagnostics) {
	if policy == nil {
		var d diag.Diagnostics
		d.AddError("unexpected nil policy", "received nil DataRetentionPolicyable from API")
		return modelTFEDataRetentionPolicy{}, d
	}

	pType := policy.GetTypeEscaped()
	if pType != nil && *pType == models.DATARETENTIONPOLICYDONTDELETES_DATARETENTIONPOLICY_TYPE {
		var emptyDiag diag.Diagnostics
		return dontDeleteFromAPIResponse(priorState.Organization, priorState.WorkspaceID, policy), emptyDiag
	}

	return deleteOlderThanFromAPIResponse(ctx, priorState.Organization, priorState.WorkspaceID, priorState.DeleteOlderThan, policy)
}

func hasGranularFields(model modelTFEDeleteOlderThan) bool {
	return !model.DeleteStateVersions.IsNull() ||
		!model.DeleteConfigurationVersions.IsNull() ||
		!model.DeleteRunDataAndLogs.IsNull() ||
		!model.StateVersionsDeleteAfterNDays.IsNull() ||
		!model.ConfigurationVersionsDeleteAfterNDays.IsNull() ||
		!model.RunDataAndLogsDeleteAfterNDays.IsNull() ||
		!model.StateVersionsKeepLatestCount.IsNull() ||
		!model.ConfigurationVersionsKeepLatestCount.IsNull() ||
		!model.RunDataKeepLatestCount.IsNull()
}

func newDeleteOlderEnvelope(model modelTFEDeleteOlderThan) (models.DataRetentionPolicyEnvelopeable, error) {
	attrs := models.NewDataRetentionPolicy_attributes()

	if !hasGranularFields(model) && !model.Days.IsNull() {
		days, _ := model.Days.ValueBigFloat().Int64()
		days32 := int32(days)
		attrs.SetDeleteOlderThanNDays(&days32)
	}
	attrs.SetDeleteStateVersions(typesBoolToPtr(model.DeleteStateVersions))
	attrs.SetDeleteConfigurationVersions(typesBoolToPtr(model.DeleteConfigurationVersions))
	attrs.SetDeleteRunDataAndLogs(typesBoolToPtr(model.DeleteRunDataAndLogs))
	attrs.SetStateVersionsDeleteAfterNDays(numberToInt32Ptr(model.StateVersionsDeleteAfterNDays))
	attrs.SetConfigurationVersionsDeleteAfterNDays(numberToInt32Ptr(model.ConfigurationVersionsDeleteAfterNDays))
	attrs.SetRunDataAndLogsDeleteAfterNDays(numberToInt32Ptr(model.RunDataAndLogsDeleteAfterNDays))
	attrs.SetStateVersionsKeepLatestCount(numberToInt32Ptr(model.StateVersionsKeepLatestCount))
	attrs.SetConfigurationVersionsKeepLatestCount(numberToInt32Ptr(model.ConfigurationVersionsKeepLatestCount))
	attrs.SetRunDataKeepLatestCount(numberToInt32Ptr(model.RunDataKeepLatestCount))

	policy := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	policy.SetTypeEscaped(&pType)
	policy.SetAttributes(attrs)

	env := models.NewDataRetentionPolicyEnvelope()
	env.SetData(policy)

	return env, nil
}

func newDontDeleteEnvelope() models.DataRetentionPolicyEnvelopeable {
	policy := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDONTDELETES_DATARETENTIONPOLICY_TYPE
	policy.SetTypeEscaped(&pType)
	policy.SetAttributes(models.NewDataRetentionPolicy_attributes())

	env := models.NewDataRetentionPolicyEnvelope()
	env.SetData(policy)
	return env
}

func getPolicyIDFromV2(policy models.DataRetentionPolicyable) (string, error) {
	if policy == nil || policy.GetId() == nil {
		return "", fmt.Errorf("policy has no ID")
	}
	return *policy.GetId(), nil
}
