// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOutputsDataSourceSchemaIncludesComputedHostname(t *testing.T) {
	t.Parallel()

	dataSource := &outputsDataSource{}
	response := &datasource.SchemaResponse{}
	dataSource.Schema(context.Background(), datasource.SchemaRequest{}, response)

	hostname, ok := response.Schema.Attributes["hostname"].(frameworkschema.StringAttribute)
	if !ok {
		t.Fatal("hostname is not a string attribute")
	}
	if !hostname.Computed {
		t.Fatal("hostname is not computed")
	}
}

func TestOutputsDataSourceStateFromOutputsIncludesConfiguredHostname(t *testing.T) {
	t.Parallel()

	state := outputsDataSourceStateFromOutputs(
		&tfe.Workspace{
			Name: "platform-core",
			Organization: &tfe.Organization{
				Name: "hashicorp",
			},
		},
		"tfcdev-1d567332.ngrok.app",
		types.DynamicNull(),
		types.DynamicNull(),
	)

	if actual := state.Hostname.ValueString(); actual != "tfcdev-1d567332.ngrok.app" {
		t.Fatalf("hostname = %q, want %q", actual, "tfcdev-1d567332.ngrok.app")
	}
	if actual := state.Organization.ValueString(); actual != "hashicorp" {
		t.Fatalf("organization = %q, want %q", actual, "hashicorp")
	}
	if actual := state.Workspace.ValueString(); actual != "platform-core" {
		t.Fatalf("workspace = %q, want %q", actual, "platform-core")
	}
}
