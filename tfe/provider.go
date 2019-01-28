package tfe

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/svchost"
	"github.com/hashicorp/terraform/svchost/auth"
	"github.com/hashicorp/terraform/svchost/disco"
	"github.com/hashicorp/terraform/terraform"
)

const defaultHostname = "app.terraform.io"

var serviceIDs = []string{"tfe.v2.1", "tfe.v2"}

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
				Default:     defaultHostname,
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["token"],
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
			"tfe_oauth_client":       resourceTFEOAuthClient(),
			"tfe_organization":       resourceTFEOrganization(),
			"tfe_organization_token": resourceTFEOrganizationToken(),
			"tfe_policy_set":         resourceTFEPolicySet(),
			"tfe_sentinel_policy":    resourceTFESentinelPolicy(),
			"tfe_ssh_key":            resourceTFESSHKey(),
			"tfe_team":               resourceTFETeam(),
			"tfe_team_access":        resourceTFETeamAccess(),
			"tfe_team_member":        resourceTFETeamMember(),
			"tfe_team_members":       resourceTFETeamMembers(),
			"tfe_team_token":         resourceTFETeamToken(),
			"tfe_workspace":          resourceTFEWorkspace(),
			"tfe_variable":           resourceTFEVariable(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Get the hostname and token.
	hostname := d.Get("hostname").(string)
	token := d.Get("token").(string)

	// Parse the hostname for comparison,
	host, err := svchost.ForComparison(hostname)
	if err != nil {
		return nil, err
	}

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config)
	services := disco.NewWithCredentialsSource(credsSrc)
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

	// Discover the full Terraform Enterprise service address.
	var address *url.URL
	for _, serviceID := range serviceIDs {
		addr, err := services.DiscoverServiceURL(host, serviceID)
		if err != nil {
			return nil, err
		}
		if addr != nil {
			address = addr
			break
		}
	}

	// Check if we were able to discover a service address.
	if address == nil {
		return nil, fmt.Errorf("host %s does not provide a Terraform Enterprise API", host)
	}

	// Only try to get to the token from the credentials source if no token
	// was explicitly set in the provider configuration.
	if token == "" {
		creds, err := services.CredentialsForHost(host)
		if err != nil {
			log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", host, err)
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

	// Create s new TFE client.
	return tfe.NewClient(cfg)
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

var descriptions = map[string]string{
	"hostname": "The Terraform Enterprise hostname to connect to. Defaults to app.terraform.io.",
	"token": "The token used to authenticate with Terraform Enterprise. We recommend omitting\n" +
		"the token which can be set as credentials in the CLI config file.",
}
