package tfe

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOAuthClientRead,
		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_url": {
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

	ocID := d.Get("oauth_client_id").(string)

	oc, err := tfeClient.OAuthClients.Read(ctx, ocID)
	if err != nil {
		return fmt.Errorf("Error retrieving OAuth client: %w", err)
	}

	d.SetId(oc.ID)
	_ = d.Set("api_url", oc.APIURL)
	_ = d.Set("http_url", oc.HTTPURL)

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
