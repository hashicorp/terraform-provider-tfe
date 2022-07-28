package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "An OAuth Client represents the connection between an organization and a VCS provider." +
			"\n\n -> **Note:** This resource does not currently support creation of Azure DevOps Services OAuth clients.",

		Create: resourceTFEOAuthClientCreate,
		Read:   resourceTFEOAuthClientRead,
		Delete: resourceTFEOAuthClientDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Display name for the OAuth Client. Defaults to the `service_provider` if not supplied.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"organization": {
				Description: "Name of the Terraform organization.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"api_url": {
				Description: "The base URL of your VCS provider's API (e.g.`https://api.github.com` or `https://ghe.example.com/api/v3`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"http_url": {
				Description: "The homepage of your VCS provider (e.g.`https://github.com` or `https://ghe.example.com`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"key": {
				Description: "The OAuth Client key can refer to a Consumer Key, Application Key, or another type of client key for the VCS provider.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"oauth_token": {
				Description: "The token string you were given by your VCS provider, e.g. `ghp_xxxxxxxxxxxxxxx` for a GitHub personal access token. For more information on how to generate this token string for your VCS provider, see the [Create an OAuth Client](https://www.terraform.io/docs/cloud/api/oauth-clients.html#create-an-oauth-client) documentation.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
			},

			"private_key": {
				Description: "(Required for `ado_server`) The text of the private key associated with your Azure DevOps Server account",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"secret": {
				Description: "(Required for `bitbucket_server`) The OAuth Client secret is used for BitBucket Server, this secret is the " +
					"the text of the SSH private key associated with your BitBucket Server Application Link.",
				Type:      schema.TypeString,
				ForceNew:  true,
				Sensitive: true,
				Optional:  true,
			},

			"rsa_public_key": {
				Description: "(Required for `bitbucket_server`) Required for BitBucket " +
					"Server in conjunction with the secret. Not used for any other providers. The " +
					"text of the SSH public key associated with your BitBucket Server Application Link.",
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				// this field is only for BitBucket Server, and requires these other
				RequiredWith: []string{"secret", "key"},
			},

			"service_provider": {
				Description: "The VCS provider being connected with. Valid options are `ado_server`, `ado_services`, `bitbucket_hosted`, `bitbucket_server`, `github`, `github_enterprise`, `gitlab_hosted`,`gitlab_community_edition`, or `gitlab_enterprise_edition`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderBitbucketServer),
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
				Description: "The ID of the OAuth token associated with the OAuth client.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceTFEOAuthClientCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization and provider.
	organization := d.Get("organization").(string)
	name := d.Get("name").(string)
	privateKey := d.Get("private_key").(string)
	rsaPublicKey := d.Get("rsa_public_key").(string)
	key := d.Get("key").(string)
	secret := d.Get("secret").(string)
	serviceProvider := tfe.ServiceProviderType(d.Get("service_provider").(string))

	if serviceProvider == tfe.ServiceProviderAzureDevOpsServer && privateKey == "" {
		return fmt.Errorf("private_key is required for service_provider %s", serviceProvider)
	}

	// Create a new options struct.
	// The tfe.OAuthClientCreateOptions has omitempty for these values, so if it
	// is empty, then it will be ignored in the create request
	options := tfe.OAuthClientCreateOptions{
		Name:            tfe.String(name),
		APIURL:          tfe.String(d.Get("api_url").(string)),
		HTTPURL:         tfe.String(d.Get("http_url").(string)),
		OAuthToken:      tfe.String(d.Get("oauth_token").(string)),
		Key:             tfe.String(key),
		ServiceProvider: tfe.ServiceProvider(serviceProvider),
	}

	if serviceProvider == tfe.ServiceProviderAzureDevOpsServer {
		options.PrivateKey = tfe.String(privateKey)
	}
	if serviceProvider == tfe.ServiceProviderBitbucketServer {
		options.RSAPublicKey = tfe.String(rsaPublicKey)
		options.Secret = tfe.String(secret)
	}
	if serviceProvider == tfe.ServiceProviderBitbucket {
		options.Secret = tfe.String(secret)
	}

	log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
	oc, err := tfeClient.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %w", organization, err)
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
		return fmt.Errorf("Error deleting OAuth client %s: %w", d.Id(), err)
	}

	return nil
}
