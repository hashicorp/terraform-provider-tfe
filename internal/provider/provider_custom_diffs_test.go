package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type getter struct {
	val types.String
}

func (g *getter) GetAttribute(_ context.Context, _ path.Path, target interface{}) diag.Diagnostics {
	*(target.(*basetypes.StringValue)) = g.val
	return diag.Diagnostics{}
}

func TestModifyPlanForDefaultOrganizationChange(t *testing.T) {
	// if configOrg.IsNull() && !plannedOrg.IsNull() && providerDefaultOrg != plannedOrg.ValueString() {
	testCases := map[string]struct {
		providerDefaultOrg      string
		planValue               types.String
		configValue             types.String
		expectedPlanValue       string
		expectedRequiresReplace bool
	}{
		"No change in provider org": {
			providerDefaultOrg:      "foo",
			planValue:               types.StringValue("foo"),
			configValue:             types.StringNull(),
			expectedPlanValue:       "foo",
			expectedRequiresReplace: false,
		},
		"Change in provider org": {
			providerDefaultOrg:      "bar",
			planValue:               types.StringValue("foo"),
			configValue:             types.StringNull(),
			expectedPlanValue:       "bar",
			expectedRequiresReplace: true,
		},
		"Config org changed": {
			providerDefaultOrg: "foo",
			planValue:          types.StringValue("bar"),
			configValue:        types.StringValue("bar"),
			expectedPlanValue:  "bar",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fakeState := tftypes.NewValue(tftypes.Object{}, make(map[string]tftypes.Value))

			fakeSchema := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"organization": schema.StringAttribute{
						Computed:    true,
						Optional:    true,
						Description: "Test organization",
					},
				},
			}

			fakePlan := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"organization": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"organization": tftypes.NewValue(tftypes.String, tc.planValue.ValueString()),
				},
			)

			fakeResponse := &resource.ModifyPlanResponse{
				Plan:            tfsdk.Plan{Schema: fakeSchema, Raw: fakePlan},
				RequiresReplace: make(path.Paths, 0),
				Diagnostics:     diag.Diagnostics{},
			}

			c := context.TODO()

			modifyPlanForDefaultOrganizationChange(
				c,
				tc.providerDefaultOrg,
				tfsdk.State{Raw: fakeState},
				&getter{val: tc.configValue},
				&getter{val: tc.planValue},
				fakeResponse,
			)

			orgPath := path.Root("organization")
			var value types.String
			fakeResponse.Plan.GetAttribute(c, orgPath, &value)
			if fakeResponse.Diagnostics.HasError() {
				t.Fatalf("Expected no errors, got %v", fakeResponse.Diagnostics)
			}

			if value.ValueString() != tc.expectedPlanValue {
				t.Fatalf("Expected plan value to be %q, got %q", tc.expectedPlanValue, value.ValueString())
			}

			if tc.expectedRequiresReplace && len(fakeResponse.RequiresReplace) == 0 {
				t.Fatal("Expected RequiresReplace to be set, but it was not")
			}
		})
	}
}
