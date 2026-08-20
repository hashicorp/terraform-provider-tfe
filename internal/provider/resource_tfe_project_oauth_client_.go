// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEProjectOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Adds and removes OAuth clients from a project.",
		Create:      resourceTFEProjectOauthClientCreate,
		Read:        resourceTFEProjectOauthClientRead,
		Delete:      resourceTFEProjectOauthClientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEProjectOauthClientImporter,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the oauth client attachment. ID format: `<project-id>_<oauth-client-id>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"oauth_client_id": {
				Description: "The ID of the OAuth client to attach to the project.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"project_id": {
				Description: "The ID of the project to attach the OAuth client to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTFEProjectOauthClientCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	oauthClientID := d.Get("oauth_client_id").(string)
	projectID := d.Get("project_id").(string)

	body := makeProjectIdentifierArrayDocument([]interface{}{projectID})
	err := config.ClientV2.API.OauthClients().ByOauth_client_id(oauthClientID).Relationships().Projects().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"error attaching oauth client id %s to project %s: %w", oauthClientID, projectID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", projectID, oauthClientID))

	return resourceTFEProjectOauthClientRead(d, meta)
}

func resourceTFEProjectOauthClientRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	oauthClientID := d.Get("oauth_client_id").(string)
	projectID := d.Get("project_id").(string)

	log.Printf("[DEBUG] Read configuration of project oauth client: %s", oauthClientID)
	ocEnv, err := config.ClientV2.API.OauthClients().ByOauth_client_id(oauthClientID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] Oauth client %s no longer exists", oauthClientID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of oauth client %s: %w", oauthClientID, err)
	}

	oc := ocEnv.GetData()
	rels := oc.GetRelationships()

	isProjectAttached := false
	if rels != nil && rels.GetProjects() != nil {
		for _, proj := range rels.GetProjects().GetData() {
			if valueOrZero(proj.GetId()) == projectID {
				isProjectAttached = true
				d.Set("project_id", projectID)
				break
			}
		}
	}

	if !isProjectAttached {
		log.Printf("[DEBUG] Project %s not attached to oauth client %s. Removing from state.", projectID, oauthClientID)
		d.SetId("")
		return nil
	}

	d.Set("oauth_client_id", oauthClientID)
	return nil
}

func resourceTFEProjectOauthClientDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	oauthClientID := d.Get("oauth_client_id").(string)
	projectID := d.Get("project_id").(string)

	log.Printf("[DEBUG] Detaching project (%s) from oauth client (%s)", projectID, oauthClientID)
	body := makeProjectIdentifierArrayDocument([]interface{}{projectID})
	err := config.ClientV2.API.OauthClients().ByOauth_client_id(oauthClientID).Relationships().Projects().Delete(ctx, body, nil)
	if err != nil {
		return fmt.Errorf(
			"error detaching project %s from oauth client %s: %w", projectID, oauthClientID, err)
	}

	return nil
}

func resourceTFEProjectOauthClientImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The format of the import ID is <ORGANIZATION/PROJECT ID/OAUTHCLIENT NAME>
	splitID := strings.SplitN(d.Id(), "/", 3)
	if len(splitID) != 3 {
		return nil, fmt.Errorf(
			"invalid project oauth client input format: %s (expected <ORGANIZATION>/<PROJECT ID>/<OAUTHCLIENT NAME>)",
			splitID,
		)
	}

	organization, projectID, oauthClientName := splitID[0], splitID[1], splitID[2]

	config := meta.(ConfiguredClient)

	// Ensure the named project exists before fetching all the oauth clients in the org
	_, err := config.ClientV2.API.Projects().ByProject_id(projectID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of project %s in organization %s: %w", projectID, organization, err)
	}

	pageSize := int32(100)
	queryParams := &organizations.ItemOauthClientsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}
	for {
		list, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).OauthClients().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, fmt.Errorf("error retrieving organization's list of oauth clients: %w", err)
		}
		for _, oauthClient := range list.GetData() {
			ocAttrs := oauthClient.GetAttributes()
			if ocAttrs == nil || valueOrZero(ocAttrs.GetName()) != oauthClientName {
				continue
			}

			rels := oauthClient.GetRelationships()
			if rels == nil || rels.GetProjects() == nil {
				continue
			}
			for _, proj := range rels.GetProjects().GetData() {
				if valueOrZero(proj.GetId()) != projectID {
					continue
				}

				d.Set("project_id", projectID)
				d.Set("oauth_client_id", valueOrZero(oauthClient.GetId()))
				d.SetId(fmt.Sprintf("%s_%s", projectID, valueOrZero(oauthClient.GetId())))
				return []*schema.ResourceData{d}, nil
			}
		}

		nextPage := nextPageFromMeta(list.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}

	return nil, fmt.Errorf("project %s has not been assigned to oauth client %s", projectID, oauthClientName)
}
