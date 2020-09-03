package tfe

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEOAuthClientRead,
		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssh_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_url": {
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
		return fmt.Errorf("Error retrieving OAuth client: %v", err)
	}

	tokenID := oc.OAuthTokens[0].ID
	d.SetId(oc.ID)
	_ = d.Set("ssh_key", oc.RSAPublicKey)
	_ = d.Set("token_id", tokenID)
	_ = d.Set("api_url", oc.APIURL)
	_ = d.Set("http_url", oc.HTTPURL)

	return nil
}
