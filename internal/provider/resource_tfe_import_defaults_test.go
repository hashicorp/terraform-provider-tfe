// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Only the workspace lookup is implemented. Unexpected API operations fail
// through the nil embedded interface rather than making a request.
type importDefaultsReadOnlyWorkspaces struct {
	tfe.Workspaces
	id    string
	reads int
}

func (w *importDefaultsReadOnlyWorkspaces) Read(context.Context, string, string) (*tfe.Workspace, error) {
	w.reads++
	return &tfe.Workspace{ID: w.id}, nil
}

func TestResourceTFEVariableImportState_DefaultBaseline(t *testing.T) {
	ctx := context.Background()
	for _, container := range []string{"ws-abcdefghijklmnop", "varset-abcdefghijklmnop"} {
		for _, byIdentity := range []bool{false, true} {
			name := container
			if byIdentity {
				name += "/identity"
			} else {
				name += "/legacy"
			}
			t.Run(name, func(t *testing.T) {
				workspaces := &importDefaultsReadOnlyWorkspaces{id: container}
				r := &resourceTFEVariable{config: ConfiguredClient{Client: &tfe.Client{Workspaces: workspaces}}}
				schemaResp := &fwresource.SchemaResponse{}
				r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
				req := fwresource.ImportStateRequest{}
				if byIdentity {
					identitySchema := &fwresource.IdentitySchemaResponse{}
					r.IdentitySchema(ctx, fwresource.IdentitySchemaRequest{}, identitySchema)
					req.Identity = &tfsdk.ResourceIdentity{Schema: identitySchema.IdentitySchema}
					if diags := req.Identity.Set(ctx, &modelTFEVariableIdentity{
						ID:             types.StringValue("var-abcdefghijklmnop"),
						ConfigurableID: types.StringValue(container),
						Hostname:       types.StringValue("example.invalid"),
					}); diags.HasError() {
						t.Fatal(diags)
					}
				} else {
					req.ID = "invented/" + container + "/var-abcdefghijklmnop"
				}
				resp := fwresource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
				r.ImportState(ctx, req, &resp)
				if resp.Diagnostics.HasError() {
					t.Fatal(resp.Diagnostics)
				}
				wantReads := 0
				if !byIdentity && container[:3] == "ws-" {
					wantReads = 1
				}
				if workspaces.reads != wantReads {
					t.Fatalf("unexpected API read count: %d", workspaces.reads)
				}
				var state modelTFEVariable
				if diags := resp.State.Get(ctx, &state); diags.HasError() {
					t.Fatal(diags)
				}
				if state.Value.IsNull() || state.Value.IsUnknown() || state.Value.ValueString() != "" {
					t.Fatalf("import must establish the known empty schema default, got %s", state.Value)
				}

				// A masked API read must carry this baseline forward, without copying an
				// API value into state. Non-sensitive reads must still use the API value.
				for _, sensitive := range []bool{false, true} {
					var read modelTFEVariable
					if byVariableSet := !state.VariableSetID.IsNull(); byVariableSet {
						read = modelFromTFEVariableSetVariable(tfe.VariableSetVariable{
							ID: "var-abcdefghijklmnop", Value: "invented-api-value", Sensitive: sensitive,
							VariableSet: &tfe.VariableSet{ID: container},
						}, state.Value, types.Int64Null())
					} else {
						read = modelFromTFEVariable(tfe.Variable{
							ID: "var-abcdefghijklmnop", Value: "invented-api-value", Sensitive: sensitive,
							Workspace: &tfe.Workspace{ID: container},
						}, state.Value, types.Int64Null())
					}
					if sensitive {
						if !read.Value.Equal(state.Value) || !read.ReadableValue.IsNull() {
							t.Fatal("masked read changed the imported baseline or exposed a value")
						}
						planned := read
						planned.Description = types.StringValue("changed metadata")
						if value := r.determineValueForUpdate(planned, read, modelTFEVariable{}); value != nil {
							t.Fatal("metadata update would send a value")
						}
					} else if read.Value.ValueString() != "invented-api-value" || read.ReadableValue.ValueString() != "invented-api-value" {
						t.Fatal("non-sensitive read lost its API value")
					}
				}
			})
		}
	}
}

type importDefaultsGlobalWorkspace struct {
	tfe.Workspaces
	reads int
}

func (w *importDefaultsGlobalWorkspace) ReadByIDWithOptions(context.Context, string, *tfe.WorkspaceReadOptions) (*tfe.Workspace, error) {
	w.reads++
	return &tfe.Workspace{
		ID: "ws-abcdefghijklmnop", Name: "invented", GlobalRemoteState: true,
		Organization:     &tfe.Organization{Name: "invented"},
		TerraformVersion: "1.15.8", ExecutionMode: "remote",
	}, nil
}

func TestResourceTFEWorkspaceRead_GlobalRemoteStateConsumersKnown(t *testing.T) {
	for _, existingConsumers := range []bool{false, true} {
		name := "unset imported consumers"
		if existingConsumers {
			name = "existing explicit consumers preserved"
		}
		t.Run(name, func(t *testing.T) {
			r := resourceTFEWorkspace()
			d := r.Data(&terraform.InstanceState{ID: "ws-abcdefghijklmnop", Attributes: map[string]string{"id": "ws-abcdefghijklmnop"}})
			if existingConsumers {
				if err := d.Set("remote_state_consumer_ids", []string{"ws-ponmlkjihgfedcba"}); err != nil {
					t.Fatal(err)
				}
			}
			workspaces := &importDefaultsGlobalWorkspace{}
			if err := resourceTFEWorkspaceRead(d, ConfiguredClient{Client: &tfe.Client{Workspaces: workspaces}, Hostname: "example.invalid"}); err != nil {
				t.Fatal(err)
			}
			if workspaces.reads != 1 {
				t.Fatalf("expected one read, got %d", workspaces.reads)
			}
			attrs := d.State().Attributes
			if attrs["global_remote_state"] != "true" {
				t.Fatal("global sharing must preserve its grant")
			}
			if existingConsumers {
				consumers := d.Get("remote_state_consumer_ids").(*schema.Set).List()
				if len(consumers) != 1 || consumers[0] != "ws-ponmlkjihgfedcba" {
					t.Fatal("global sharing must preserve existing explicit consumers in state")
				}
			} else if attrs["remote_state_consumer_ids.#"] != "0" {
				t.Fatal("import must establish a known empty computed consumer set")
			}
		})
	}
}
