// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	absser "github.com/microsoft/kiota-abstractions-go/serialization"
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
	Days                                  types.Int64 `tfsdk:"days"`
	DeleteStateVersions                   types.Bool  `tfsdk:"delete_state_versions"`
	DeleteConfigurationVersions           types.Bool  `tfsdk:"delete_configuration_versions"`
	DeleteRunDataAndLogs                  types.Bool  `tfsdk:"delete_run_data_and_logs"`
	StateVersionsDeleteAfterNDays         types.Int64 `tfsdk:"state_versions_delete_after_n_days"`
	ConfigurationVersionsDeleteAfterNDays types.Int64 `tfsdk:"configuration_versions_delete_after_n_days"`
	RunDataAndLogsDeleteAfterNDays        types.Int64 `tfsdk:"run_data_and_logs_delete_after_n_days"`
	StateVersionsKeepLatestCount          types.Int64 `tfsdk:"state_versions_keep_latest_count"`
	ConfigurationVersionsKeepLatestCount  types.Int64 `tfsdk:"configuration_versions_keep_latest_count"`
	RunDataKeepLatestCount                types.Int64 `tfsdk:"run_data_keep_latest_count"`
}

func (m modelTFEDeleteOlderThan) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"days":                                       types.Int64Type,
		"delete_state_versions":                      types.BoolType,
		"delete_configuration_versions":              types.BoolType,
		"delete_run_data_and_logs":                   types.BoolType,
		"state_versions_delete_after_n_days":         types.Int64Type,
		"configuration_versions_delete_after_n_days": types.Int64Type,
		"run_data_and_logs_delete_after_n_days":      types.Int64Type,
		"state_versions_keep_latest_count":           types.Int64Type,
		"configuration_versions_keep_latest_count":   types.Int64Type,
		"run_data_keep_latest_count":                 types.Int64Type,
	}
}

func DontDeleteEmptyObject() basetypes.ObjectValue {
	// ObjectValueMust is safe here: the attribute types are a static empty map
	// and this can never fail.
	return types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})
}

func int32PtrToInt64(v *int32) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

// int64ToInt32Ptr narrows a configured value to the int32 the go-tfe v2 setters expect,
// returning nil for a null or unknown value. An out-of-range value is rejected rather than
// clamped or wrapped. Terraform Enterprise doesn't range-check every retention field, so a
// changed value could be stored and acted on before the apply fails with "Provider produced
// inconsistent result after apply".
//
// The schema validators reject these at plan time, which is where a practitioner should see
// the error. This guard is what additionally makes the conversion provably in range for
// gosec, which cannot see a runtime validator.
func int64ToInt32Ptr(name string, n types.Int64) (*int32, error) {
	if n.IsNull() || n.IsUnknown() {
		return nil, nil
	}
	v := n.ValueInt64()
	if v < math.MinInt32 || v > math.MaxInt32 {
		return nil, fmt.Errorf("%s must be between %d and %d, got %d", name, math.MinInt32, math.MaxInt32, v)
	}
	i := int32(v)
	return &i, nil
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

// explicitDeleteFlag returns the configured delete flag, treating an unset flag as false.
func explicitDeleteFlag(b types.Bool) *bool {
	v := !b.IsNull() && !b.IsUnknown() && b.ValueBool()
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

	days := types.Int64Null()
	deleteStateVersions := types.BoolNull()
	deleteConfigurationVersions := types.BoolNull()
	deleteRunDataAndLogs := types.BoolNull()
	stateVersionsDeleteAfterNDays := types.Int64Null()
	configurationVersionsDeleteAfterNDays := types.Int64Null()
	runDataAndLogsDeleteAfterNDays := types.Int64Null()
	stateVersionsKeepLatestCount := types.Int64Null()
	configurationVersionsKeepLatestCount := types.Int64Null()
	runDataKeepLatestCount := types.Int64Null()

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
			stateVersionsDeleteAfterNDays = int32PtrToInt64(attrs.GetStateVersionsDeleteAfterNDays())
			configurationVersionsDeleteAfterNDays = int32PtrToInt64(attrs.GetConfigurationVersionsDeleteAfterNDays())
			runDataAndLogsDeleteAfterNDays = int32PtrToInt64(attrs.GetRunDataAndLogsDeleteAfterNDays())
			stateVersionsKeepLatestCount = int32PtrToInt64(attrs.GetStateVersionsKeepLatestCount())
			configurationVersionsKeepLatestCount = int32PtrToInt64(attrs.GetConfigurationVersionsKeepLatestCount())
			runDataKeepLatestCount = int32PtrToInt64(attrs.GetRunDataKeepLatestCount())
		} else {
			// Server has only days — use it
			days = int32PtrToInt64(attrs.GetDeleteOlderThanNDays())
		}
	} else if userSetDaysOnly {
		// User sent only "days" in their configuration or prior state
		// Even if we are on TFE >= 2.1.0, ignote the server-backfilled granular fields, preserve original "days", to avoid drift
		days = int32PtrToInt64(attrs.GetDeleteOlderThanNDays())
	} else {
		// User sent granular fields — use them, ignore coalesced "days" from the server.
		// For delete flags: prefer the configured value (which may be explicit false) over
		// the server value. The server may backfill false for flags the user omitted, which
		// would cause drift since the plan has null for those flags.
		deleteStateVersions = mergeDeleteFlag(configuredDot.DeleteStateVersions, attrs.GetDeleteStateVersions())
		deleteConfigurationVersions = mergeDeleteFlag(configuredDot.DeleteConfigurationVersions, attrs.GetDeleteConfigurationVersions())
		deleteRunDataAndLogs = mergeDeleteFlag(configuredDot.DeleteRunDataAndLogs, attrs.GetDeleteRunDataAndLogs())
		stateVersionsDeleteAfterNDays = int32PtrToInt64(attrs.GetStateVersionsDeleteAfterNDays())
		configurationVersionsDeleteAfterNDays = int32PtrToInt64(attrs.GetConfigurationVersionsDeleteAfterNDays())
		runDataAndLogsDeleteAfterNDays = int32PtrToInt64(attrs.GetRunDataAndLogsDeleteAfterNDays())
		stateVersionsKeepLatestCount = int32PtrToInt64(attrs.GetStateVersionsKeepLatestCount())
		configurationVersionsKeepLatestCount = int32PtrToInt64(attrs.GetConfigurationVersionsKeepLatestCount())
		runDataKeepLatestCount = int32PtrToInt64(attrs.GetRunDataKeepLatestCount())
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

	if hasGranularFields(model) {
		if err := setGranularAttributes(attrs, model); err != nil {
			return nil, err
		}
	} else {
		days, err := int64ToInt32Ptr("days", model.Days)
		if err != nil {
			return nil, err
		}
		attrs.SetDeleteOlderThanNDays(days)
	}

	policy := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	policy.SetTypeEscaped(&pType)
	policy.SetAttributes(attrs)

	env := models.NewDataRetentionPolicyEnvelope()
	env.SetData(policy)

	return env, nil
}

// setGranularAttributes sets every per-artifact-type field on attrs. TFE applies these
// fields as a partial update, keeping any value omitted from the request, so unset delete
// flags are sent as false and unset windows and counts are sent as explicit nulls. This
// makes the request describe the whole policy, so fields removed from the configuration
// are cleared on the server.
func setGranularAttributes(attrs models.DataRetentionPolicy_attributesable, model modelTFEDeleteOlderThan) error {
	attrs.SetDeleteStateVersions(explicitDeleteFlag(model.DeleteStateVersions))
	attrs.SetDeleteConfigurationVersions(explicitDeleteFlag(model.DeleteConfigurationVersions))
	attrs.SetDeleteRunDataAndLogs(explicitDeleteFlag(model.DeleteRunDataAndLogs))

	// The generated model omits nil fields when serializing, so explicit nulls are sent
	// through the additional data map instead.
	nulls := map[string]any{}
	if err := setInt64OrNull(model.StateVersionsDeleteAfterNDays, attrs.SetStateVersionsDeleteAfterNDays, "state-versions-delete-after-n-days", nulls); err != nil {
		return err
	}
	if err := setInt64OrNull(model.ConfigurationVersionsDeleteAfterNDays, attrs.SetConfigurationVersionsDeleteAfterNDays, "configuration-versions-delete-after-n-days", nulls); err != nil {
		return err
	}
	if err := setInt64OrNull(model.RunDataAndLogsDeleteAfterNDays, attrs.SetRunDataAndLogsDeleteAfterNDays, "run-data-and-logs-delete-after-n-days", nulls); err != nil {
		return err
	}
	if err := setInt64OrNull(model.StateVersionsKeepLatestCount, attrs.SetStateVersionsKeepLatestCount, "state-versions-keep-latest-count", nulls); err != nil {
		return err
	}
	if err := setInt64OrNull(model.ConfigurationVersionsKeepLatestCount, attrs.SetConfigurationVersionsKeepLatestCount, "configuration-versions-keep-latest-count", nulls); err != nil {
		return err
	}
	if err := setInt64OrNull(model.RunDataKeepLatestCount, attrs.SetRunDataKeepLatestCount, "run-data-keep-latest-count", nulls); err != nil {
		return err
	}
	attrs.SetAdditionalData(nulls)

	return nil
}

// setInt64OrNull sets value using set if it is known, or records key in nulls so that it
// is sent as an explicit null.
func setInt64OrNull(value types.Int64, set func(*int32), key string, nulls map[string]any) error {
	v, err := int64ToInt32Ptr(key, value)
	if err != nil {
		return err
	}
	if v == nil {
		nulls[key] = absser.NewUntypedNull()
		return nil
	}
	set(v)
	return nil
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
