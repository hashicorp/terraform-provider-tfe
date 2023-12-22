// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func customizeDiffIfProviderDefaultOrganizationChanged(c context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	config := meta.(ConfiguredClient)

	configOrg := diff.GetRawConfig().GetAttr("organization")
	plannedOrg := diff.Get("organization").(string)

	if configOrg.IsNull() && config.Organization != plannedOrg {
		// There is no organization configured on the resource, yet the provider org is different from
		// the planned organization. We must conclude that the provider default organization changed.
		if err := diff.SetNew("organization", config.Organization); err != nil {
			return err
		}
	}
	return nil
}

func modifyPlanForDefaultOrganizationChange(ctx context.Context, providerDefaultOrg string, state tfsdk.State, configAttributes, planAttributes AttrGettable, resp *resource.ModifyPlanResponse) {
	if state.Raw.IsNull() {
		return
	}

	orgPath := path.Root("organization")

	var configOrg, plannedOrg types.String
	resp.Diagnostics.Append(configAttributes.GetAttribute(ctx, orgPath, &configOrg)...)
	resp.Diagnostics.Append(planAttributes.GetAttribute(ctx, orgPath, &plannedOrg)...)

	if configOrg.IsNull() && !plannedOrg.IsNull() && providerDefaultOrg != plannedOrg.ValueString() {
		// There is no organization configured on the resource, yet the provider org is different from
		// the planned organization value. We must conclude that the provider default organization changed.
		resp.Plan.SetAttribute(ctx, orgPath, types.StringValue(providerDefaultOrg))
		resp.RequiresReplace.Append(orgPath)
	}
}
