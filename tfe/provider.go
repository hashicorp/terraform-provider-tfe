package tfe

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
	providerVersion "github.com/terraform-providers/terraform-provider-tfe/version"
)

const defaultHostname = "app.terraform.io"

var tfeServiceIDs = []string{"tfe.v2.2"}

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

// ctx is used as default context.Context when making TFE calls.
var ctx = context.Background()

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["hostname"],
				DefaultFunc: schema.EnvDefaultFunc("TFE_HOSTNAME", defaultHostname),
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["token"],
				DefaultFunc: schema.EnvDefaultFunc("TFE_TOKEN", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"tfe_ssh_key":       dataSourceTFESSHKey(),
			"tfe_team":          dataSourceTFETeam(),
			"tfe_team_access":   dataSourceTFETeamAccess(),
			"tfe_workspace":     dataSourceTFEWorkspace(),
			"tfe_workspace_ids": dataSourceTFEWorkspaceIDs(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"tfe_notification_configuration": resourceTFENotificationConfiguration(),
			"tfe_oauth_client":               resourceTFEOAuthClient(),
			"tfe_organization":               resourceTFEOrganization(),
			"tfe_organization_membership":    resourceTFEOrganizationMembership(),
			"tfe_organization_token":         resourceTFEOrganizationToken(),
			"tfe_policy_set":                 resourceTFEPolicySet(),
			"tfe_policy_set_parameter":       resourceTFEPolicySetParameter(),
			"tfe_run_trigger":                resourceTFERunTrigger(),
			"tfe_sentinel_policy":            resourceTFESentinelPolicy(),
			"tfe_ssh_key":                    resourceTFESSHKey(),
			"tfe_team":                       resourceTFETeam(),
			"tfe_team_access":                resourceTFETeamAccess(),
			"tfe_team_organization_member":   resourceTFETeamOrganizationMember(),
			"tfe_team_member":                resourceTFETeamMember(),
			"tfe_team_members":               resourceTFETeamMembers(),
			"tfe_team_token":                 resourceTFETeamToken(),
			"tfe_workspace":                  resourceTFEWorkspace(),
			"tfe_variable":                   resourceTFEVariable(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Parse the hostname for comparison,
	hostname, err := svchost.ForComparison(d.Get("hostname").(string))
	if err != nil {
		return nil, err
	}

	providerUaString := fmt.Sprintf("terraform-provider-tfe/%s", providerVersion.ProviderVersion)

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config)
	services := disco.NewWithCredentialsSource(credsSrc)
	services.SetUserAgent(providerUaString)
	services.Transport = logging.NewTransport("TFE Discovery", services.Transport)

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

	// Get the token from the config.
	token := d.Get("token").(string)

	// Only try to get to the token from the credentials source if no token
	// was explicitly set in the provider configuration.
	if token == "" {
		creds, err := services.CredentialsForHost(hostname)
		if err != nil {
			log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
		}
		if creds != nil {
			token = creds.Token()
		}
	}

	// If we still don't have a token at this point, we return an error.
	if token == "" {
		return nil, fmt.Errorf("required token could not be found")
	}

	httpClient := tfe.DefaultConfig().HTTPClient
	httpClient.Transport = logging.NewTransport("TFE", httpClient.Transport)

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
func cliConfig() *Config {
	config := &Config{}

	// Detect the CLI config file path.
	configFilePath := os.Getenv("TERRAFORM_CONFIG")
	if configFilePath == "" {
		filePath, err := configFile()
		if err != nil {
			log.Printf("[ERROR] Error detecting default CLI config file path: %s", err)
			return config
		}
		configFilePath = filePath
	}

	// Read the CLI config file content.
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Printf("[ERROR] Error reading the CLI config file %s: %v", configFilePath, err)
		return config
	}

	// Parse the CLI config file content.
	obj, err := hcl.Parse(string(content))
	if err != nil {
		log.Printf("[ERROR] Error parsing the CLI config file %s: %v", configFilePath, err)
		return config
	}

	// Decode the CLI config file content.
	if err := hcl.DecodeObject(&config, obj); err != nil {
		log.Printf("[ERROR] Error decoding the CLI config file %s: %v", configFilePath, err)
	}

	return config
}

func credentialsSource(config *Config) auth.CredentialsSource {
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
		"Failed to check version constraints: %v\n\n"+
			"Checking version constraints is considered optional, but this is an\n"+
			"unexpected error which should be reported.",
		err,
	)
}

var descriptions = map[string]string{
	"hostname": "The Terraform Enterprise hostname to connect to. Defaults to app.terraform.io.",
	"token": "The token used to authenticate with Terraform Enterprise. We recommend omitting\n" +
		"the token which can be set as credentials in the CLI config file.",
}
