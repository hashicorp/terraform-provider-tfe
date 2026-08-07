// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFERunTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "Manages run triggers." +
			"\n\nHCP Terraform provides a way to connect your workspace to one or more workspaces within your organization, known as \"source workspaces\". These connections, called run triggers, allow runs to queue automatically in your workspace on successful apply of runs in any of the source workspaces. You can connect your workspace to up to 20 source workspaces.",

		Create: resourceTFERunTriggerCreate,
		Read:   resourceTFERunTriggerRead,
		Delete: resourceTFERunTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the run trigger.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"workspace_id": {
				Description: "The ID of the workspace that owns the run trigger. This is the workspace where runs will be triggered.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"sourceable_id": {
				Description: "The ID of the sourceable. The sourceable must be a workspace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

// newRunTriggerCreateEnvelope builds the request body for creating a run trigger. The sourceable
// relationship is the only writable part of a run trigger; everything else is server-computed.
func newRunTriggerCreateEnvelope(sourceableID string) *models.RunTriggersEnvelope {
	sourceableData := models.NewWorkspacesHasOne_data()
	sourceableData.SetId(&sourceableID)
	wsType := models.WORKSPACES_WORKSPACESIDENTIFIER_TYPE
	sourceableData.SetTypeEscaped(&wsType)
	sourceable := models.NewWorkspacesHasOne()
	sourceable.SetData(sourceableData)

	relationships := models.NewRunTriggers_relationships()
	relationships.SetSourceable(sourceable)

	data := models.NewRunTriggers()
	data.SetRelationships(relationships)
	rtType := models.RUNTRIGGERS_RUNTRIGGERS_TYPE
	data.SetTypeEscaped(&rtType)

	envelope := models.NewRunTriggersEnvelope()
	envelope.SetData(data)
	return envelope
}

// isRunTriggerCreationLocked reports whether err represents the API's "Run Trigger creation
// locked" response. The generated client's err.Error() only ever contains the generic HTTP status
// text (e.g. "409 Conflict"); the JSON:API error detail text is only reachable via *tfe.APIError.Details.
func isRunTriggerCreationLocked(err error) bool {
	var apiErr *tfe.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, detail := range apiErr.Details {
		if strings.Contains(detail, "Run Trigger creation locked") {
			return true
		}
	}
	return false
}

func resourceTFERunTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get attributes
	workspaceID := d.Get("workspace_id").(string)
	sourceableID := d.Get("sourceable_id").(string)

	envelope := newRunTriggerCreateEnvelope(sourceableID)

	log.Printf("[DEBUG] Create run trigger on workspace %s with sourceable %s", workspaceID, sourceableID)
	err := retry.Retry(1*time.Minute, func() *retry.RetryError {
		runTriggerEnvelope, err := config.ClientV2.API.Workspaces().ByWorkspace_id(workspaceID).RunTriggers().Post(ctx, envelope, nil)
		if err == nil {
			if runTriggerEnvelope == nil || runTriggerEnvelope.GetData() == nil || runTriggerEnvelope.GetData().GetId() == nil {
				return retry.NonRetryableError(errors.New("no run trigger data was returned by the API"))
			}
			d.SetId(*runTriggerEnvelope.GetData().GetId())
			return nil
		}

		if isRunTriggerCreationLocked(err) {
			log.Printf("[DEBUG] Run triggers are locked for workspace %s, will retry", workspaceID)
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})

	if err != nil {
		return fmt.Errorf("Error creating run trigger on workspace %s with sourceable %s: %w", workspaceID, sourceableID, err)
	}

	return resourceTFERunTriggerRead(d, meta)
}

func resourceTFERunTriggerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read run trigger: %s", d.Id())
	runTriggerEnvelope, err := config.ClientV2.API.RunTriggers().ById(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] run trigger %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading run trigger %s: %w", d.Id(), err)
	}
	if runTriggerEnvelope == nil || runTriggerEnvelope.GetData() == nil {
		log.Printf("[DEBUG] run trigger %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	relationships := runTriggerEnvelope.GetData().GetRelationships()
	d.Set("workspace_id", runTriggerWorkspaceID(relationships))
	d.Set("sourceable_id", runTriggerSourceableID(relationships))

	return nil
}

func runTriggerWorkspaceID(relationships models.RunTriggers_relationshipsable) string {
	if relationships == nil || relationships.GetWorkspace() == nil || relationships.GetWorkspace().GetData() == nil {
		return ""
	}
	return valueOrZero(relationships.GetWorkspace().GetData().GetId())
}

func runTriggerSourceableID(relationships models.RunTriggers_relationshipsable) string {
	if relationships == nil || relationships.GetSourceable() == nil || relationships.GetSourceable().GetData() == nil {
		return ""
	}
	return valueOrZero(relationships.GetSourceable().GetData().GetId())
}

func resourceTFERunTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete run trigger: %s", d.Id())
	err := config.ClientV2.API.RunTriggers().ById(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting run trigger %s: %w", d.Id(), err)
	}

	return nil
}
