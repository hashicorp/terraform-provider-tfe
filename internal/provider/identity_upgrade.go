// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
)

// hostnameIdentityUpgrader upgrades an identity whose hostname was derived from
// the client discovered URL. That host can be a local proxy with an ephemeral port,
// so the stored hostname is replaced with the stable configured hostname.
// All other attributes are preserved, so priorSchema must have the same
// attributes as the current identity schema.
func hostnameIdentityUpgrader(priorSchema identityschema.Schema, hostname string) resource.IdentityUpgrader {
	return resource.IdentityUpgrader{
		PriorSchema: &priorSchema,
		IdentityUpgrader: func(ctx context.Context, req resource.UpgradeIdentityRequest, resp *resource.UpgradeIdentityResponse) {
			resp.Identity.Raw = req.Identity.Raw

			// Without a configured provider there is no hostname to migrate to.
			if hostname == "" || req.Identity.Raw.IsFullyNull() {
				return
			}

			resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("hostname"), hostname)...)
		},
	}
}
