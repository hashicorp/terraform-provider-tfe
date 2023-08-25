// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	providerVersion "github.com/hashicorp/terraform-provider-tfe/version"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
)

const defaultHostname = "app.terraform.io"
const defaultSSLSkipVerify = false

var (
	tfeServiceIDs          = []string{"tfe.v2.2"}
	errMissingAuthToken    = errors.New("required token could not be found. Please set the token using an input variable in the provider configuration block or by using the TFE_TOKEN environment variable")
	errMissingOrganization = errors.New("no organization was specified on the resource or provider")
)

// Config is the structure of the configuration for the Terraform CLI.
type Config struct {
	Hosts       map[string]*ConfigHost            `hcl:"host"`
	Credentials map[string]map[string]interface{} `hcl:"credentials"`
}

// ConfigHost is the structure of the "host" nested block within the CLI
// configuration, which can be used to override the default service host
// discovery behavior for a particular hostname.
type ConfigHost struct {
	Services map[string]interface{} `hcl:"services"`
}

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
			"tfe_admin_organization_settings":   resourceTFEAdminOrganizationSettings(),
			"tfe_agent_pool":                    resourceTFEAgentPool(),
			"tfe_agent_pool_allowed_workspaces": resourceTFEAgentPoolAllowedWorkspaces(),
			"tfe_agent_token":                   resourceTFEAgentToken(),
			"tfe_notification_configuration":    resourceTFENotificationConfiguration(),
			"tfe_oauth_client":                  resourceTFEOAuthClient(),
			"tfe_organization":                  resourceTFEOrganization(),
			"tfe_organization_membership":       resourceTFEOrganizationMembership(),
			"tfe_organization_module_sharing":   resourceTFEOrganizationModuleSharing(),
			"tfe_organization_run_task":         resourceTFEOrganizationRunTask(),
			"tfe_organization_token":            resourceTFEOrganizationToken(),
			"tfe_policy":                        resourceTFEPolicy(),
			"tfe_policy_set":                    resourceTFEPolicySet(),
			"tfe_policy_set_parameter":          resourceTFEPolicySetParameter(),
			"tfe_project":                       resourceTFEProject(),
			"tfe_project_policy_set":            resourceTFEProjectPolicySet(),
			"tfe_project_variable_set":          resourceTFEProjectVariableSet(),
			"tfe_registry_module":               resourceTFERegistryModule(),
			"tfe_no_code_module":                resourceTFENoCodeModule(),
			"tfe_run_trigger":                   resourceTFERunTrigger(),
			"tfe_sentinel_policy":               resourceTFESentinelPolicy(),
			"tfe_ssh_key":                       resourceTFESSHKey(),
			"tfe_team":                          resourceTFETeam(),
			"tfe_team_access":                   resourceTFETeamAccess(),
			"tfe_team_organization_member":      resourceTFETeamOrganizationMember(),
			"tfe_team_organization_members":     resourceTFETeamOrganizationMembers(),
			"tfe_team_project_access":           resourceTFETeamProjectAccess(),
			"tfe_team_member":                   resourceTFETeamMember(),
			"tfe_team_members":                  resourceTFETeamMembers(),
			"tfe_team_token":                    resourceTFETeamToken(),
			"tfe_terraform_version":             resourceTFETerraformVersion(),
			"tfe_workspace":                     resourceTFEWorkspace(),
			"tfe_workspace_run_task":            resourceTFEWorkspaceRunTask(),
			"tfe_variable_set":                  resourceTFEVariableSet(),
			"tfe_workspace_variable_set":        resourceTFEWorkspaceVariableSet(),
			"tfe_workspace_policy_set":          resourceTFEWorkspacePolicySet(),
			"tfe_workspace_run":                 resourceTFEWorkspaceRun(),
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

	return getClient(hostname, token, insecure)
}

func getTokenFromEnv() string {
	log.Printf("[DEBUG] TFE_TOKEN used for token value")
	return os.Getenv("TFE_TOKEN")
}

func getTokenFromCreds(services *disco.Disco, hostname svchost.Hostname) string {
	log.Printf("[DEBUG] Attempting to fetch token from Terraform CLI configuration for hostname %s...", hostname)
	creds, err := services.CredentialsForHost(hostname)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
	}
	if creds != nil {
		return creds.Token()
	}
	return ""
}

// getClient encapsulates the logic for configuring a go-tfe client instance for
// the provider, including fallback to values from environment variables. This
// is useful because we're muxing multiple provider servers together and each
// one needs an identically configured client.
func getClient(tfeHost, token string, insecure bool) (*tfe.Client, error) {
	h := tfeHost
	if tfeHost == "" {
		if os.Getenv("TFE_HOSTNAME") != "" {
			h = os.Getenv("TFE_HOSTNAME")
		} else {
			h = defaultHostname
		}
	}

	log.Printf("[DEBUG] Configuring client for host %q", h)

	// Parse the hostname for comparison,
	hostname, err := svchost.ForComparison(h)
	if err != nil {
		return nil, err
	}

	providerUaString := fmt.Sprintf("terraform-provider-tfe/%s", providerVersion.ProviderVersion)

	httpClient := tfe.DefaultConfig().HTTPClient

	// Make sure the transport has a TLS config.
	transport := httpClient.Transport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// If ssl_skip_verify is false, it is either set that way in configuration or unset. Check
	// the environment to see if it was set to true there.  Strictly speaking, this means that
	// the env var can override an explicit 'false' in configuration (which is not true of the
	// other settings), but that's how it goes with a boolean zero value.
	if !insecure && os.Getenv("TFE_SSL_SKIP_VERIFY") != "" {
		v := os.Getenv("TFE_SSL_SKIP_VERIFY")
		insecure, err = strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
	}

	// Configure the certificate verification options.
	if insecure {
		log.Printf("[DEBUG] Warning: Client configured to skip certificate verifications")
	}
	transport.TLSClientConfig.InsecureSkipVerify = insecure

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config)
	services := disco.NewWithCredentialsSource(credsSrc)
	services.SetUserAgent(providerUaString)
	services.Transport = NewLoggingTransport("TFE Discovery", transport)

	// Add any static host configurations service discovery object.
	for userHost, hostConfig := range config.Hosts {
		host, err := svchost.ForComparison(userHost)
		if err != nil {
			// ignore invalid hostnames.
			continue
		}
		services.ForceHostServices(host, hostConfig.Services)
	}

	// Discover the Terraform Enterprise address.
	host, err := services.Discover(hostname)
	if err != nil {
		return nil, err
	}

	// Get the full Terraform Enterprise service address.
	var address *url.URL
	var discoErr error
	for _, tfeServiceID := range tfeServiceIDs {
		service, err := host.ServiceURL(tfeServiceID)
		if _, ok := err.(*disco.ErrVersionNotSupported); !ok && err != nil {
			return nil, err
		}
		// If discoErr is nil we save the first error. When multiple services
		// are checked and we found one that didn't give an error we need to
		// reset the discoErr. So if err is nil, we assign it as well.
		if discoErr == nil || err == nil {
			discoErr = err
		}
		if service != nil {
			address = service
			break
		}
	}

	if providerVersion.ProviderVersion != "dev" {
		// We purposefully ignore the error and return the previous error, as
		// checking for version constraints is considered optional.
		constraints, _ := host.VersionConstraints(tfeServiceIDs[0], "tfe-provider")

		// First check any constraints we might have received.
		if constraints != nil {
			if err := checkConstraints(constraints); err != nil {
				return nil, err
			}
		}
	}

	// When we don't have any constraints errors, also check for discovery
	// errors before we continue.
	if discoErr != nil {
		return nil, discoErr
	}

	// If a token wasn't set in the provider configuration block, try and fetch it
	// from the environment or from Terraform's CLI configuration or configured credential helper.
	if token == "" {
		if os.Getenv("TFE_TOKEN") != "" {
			token = getTokenFromEnv()
		} else {
			token = getTokenFromCreds(services, hostname)
		}
	}

	// If we still don't have a token at this point, we return an error.
	if token == "" {
		return nil, errMissingAuthToken
	}

	// Wrap the configured transport to enable logging.
	httpClient.Transport = NewLoggingTransport("TFE", transport)

	// Create a new TFE client config
	cfg := &tfe.Config{
		Address:    address.String(),
		Token:      token,
		HTTPClient: httpClient,
	}

	// Create a new TFE client.
	client, err := tfe.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	client.RetryServerErrors(true)
	return client, nil
}

// cliConfig tries to find and parse the configuration of the Terraform CLI.
// This is an optional step, so any errors are ignored.
func cliConfig() Config {
	mainConfig := Config{}
	credentialsConfig := Config{}
	combinedConfig := Config{}

	// Main CLI config file; might contain manually-entered credentials, and/or
	// some host service discovery objects. Location is configurable via
	// environment variables.
	configFilePath := locateConfigFile()
	if configFilePath != "" {
		mainConfig = readCliConfigFile(configFilePath)
	}

	// Credentials file; might contain credentials auto-configured by terraform
	// login. Location isn't configurable.
	credentialsFilePath, err := credentialsFile()
	if err != nil {
		log.Printf("[ERROR] Error detecting default credentials file path: %s", err)
	} else {
		credentialsConfig = readCliConfigFile(credentialsFilePath)
	}

	// Use host service discovery configs from main config file.
	combinedConfig.Hosts = mainConfig.Hosts

	// Combine both sets of credentials. Per Terraform's own behavior, the main
	// config file overrides the credentials file if they have any overlapping
	// hostnames.
	combinedConfig.Credentials = credentialsConfig.Credentials
	if combinedConfig.Credentials == nil {
		combinedConfig.Credentials = make(map[string]map[string]interface{})
	}
	for host, creds := range mainConfig.Credentials {
		combinedConfig.Credentials[host] = creds
	}

	return combinedConfig
}

func locateConfigFile() string {
	// To find the main CLI config file, follow Terraform's own logic: try
	// TF_CLI_CONFIG_FILE, then try TERRAFORM_CONFIG, then try the default
	// location.

	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		return os.Getenv("TF_CLI_CONFIG_FILE")
	}

	if os.Getenv("TERRAFORM_CONFIG") != "" {
		return os.Getenv("TERRAFORM_CONFIG")
	}
	filePath, err := configFile()
	if err != nil {
		log.Printf("[ERROR] Error detecting default CLI config file path: %s", err)
		return ""
	}

	return filePath
}

func readCliConfigFile(configFilePath string) Config {
	config := Config{}

	// Read the CLI config file content.
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("[ERROR] Error reading CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Parse the CLI config file content.
	obj, err := hcl.Parse(string(content))
	if err != nil {
		log.Printf("[ERROR] Error parsing CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Decode the CLI config file content.
	if err := hcl.DecodeObject(&config, obj); err != nil {
		log.Printf("[ERROR] Error decoding CLI config or credentials file %s: %v", configFilePath, err)
	}

	return config
}

func credentialsSource(config Config) auth.CredentialsSource {
	creds := auth.NoCredentials

	// Add all configured credentials to the credentials source.
	if len(config.Credentials) > 0 {
		staticTable := map[svchost.Hostname]map[string]interface{}{}
		for userHost, creds := range config.Credentials {
			host, err := svchost.ForComparison(userHost)
			if err != nil {
				// We expect the config was already validated by the time we get
				// here, so we'll just ignore invalid hostnames.
				continue
			}
			staticTable[host] = creds
		}
		creds = auth.StaticCredentialsSource(staticTable)
	}

	return creds
}

// checkConstraints checks service version constrains against our own
// version and returns rich and informational diagnostics in case any
// incompatibilities are detected.
func checkConstraints(c *disco.Constraints) error {
	if c == nil || c.Minimum == "" || c.Maximum == "" {
		return nil
	}

	// Generate a parsable constraints string.
	excluding := ""
	if len(c.Excluding) > 0 {
		excluding = fmt.Sprintf(", != %s", strings.Join(c.Excluding, ", != "))
	}
	constStr := fmt.Sprintf(">= %s%s, <= %s", c.Minimum, excluding, c.Maximum)

	// Create the constraints to check against.
	constraints, err := version.NewConstraint(constStr)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Create the version to check.
	v, err := version.NewVersion(providerVersion.ProviderVersion)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Return if we satisfy all constraints.
	if constraints.Check(v) {
		return nil
	}

	// Find out what action (upgrade/downgrade) we should advice.
	minimum, err := version.NewVersion(c.Minimum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	maximum, err := version.NewVersion(c.Maximum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	var excludes []*version.Version
	for _, exclude := range c.Excluding {
		v, err := version.NewVersion(exclude)
		if err != nil {
			return checkConstraintsWarning(err)
		}
		excludes = append(excludes, v)
	}

	// Sort all the excludes.
	sort.Sort(version.Collection(excludes))

	var action, toVersion string
	switch {
	case minimum.GreaterThan(v):
		action = "upgrade"
		toVersion = ">= " + minimum.String()
	case maximum.LessThan(v):
		action = "downgrade"
		toVersion = "<= " + maximum.String()
	case len(excludes) > 0:
		// Get the latest excluded version.
		action = "upgrade"
		toVersion = "> " + excludes[len(excludes)-1].String()
	}

	switch {
	case len(excludes) == 1:
		excluding = fmt.Sprintf(", excluding version %s", excludes[0].String())
	case len(excludes) > 1:
		var vs []string
		for _, v := range excludes {
			vs = append(vs, v.String())
		}
		excluding = fmt.Sprintf(", excluding versions %s", strings.Join(vs, ", "))
	default:
		excluding = ""
	}

	summary := fmt.Sprintf("Incompatible TFE provider version v%s", v.String())
	details := fmt.Sprintf(
		"The configured Terraform Enterprise backend is compatible with TFE provider\n"+
			"versions >= %s, <= %s%s.", c.Minimum, c.Maximum, excluding,
	)

	if action != "" && toVersion != "" {
		summary = fmt.Sprintf("Please %s the TFE provider to %s", action, toVersion)
	}

	// Return the customized and informational error message.
	return fmt.Errorf("%s\n\n%s", summary, details)
}

func checkConstraintsWarning(err error) error {
	return fmt.Errorf(
		"failed to check version constraints: %v\n\n"+
			"checking version constraints is considered optional, but this is an\n"+
			"unexpected error which should be reported",
		err,
	)
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
