// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDeleteOlderThanValidator_StateVersionsRequiresWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_state_versions = true, but no window or keep-latest
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "you must set either state_versions_delete_after_n_days or state_versions_keep_latest_count")
}

func TestDeleteOlderThanValidator_ZeroValueRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberValue(big.NewFloat(0)),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "must be at least 1")
}

func TestDeleteOlderThanValidator_NegativeValueRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberValue(big.NewFloat(-5)),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "must be at least 1")
}

func TestDeleteOlderThanValidator_DaysZeroRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	attrs := map[string]attr.Value{
		"days":                                       types.NumberValue(big.NewFloat(0)),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "must be at least 1")
}

func TestDeleteOlderThanValidator_ConfigurationVersionsRequiresWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_configuration_versions = true, but no window or keep-latest
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolValue(true),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "configuration_versions_delete_after_n_days or configuration_versions_keep_latest_count")
}

func TestDeleteOlderThanValidator_RunDataRequiresWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_run_data_and_logs = true, but no window or keep-latest
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolValue(true),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "run_data_and_logs_delete_after_n_days or run_data_keep_latest_count")
}

func TestDeleteOlderThanValidator_ValidWithWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_state_versions = true WITH state_versions_delete_after_n_days
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberValue(big.NewFloat(30)),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestDeleteOlderThanValidator_ValidWithKeepLatest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_state_versions = true WITH state_versions_keep_latest_count
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberValue(big.NewFloat(5)),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestDeleteOlderThanValidator_ValidWhenDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// delete_state_versions = false (or null), no window required, but this config does nothing
	// so the validator should reject it
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(false),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "At least one artifact type must be enabled")
}

func TestDeleteOlderThanValidator_WindowRequiresDeleteFlag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// state_versions_delete_after_n_days set, but delete_state_versions = false
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolValue(false),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberValue(big.NewFloat(30)),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "requires delete_state_versions to be true")
}

func TestDeleteOlderThanValidator_KeepLatestRequiresDeleteFlag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// configuration_versions_keep_latest_count set, but delete flag not enabled
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberValue(big.NewFloat(5)),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "requires delete_configuration_versions to be true")
}

func TestDeleteOlderThanValidator_RunDataWindowRequiresDeleteFlag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// run_data_and_logs_delete_after_n_days set with delete flag explicitly false
	attrs := map[string]attr.Value{
		"days":                                       types.NumberNull(),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolValue(false),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberValue(big.NewFloat(60)),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "requires delete_run_data_and_logs to be true")
}

func TestDeleteOlderThanValidator_DaysAndGranularFieldsMutuallyExclusive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// Both 'days' and granular fields set
	attrs := map[string]attr.Value{
		"days":                                       types.NumberValue(big.NewFloat(30)),
		"delete_state_versions":                      types.BoolValue(true),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberValue(big.NewFloat(45)),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberNull(),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "cannot be used together with granular fields")
}

func TestDeleteOlderThanValidator_DaysWithKeepLatestConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ValidateDeleteOlderThan()

	// 'days' with only keep-latest count is invalid
	attrs := map[string]attr.Value{
		"days":                                       types.NumberValue(big.NewFloat(30)),
		"delete_state_versions":                      types.BoolNull(),
		"delete_configuration_versions":              types.BoolNull(),
		"delete_run_data_and_logs":                   types.BoolNull(),
		"state_versions_delete_after_n_days":         types.NumberNull(),
		"configuration_versions_delete_after_n_days": types.NumberNull(),
		"run_data_and_logs_delete_after_n_days":      types.NumberNull(),
		"state_versions_keep_latest_count":           types.NumberValue(big.NewFloat(5)),
		"configuration_versions_keep_latest_count":   types.NumberNull(),
		"run_data_keep_latest_count":                 types.NumberNull(),
	}

	req := validator.ObjectRequest{
		Path:        path.Root("delete_older_than"),
		ConfigValue: types.ObjectValueMust(modelTFEDeleteOlderThan{}.AttributeTypes(), attrs),
	}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "cannot be used together with granular fields")
}
