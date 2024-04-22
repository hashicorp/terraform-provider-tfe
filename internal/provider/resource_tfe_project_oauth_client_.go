// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEProjectOAuthClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEProjectOauthClientCreate,
		Read:   resourceTFEProjectOauthClientRead,
		Delete: resourceTFEProjectOauthClientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEProjectOauthClientImporter,
		},

		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEProjectOauthClientCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	oauthClientID := d.Get("oauth_client_id").(string)
	projectID := d.Get("project_id").(string)

	oauthClientAddProjectsOptions := tfe.OAuthClientAddProjectsOptions{}
	oauthClientAddProjectsOptions.Projects = append(oauthClientAddProjectsOptions.Projects, &tfe.Project{ID: projectID})

	err := config.Client.OAuthClients.AddProjects(ctx, oauthClientID, oauthClientAddProjectsOptions)
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
	oauthClient, err := config.Client.OAuthClients.ReadWithOptions(ctx, oauthClientID, &tfe.OAuthClientReadOptions{
		Include: []tfe.OAuthClientIncludeOpt{tfe.OauthClientProjects},
	})
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Oauth client %s no longer exists", oauthClientID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading configuration of oauth client %s: %w", oauthClientID, err)
	}

	isProjectAttached := false
	for _, project := range oauthClient.Projects {
		if project.ID == projectID {
			isProjectAttached = true
			d.Set("project_id", projectID)
			break
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
	oauthClientRemoveProjectsOptions := tfe.OAuthClientRemoveProjectsOptions{}
	oauthClientRemoveProjectsOptions.Projects = append(oauthClientRemoveProjectsOptions.Projects, &tfe.Project{ID: projectID})

	err := config.Client.OAuthClients.RemoveProjects(ctx, oauthClientID, oauthClientRemoveProjectsOptions)
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
	_, err := config.Client.Projects.Read(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration of project %s in organization %s: %w", projectID, organization, err)
	}

	options := &tfe.OAuthClientListOptions{Include: []tfe.OAuthClientIncludeOpt{tfe.OauthClientProjects}}
	for {
		list, err := config.Client.OAuthClients.List(ctx, organization, options)
		if err != nil {
			return nil, fmt.Errorf("error retrieving organization's list of oauth clients: %w", err)
		}
		for _, oauthClient := range list.Items {
			if *oauthClient.Name != oauthClientName {
				continue
			}

			for _, project := range oauthClient.Projects {
				if project.ID != projectID {
					continue
				}

				d.Set("project_id", project.ID)
				d.Set("oauth_client_id", oauthClient.ID)
				d.SetId(fmt.Sprintf("%s_%s", project.ID, oauthClient.ID))

				return []*schema.ResourceData{d}, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if list.CurrentPage >= list.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = list.NextPage
	}

	return nil, fmt.Errorf("project %s has not been assigned to oauth client %s", projectID, oauthClientName)
}
