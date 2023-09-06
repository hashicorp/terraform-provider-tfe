// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-tfe/internal/client"
)

const defaultSSLSkipVerify = false

var (
	errMissingOrganization = errors.New("no organization was specified on the resource or provider")
)

// ConfiguredClient wraps the tfe.Client the provider uses, plus the default
// organization name to be used by resources that need an organization but don't
// specify one.
type ConfiguredClient struct {
	Client       *tfe.Client
	Organization string
}

func (c ConfiguredClient) schemaOrDefaultOrganization(resource *schema.ResourceData) (string, error) {
	return c.schemaOrDefaultOrganizationKey(resource, "organization")
}

func (c ConfiguredClient) schemaOrDefaultOrganizationKey(resource *schema.ResourceData, key string) (string, error) {
	schemaOrg, got := resource.GetOk(key)
	if got {
		return schemaOrg.(string), nil
	}
	if c.Organization == "" {
		return "", errMissingOrganization
	}
	return c.Organization, nil
}

// ctx is used as default context.Context when making TFE calls.
var ctx = context.Background()

// Provider returns a schema.Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		// Note that defaults and fallbacks which are usually handled by DefaultFunc here are
		// instead handled when fetching a TFC/E client in getClient(). This is because the this
		// provider is actually two muxed providers which must respect the same logic for fetching
		// those values in each schema.
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["hostname"],
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["token"],
			},

			"ssl_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: descriptions["ssl_skip_verify"],
			},

			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["organization"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"tfe_organizations":           dataSourceTFEOrganizations(),
			"tfe_organization":            dataSourceTFEOrganization(),
			"tfe_agent_pool":              dataSourceTFEAgentPool(),
			"tfe_ip_ranges":               dataSourceTFEIPRanges(),
			"tfe_oauth_client":            dataSourceTFEOAuthClient(),
			"tfe_organization_membership": dataSourceTFEOrganizationMembership(),
			"tfe_organization_run_task":   dataSourceTFEOrganizationRunTask(),
			"tfe_organization_tags":       dataSourceTFEOrganizationTags(),
			"tfe_project":                 dataSourceTFEProject(),
			"tfe_slug":                    dataSourceTFESlug(),
			"tfe_ssh_key":                 dataSourceTFESSHKey(),
			"tfe_team":                    dataSourceTFETeam(),
			"tfe_teams":                   dataSourceTFETeams(),
			"tfe_team_access":             dataSourceTFETeamAccess(),
			"tfe_team_project_access":     dataSourceTFETeamProjectAccess(),
			"tfe_workspace":               dataSourceTFEWorkspace(),
			"tfe_workspace_ids":           dataSourceTFEWorkspaceIDs(),
			"tfe_workspace_run_task":      dataSourceTFEWorkspaceRunTask(),
			"tfe_variables":               dataSourceTFEWorkspaceVariables(),
			"tfe_variable_set":            dataSourceTFEVariableSet(),
			"tfe_policy_set":              dataSourceTFEPolicySet(),
			"tfe_organization_members":    dataSourceTFEOrganizationMembers(),
			"tfe_github_app_installation": dataSourceTFEGHAInstallation(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"tfe_admin_organization_settings":    resourceTFEAdminOrganizationSettings(),
			"tfe_agent_pool":                     resourceTFEAgentPool(),
			"tfe_agent_pool_allowed_workspaces":  resourceTFEAgentPoolAllowedWorkspaces(),
			"tfe_agent_token":                    resourceTFEAgentToken(),
			"tfe_notification_configuration":     resourceTFENotificationConfiguration(),
			"tfe_oauth_client":                   resourceTFEOAuthClient(),
			"tfe_organization":                   resourceTFEOrganization(),
			"tfe_organization_membership":        resourceTFEOrganizationMembership(),
			"tfe_organization_module_sharing":    resourceTFEOrganizationModuleSharing(),
			"tfe_organization_run_task":          resourceTFEOrganizationRunTask(),
			"tfe_organization_token":             resourceTFEOrganizationToken(),
			"tfe_policy":                         resourceTFEPolicy(),
			"tfe_policy_set":                     resourceTFEPolicySet(),
			"tfe_policy_set_parameter":           resourceTFEPolicySetParameter(),
			"tfe_project":                        resourceTFEProject(),
			"tfe_project_policy_set":             resourceTFEProjectPolicySet(),
			"tfe_project_variable_set":           resourceTFEProjectVariableSet(),
			"tfe_registry_module":                resourceTFERegistryModule(),
			"tfe_no_code_module":                 resourceTFENoCodeModule(),
			"tfe_run_trigger":                    resourceTFERunTrigger(),
			"tfe_sentinel_policy":                resourceTFESentinelPolicy(),
			"tfe_ssh_key":                        resourceTFESSHKey(),
			"tfe_team":                           resourceTFETeam(),
			"tfe_team_access":                    resourceTFETeamAccess(),
			"tfe_team_organization_member":       resourceTFETeamOrganizationMember(),
			"tfe_team_organization_members":      resourceTFETeamOrganizationMembers(),
			"tfe_team_project_access":            resourceTFETeamProjectAccess(),
			"tfe_team_member":                    resourceTFETeamMember(),
			"tfe_team_members":                   resourceTFETeamMembers(),
			"tfe_team_token":                     resourceTFETeamToken(),
			"tfe_terraform_version":              resourceTFETerraformVersion(),
			"tfe_workspace":                      resourceTFEWorkspace(),
			"tfe_workspace_run_task":             resourceTFEWorkspaceRunTask(),
			"tfe_variable_set":                   resourceTFEVariableSet(),
			"tfe_workspace_policy_set":           resourceTFEWorkspacePolicySet(),
			"tfe_workspace_policy_set_exclusion": resourceTFEWorkspacePolicySetExclusion(),
			"tfe_workspace_run":                  resourceTFEWorkspaceRun(),
			"tfe_workspace_variable_set":         resourceTFEWorkspaceVariableSet(),
		},
		ConfigureContextFunc: configure(),
	}
}

func configure() schema.ConfigureContextFunc {
	return func(ctx context.Context, rd *schema.ResourceData) (any, diag.Diagnostics) {
		providerOrganization := rd.Get("organization").(string)
		if providerOrganization == "" {
			providerOrganization = os.Getenv("TFE_ORGANIZATION")
		}

		client, err := configureClient(rd)
		if err != nil {
			return nil, diag.Errorf("failed to create SDK client: %s", err)
		}

		return ConfiguredClient{
			client,
			providerOrganization,
		}, nil
	}
}

func configureClient(d *schema.ResourceData) (*tfe.Client, error) {
	hostname := d.Get("hostname").(string)
	token := d.Get("token").(string)
	insecure := d.Get("ssl_skip_verify").(bool)

	return client.GetClient(hostname, token, insecure)
}

var descriptions = map[string]string{
	"hostname": "The Terraform Enterprise hostname to connect to. Defaults to app.terraform.io.",
	"token": "The token used to authenticate with Terraform Enterprise. We recommend omitting\n" +
		"the token which can be set as credentials in the CLI config file.",
	"ssl_skip_verify": "Whether or not to skip certificate verifications.",
	"organization": "The organization to apply to a resource if one is not defined on\n" +
		"the resource itself",
}

// A commonly used helper method to check if the error
// returned was tfe.ErrResourceNotFound
func isErrResourceNotFound(err error) bool {
	return errors.Is(err, tfe.ErrResourceNotFound)
}
