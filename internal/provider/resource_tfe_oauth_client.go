// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOAuthClientCreate,
		Read:   resourceTFEOAuthClientRead,
		Delete: resourceTFEOAuthClientDelete,
		Update: resourceTFEOAuthClientUpdate,

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Display name for the OAuth Client. Defaults to the service_provider if not supplied.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"api_url": {
				Description: "The base URL of your VCS provider's API (e.g. https://api.github.com or https://ghe.example.com/api/v3).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"http_url": {
				Description: "The homepage of your VCS provider (e.g. https://github.com or https://ghe.example.com).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"key": {
				Description: "The OAuth Client key. Can refer to a Consumer Key, Application Key, or another type of client key for the VCS provider.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"oauth_token": {
				Description: "The token string you were given by your VCS provider, e.g. a GitHub personal access token.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},

			"private_key": {
				Description: "The text of the private key associated with your Azure DevOps Server account. Required for ado_server.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"secret": {
				Description: "The OAuth Client secret. For Bitbucket Data Center, this secret is the text of the SSH private key associated with your Bitbucket Data Center Application Link. Required for bitbucket_data_center.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"rsa_public_key": {
				Description: "The text of the SSH public key associated with your Bitbucket Data Center Application Link. Required for Bitbucket Data Center in conjunction with the secret. Not used for any other providers.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				// this field is only for BitBucket Data Center, and requires these other
				RequiredWith: []string{"secret", "key"},
			},

			"service_provider": {
				Description: "The VCS provider being connected with. Valid options are ado_server, ado_services, bitbucket_data_center, bitbucket_hosted, bitbucket_server (deprecated), github, github_enterprise, gitlab_hosted, gitlab_community_edition, or gitlab_enterprise_edition.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.ServiceProviderAzureDevOpsServer),
						string(tfe.ServiceProviderAzureDevOpsServices),
						string(tfe.ServiceProviderBitbucket),
						string(tfe.ServiceProviderBitbucketServer),
						string(tfe.ServiceProviderBitbucketDataCenter),
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
			"agent_pool_id": {
				Description: "An existing agent pool ID within the organization that has Private VCS support enabled.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"organization_scoped": {
				Description: "Whether or not the OAuth client is scoped to all projects and workspaces in the organization. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceTFEOAuthClientCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization and provider.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}
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
		Name:               tfe.String(name),
		APIURL:             tfe.String(d.Get("api_url").(string)),
		HTTPURL:            tfe.String(d.Get("http_url").(string)),
		OAuthToken:         tfe.String(d.Get("oauth_token").(string)),
		Key:                tfe.String(key),
		ServiceProvider:    tfe.ServiceProvider(serviceProvider),
		OrganizationScoped: tfe.Bool(d.Get("organization_scoped").(bool)),
	}

	if serviceProvider == tfe.ServiceProviderAzureDevOpsServer {
		options.PrivateKey = tfe.String(privateKey)
	}
	if serviceProvider == tfe.ServiceProviderBitbucketServer || serviceProvider == tfe.ServiceProviderBitbucketDataCenter {
		options.RSAPublicKey = tfe.String(rsaPublicKey)
		options.Secret = tfe.String(secret)
	}
	if serviceProvider == tfe.ServiceProviderBitbucket {
		options.Secret = tfe.String(secret)
	}
	if v, ok := d.GetOk("agent_pool_id"); ok && v.(string) != "" {
		options.AgentPool = &tfe.AgentPool{ID: *tfe.String(v.(string))}
	}

	log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
	oc, err := config.Client.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %w", organization, err)
	}

	d.SetId(oc.ID)

	if len(oc.OAuthTokens) > 0 {
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	} else {
		d.Set("oauth_token_id", "")
	}

	return resourceTFEOAuthClientRead(d, meta)
}

func resourceTFEOAuthClientRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of OAuth client: %s", d.Id())
	oc, err := config.Client.OAuthClients.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] OAuth client %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	// Update the config.
	d.Set("organization", oc.Organization.Name)
	d.Set("api_url", oc.APIURL)
	d.Set("http_url", oc.HTTPURL)
	d.Set("service_provider", string(oc.ServiceProvider))
	d.Set("organization_scoped", oc.OrganizationScoped)

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

func resourceTFEOAuthClientDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete OAuth client: %s", d.Id())
	err := config.Client.OAuthClients.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting OAuth client %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOAuthClientUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Create a new options struct.
	options := tfe.OAuthClientUpdateOptions{
		OrganizationScoped: tfe.Bool(d.Get("organization_scoped").(bool)),
		OAuthToken:         tfe.String(d.Get("oauth_token").(string)),
	}

	log.Printf("[DEBUG] Update OAuth client %s", d.Id())
	_, err := config.Client.OAuthClients.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating OAuth client %s: %w", d.Id(), err)
	}

	return resourceTFEOAuthClientRead(d, meta)
}
