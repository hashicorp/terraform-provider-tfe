package provider

import (
	"context"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"math/big"
)

type modelTFEDataRetentionPolicy struct {
	ID              types.String `tfsdk:"id"`
	Organization    types.String `tfsdk:"organization"`
	WorkspaceId     types.String `tfsdk:"workspace_id"`
	DeleteOlderThan types.Object `tfsdk:"delete_older_than"`
	DontDelete      types.Object `tfsdk:"dont_delete"`
}

type modelTFEDeleteOlderThan struct {
	Days types.Number `tfsdk:"days"`
}

func (m modelTFEDeleteOlderThan) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"days": types.NumberType,
	}
}

func DontDeleteEmptyObject() basetypes.ObjectValue {
	object, diags := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	if diags.HasError() {
		panic(diags.Errors())
	}
	return object
}

func modelFromTFEDataRetentionPolicyDeleteOlder(ctx context.Context, model modelTFEDataRetentionPolicy, deleteOlder *tfe.DataRetentionPolicyDeleteOlder) (modelTFEDataRetentionPolicy, diag.Diagnostics) {
	deleteOlderThan := modelTFEDeleteOlderThan{
		Days: types.NumberValue(big.NewFloat(float64(deleteOlder.DeleteOlderThanNDays))),
	}
	deleteOlderThanObject, diags := types.ObjectValueFrom(ctx, deleteOlderThan.AttributeTypes(), deleteOlderThan)

	return modelTFEDataRetentionPolicy{
		ID:              types.StringValue(deleteOlder.ID),
		Organization:    model.Organization,
		WorkspaceId:     model.WorkspaceId,
		DeleteOlderThan: deleteOlderThanObject,
		DontDelete:      types.ObjectNull(map[string]attr.Type{}),
	}, diags
}

func modelFromTFEDataRetentionPolicyDontDelete(model modelTFEDataRetentionPolicy, dontDelete *tfe.DataRetentionPolicyDontDelete) modelTFEDataRetentionPolicy {
	return modelTFEDataRetentionPolicy{
		ID:              types.StringValue(dontDelete.ID),
		Organization:    model.Organization,
		WorkspaceId:     model.WorkspaceId,
		DeleteOlderThan: types.ObjectNull(modelTFEDeleteOlderThan{}.AttributeTypes()),
		DontDelete:      DontDeleteEmptyObject(),
	}
}

func modelFromTFELegacyDataRetentionPolicy(ctx context.Context, model modelTFEDataRetentionPolicy, legacy *tfe.DataRetentionPolicy) (modelTFEDataRetentionPolicy, diag.Diagnostics) {
	deleteOlderThan := modelTFEDeleteOlderThan{
		Days: types.NumberValue(big.NewFloat(float64(legacy.DeleteOlderThanNDays))),
	}
	deleteOlderThanObject, diags := types.ObjectValueFrom(ctx, deleteOlderThan.AttributeTypes(), deleteOlderThan)

	return modelTFEDataRetentionPolicy{
		ID:              types.StringValue(legacy.ID),
		Organization:    model.Organization,
		WorkspaceId:     model.WorkspaceId,
		DeleteOlderThan: deleteOlderThanObject,
		DontDelete:      types.ObjectNull(map[string]attr.Type{}),
	}, diags
}

func modelFromTFEDataRetentionPolicyChoice(ctx context.Context, model modelTFEDataRetentionPolicy, choice *tfe.DataRetentionPolicyChoice) (modelTFEDataRetentionPolicy, diag.Diagnostics) {
	if choice.DataRetentionPolicyDeleteOlder != nil {
		return modelFromTFEDataRetentionPolicyDeleteOlder(ctx, model, choice.DataRetentionPolicyDeleteOlder)
	}

	var emptyDiag []diag.Diagnostic
	if choice.DataRetentionPolicyDontDelete != nil {
		return modelFromTFEDataRetentionPolicyDontDelete(model, choice.DataRetentionPolicyDontDelete), emptyDiag
	}

	legacyPolicy := choice.ConvertToLegacyStruct()
	return modelFromTFELegacyDataRetentionPolicy(ctx, model, legacyPolicy)
}
