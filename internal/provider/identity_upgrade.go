// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// hostnameIdentityUpgrader upgrades an identity whose hostname was derived from
// the client discovered URL. That host can be a local proxy with an ephemeral port,
// so the stored hostname is replaced with the stable configured hostname.
// All other attributes are preserved, so priorSchema must have the same
// attributes as the current identity schema.
func hostnameIdentityUpgrader(priorSchema identityschema.Schema, hostname string) resource.IdentityUpgrader {
	return resource.IdentityUpgrader{
		// PriorSchema is intentionally unset. Terraform also requests this upgrade
		// for state that has no stored identity (e.g. written by older provider versions
		// without identity support), which the framework fails to decode, so the
		// raw identity is decoded here instead.
		IdentityUpgrader: func(ctx context.Context, req resource.UpgradeIdentityRequest, resp *resource.UpgradeIdentityResponse) {
			priorType, ok := priorSchema.Type().TerraformType(ctx).(tftypes.Object)
			if !ok {
				resp.Diagnostics.AddError("Unable to upgrade resource identity", "The prior identity schema is not an object. This is a bug in the provider.")
				return
			}

			// A fully null identity is allowed to change, so Read backfills it.
			if req.RawIdentity == nil || (len(req.RawIdentity.JSON) == 0 && len(req.RawIdentity.Flatmap) == 0) {
				resp.Identity.Raw = fullyNullObject(priorType)
				return
			}

			prior, err := req.RawIdentity.UnmarshalWithOpts(priorType, tfprotov6.UnmarshalOpts{
				ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true},
			})
			if err != nil {
				resp.Diagnostics.AddError("Unable to upgrade resource identity", fmt.Sprintf("Unable to read the stored version 0 identity: %s", err))
				return
			}
			resp.Identity.Raw = prior

			// Without a configured provider there is no hostname to migrate to.
			if hostname == "" || prior.IsFullyNull() {
				return
			}

			resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("hostname"), hostname)...)
		},
	}
}

func fullyNullObject(objectType tftypes.Object) tftypes.Value {
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	return tftypes.NewValue(objectType, attributes)
}
