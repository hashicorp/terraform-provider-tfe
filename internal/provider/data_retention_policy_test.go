// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	assert.Equal(t, types.NumberType, attrTypes["days"])
	assert.Equal(t, types.BoolType, attrTypes["delete_state_versions"])
	assert.Equal(t, types.BoolType, attrTypes["delete_configuration_versions"])
	assert.Equal(t, types.BoolType, attrTypes["delete_run_data_and_logs"])
	assert.Equal(t, types.NumberType, attrTypes["state_versions_delete_after_n_days"])
	assert.Equal(t, types.NumberType, attrTypes["configuration_versions_delete_after_n_days"])
	assert.Equal(t, types.NumberType, attrTypes["run_data_and_logs_delete_after_n_days"])
	assert.Equal(t, types.NumberType, attrTypes["state_versions_keep_latest_count"])
	assert.Equal(t, types.NumberType, attrTypes["configuration_versions_keep_latest_count"])
	assert.Equal(t, types.NumberType, attrTypes["run_data_keep_latest_count"])
}

func TestModelTFEDeleteOlderThan_AttributeTypes_NoExtraKeys(t *testing.T) {
	t.Parallel()

	attrTypes := modelTFEDeleteOlderThan{}.AttributeTypes()
	expected := map[string]attr.Type{
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

	stateVal, _ := dot.StateVersionsKeepLatestCount.ValueBigFloat().Int64()
	cvVal, _ := dot.ConfigurationVersionsKeepLatestCount.ValueBigFloat().Int64()
	runVal, _ := dot.RunDataKeepLatestCount.ValueBigFloat().Int64()

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

	svDays, _ := dot.StateVersionsDeleteAfterNDays.ValueBigFloat().Int64()
	cvDays, _ := dot.ConfigurationVersionsDeleteAfterNDays.ValueBigFloat().Int64()
	assert.Equal(t, int64(30), svDays)
	assert.Equal(t, int64(60), cvDays)
	assert.True(t, dot.RunDataAndLogsDeleteAfterNDays.IsNull())

	svKeep, _ := dot.StateVersionsKeepLatestCount.ValueBigFloat().Int64()
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

	stateVal, _ := dot.StateVersionsKeepLatestCount.ValueBigFloat().Int64()
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

	stateVal, accuracy := dot.StateVersionsKeepLatestCount.ValueBigFloat().Int64()
	assert.Equal(t, int64(2147483647), stateVal)
	assert.Equal(t, big.Exact, accuracy)
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

	days, _ := dot.Days.ValueBigFloat().Int64()
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
		Days:                                  types.NumberValue(big.NewFloat(30)),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolValue(true),
		DeleteRunDataAndLogs:                  types.BoolValue(false),
		StateVersionsDeleteAfterNDays:         types.NumberValue(big.NewFloat(30)),
		ConfigurationVersionsDeleteAfterNDays: types.NumberValue(big.NewFloat(60)),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberValue(big.NewFloat(5)),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
		Days:                                  types.NumberValue(big.NewFloat(30)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
		Days:                                  types.NumberValue(big.NewFloat(1)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
		Days:                                  types.NumberValue(big.NewFloat(30)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
	}
	assert.False(t, hasGranularFields(dot))
}

func TestHasGranularFields_TrueForEachField(t *testing.T) {
	t.Parallel()

	base := modelTFEDeleteOlderThan{
		Days:                                  types.NumberValue(big.NewFloat(30)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
	}

	cases := []struct {
		name  string
		mutFn func(*modelTFEDeleteOlderThan)
	}{
		{"delete_state_versions", func(d *modelTFEDeleteOlderThan) { d.DeleteStateVersions = types.BoolValue(true) }},
		{"delete_configuration_versions", func(d *modelTFEDeleteOlderThan) { d.DeleteConfigurationVersions = types.BoolValue(true) }},
		{"delete_run_data_and_logs", func(d *modelTFEDeleteOlderThan) { d.DeleteRunDataAndLogs = types.BoolValue(false) }},
		{"state_versions_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.StateVersionsDeleteAfterNDays = types.NumberValue(big.NewFloat(30))
		}},
		{"configuration_versions_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.ConfigurationVersionsDeleteAfterNDays = types.NumberValue(big.NewFloat(30))
		}},
		{"run_data_and_logs_delete_after_n_days", func(d *modelTFEDeleteOlderThan) {
			d.RunDataAndLogsDeleteAfterNDays = types.NumberValue(big.NewFloat(30))
		}},
		{"state_versions_keep_latest_count", func(d *modelTFEDeleteOlderThan) { d.StateVersionsKeepLatestCount = types.NumberValue(big.NewFloat(5)) }},
		{"configuration_versions_keep_latest_count", func(d *modelTFEDeleteOlderThan) {
			d.ConfigurationVersionsKeepLatestCount = types.NumberValue(big.NewFloat(5))
		}},
		{"run_data_keep_latest_count", func(d *modelTFEDeleteOlderThan) { d.RunDataKeepLatestCount = types.NumberValue(big.NewFloat(5)) }},
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
		Days:                                  types.NumberValue(big.NewFloat(99)),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberValue(big.NewFloat(30)),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
		Days:                                  types.NumberValue(big.NewFloat(42)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
		Days:                                  types.NumberValue(big.NewFloat(30)),
		DeleteStateVersions:                   types.BoolNull(),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolNull(),
		StateVersionsDeleteAfterNDays:         types.NumberNull(),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberNull(),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberNull(),
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
	days, _ := dot.Days.ValueBigFloat().Int64()
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
		Days:                                  types.NumberNull(),
		DeleteStateVersions:                   types.BoolValue(true),
		DeleteConfigurationVersions:           types.BoolNull(),
		DeleteRunDataAndLogs:                  types.BoolValue(true),
		StateVersionsDeleteAfterNDays:         types.NumberValue(big.NewFloat(14)),
		ConfigurationVersionsDeleteAfterNDays: types.NumberNull(),
		RunDataAndLogsDeleteAfterNDays:        types.NumberNull(),
		StateVersionsKeepLatestCount:          types.NumberValue(big.NewFloat(5)),
		ConfigurationVersionsKeepLatestCount:  types.NumberNull(),
		RunDataKeepLatestCount:                types.NumberValue(big.NewFloat(3)),
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

	svDays, _ := dot.StateVersionsDeleteAfterNDays.ValueBigFloat().Int64()
	assert.Equal(t, int64(14), svDays)

	svKeep, _ := dot.StateVersionsKeepLatestCount.ValueBigFloat().Int64()
	assert.Equal(t, int64(5), svKeep)

	runKeepVal, _ := dot.RunDataKeepLatestCount.ValueBigFloat().Int64()
	assert.Equal(t, int64(3), runKeepVal)

	assert.True(t, dot.DeleteConfigurationVersions.IsNull())
	assert.True(t, dot.ConfigurationVersionsDeleteAfterNDays.IsNull())
	assert.True(t, dot.ConfigurationVersionsKeepLatestCount.IsNull())
}
