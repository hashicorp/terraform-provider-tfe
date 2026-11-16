// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = deleteOlderThanValidator{}

type deleteOlderThanValidator struct{}

// ValidateDeleteOlderThan returns a validator that ensures:
//  1. The 'days' field and all granular fields are mutually exclusive
//  2. Artifact-type deletion flags have corresponding retention windows or keep-latest counts configured
//  3. Retention windows or keep-latest counts require their corresponding deletion flags to be enabled
//  4. The block sets either 'days' or at least one granular field
func ValidateDeleteOlderThan() validator.Object {
	return deleteOlderThanValidator{}
}

func (v deleteOlderThanValidator) Description(ctx context.Context) string {
	return "Validates that 'days' and all granular fields are mutually exclusive, and that each enabled artifact type has either a retention window or keep-latest count configured"
}

func (v deleteOlderThanValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v deleteOlderThanValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Extract the delete_older_than attributes
	attrs := req.ConfigValue.Attributes()

	days := isInt64AttrSet(attrs, "days")

	deleteStateVersions := getBoolAttr(attrs, "delete_state_versions")
	deleteConfigurationVersions := getBoolAttr(attrs, "delete_configuration_versions")
	deleteRunDataAndLogs := getBoolAttr(attrs, "delete_run_data_and_logs")

	stateVersionsDeleteAfterNDays := isInt64AttrSet(attrs, "state_versions_delete_after_n_days")
	configurationVersionsDeleteAfterNDays := isInt64AttrSet(attrs, "configuration_versions_delete_after_n_days")
	runDataAndLogsDeleteAfterNDays := isInt64AttrSet(attrs, "run_data_and_logs_delete_after_n_days")

	stateVersionsKeepLatestCount := isInt64AttrSet(attrs, "state_versions_keep_latest_count")
	configurationVersionsKeepLatestCount := isInt64AttrSet(attrs, "configuration_versions_keep_latest_count")
	runDataKeepLatestCount := isInt64AttrSet(attrs, "run_data_keep_latest_count")

	// Validate that 'days' and all granular fields are mutually exclusive.
	hasGranularFields := deleteStateVersions != nil ||
		deleteConfigurationVersions != nil ||
		deleteRunDataAndLogs != nil ||
		stateVersionsDeleteAfterNDays ||
		configurationVersionsDeleteAfterNDays ||
		runDataAndLogsDeleteAfterNDays ||
		stateVersionsKeepLatestCount ||
		configurationVersionsKeepLatestCount ||
		runDataKeepLatestCount

	if days && hasGranularFields {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("days"),
			"Conflicting configuration",
			"The 'days' field cannot be used together with granular fields. Use either 'days' alone (deprecated) or the granular per-artifact fields.",
		)
		return
	}

	// The following checks only apply when using granular fields.
	if !hasGranularFields {
		// An unknown value may still turn out to be set, so only treat the block as empty
		// when every value is known.
		if !days && !hasUnknownAttr(attrs) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing required field",
				"The delete_older_than block must set either 'days' or at least one per-artifact-type field.",
			)
		}
		return
	}

	// At least one artifact type must be enabled for deletion when granular fields are present
	// and no windows or keep-latest counts are configured. If windows or keep-latest counts are
	// set, the per-attribute checks below will produce more specific errors.
	noWindowsOrKeepLatest := !stateVersionsDeleteAfterNDays &&
		!configurationVersionsDeleteAfterNDays &&
		!runDataAndLogsDeleteAfterNDays &&
		!stateVersionsKeepLatestCount &&
		!configurationVersionsKeepLatestCount &&
		!runDataKeepLatestCount
	if noWindowsOrKeepLatest &&
		(deleteStateVersions == nil || !*deleteStateVersions) &&
		(deleteConfigurationVersions == nil || !*deleteConfigurationVersions) &&
		(deleteRunDataAndLogs == nil || !*deleteRunDataAndLogs) {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"At least one artifact type must be enabled for deletion. Set delete_state_versions, delete_configuration_versions, or delete_run_data_and_logs to true, or remove the delete_older_than block.",
		)
		return
	}

	// Validate state versions
	if deleteStateVersions != nil && *deleteStateVersions {
		if !stateVersionsDeleteAfterNDays && !stateVersionsKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("delete_state_versions"),
				"Missing required field",
				"When delete_state_versions is true, you must set either state_versions_delete_after_n_days or state_versions_keep_latest_count.",
			)
		}
	} else {
		// delete_state_versions is false or unset, so ensure we dont have a retention window or keep latest count
		if stateVersionsDeleteAfterNDays {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("state_versions_delete_after_n_days"),
				"Invalid configuration",
				"Setting state_versions_delete_after_n_days requires delete_state_versions to be true.",
			)
		}
		if stateVersionsKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("state_versions_keep_latest_count"),
				"Invalid configuration",
				"Setting state_versions_keep_latest_count requires delete_state_versions to be true.",
			)
		}
	}

	// Validate configuration versions
	if deleteConfigurationVersions != nil && *deleteConfigurationVersions {
		if !configurationVersionsDeleteAfterNDays && !configurationVersionsKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("delete_configuration_versions"),
				"Missing required field",
				"When delete_configuration_versions is true, you must set either configuration_versions_delete_after_n_days or configuration_versions_keep_latest_count.",
			)
		}
	} else {
		// delete_configuration_versions is false or unset, so ensure we dont have a retention window or keep latest count
		if configurationVersionsDeleteAfterNDays {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("configuration_versions_delete_after_n_days"),
				"Invalid configuration",
				"Setting configuration_versions_delete_after_n_days requires delete_configuration_versions to be true.",
			)
		}
		if configurationVersionsKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("configuration_versions_keep_latest_count"),
				"Invalid configuration",
				"Setting configuration_versions_keep_latest_count requires delete_configuration_versions to be true.",
			)
		}
	}

	// Validate run data and logs
	if deleteRunDataAndLogs != nil && *deleteRunDataAndLogs {
		if !runDataAndLogsDeleteAfterNDays && !runDataKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("delete_run_data_and_logs"),
				"Missing required field",
				"When delete_run_data_and_logs is true, you must set either run_data_and_logs_delete_after_n_days or run_data_keep_latest_count.",
			)
		}
	} else {
		if runDataAndLogsDeleteAfterNDays {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("run_data_and_logs_delete_after_n_days"),
				"Invalid configuration",
				"Setting run_data_and_logs_delete_after_n_days requires delete_run_data_and_logs to be true.",
			)
		}
		if runDataKeepLatestCount {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("run_data_keep_latest_count"),
				"Invalid configuration",
				"Setting run_data_keep_latest_count requires delete_run_data_and_logs to be true.",
			)
		}
	}
}

func getBoolAttr(attrs map[string]attr.Value, name string) *bool {
	val, ok := attrs[name]
	if !ok {
		return nil
	}
	boolVal, ok := val.(types.Bool)
	if !ok || boolVal.IsNull() || boolVal.IsUnknown() {
		return nil
	}
	v := boolVal.ValueBool()
	return &v
}

func hasUnknownAttr(attrs map[string]attr.Value) bool {
	for _, a := range attrs {
		if a.IsUnknown() {
			return true
		}
	}
	return false
}

func isInt64AttrSet(attrs map[string]attr.Value, name string) bool {
	val, ok := attrs[name]
	if !ok {
		return false
	}
	intVal, ok := val.(types.Int64)
	if !ok || intVal.IsNull() || intVal.IsUnknown() {
		return false
	}
	return true
}
