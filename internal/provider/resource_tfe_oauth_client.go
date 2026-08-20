// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an OAuth client, which represents the connection between an organization and a VCS provider." +
			"\n\n-> **Note:** This resource does not currently support creation of Azure DevOps Services OAuth clients.",

		Create: resourceTFEOAuthClientCreate,
		Read:   resourceTFEOAuthClientRead,
		Delete: resourceTFEOAuthClientDelete,
		Update: resourceTFEOAuthClientUpdate,

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the OAuth client.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Display name for the OAuth Client. Defaults to the `service_provider` if not supplied.",
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
				Description: "The token string you were given by your VCS provider, e.g. `ghp_xxxxxxxxxxxxxxx` for a GitHub personal access token. For more information on how to generate this token string for your VCS provider, see the [Create an OAuth Client](https://developer.hashicorp.com/terraform/cloud-docs/api-docs/oauth-clients#create-an-oauth-client) documentation.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},

			"private_key": {
				Description: "The text of the private key associated with your Azure DevOps Server account. Required for `ado_server`.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Sensitive:   true,
				Optional:    true,
			},

			"secret": {
				Description: "The OAuth Client secret, used for Bitbucket Data Center. This secret is the text of the SSH private key associated with your Bitbucket Data Center Application Link. Required for `bitbucket_data_center`.",
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
				Description: "The VCS provider being connected with. Valid options are `ado_server`, `ado_services`, `bitbucket_data_center`, `bitbucket_hosted`, `bitbucket_server`(deprecated), `github`, `github_enterprise`, `gitlab_hosted`, `gitlab_community_edition`, or `gitlab_enterprise_edition`.",
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
				Description: "Whether or not the OAuth client is scoped to all projects and workspaces in the organization. Defaults to `true`.",
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

	attrs := models.NewOauthClients_attributes()
	attrs.SetName(ptr(name))
	attrs.SetApiUrl(ptr(d.Get("api_url").(string)))
	attrs.SetHttpUrl(ptr(d.Get("http_url").(string)))
	attrs.SetOauthTokenString(ptr(d.Get("oauth_token").(string)))
	attrs.SetKey(ptr(key))
	attrs.SetServiceProvider(ptr(string(serviceProvider)))
	attrs.SetOrganizationScoped(ptr(d.Get("organization_scoped").(bool)))

	if serviceProvider == tfe.ServiceProviderAzureDevOpsServer {
		attrs.SetPrivateKey(ptr(privateKey))
	}
	if serviceProvider == tfe.ServiceProviderBitbucketServer || serviceProvider == tfe.ServiceProviderBitbucketDataCenter {
		attrs.SetRsaPublicKey(ptr(rsaPublicKey))
		attrs.SetSecret(ptr(secret))
	}
	if serviceProvider == tfe.ServiceProviderBitbucket {
		attrs.SetSecret(ptr(secret))
	}

	body := models.NewOauthClients()
	body.SetTypeEscaped(ptr(models.OAUTHCLIENTS_OAUTHCLIENTS_TYPE))
	body.SetAttributes(attrs)
	if v, ok := d.GetOk("agent_pool_id"); ok && v.(string) != "" {
		agentPoolData := models.NewAgentPoolsHasOne_data()
		agentPoolData.SetId(ptr(v.(string)))
		agentPoolData.SetTypeEscaped(ptr(models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE))
		agentPool := models.NewAgentPoolsHasOne()
		agentPool.SetData(agentPoolData)
		rels := models.NewOauthClients_relationships()
		rels.SetAgentPool(agentPool)
		body.SetRelationships(rels)
	}
	env := models.NewOauthClientsEnvelope()
	env.SetData(body)

	log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
	ocEnv, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).OauthClients().Post(ctx, env, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %w", organization, err)
	}

	oc := ocEnv.GetData()
	d.SetId(valueOrZero(oc.GetId()))

	d.Set("oauth_token_id", "")
	if rels := oc.GetRelationships(); rels != nil && rels.GetOauthTokens() != nil {
		tokens := rels.GetOauthTokens().GetData()
		if len(tokens) > 0 {
			d.Set("oauth_token_id", valueOrZero(tokens[0].GetId()))
		}
	}

	return resourceTFEOAuthClientRead(d, meta)
}

func resourceTFEOAuthClientRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of OAuth client: %s", d.Id())
	ocEnv, err := config.ClientV2.API.OauthClients().ByOauth_client_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] OAuth client %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	oc := ocEnv.GetData()
	attrs := oc.GetAttributes()
	rels := oc.GetRelationships()

	if attrs != nil {
		if rels != nil && rels.GetOrganization() != nil && rels.GetOrganization().GetData() != nil {
			d.Set("organization", valueOrZero(rels.GetOrganization().GetData().GetId()))
		}
		d.Set("api_url", valueOrZero(attrs.GetApiUrl()))
		d.Set("http_url", valueOrZero(attrs.GetHttpUrl()))
		d.Set("service_provider", valueOrZero(attrs.GetServiceProvider()))
		d.Set("organization_scoped", attrs.GetOrganizationScoped())
	}

	d.Set("oauth_token_id", "")
	if rels != nil && rels.GetOauthTokens() != nil {
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

	return nil
}

func resourceTFEOAuthClientDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete OAuth client: %s", d.Id())
	err := config.ClientV2.API.OauthClients().ByOauth_client_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting OAuth client %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOAuthClientUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	attrs := models.NewOauthClients_attributes()
	attrs.SetOrganizationScoped(ptr(d.Get("organization_scoped").(bool)))
	attrs.SetOauthTokenString(ptr(d.Get("oauth_token").(string)))
	body := models.NewOauthClients()
	body.SetTypeEscaped(ptr(models.OAUTHCLIENTS_OAUTHCLIENTS_TYPE))
	body.SetAttributes(attrs)
	env := models.NewOauthClientsEnvelope()
	env.SetData(body)

	log.Printf("[DEBUG] Update OAuth client %s", d.Id())
	_, err := config.ClientV2.API.OauthClients().ByOauth_client_id(d.Id()).Patch(ctx, env, nil)
	if err != nil {
		return fmt.Errorf("Error updating OAuth client %s: %w", d.Id(), err)
	}

	return resourceTFEOAuthClientRead(d, meta)
}
