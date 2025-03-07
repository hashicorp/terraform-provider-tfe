// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource = &AgentTokenEphemeralResource{}
)

func NewAgentTokenEphemeralResource() ephemeral.EphemeralResource {
	return &AgentTokenEphemeralResource{}
}

type AgentTokenEphemeralResource struct {
	config ConfiguredClient
}

type AgentTokenEphemeralResourceModel struct {
	AgentPoolID types.String `tfsdk:"agent_pool_id"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
}

// defines a schema describing what data is available in the ephemeral resource's configuration and result data.
func (e *AgentTokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This ephemeral resource can be used to retrieve an agent token without saving its value in state.",
		Attributes: map[string]schema.Attribute{
			"agent_pool_id": schema.StringAttribute{
				Description: `ID of the agent. If omitted, agent must be defined in the provider config.`,
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: `Description of the agent token.`,
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: `The generated token.`,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (e *AgentTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)

		return
	}
	e.config = client
}

func (e *AgentTokenEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_token" // tfe_agent_token
}

// The request contains the configuration supplied to Terraform for the ephemeral resource. The response contains the ephemeral result data. The data is defined by the schema of the ephemeral resource.
func (e *AgentTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data AgentTokenEphemeralResourceModel

	// Read Terraform config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agentPoolID := data.AgentPoolID.ValueString()
	description := data.Description.ValueString()

	options := tfe.AgentTokenCreateOptions{
		Description: tfe.String(description),
	}

	log.Printf("[DEBUG] Create new agent token for agent pool ID: %s", agentPoolID)
	log.Printf("[DEBUG] Create new agent token with description: %s", description)

	result, err := e.config.Client.AgentTokens.Create(ctx, agentPoolID, options)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create agent token", err.Error())
		return
	}

	data = ephemeralResourceModelFromTFEagentToken(agentPoolID, result)

	// Save to ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

// ephemeralResourceModelFromTFEagentToken builds a agentTokenEphemeralResourceModel struct from a
// tfe.agentToken value.
func ephemeralResourceModelFromTFEagentToken(id string, v *tfe.AgentToken) AgentTokenEphemeralResourceModel {
	return AgentTokenEphemeralResourceModel{
		AgentPoolID: types.StringValue(id),
		Description: types.StringValue(v.Description),
		Token:       types.StringValue(v.Token),
	}
}
