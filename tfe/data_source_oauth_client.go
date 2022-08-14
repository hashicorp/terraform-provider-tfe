package tfe

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
	tfeClient := meta.(*tfe.Client)

	var oc *tfe.OAuthClient
	var err error

	if v, ok := d.GetOk("oauth_client_id"); ok {
		oc, err = tfeClient.OAuthClients.Read(ctx, v.(string))
		if err != nil {
			return fmt.Errorf("Error retrieving OAuth client: %w", err)
		}
	} else {
		// search by name or service provider within a specific organization instead
		organization := d.Get("organization").(string)

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

		// Paginate through all OAuthClients in the organization; if multiple pages
		// of results are returned by the API, use the options variable to increment
		// the page number until all results have been retrieved.
		//
		// Within the pagination loop, loop again through each result on each page.
		// If 'name' was set, then match against the 'Name' field. If 'service_provider'
		// was set, then match against the 'ServiceProvider' field. If both are set,
		// then both must match. All matches are added to the ocMatches slice.
		//
		// At the end of the loop, if zero or more than one matches were found, an
		// error is returned. Otherwise, only one match was found, and execution
		// proceeds beyond the loop to build the resource object.
		//
		var ocMatches []*tfe.OAuthClient
		options := &tfe.OAuthClientListOptions{}
		for {
			ocList, err := tfeClient.OAuthClients.List(ctx, organization, options)
			if err != nil {
				return fmt.Errorf("Error retrieving OAuth Clients: %w", err)
			}

			for _, item := range ocList.Items {
				if name != "" && serviceProvider != "" {
					if item.Name != nil && *item.Name == name && item.ServiceProvider == serviceProvider {
						ocMatches = append(ocMatches, item)
					}
				} else if name != "" {
					if item.Name != nil && *item.Name == name {
						ocMatches = append(ocMatches, item)
					}
				} else if serviceProvider != "" {
					if item.ServiceProvider == serviceProvider {
						ocMatches = append(ocMatches, item)
					}
				}
			}

			// Exit the loop when we've seen all pages.
			if ocList.CurrentPage >= ocList.TotalPages {
				break
			}

			// Update the page number to get the next page.
			options.PageNumber = ocList.NextPage
		}
		if len(ocMatches) == 0 {
			return fmt.Errorf("No OAuthClients found matching the given parameters")
		}
		if len(ocMatches) > 1 {
			return fmt.Errorf("Too many OAuthClients were found to match the given parameters. Please narrow your search.")
		}
		oc = ocMatches[0]
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
	d.Set("organization", oc.Organization.Name)
	d.Set("service_provider", oc.ServiceProvider)
	d.Set("service_provider_display_name", oc.ServiceProviderName)

	switch len(oc.OAuthTokens) {
	case 0:
		d.Set("oauth_token_id", "")
	case 1:
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	default:
		return fmt.Errorf("Unexpected number of OAuth tokens: %d", len(oc.OAuthTokens)) // nolint:golint,errorlint
	}

	return nil
}
