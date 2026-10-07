// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const testIdentityHostname = "tfe.example.com"

type identityUpgradeResource interface {
	resource.ResourceWithIdentity
	resource.ResourceWithUpgradeIdentity
}

func currentIdentitySchema(t *testing.T, ctx context.Context, r identityUpgradeResource) resource.IdentitySchemaResponse {
	t.Helper()
	resp := resource.IdentitySchemaResponse{}
	r.IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected identity schema diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func runIdentityUpgradeV0(t *testing.T, ctx context.Context, r identityUpgradeResource, prior map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	upgrader, ok := r.UpgradeIdentity(ctx)[0]
	if !ok {
		t.Fatal("expected an identity upgrader for version 0")
	}

	priorType := upgrader.PriorSchema.Type().TerraformType(ctx)
	req := resource.UpgradeIdentityRequest{
		Identity: &tfsdk.ResourceIdentity{Raw: tftypes.NewValue(priorType, prior), Schema: *upgrader.PriorSchema},
	}
	resp := resource.UpgradeIdentityResponse{
		Identity: &tfsdk.ResourceIdentity{Schema: currentIdentitySchema(t, ctx, r).IdentitySchema},
	}
	upgrader.IdentityUpgrader(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected identity upgrade diagnostics: %v", resp.Diagnostics)
	}

	return resp.Identity.Raw
}

func TestIdentityUpgradeV0ToV1_PriorSchemaHasSameAttributesAsV1Schema(t *testing.T) {
	ctx := context.Background()
	resources := map[string]identityUpgradeResource{
		"tfe_project":           &resourceTFEProject{},
		"tfe_stack":             &resourceTFEStack{},
		"tfe_variable":          &resourceTFEVariable{},
		"tfe_registry_provider": &resourceTFERegistryProvider{},
	}

	for name, r := range resources {
		t.Run(name, func(t *testing.T) {
			current := currentIdentitySchema(t, ctx, r).IdentitySchema
			if current.Version != 1 {
				t.Fatalf("expected identity schema version 1, got %d", current.Version)
			}

			prior := r.UpgradeIdentity(ctx)[0].PriorSchema
			if prior == nil {
				t.Fatal("expected a prior schema for version 0")
			}
			// The upgrader copies the prior raw identity as-is, which only works
			// when both versions have the same attributes.
			if !prior.Type().TerraformType(ctx).Equal(current.Type().TerraformType(ctx)) {
				t.Fatalf("version 0 attributes must match the version 1 attributes: %s != %s",
					prior.Type().TerraformType(ctx), current.Type().TerraformType(ctx))
			}
		})
	}
}

func TestIdentityUpgradeV0ToV1_ReplacesStoredHostnameWithConfiguredHostname(t *testing.T) {
	ctx := context.Background()

	cases := map[string]string{
		"stored tfc-agent HYOK proxy address": "127.0.0.1:39553",
		"stored generic localterraform.com":   "localterraform.com",
		"stored discovered API host":          "api.tfe.example.com",
	}

	for name, priorHostname := range cases {
		t.Run(name, func(t *testing.T) {
			r := &resourceTFEVariable{config: ConfiguredClient{Hostname: testIdentityHostname}}

			got := runIdentityUpgradeV0(t, ctx, r, map[string]tftypes.Value{
				"id":              tftypes.NewValue(tftypes.String, "var-123"),
				"configurable_id": tftypes.NewValue(tftypes.String, "ws-123"),
				"hostname":        tftypes.NewValue(tftypes.String, priorHostname),
			})

			var identity modelTFEVariableIdentity
			if diags := (&tfsdk.ResourceIdentity{Raw: got, Schema: currentIdentitySchema(t, ctx, r).IdentitySchema}).Get(ctx, &identity); diags.HasError() {
				t.Fatalf("unexpected diagnostics reading upgraded identity: %v", diags)
			}

			if identity.Hostname.ValueString() != testIdentityHostname {
				t.Fatalf("expected hostname %q, got %q", testIdentityHostname, identity.Hostname.ValueString())
			}
			if identity.ID.ValueString() != "var-123" || identity.ConfigurableID.ValueString() != "ws-123" {
				t.Fatalf("expected other identity attributes to be preserved, got %+v", identity)
			}
		})
	}
}

func TestIdentityUpgradeV0ToV1_PreservesIdentityWhenProviderIsNotConfigured(t *testing.T) {
	ctx := context.Background()
	r := &resourceTFEProject{config: ConfiguredClient{}}
	prior := map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "prj-123"),
		"hostname": tftypes.NewValue(tftypes.String, "127.0.0.1:39553"),
	}

	assertIdentityUpgradeV0Preserved(t, ctx, r, prior)
}

func TestIdentityUpgradeV0ToV1_PreservesIdentityWhenPriorIdentityIsNull(t *testing.T) {
	ctx := context.Background()
	r := &resourceTFEProject{config: ConfiguredClient{Hostname: testIdentityHostname}}
	prior := map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, nil),
		"hostname": tftypes.NewValue(tftypes.String, nil),
	}

	assertIdentityUpgradeV0Preserved(t, ctx, r, prior)
}

func assertIdentityUpgradeV0Preserved(t *testing.T, ctx context.Context, r identityUpgradeResource, prior map[string]tftypes.Value) {
	t.Helper()
	upgrader := r.UpgradeIdentity(ctx)[0]
	expected := tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), prior)

	got := runIdentityUpgradeV0(t, ctx, r, prior)

	if !got.Equal(expected) {
		t.Fatalf("expected identity to be preserved as %s, got %s", expected, got)
	}
}
