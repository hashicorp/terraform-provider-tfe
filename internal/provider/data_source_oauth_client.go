// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-tfe"
	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an OAuth client.",
		Read:        dataSourceTFEOAuthClientRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The OAuth client ID. This will match `oauth_client_id`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

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
				Description:  "Name of the OAuth client (may be `null`).",
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
				Description: "Whether or not the OAuth client is scoped to all workspaces and projects in the organization.",
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

	var ocID string

	switch v, ok := d.GetOk("oauth_client_id"); {
	case ok:
		ocID = v.(string)
	default:
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

		id, err := fetchOAuthClientByNameOrServiceProvider(ctx, config, organization, name, serviceProvider)
		if err != nil {
			return err
		}
		ocID = id
	}

	ocEnv, err := config.ClientV2.API.OauthClients().ByOauth_client_id(ocID).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			return fmt.Errorf("OAuth client %s not found: %w", ocID, err)
		}
		return fmt.Errorf("Error retrieving OAuth client: %w", err)
	}

	oc := ocEnv.GetData()
	attrs := oc.GetAttributes()
	rels := oc.GetRelationships()

	d.SetId(ocID)
	d.Set("oauth_client_id", ocID)

	if attrs != nil {
		d.Set("api_url", valueOrZero(attrs.GetApiUrl()))
		d.Set("callback_url", valueOrZero(attrs.GetCallbackUrl()))
		if attrs.GetCreatedAt() != nil {
			d.Set("created_at", attrs.GetCreatedAt().Format(time.RFC3339))
		}
		d.Set("http_url", valueOrZero(attrs.GetHttpUrl()))
		if n := attrs.GetName(); n != nil {
			d.Set("name", *n)
		}
		d.Set("service_provider", valueOrZero(attrs.GetServiceProvider()))
		d.Set("service_provider_display_name", valueOrZero(attrs.GetServiceProviderDisplayName()))
		d.Set("organization_scoped", attrs.GetOrganizationScoped() != nil && *attrs.GetOrganizationScoped())
	}

	if rels != nil {
		if rels.GetOauthTokens() != nil {
			tokens := rels.GetOauthTokens().GetData()
			switch len(tokens) {
			case 0:
				d.Set("oauth_token_id", "")
			case 1:
				d.Set("oauth_token_id", valueOrZero(tokens[0].GetId()))
			default:
				return fmt.Errorf("unexpected number of OAuth tokens: %d", len(tokens))
			}
		}

		if rels.GetProjects() != nil {
			var projectIDs []interface{}
			for _, proj := range rels.GetProjects().GetData() {
				projectIDs = append(projectIDs, valueOrZero(proj.GetId()))
			}
			d.Set("project_ids", projectIDs)
		}
	}

	return nil
}
