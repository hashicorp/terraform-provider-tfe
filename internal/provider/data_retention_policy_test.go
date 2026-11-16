// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	jsonserialization "github.com/microsoft/kiota-serialization-json-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func int32Ptr(v int32) *int32 { return &v }
func boolPtr(v bool) *bool    { return &v }

func newDeleteOlderV2Policy(id string, days int32, stateKeep, cvKeep, runKeep *int32) models.DataRetentionPolicyable {
	p := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	p.SetTypeEscaped(&pType)
	p.SetId(&id)

	attrs := models.NewDataRetentionPolicy_attributes()
	attrs.SetDeleteOlderThanNDays(&days)
	attrs.SetStateVersionsKeepLatestCount(stateKeep)
	attrs.SetConfigurationVersionsKeepLatestCount(cvKeep)
	attrs.SetRunDataKeepLatestCount(runKeep)
	p.SetAttributes(attrs)
	return p
}

func newFullDeleteOlderV2Policy(
	id string,
	days int32,
	delSV, delCV, delRun *bool,
	svDays, cvDays, runDays *int32,
	svKeep, cvKeep, runKeep *int32,
) models.DataRetentionPolicyable {
	p := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	p.SetTypeEscaped(&pType)
	p.SetId(&id)

	attrs := models.NewDataRetentionPolicy_attributes()
	attrs.SetDeleteOlderThanNDays(&days)
	attrs.SetDeleteStateVersions(delSV)
	attrs.SetDeleteConfigurationVersions(delCV)
	attrs.SetDeleteRunDataAndLogs(delRun)
	attrs.SetStateVersionsDeleteAfterNDays(svDays)
	attrs.SetConfigurationVersionsDeleteAfterNDays(cvDays)
	attrs.SetRunDataAndLogsDeleteAfterNDays(runDays)
	attrs.SetStateVersionsKeepLatestCount(svKeep)
	attrs.SetConfigurationVersionsKeepLatestCount(cvKeep)
	attrs.SetRunDataKeepLatestCount(runKeep)
	p.SetAttributes(attrs)
	return p
}

func newDontDeleteV2Policy(id string) models.DataRetentionPolicyable {
	p := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDONTDELETES_DATARETENTIONPOLICY_TYPE
	p.SetTypeEscaped(&pType)
	p.SetId(&id)
	p.SetAttributes(models.NewDataRetentionPolicy_attributes())
	return p
}

func TestModelTFEDeleteOlderThan_AttributeTypes(t *testing.T) {
	t.Parallel()

	attrTypes := modelTFEDeleteOlderThan{}.AttributeTypes()

	assert.Equal(t, types.Int64Type, attrTypes["days"])
	assert.Equal(t, types.BoolType, attrTypes["delete_state_versions"])
	assert.Equal(t, types.BoolType, attrTypes["delete_configuration_versions"])
	assert.Equal(t, types.BoolType, attrTypes["delete_run_data_and_logs"])
	assert.Equal(t, types.Int64Type, attrTypes["state_versions_delete_after_n_days"])
	assert.Equal(t, types.Int64Type, attrTypes["configuration_versions_delete_after_n_days"])
	assert.Equal(t, types.Int64Type, attrTypes["run_data_and_logs_delete_after_n_days"])
	assert.Equal(t, types.Int64Type, attrTypes["state_versions_keep_latest_count"])
	assert.Equal(t, types.Int64Type, attrTypes["configuration_versions_keep_latest_count"])
	assert.Equal(t, types.Int64Type, attrTypes["run_data_keep_latest_count"])
}

func TestModelTFEDeleteOlderThan_AttributeTypes_NoExtraKeys(t *testing.T) {
	t.Parallel()

	attrTypes := modelTFEDeleteOlderThan{}.AttributeTypes()
	expected := map[string]attr.Type{
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
	assert.Equal(t, expected, attrTypes)
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_NilKeepLatestFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := newDeleteOlderV2Policy("drp-123", 42, nil, nil, nil)
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	assert.True(t, dot.StateVersionsKeepLatestCount.IsNull())
	assert.True(t, dot.ConfigurationVersionsKeepLatestCount.IsNull())
	assert.True(t, dot.RunDataKeepLatestCount.IsNull())
	assert.True(t, dot.DeleteStateVersions.IsNull())
	assert.True(t, dot.DeleteConfigurationVersions.IsNull())
	assert.True(t, dot.DeleteRunDataAndLogs.IsNull())
	assert.True(t, dot.StateVersionsDeleteAfterNDays.IsNull())
	assert.True(t, dot.ConfigurationVersionsDeleteAfterNDays.IsNull())
	assert.True(t, dot.RunDataAndLogsDeleteAfterNDays.IsNull())
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_PopulatedKeepLatestFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	trueVal := true
	policy := newFullDeleteOlderV2Policy("drp-456", 30, &trueVal, &trueVal, &trueVal, nil, nil, nil, int32Ptr(5), int32Ptr(10), int32Ptr(3))
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	stateVal := dot.StateVersionsKeepLatestCount.ValueInt64()
	cvVal := dot.ConfigurationVersionsKeepLatestCount.ValueInt64()
	runVal := dot.RunDataKeepLatestCount.ValueInt64()

	assert.Equal(t, int64(5), stateVal)
	assert.Equal(t, int64(10), cvVal)
	assert.Equal(t, int64(3), runVal)
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_PerArtifactFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := newFullDeleteOlderV2Policy(
		"drp-full", 0,
		boolPtr(true), boolPtr(true), boolPtr(false),
		int32Ptr(30), int32Ptr(60), nil,
		int32Ptr(5), nil, nil,
	)
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	assert.True(t, dot.DeleteStateVersions.ValueBool())
	assert.True(t, dot.DeleteConfigurationVersions.ValueBool())
	assert.False(t, dot.DeleteRunDataAndLogs.ValueBool())

	svDays := dot.StateVersionsDeleteAfterNDays.ValueInt64()
	cvDays := dot.ConfigurationVersionsDeleteAfterNDays.ValueInt64()
	assert.Equal(t, int64(30), svDays)
	assert.Equal(t, int64(60), cvDays)
	assert.True(t, dot.RunDataAndLogsDeleteAfterNDays.IsNull())

	svKeep := dot.StateVersionsKeepLatestCount.ValueInt64()
	assert.Equal(t, int64(5), svKeep)
	assert.True(t, dot.ConfigurationVersionsKeepLatestCount.IsNull())
	assert.True(t, dot.RunDataKeepLatestCount.IsNull())
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_ZeroIsPreservedNotNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	trueVal := true
	policy := newFullDeleteOlderV2Policy("drp-789", 0, &trueVal, &trueVal, &trueVal, nil, nil, nil, int32Ptr(0), int32Ptr(0), int32Ptr(0))
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	assert.False(t, dot.StateVersionsKeepLatestCount.IsNull())
	assert.False(t, dot.ConfigurationVersionsKeepLatestCount.IsNull())
	assert.False(t, dot.RunDataKeepLatestCount.IsNull())

	stateVal := dot.StateVersionsKeepLatestCount.ValueInt64()
	assert.Equal(t, int64(0), stateVal)
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_MaxInt32NoOverflow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	var maxInt32 int32 = 2147483647
	trueVal2 := true
	policy := newFullDeleteOlderV2Policy("drp-max", 0, &trueVal2, nil, nil, nil, nil, nil, &maxInt32, nil, nil)
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	assert.Equal(t, int64(2147483647), dot.StateVersionsKeepLatestCount.ValueInt64())
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_DaysStillMappedCorrectly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := newDeleteOlderV2Policy("drp-days", 99, nil, nil, nil)
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	assert.Equal(t, "drp-days", result.ID.ValueString())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	days := dot.Days.ValueInt64()
	assert.Equal(t, int64(99), days)
}

func TestModelFromTFEDataRetentionPolicyDontDelete_NoKeepLatestFields(t *testing.T) {
	t.Parallel()

	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := newDontDeleteV2Policy("drp-dont-123")
	result := dontDeleteFromAPIResponse(prior.Organization, prior.WorkspaceID, policy)

	assert.Equal(t, "drp-dont-123", result.ID.ValueString())
	assert.True(t, result.DeleteOlderThan.IsNull())
	assert.False(t, result.DontDelete.IsNull())
}

func TestNewDeleteOlderEnvelope_AllFieldsSet(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(30),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolValue(true),
		DeleteRunDataAndLogs:                  types.BoolValue(false),
		StateVersionsDeleteAfterNDays:         types.Int64Value(30),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Value(60),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Value(5),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	attrs := env.GetData().GetAttributes()
	require.NotNil(t, attrs)

	assert.Nil(t, attrs.GetDeleteOlderThanNDays(), "days should be suppressed when granular fields are present")
	assert.Equal(t, true, *attrs.GetDeleteStateVersions())
	assert.Equal(t, true, *attrs.GetDeleteConfigurationVersions())
	assert.Equal(t, false, *attrs.GetDeleteRunDataAndLogs())
	assert.Equal(t, int32(30), *attrs.GetStateVersionsDeleteAfterNDays())
	assert.Equal(t, int32(60), *attrs.GetConfigurationVersionsDeleteAfterNDays())
	assert.Nil(t, attrs.GetRunDataAndLogsDeleteAfterNDays())
	assert.Equal(t, int32(5), *attrs.GetStateVersionsKeepLatestCount())
	assert.Nil(t, attrs.GetConfigurationVersionsKeepLatestCount())
	assert.Nil(t, attrs.GetRunDataKeepLatestCount())
}

func TestNewDeleteOlderEnvelope_NullKeepLatestFieldsOmitted(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(30),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	attrs := env.GetData().GetAttributes()
	require.NotNil(t, attrs)

	assert.Nil(t, attrs.GetDeleteStateVersions())
	assert.Nil(t, attrs.GetDeleteConfigurationVersions())
	assert.Nil(t, attrs.GetDeleteRunDataAndLogs())
	assert.Nil(t, attrs.GetStateVersionsDeleteAfterNDays())
	assert.Nil(t, attrs.GetConfigurationVersionsDeleteAfterNDays())
	assert.Nil(t, attrs.GetRunDataAndLogsDeleteAfterNDays())
	assert.Nil(t, attrs.GetStateVersionsKeepLatestCount())
	assert.Nil(t, attrs.GetConfigurationVersionsKeepLatestCount())
	assert.Nil(t, attrs.GetRunDataKeepLatestCount())
}

func TestNewDontDeleteEnvelope_TypeDiscriminator(t *testing.T) {
	t.Parallel()

	env := newDontDeleteEnvelope()

	require.NotNil(t, env.GetData())
	pType := env.GetData().GetTypeEscaped()
	require.NotNil(t, pType)
	assert.Equal(t, models.DATARETENTIONPOLICYDONTDELETES_DATARETENTIONPOLICY_TYPE, *pType)
}

func TestDeleteOlderEnvelope_TypeDiscriminator(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(1),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	pType := env.GetData().GetTypeEscaped()
	require.NotNil(t, pType)
	assert.Equal(t, models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE, *pType)
}

func TestModelFromTFEDataRetentionPolicyV2_NilPolicyReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	_, diags := dataRetentionPolicyFromAPIResponse(ctx, prior, nil)
	assert.True(t, diags.HasError())
}

func TestModelFromTFEDataRetentionPolicyV2_NilTypeDefaultsToDeleteOlder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := models.NewDataRetentionPolicy()
	id := "drp-nil-type"
	policy.SetId(&id)
	attrs := models.NewDataRetentionPolicy_attributes()
	days := int32(14)
	attrs.SetDeleteOlderThanNDays(&days)
	policy.SetAttributes(attrs)

	result, diags := dataRetentionPolicyFromAPIResponse(ctx, prior, policy)
	require.False(t, diags.HasError())

	assert.False(t, result.DeleteOlderThan.IsNull())
	assert.True(t, result.DontDelete.IsNull())
}

func TestModelFromTFEDataRetentionPolicyV2_DontDeleteRouting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := modelTFEDataRetentionPolicy{
		Organization: types.StringValue("my-org"),
		WorkspaceID:  types.StringNull(),
	}

	policy := newDontDeleteV2Policy("drp-dont-routing")
	result, diags := dataRetentionPolicyFromAPIResponse(ctx, prior, policy)
	require.False(t, diags.HasError())

	assert.True(t, result.DeleteOlderThan.IsNull())
	assert.False(t, result.DontDelete.IsNull())
	assert.Equal(t, "drp-dont-routing", result.ID.ValueString())
}

func TestHasGranularFields_FalseWhenAllNull(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(30),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}
	assert.False(t, hasGranularFields(dot))
}

func TestHasGranularFields_TrueForEachField(t *testing.T) {
	t.Parallel()

	base := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(30),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	cases := []struct {
		name  string
		mutFn func(*modelTFEDeleteOlderThan)
	}{
		{"delete_state_versions", func(d *modelTFEDeleteOlderThan) { d.DeleteStateVersions = types.BoolValue(true) }},
		{"delete_configuration_versions", func(d *modelTFEDeleteOlderThan) { d.DeleteConfigurationVersions = types.BoolValue(true) }},
		{"delete_run_data_and_logs", func(d *modelTFEDeleteOlderThan) { d.DeleteRunDataAndLogs = types.BoolValue(false) }},
		{"state_versions_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.StateVersionsDeleteAfterNDays = types.Int64Value(30)
		}},
		{"configuration_versions_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.ConfigurationVersionsDeleteAfterNDays = types.Int64Value(30)
		}},
		{"run_data_and_logs_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.RunDataAndLogsDeleteAfterNDays = types.Int64Value(30)
		}},
		{"state_versions_keep_latest_count", func(d *modelTFEDeleteOlderThan) { d.StateVersionsKeepLatestCount = types.Int64Value(5) }},
		{"configuration_versions_keep_latest_count", func(d *modelTFEDeleteOlderThan) {
			d.ConfigurationVersionsKeepLatestCount = types.Int64Value(5)
		}},
		{"run_data_keep_latest_count", func(d *modelTFEDeleteOlderThan) { d.RunDataKeepLatestCount = types.Int64Value(5) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := base
			tc.mutFn(&d)
			assert.True(t, hasGranularFields(d), "expected hasGranularFields=true when %s is set", tc.name)
		})
	}
}

func TestNewDeleteOlderEnvelope_DaysSuppressedWhenGranularFieldsPresent(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(99),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Value(30),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	attrs := env.GetData().GetAttributes()
	assert.Nil(t, attrs.GetDeleteOlderThanNDays(), "days should be suppressed when granular fields are present")
	assert.Equal(t, int32(30), *attrs.GetStateVersionsDeleteAfterNDays())
}

func TestNewDeleteOlderEnvelope_DaysPassedWhenNoGranularFields(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(42),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	attrs := env.GetData().GetAttributes()
	require.NotNil(t, attrs.GetDeleteOlderThanNDays(), "days should be sent when no granular fields are set")
	assert.Equal(t, int32(42), *attrs.GetDeleteOlderThanNDays())
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_IgnoresBackfilledGranularFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Simulate user config with only "days: 30" set
	priorDeleteOlderThan := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(30),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}
	priorDeleteOlderThanObject, _ := types.ObjectValueFrom(ctx, priorDeleteOlderThan.AttributeTypes(), priorDeleteOlderThan)

	prior := modelTFEDataRetentionPolicy{
		Organization:    types.StringValue("my-org"),
		WorkspaceID:     types.StringNull(),
		DeleteOlderThan: priorDeleteOlderThanObject,
	}

	// Simulate API response where server backfilled granular fields
	deleteStateVersions := true
	deleteConfigurationVersions := true
	stateVersionsWindow := int32(30)
	configurationVersionsWindow := int32(30)
	id := "drp-backfilled"

	policy := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	policy.SetTypeEscaped(&pType)
	policy.SetId(&id)

	attrs := models.NewDataRetentionPolicy_attributes()
	attrs.SetDeleteOlderThanNDays(&stateVersionsWindow)
	attrs.SetDeleteStateVersions(&deleteStateVersions)
	attrs.SetDeleteConfigurationVersions(&deleteConfigurationVersions)
	attrs.SetStateVersionsDeleteAfterNDays(&stateVersionsWindow)
	attrs.SetConfigurationVersionsDeleteAfterNDays(&configurationVersionsWindow)
	policy.SetAttributes(attrs)

	// Read should ignore the backfilled granular fields and preserve "days: 30"
	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	// Should preserve original "days" value
	require.False(t, dot.Days.IsNull(), "days should not be null")
	days := dot.Days.ValueInt64()
	assert.Equal(t, int64(30), days)

	// Should ignore backfilled granular fields
	assert.True(t, dot.DeleteStateVersions.IsNull(), "delete_state_versions should be null")
	assert.True(t, dot.DeleteConfigurationVersions.IsNull(), "delete_configuration_versions should be null")
	assert.True(t, dot.StateVersionsDeleteAfterNDays.IsNull(), "state_versions_delete_after_n_days should be null")
	assert.True(t, dot.ConfigurationVersionsDeleteAfterNDays.IsNull(), "configuration_versions_delete_after_n_days should be null")
}

func TestModelFromTFEDataRetentionPolicyDeleteOlder_GranularPriorState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	deleteStateVersions := true
	deleteRunDataAndLogs := true
	stateWindow := int32(14)
	stateKeep := int32(5)
	runKeep := int32(3)

	priorDeleteOlderThan := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Null(),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolValue(true),
		StateVersionsDeleteAfterNDays:         types.Int64Value(14),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Value(5),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Value(3),
	}
	priorDeleteOlderThanObject, diags := types.ObjectValueFrom(ctx, priorDeleteOlderThan.AttributeTypes(), priorDeleteOlderThan)
	require.False(t, diags.HasError())

	prior := modelTFEDataRetentionPolicy{
		Organization:    types.StringValue("my-org"),
		WorkspaceID:     types.StringNull(),
		DeleteOlderThan: priorDeleteOlderThanObject,
	}

	id := "drp-granular"
	policy := models.NewDataRetentionPolicy()
	pType := models.DATARETENTIONPOLICYDELETEOLDERS_DATARETENTIONPOLICY_TYPE
	policy.SetTypeEscaped(&pType)
	policy.SetId(&id)

	attrs := models.NewDataRetentionPolicy_attributes()
	attrs.SetDeleteStateVersions(&deleteStateVersions)
	attrs.SetDeleteRunDataAndLogs(&deleteRunDataAndLogs)
	attrs.SetStateVersionsDeleteAfterNDays(&stateWindow)
	attrs.SetStateVersionsKeepLatestCount(&stateKeep)
	attrs.SetRunDataKeepLatestCount(&runKeep)
	policy.SetAttributes(attrs)

	result, diags := deleteOlderThanFromAPIResponse(ctx, prior.Organization, prior.WorkspaceID, prior.DeleteOlderThan, policy)
	require.False(t, diags.HasError())

	var dot modelTFEDeleteOlderThan
	diags = result.DeleteOlderThan.As(ctx, &dot, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	assert.True(t, dot.Days.IsNull(), "days should be null for granular config")

	assert.False(t, dot.DeleteStateVersions.IsNull())
	assert.Equal(t, true, dot.DeleteStateVersions.ValueBool())

	assert.False(t, dot.DeleteRunDataAndLogs.IsNull())
	assert.Equal(t, true, dot.DeleteRunDataAndLogs.ValueBool())

	svDays := dot.StateVersionsDeleteAfterNDays.ValueInt64()
	assert.Equal(t, int64(14), svDays)

	svKeep := dot.StateVersionsKeepLatestCount.ValueInt64()
	assert.Equal(t, int64(5), svKeep)

	runKeepVal := dot.RunDataKeepLatestCount.ValueInt64()
	assert.Equal(t, int64(3), runKeepVal)

	assert.True(t, dot.DeleteConfigurationVersions.IsNull())
	assert.True(t, dot.ConfigurationVersionsDeleteAfterNDays.IsNull())
	assert.True(t, dot.ConfigurationVersionsKeepLatestCount.IsNull())
}

func serializeEnvelope(t *testing.T, env models.DataRetentionPolicyEnvelopeable) map[string]any {
	t.Helper()

	writer := jsonserialization.NewJsonSerializationWriter()
	require.NoError(t, writer.WriteObjectValue("", env))
	content, err := writer.GetSerializedContent()
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(content, &body))
	data, ok := body["data"].(map[string]any)
	require.True(t, ok, "expected data object in %s", content)
	attrs, ok := data["attributes"].(map[string]any)
	require.True(t, ok, "expected attributes object in %s", content)
	return attrs
}

func TestNewDeleteOlderEnvelope_GranularSendsExplicitFalseAndNulls(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Null(),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Value(30),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	attrs := serializeEnvelope(t, env)

	assert.NotContains(t, attrs, "delete-older-than-n-days")
	assert.Equal(t, true, attrs["delete-state-versions"])
	assert.Equal(t, false, attrs["delete-configuration-versions"])
	assert.Equal(t, false, attrs["delete-run-data-and-logs"])
	assert.InDelta(t, 30, attrs["state-versions-delete-after-n-days"], 0)

	for _, key := range []string{
		"configuration-versions-delete-after-n-days",
		"run-data-and-logs-delete-after-n-days",
		"state-versions-keep-latest-count",
		"configuration-versions-keep-latest-count",
		"run-data-keep-latest-count",
	} {
		value, present := attrs[key]
		assert.True(t, present, "expected %s to be sent as an explicit null", key)
		assert.Nil(t, value, "expected %s to be null", key)
	}
}

func TestNewDeleteOlderEnvelope_DaysOnlySendsOnlyDays(t *testing.T) {
	t.Parallel()

	dot := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(42),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}

	env, err := newDeleteOlderEnvelope(dot)
	require.NoError(t, err)

	// TFE rejects requests that combine days with any per-artifact-type key, including nulls.
	assert.Equal(t, map[string]any{"delete-older-than-n-days": float64(42)}, serializeEnvelope(t, env))
}

// TestInt64ToInt32Ptr checks the narrowing to the int32 the go-tfe v2 setters expect. An
// out-of-range value must be rejected rather than wrapped. The schema validators catch these
// at plan time; this guard also proves the narrowing in range for gosec.
func TestInt64ToInt32Ptr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		in      types.Int64
		want    *int32
		wantErr bool
	}{
		{name: "null is omitted", in: types.Int64Null(), want: nil},
		{name: "unknown is omitted", in: types.Int64Unknown(), want: nil},
		{name: "a typical value is in range", in: types.Int64Value(30), want: int32Ptr(30)},
		{name: "the upper bound is in range", in: types.Int64Value(math.MaxInt32), want: int32Ptr(math.MaxInt32)},
		{name: "the lower bound is in range", in: types.Int64Value(math.MinInt32), want: int32Ptr(math.MinInt32)},
		{name: "above int32 is rejected, not wrapped", in: types.Int64Value(math.MaxInt32 + 1), wantErr: true},
		{name: "below int32 is rejected, not wrapped", in: types.Int64Value(math.MinInt32 - 1), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := int64ToInt32Ptr("state-versions-delete-after-n-days", tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "state-versions-delete-after-n-days")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewDeleteOlderEnvelope_OutOfRangeReturnsError(t *testing.T) {
	t.Parallel()

	granular := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Null(),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Value(math.MaxInt32 + 1),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}
	_, err := newDeleteOlderEnvelope(granular)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state-versions-delete-after-n-days")

	daysOnly := modelTFEDeleteOlderThan{
		Days:                                  types.Int64Value(math.MaxInt32 + 1),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.Int64Null(),
		ConfigurationVersionsDeleteAfterNDays: types.Int64Null(),
		RunDataAndLogsDeleteAfterNDays:        types.Int64Null(),
		StateVersionsKeepLatestCount:          types.Int64Null(),
		ConfigurationVersionsKeepLatestCount:  types.Int64Null(),
		RunDataKeepLatestCount:                types.Int64Null(),
	}
	_, err = newDeleteOlderEnvelope(daysOnly)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "days")
}

// TestDataRetentionPolicySchema_NumericAttributesAreRangeChecked runs the schema validators on
// each numeric attribute, so out-of-range values are rejected at plan time before they can
// be narrowed to int32.
func TestDataRetentionPolicySchema_NumericAttributesAreRangeChecked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaResp := &resource.SchemaResponse{}
	NewDataRetentionPolicyResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	block, ok := schemaResp.Schema.Blocks["delete_older_than"].(schema.SingleNestedBlock)
	require.True(t, ok)

	for _, name := range []string{
		"days",
		"state_versions_delete_after_n_days",
		"configuration_versions_delete_after_n_days",
		"run_data_and_logs_delete_after_n_days",
		"state_versions_keep_latest_count",
		"configuration_versions_keep_latest_count",
		"run_data_keep_latest_count",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attribute, ok := block.Attributes[name].(schema.Int64Attribute)
			require.True(t, ok, "expected %s to be an Int64Attribute", name)

			validate := func(v int64) bool {
				resp := &validator.Int64Response{}
				for _, val := range attribute.Int64Validators() {
					val.ValidateInt64(ctx, validator.Int64Request{
						Path:        path.Root("delete_older_than").AtName(name),
						ConfigValue: types.Int64Value(v),
					}, resp)
				}
				return !resp.Diagnostics.HasError()
			}

			assert.True(t, validate(1), "1 should be accepted")
			assert.True(t, validate(math.MaxInt32), "the int32 maximum should be accepted")
			assert.False(t, validate(0), "0 should be rejected")
			assert.False(t, validate(-1), "negative values should be rejected")
			assert.False(t, validate(math.MaxInt32+1), "values above the int32 maximum should be rejected")
		})
	}
}
