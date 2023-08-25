// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			"key": {
				Type:      schema.TypeString,
				ForceNew:  true,
				Sensitive: true,
				Optional:  true,
			},

			"oauth_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"private_key": {
				Type:      schema.TypeString,
				ForceNew:  true,
				Sensitive: true,
				Optional:  true,
			},

			"secret": {
				Type:      schema.TypeString,
				ForceNew:  true,
				Sensitive: true,
				Optional:  true,
			},

			"rsa_public_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				// this field is only for BitBucket Server, and requires these other
				RequiredWith: []string{"secret", "key"},
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
				Type:     schema.TypeString,
				Computed: true,
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
	oc, err := config.Client.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %w", organization, err)
	}

	d.SetId(oc.ID)

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
