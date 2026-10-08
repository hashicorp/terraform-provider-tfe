// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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

func runIdentityUpgradeV0(t *testing.T, ctx context.Context, r identityUpgradeResource, rawIdentity *tfprotov6.RawState) tftypes.Value {
	t.Helper()

	upgrader, ok := r.UpgradeIdentity(ctx)[0]
	if !ok {
		t.Fatal("expected an identity upgrader for version 0")
	}

	resp := resource.UpgradeIdentityResponse{
		Identity: &tfsdk.ResourceIdentity{Schema: currentIdentitySchema(t, ctx, r).IdentitySchema},
	}
	upgrader.IdentityUpgrader(ctx, resource.UpgradeIdentityRequest{RawIdentity: rawIdentity}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected identity upgrade diagnostics: %v", resp.Diagnostics)
	}

	return resp.Identity.Raw
}

func storedIdentityJSON(t *testing.T, identity map[string]any) *tfprotov6.RawState {
	t.Helper()
	b, err := json.Marshal(identity)
	if err != nil {
		t.Fatalf("unable to marshal identity: %v", err)
	}
	return &tfprotov6.RawState{JSON: b}
}

func TestIdentityUpgradeV0ToV1_UpgradesEveryResourceToV1Schema(t *testing.T) {
	ctx := context.Background()
	resources := map[string]identityUpgradeResource{
		"tfe_project":           &resourceTFEProject{config: ConfiguredClient{Hostname: testIdentityHostname}},
		"tfe_stack":             &resourceTFEStack{config: ConfiguredClient{Hostname: testIdentityHostname}},
		"tfe_variable":          &resourceTFEVariable{config: ConfiguredClient{Hostname: testIdentityHostname}},
		"tfe_registry_provider": &resourceTFERegistryProvider{config: ConfiguredClient{Hostname: testIdentityHostname}},
	}

	for name, r := range resources {
		t.Run(name, func(t *testing.T) {
			current := currentIdentitySchema(t, ctx, r).IdentitySchema
			if current.Version != 1 {
				t.Fatalf("expected identity schema version 1, got %d", current.Version)
			}

			stored := map[string]any{}
			for attribute := range current.Attributes {
				stored[attribute] = "stored-" + attribute
			}
			stored["hostname"] = "127.0.0.1:39553"

			got := runIdentityUpgradeV0(t, ctx, r, storedIdentityJSON(t, stored))

			// The upgraded identity must match the v1 schema, which only holds
			// while v0 and v1 have the same attributes.
			if !got.Type().Equal(current.Type().TerraformType(ctx)) {
				t.Fatalf("upgraded identity type %s does not match the v1 schema %s", got.Type(), current.Type().TerraformType(ctx))
			}

			var attributes map[string]tftypes.Value
			if err := got.As(&attributes); err != nil {
				t.Fatalf("unable to read upgraded identity: %v", err)
			}
			for attribute, value := range attributes {
				want := stored[attribute]
				if attribute == "hostname" {
					want = testIdentityHostname
				}
				if !value.Equal(tftypes.NewValue(tftypes.String, want)) {
					t.Fatalf("expected %s to be %q, got %s", attribute, want, value)
				}
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

			got := runIdentityUpgradeV0(t, ctx, r, storedIdentityJSON(t, map[string]any{
				"id":              "var-123",
				"configurable_id": "ws-123",
				"hostname":        priorHostname,
			}))

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

	got := runIdentityUpgradeV0(t, ctx, r, storedIdentityJSON(t, map[string]any{
		"id":       "prj-123",
		"hostname": "127.0.0.1:39553",
	}))

	assertProjectIdentity(t, ctx, r, got, "prj-123", "127.0.0.1:39553")
}

func TestIdentityUpgradeV0ToV1_PreservesIdentityWhenPriorIdentityIsNull(t *testing.T) {
	ctx := context.Background()
	r := &resourceTFEProject{config: ConfiguredClient{Hostname: testIdentityHostname}}

	got := runIdentityUpgradeV0(t, ctx, r, storedIdentityJSON(t, map[string]any{
		"id":       nil,
		"hostname": nil,
	}))

	if !got.IsFullyNull() {
		t.Fatalf("expected a fully null identity, got %s", got)
	}
}

// State written by older provider versions without identity support has no stored
// identity, but Terraform still requests the upgrade from version 0.
func TestIdentityUpgradeV0ToV1_ReturnsFullyNullIdentityWhenNoIdentityWasStored(t *testing.T) {
	ctx := context.Background()
	r := &resourceTFEProject{config: ConfiguredClient{Hostname: testIdentityHostname}}

	cases := map[string]*tfprotov6.RawState{
		"raw identity without data": {},
		"no raw identity":           nil,
	}

	for name, rawIdentity := range cases {
		t.Run(name, func(t *testing.T) {
			got := runIdentityUpgradeV0(t, ctx, r, rawIdentity)

			// A null object would be rejected by the framework, while a fully
			// null one is accepted and backfilled by Read.
			if got.IsNull() || !got.IsFullyNull() {
				t.Fatalf("expected an object with only null attributes, got %s", got)
			}
			if !got.Type().Equal(currentIdentitySchema(t, ctx, r).IdentitySchema.Type().TerraformType(ctx)) {
				t.Fatalf("expected the v1 identity type, got %s", got.Type())
			}
		})
	}
}

func assertProjectIdentity(t *testing.T, ctx context.Context, r *resourceTFEProject, got tftypes.Value, id, hostname string) {
	t.Helper()
	var identity modelProjectIdentity
	if diags := (&tfsdk.ResourceIdentity{Raw: got, Schema: currentIdentitySchema(t, ctx, r).IdentitySchema}).Get(ctx, &identity); diags.HasError() {
		t.Fatalf("unexpected diagnostics reading upgraded identity: %v", diags)
	}
	if identity.ID.ValueString() != id || identity.Hostname.ValueString() != hostname {
		t.Fatalf("expected identity {id: %q, hostname: %q}, got %+v", id, hostname, identity)
	}
}
