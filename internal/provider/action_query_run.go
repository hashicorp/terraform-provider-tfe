// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ action.Action                   = &actionTFEQueryRun{}
	_ action.ActionWithConfigure      = &actionTFEQueryRun{}
	_ action.ActionWithValidateConfig = &actionTFEQueryRun{}
)

func NewQueryRunAction() action.Action {
	return &actionTFEQueryRun{}
}

type actionTFEQueryRun struct {
	config ConfiguredClient
}

type actionTFEQueryRunModel struct {
	ConfigurationVersionID types.String `tfsdk:"configuration_version_id"`
	Variables              types.Map    `tfsdk:"variables"`
	WaitForLatestConfig    types.Bool   `tfsdk:"wait_for_latest_configuration"`
	WorkspaceID            types.String `tfsdk:"workspace_id"`
}

func (a *actionTFEQueryRun) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected action Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on Github.", req.ProviderData),
		)
	}
	a.config = client
}

func (a *actionTFEQueryRun) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query_run"
}

func (a *actionTFEQueryRun) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"configuration_version_id": schema.StringAttribute{
				Optional: true,
			},
			"variables": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"wait_for_latest_configuration": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (a *actionTFEQueryRun) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	var data actionTFEQueryRunModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ConfigurationVersionID.IsNull() && !data.WaitForLatestConfig.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("configuration_version_id"),
			"Configuration Version ID is required",
			"Expected a configuration_version_id to be set or wait_for_latest_configuration set to true",
		)
		return
	}

	// If a config version is specified and wait_for_latest_configuration is also specified
	// throw a diagnostic warning.
	if !data.ConfigurationVersionID.IsNull() && data.WaitForLatestConfig.ValueBool() {
		resp.Diagnostics.AddAttributeWarning(path.Root("wait_for_latest_configuration"),
			"Attribute is ignored",
			"A configuration_version_id is specified, ignoring wait_for_latest_configuration",
		)
	}
}

func (a *actionTFEQueryRun) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data actionTFEQueryRunModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := a.config.Client
	workspaceID := data.WorkspaceID.ValueString()
	var configVersionID string

	// Extracting the nested block into a helper to satisfy the nestif linter
	if data.WaitForLatestConfig.ValueBool() {
		id, ok := a.waitForLatestConfigVersion(ctx, client, workspaceID, resp)
		if !ok {
			return // Diagnostics were already appended by the helper
		}
		configVersionID = id
	} else {
		configVersionID = data.ConfigurationVersionID.ValueString()
	}

	resp.SendProgress(action.InvokeProgressEvent{Message: "Creating Query Run..."})

	var variableList []*tfe.RunVariable
	if !data.Variables.IsNull() {
		var vars map[string]string
		resp.Diagnostics.Append(data.Variables.ElementsAs(ctx, &vars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range vars {
			variableList = append(variableList, &tfe.RunVariable{
				Key:   k,
				Value: v,
			})
		}
	}

	createOpts := tfe.QueryRunCreateOptions{
		Workspace: &tfe.Workspace{ID: workspaceID},
		ConfigurationVersion: &tfe.ConfigurationVersion{
			ID: configVersionID,
		},
		Variables: variableList,
		Source:    tfe.QueryRunSourceAPI,
	}

	run, err := client.QueryRuns.Create(ctx, createOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating query run", err.Error())
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Query run %s created. Status: %s", run.ID, run.Status),
	})

	pollTicker := time.NewTicker(3 * time.Second)
	defer pollTicker.Stop()
	lastStatus := run.Status

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context cancelled", "Context cancelled while waiting for query run.")
			return
		case <-pollTicker.C:
			// Refresh run status
			currentRun, err := client.QueryRuns.Read(ctx, run.ID)
			if err != nil {
				resp.Diagnostics.AddError("Error reading query run status", err.Error())
				return
			}

			// If status changed, notify Terraform
			if currentRun.Status != lastStatus {
				resp.SendProgress(action.InvokeProgressEvent{
					Message: fmt.Sprintf("Query run %s status: %s", run.ID, currentRun.Status),
				})
				lastStatus = currentRun.Status
			}

			// Check for terminal states
			switch currentRun.Status {
			case tfe.QueryRunFinished:
				resp.SendProgress(action.InvokeProgressEvent{
					Message: "Query run finished successfully.",
				})
				return

			case tfe.QueryRunCanceled:
				canceledByID := "unknown"
				// Prevent a panic, just in case
				if currentRun.CanceledBy != nil {
					canceledByID = currentRun.CanceledBy.ID
				}

				resp.Diagnostics.AddError(
					"Query run canceled",
					fmt.Sprintf("Query run was canceled by %s", canceledByID),
				)
				return // Need to explicitly return to break the loop on cancel
			case tfe.QueryRunErrored:
				resp.Diagnostics.AddError(
					"Query run errored",
					fmt.Sprintf("Query finished with an error, view the query on %s for details", workspaceID),
				)
				return
			}
		}
	}
}

// waitForLatestConfigVersion polls for the latest uploaded configuration version and returns its ID.
func (a *actionTFEQueryRun) waitForLatestConfigVersion(ctx context.Context, client *tfe.Client, workspaceID string, resp *action.InvokeResponse) (string, bool) {
	resp.SendProgress(action.InvokeProgressEvent{Message: "Fetching latest configuration version..."})

	listOpts := &tfe.ConfigurationVersionListOptions{
		ListOptions: tfe.ListOptions{PageSize: 1},
	}
	cvList, err := client.ConfigurationVersions.List(ctx, workspaceID, listOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error listing configuration versions", err.Error())
		return "", false
	}

	if len(cvList.Items) == 0 {
		resp.Diagnostics.AddError("No configuration versions found", "Workspace has no configuration versions.")
		return "", false
	}

	latestCV := cvList.Items[0]
	configVersionID := latestCV.ID

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeout := time.After(20 * time.Minute)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Waiting for configuration version %s to be uploaded...", configVersionID),
	})

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Context cancelled", "Context cancelled while waiting for configuration version.")
			return "", false
		case <-timeout:
			resp.Diagnostics.AddError("Timeout", "Timed out waiting for configuration version to become uploaded.")
			return "", false
		case <-ticker.C:
			cv, err := client.ConfigurationVersions.Read(ctx, configVersionID)
			if err != nil {
				resp.Diagnostics.AddError("Error reading configuration version", err.Error())
				return "", false
			}

			if cv.Status == tfe.ConfigurationUploaded {
				return configVersionID, true
			}

			if cv.Status == tfe.ConfigurationErrored || cv.Status == tfe.ConfigurationArchived {
				resp.Diagnostics.AddError("Error creating query run", fmt.Sprintf("Can not use configuration version %s with status %s", cv.ID, cv.Status))
				return "", false
			}
		}
	}
}
