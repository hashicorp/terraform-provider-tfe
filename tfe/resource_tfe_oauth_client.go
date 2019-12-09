package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOAuthClientCreate,
		Read:   resourceTFEOAuthClientRead,
		Delete: resourceTFEOAuthClientDelete,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"api_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"oauth_token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"private_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},

			"service_provider": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderGithub),
						string(tfe.ServiceProviderGithubEE),
						string(tfe.ServiceProviderGitlab),
						string(tfe.ServiceProviderGitlabCE),
						string(tfe.ServiceProviderGitlabEE),
					},
					false,
				),
			},

			"oauth_token_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEOAuthClientCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization and provider.
	organization := d.Get("organization").(string)
	serviceProvider := tfe.ServiceProviderType(d.Get("service_provider").(string))

	// Create a new options struct.
	options := tfe.OAuthClientCreateOptions{
		APIURL:          tfe.String(d.Get("api_url").(string)),
		HTTPURL:         tfe.String(d.Get("http_url").(string)),
		OAuthToken:      tfe.String(d.Get("oauth_token").(string)),
		ServiceProvider: tfe.ServiceProvider(serviceProvider),
		PrivateKey:      tfe.String(d.Get("private_key").(string)),
	}

	log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
	oc, err := tfeClient.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %v", organization, err)
	}

	d.SetId(oc.ID)

	return resourceTFEOAuthClientRead(d, meta)
}

func resourceTFEOAuthClientRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of OAuth client: %s", d.Id())
	oc, err := tfeClient.OAuthClients.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] OAuth client %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	// Update the config.
	d.Set("api_url", oc.APIURL)
	d.Set("http_url", oc.HTTPURL)
	d.Set("organization", oc.Organization.Name)
	d.Set("service_provider", string(oc.ServiceProvider))

	switch len(oc.OAuthTokens) {
	case 0:
		d.Set("oauth_token_id", "")
	case 1:
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	default:
		return fmt.Errorf("Unexpected number of OAuth tokens: %d", len(oc.OAuthTokens))
	}

	return nil
}

func resourceTFEOAuthClientDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete OAuth client: %s", d.Id())
	err := tfeClient.OAuthClients.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting OAuth client %s: %v", d.Id(), err)
	}

	return nil
}
