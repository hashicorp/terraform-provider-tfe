// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"time"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information on a team.",

		Read: dataSourceTFETeamRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the team.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the team.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"sso_team_id": {
				Description: "The [SSO Team ID](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) of the team, if it has been defined.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"scim_linked": {
				Description: "Whether the team is linked to a SCIM group. Only populated when SCIM is enabled on the TFE instance. If not present, SCIM is not supported or not enabled on the TFE instance.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"scim_group_name": {
				Description: "The display name of the SCIM group linked to this team. Only populated when SCIM is enabled on the TFE instance. If not present, SCIM is not supported or not enabled on the TFE instance, or the team is not linked to a SCIM group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"scim_sync_paused": {
				Description: "Whether SCIM membership sync is paused for this team. Only populated when SCIM is enabled on the TFE instance. If not present, SCIM is not supported or not enabled on the TFE instance.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"scim_updated_at": {
				Description: "The timestamp of the last SCIM reconciliation for this team, in RFC3339 format. Only populated when SCIM is enabled on the TFE instance. If not present, SCIM is not supported or not enabled on the TFE instance, or the team is not linked to a SCIM group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	team, err := fetchTeamByNameV2(ctx, config.ClientV2.API, organization, name)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			return fmt.Errorf("could not find team %s/%s", organization, name)
		}
		return fmt.Errorf("Error retrieving teams: %w", err)
	}

	setTeamResourceData(d, team)
	return nil
}

// setTeamResourceData populates state with the team's attributes. SCIM fields are
// guarded by nil checks so that older TFE instances do not panic.
func setTeamResourceData(d *schema.ResourceData, team models.Teamsable) {
	d.SetId(valueOrZero(team.GetId()))

	attrs := team.GetAttributes()
	if attrs == nil {
		return
	}

	d.Set("sso_team_id", valueOrZero(attrs.GetSsoTeamId()))

	if v := attrs.GetScimLinked(); v != nil {
		d.Set("scim_linked", *v)
	}
	if v := attrs.GetScimGroupName(); v != nil {
		d.Set("scim_group_name", *v)
	}
	if v := attrs.GetScimSyncPaused(); v != nil {
		d.Set("scim_sync_paused", *v)
	}
	if v := attrs.GetScimUpdatedAt(); v != nil {
		d.Set("scim_updated_at", v.Format(time.RFC3339))
	}
}
