package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func customizeDiffIfProviderDefaultOrganizationChanged(c context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	config := meta.(ConfiguredClient)

	configOrg := diff.GetRawConfig().GetAttr("organization")
	plannedOrg := diff.Get("organization").(string)

	if configOrg.IsNull() && config.Organization != plannedOrg {
		// There is no organization configured on the resource, yet it is different from
		// the state organization. We must conclude that the provider default organization changed.
		if err := diff.SetNew("organization", config.Organization); err != nil {
			return err
		}
	}
	return nil
}
