// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		Read: dataSourceTFEOAuthClientRead,
		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"oauth_client_id", "name", "service_provider"},
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"organization"},
			},
			"service_provider": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"organization"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
						string(tfe.ServiceProviderBitbucket),
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"callback_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"oauth_token_id": {
				Type:     schema.TypeString,
				Computed: true,
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

	switch len(oc.OAuthTokens) {
	case 0:
		d.Set("oauth_token_id", "")
	case 1:
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	default:
		return fmt.Errorf("unexpected number of OAuth tokens: %d", len(oc.OAuthTokens))
	}

	return nil
}
