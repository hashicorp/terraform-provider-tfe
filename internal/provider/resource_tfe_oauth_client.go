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
	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an OAuth client, which represents the connection between an organization and a VCS provider.",

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

			"ado_org_name": {
				Description: "The Azure DevOps organization name for connections using an organization-scoped personal access token. Only valid for `ado_services`. Leave blank when using a globally-scoped personal access token.",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^$|^[A-Za-z0-9](?:[A-Za-z0-9-]{0,48}[A-Za-z0-9])?$`),
					"must be 50 characters or fewer, start and end with a letter or number, and contain only letters, numbers, and hyphens",
				),
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
	adoOrgName := d.Get("ado_org_name").(string)

	if serviceProvider == tfe.ServiceProviderAzureDevOpsServer && privateKey == "" {
		return fmt.Errorf("private_key is required for service_provider %s", serviceProvider)
	}
	if adoOrgName != "" && serviceProvider != tfe.ServiceProviderAzureDevOpsServices {
		return fmt.Errorf("ado_org_name is only valid for service_provider %s", tfe.ServiceProviderAzureDevOpsServices)
	}

	if adoOrgName != "" {
		log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
		env, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).OauthClients().Post(ctx, newOAuthClientEnvelope(d, true), nil)
		if err != nil {
			return fmt.Errorf("Error creating OAuth client for organization %s: %w", organization, err)
		}
		if env == nil || env.GetData() == nil || env.GetData().GetId() == nil {
			return fmt.Errorf("Error creating OAuth client for organization %s: API returned no data", organization)
		}

		d.SetId(*env.GetData().GetId())
		return resourceTFEOAuthClientRead(d, meta)
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

	adoOrgName, err := readOAuthClientADOOrgName(config, d.Id())
	if err != nil {
		return err
	}
	d.Set("ado_org_name", adoOrgName)

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

	if d.HasChange("ado_org_name") {
		serviceProvider := tfe.ServiceProviderType(d.Get("service_provider").(string))
		adoOrgName := d.Get("ado_org_name").(string)
		if adoOrgName != "" && serviceProvider != tfe.ServiceProviderAzureDevOpsServices {
			return fmt.Errorf("ado_org_name is only valid for service_provider %s", tfe.ServiceProviderAzureDevOpsServices)
		}

		log.Printf("[DEBUG] Update OAuth client %s", d.Id())
		_, err := config.ClientV2.API.OauthClients().ByOauth_client_id(d.Id()).Patch(ctx, newOAuthClientEnvelope(d, false), nil)
		if err != nil {
			return fmt.Errorf("Error updating OAuth client %s: %w", d.Id(), err)
		}
		return resourceTFEOAuthClientRead(d, meta)
	}

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

func newOAuthClientEnvelope(d *schema.ResourceData, create bool) models.OauthClientsEnvelopeable {
	attrs := models.NewOauthClients_attributes()
	if adoOrgName := d.Get("ado_org_name").(string); adoOrgName != "" {
		attrs.SetAdoOrgName(&adoOrgName)
	} else if !create {
		attrs.GetAdditionalData()["ado-org-name"] = nil
	}
	attrs.SetOrganizationScoped(ptr(d.Get("organization_scoped").(bool)))
	attrs.GetAdditionalData()["oauth-token-string"] = d.Get("oauth_token").(string)

	client := models.NewOauthClients()
	clientType := models.OAUTHCLIENTS_OAUTHCLIENTS_TYPE
	client.SetTypeEscaped(&clientType)
	client.SetAttributes(attrs)

	if create {
		attrs.SetName(ptr(d.Get("name").(string)))
		attrs.SetApiUrl(ptr(d.Get("api_url").(string)))
		attrs.SetHttpUrl(ptr(d.Get("http_url").(string)))
		attrs.SetKey(ptr(d.Get("key").(string)))
		attrs.SetServiceProvider(ptr(d.Get("service_provider").(string)))

		serviceProvider := tfe.ServiceProviderType(d.Get("service_provider").(string))
		if serviceProvider == tfe.ServiceProviderAzureDevOpsServer {
			attrs.GetAdditionalData()["private-key"] = d.Get("private_key").(string)
		}
		if serviceProvider == tfe.ServiceProviderBitbucketServer || serviceProvider == tfe.ServiceProviderBitbucketDataCenter {
			attrs.SetRsaPublicKey(ptr(d.Get("rsa_public_key").(string)))
			attrs.SetSecret(ptr(d.Get("secret").(string)))
		}
		if serviceProvider == tfe.ServiceProviderBitbucket {
			attrs.SetSecret(ptr(d.Get("secret").(string)))
		}

		if agentPoolID := d.Get("agent_pool_id").(string); agentPoolID != "" {
			agentPoolData := models.NewAgentPoolsHasOne_data()
			agentPoolData.SetId(&agentPoolID)
			agentPoolType := models.AGENTPOOLS_AGENTPOOLSIDENTIFIER_TYPE
			agentPoolData.SetTypeEscaped(&agentPoolType)

			agentPool := models.NewAgentPoolsHasOne()
			agentPool.SetData(agentPoolData)
			relationships := models.NewOauthClients_relationships()
			relationships.SetAgentPool(agentPool)
			client.SetRelationships(relationships)
		}
	} else {
		client.SetId(ptr(d.Id()))
	}

	envelope := models.NewOauthClientsEnvelope()
	envelope.SetData(client)
	return envelope
}

func readOAuthClientADOOrgName(config ConfiguredClient, oauthClientID string) (string, error) {
	env, err := config.ClientV2.API.OauthClients().ByOauth_client_id(oauthClientID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("Error reading Azure DevOps organization name for OAuth client %s: %w", oauthClientID, err)
	}
	if env == nil || env.GetData() == nil || env.GetData().GetAttributes() == nil {
		return "", fmt.Errorf("Error reading Azure DevOps organization name for OAuth client %s: API returned no data", oauthClientID)
	}
	return valueOrZero(env.GetData().GetAttributes().GetAdoOrgName()), nil
}
