// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an OAuth client.",
		Read:        dataSourceTFEOAuthClientRead,
		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Description:  "ID of the OAuth client.",
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"oauth_client_id", "name", "service_provider"},
			},
			"organization": {
				Description: "The name of the organization in which to search.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description:  "Name of the OAuth client.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"organization"},
			},
			"service_provider": {
				Description:  "The API identifier of the OAuth service provider. If set, must be one of: `ado_server`, `ado_services`, `bitbucket_data_center`,  `bitbucket_hosted`, `bitbucket_server`(deprecated), `github`, `github_enterprise`, `gitlab_hosted`, `gitlab_community_edition`, or `gitlab_enterprise_edition`.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"organization"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderBitbucketDataCenter),
						string(tfe.ServiceProviderBitbucketServer),
						string(tfe.ServiceProviderBitbucketServerLegacy),
						string(tfe.ServiceProviderGithub),
						string(tfe.ServiceProviderGithubEE),
						string(tfe.ServiceProviderGitlab),
						string(tfe.ServiceProviderGitlabCE),
						string(tfe.ServiceProviderGitlabEE),
					},
					false,
				)),
			},
			"service_provider_display_name": {
				Description: "The display name of the OAuth service provider.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"api_url": {
				Description: "The client's API URL.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"callback_url": {
				Description: "OAuth callback URL to provide to the OAuth service provider.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The date and time this OAuth client was created in RFC3339 format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"http_url": {
				Description: "The client's HTTP URL.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"oauth_token_id": {
				Description: "The ID of the OAuth token associated with the OAuth client.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization_scoped": {
				Description: "Whether or not the agent pool can be used by all workspaces and projects in the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"project_ids": {
				Description: "IDs of the projects that use the oauth client.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataSourceTFEOAuthClientRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	config := meta.(ConfiguredClient)

	var oc *tfe.OAuthClient
	var err error

	switch v, ok := d.GetOk("oauth_client_id"); {
	case ok:
		oc, err = config.Client.OAuthClients.Read(ctx, v.(string))
		if err != nil {
			return fmt.Errorf("Error retrieving OAuth client: %w", err)
		}
	default:
		// search by name or service provider within a specific organization instead
		organization, err := config.schemaOrDefaultOrganization(d)
		if err != nil {
			return err
		}

		var name string
		var serviceProvider tfe.ServiceProviderType
		vName, ok := d.GetOk("name")
		if ok {
			name = vName.(string)
		}
		vServiceProvider, ok := d.GetOk("service_provider")
		if ok {
			serviceProvider = tfe.ServiceProviderType(vServiceProvider.(string))
		}

		oc, err = fetchOAuthClientByNameOrServiceProvider(ctx, config.Client, organization, name, serviceProvider)
		if err != nil {
			return err
		}
	}

	d.SetId(oc.ID)
	d.Set("oauth_client_id", oc.ID)
	d.Set("api_url", oc.APIURL)
	d.Set("callback_url", oc.CallbackURL)
	d.Set("created_at", oc.CreatedAt.Format(time.RFC3339))
	d.Set("http_url", oc.HTTPURL)
	if oc.Name != nil {
		d.Set("name", *oc.Name)
	}
	d.Set("service_provider", oc.ServiceProvider)
	d.Set("service_provider_display_name", oc.ServiceProviderName)
	d.Set("organization_scoped", oc.OrganizationScoped)

	switch len(oc.OAuthTokens) {
	case 0:
		d.Set("oauth_token_id", "")
	case 1:
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	default:
		return fmt.Errorf("unexpected number of OAuth tokens: %d", len(oc.OAuthTokens))
	}

	var projectIDs []interface{}
	for _, project := range oc.Projects {
		projectIDs = append(projectIDs, project.ID)
	}
	d.Set("project_ids", projectIDs)

	return nil
}
