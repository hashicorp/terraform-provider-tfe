// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform-provider-tfe/internal/logging"
	providerVersion "github.com/hashicorp/terraform-provider-tfe/version"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
)

var (
	// TFEUserAgent is the user agent string sent with all requests made by the provider
	TFEUserAgent = fmt.Sprintf("terraform-provider-tfe/%s", providerVersion.ProviderVersion)
)

type CredentialsMap map[string]map[string]interface{}

// CLIHostConfig is the structure of the configuration for the Terraform CLI.
type CLIHostConfig struct {
	Hosts       map[string]*ConfigHost `hcl:"host"`
	Credentials CredentialsMap         `hcl:"credentials"`
}

// ConfigHost is the structure of the "host" nested block within the CLI
// configuration, which can be used to override the default service host
// discovery behavior for a particular hostname.
type ConfigHost struct {
	Services map[string]interface{} `hcl:"services"`
}

// ClientConfiguration is the refined information needed to configure a tfe.Client
type ClientConfiguration struct {
	Services   *disco.Disco
	HTTPClient *http.Client
	TFEHost    svchost.Hostname
	Token      string
	Insecure   bool
}

// Key returns a string that is comparable to other ClientConfiguration values
func (c ClientConfiguration) Key() string {
	return fmt.Sprintf("%x %s/%v", sha256.New().Sum([]byte(c.Token)), c.TFEHost, c.Insecure)
}

// cliConfig tries to find and parse the configuration of the Terraform CLI.
// This is an optional step, so any errors are ignored.
func cliConfig() CLIHostConfig {
	mainConfig := CLIHostConfig{}
	credentialsConfig := CLIHostConfig{}
	combinedConfig := CLIHostConfig{}

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

// All the errors returned by the helper methods called in this function get ignored (down the road we throw an error when all auth methods have failed.) We only use these errors to log warnings to the user.
func readCliConfigFile(configFilePath string) CLIHostConfig {
	config := CLIHostConfig{}

	// Read the CLI config file content.
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("[WARN] Unable to read CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Parse the CLI config file content.
	obj, err := hcl.Parse(string(content))
	if err != nil {
		log.Printf("[WARN] Unable to parse CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Decode the CLI config file content.
	if err := hcl.DecodeObject(&config, obj); err != nil {
		log.Printf("[WARN] Unable to decode CLI config or credentials file %s: %v", configFilePath, err)
	}

	return config
}

func credentialsSource(credentials CredentialsMap) auth.CredentialsSource {
	creds := auth.NoCredentials

	// Add all configured credentials to the credentials source.
	if len(credentials) > 0 {
		staticTable := map[svchost.Hostname]map[string]interface{}{}
		for userHost, creds := range credentials {
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

// configure accepts the provider-level configuration values and creates a
// clientConfiguration using fallback values from the environment or CLI configuration.
func configure(tfeHost, token string, insecure bool) (*ClientConfiguration, error) {
	if tfeHost == "" {
		if os.Getenv("TFE_HOSTNAME") != "" {
			tfeHost = os.Getenv("TFE_HOSTNAME")
		} else {
			tfeHost = DefaultHostname
		}
	}
	log.Printf("[DEBUG] Configuring client for host %q", tfeHost)

	// If ssl_skip_verify is false, it is either set that way in configuration or unset. Check
	// the environment to see if it was set to true there.  Strictly speaking, this means that
	// the env var can override an explicit 'false' in configuration (which is not true of the
	// other settings), but that's how it goes with a boolean zero value.
	var err error
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

	// Parse the hostname for comparison,
	hostname, err := svchost.ForComparison(tfeHost)
	if err != nil {
		return nil, err
	}

	httpClient := tfe.DefaultConfig().HTTPClient

	// Make sure the transport has a TLS config.
	transport := httpClient.Transport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	transport.TLSClientConfig.InsecureSkipVerify = insecure

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config.Credentials)
	services := disco.NewWithCredentialsSource(credsSrc)
	services.SetUserAgent(TFEUserAgent)
	services.Transport = logging.NewLoggingTransport("TFE", transport)

	// Add any static host configurations service discovery object.
	for userHost, hostConfig := range config.Hosts {
		host, err := svchost.ForComparison(userHost)
		if err != nil {
			// ignore invalid hostnames.
			continue
		}
		services.ForceHostServices(host, hostConfig.Services)
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
		return nil, ErrMissingAuthToken
	}

	return &ClientConfiguration{
		Services:   services,
		HTTPClient: httpClient,
		TFEHost:    hostname,
		Token:      token,
		Insecure:   insecure,
	}, nil
}
