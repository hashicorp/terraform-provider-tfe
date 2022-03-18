package tfe

import (
	"fmt"
	"log"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOauthClients() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOauthClientList,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
			"oauth_clients": {
				Computed: true,
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
		},
	}
}

func dataSourceTFEOauthClientList(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	organization := d.Get("organization").(string)

	log.Printf("[DEBUG] Setting OauthClients Attributes for organization %s", organization)
	err := oauthClientsPopulateFields(tfeClient, organization, d)
	if err != nil {
		return err
	}

	return nil
}

func oauthClientsPopulateFields(client *tfe.Client, organization string, d *schema.ResourceData) error {
	var oauthClients []map[string]interface{}

	log.Printf("[DEBUG] Listing oauth clients")
	options := tfe.OAuthClientListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
	}
	for {
		oauthClientList, err := client.OAuthClients.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("failed to retrieve oauth Clients from organization %s: %v", organization, err)
		}

		for _, client := range oauthClientList.Items {
			flattenedOauthClient, err := flattenOauthClient(client)
			if err != nil {
				return fmt.Errorf("failed to flatten oauthClient: %v", err)
			}
			log.Printf("[DEBUG] Flattened OauthClient: %v", flattenedOauthClient)

			oauthClients = append(oauthClients, flattenedOauthClient)
		}

		// Exit the loop when we've seen all pages.
		if oauthClientList.CurrentPage >= oauthClientList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = oauthClientList.NextPage
	}

	d.SetId("oauth-clients")
	return d.Set("oauth_clients", oauthClients)
}

func flattenOauthClient(in *tfe.OAuthClient) (map[string]interface{}, error) {
	att := make(map[string]interface{})

	if len(in.ID) > 0 {
		att["oauth_client_id"] = in.ID
	}
	if len(in.APIURL) > 0 {
		att["api_url"] = in.APIURL
	}
	if len(in.CallbackURL) > 0 {
		att["callback_url"] = in.CallbackURL
	}
	if len(in.ConnectPath) > 0 {
		att["connect_path"] = in.ConnectPath
	}
	if len(in.CreatedAt.Format(time.RFC3339)) > 0 {
		att["created_at"] = in.CreatedAt.Format(time.RFC3339)
	}
	if len(in.HTTPURL) > 0 {
		att["http_url"] = in.HTTPURL
	}
	if len(in.Key) > 0 {
		att["key"] = in.Key
	}
	if len(in.ServiceProvider) > 0 {
		att["service_provider"] = in.ServiceProvider
	}
	if len(in.ServiceProviderName) > 0 {
		att["service_provider_display_name"] = in.ServiceProviderName
	}

	att["organization_name"] = flattenOrganizationName(in.Organization)

	flattenedOauthTokenID, err := flattenOauthToken(in.OAuthTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten 'oauth_token_id': %v", err)
	}
	log.Printf("[DEBUG] Flattened oauthTokenID: %v", flattenedOauthTokenID)
	att["oauth_token_id"] = flattenedOauthTokenID

	return att, nil
}

func flattenOauthToken(in []*tfe.OAuthToken) (interface{}, error) {
	switch len(in) {
	case 0:
		return "", nil
	case 1:
		return in[0].ID, nil
	default:
		return nil, fmt.Errorf("unexpected number of OAuth tokens: %d", len(in))
	}
}

func flattenOrganizationName(in *tfe.Organization) interface{} {
	return in.Name
}
