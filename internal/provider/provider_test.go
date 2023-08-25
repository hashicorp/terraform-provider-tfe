// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-tfe/version"
	"github.com/hashicorp/terraform-svchost/disco"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccMuxedProviders map[string]func() (tfprotov5.ProviderServer, error)

func init() {
	testAccProvider = Provider()

	testAccProviders = map[string]*schema.Provider{
		"tfe": testAccProvider,
	}
	testAccMuxedProviders = map[string]func() (tfprotov5.ProviderServer, error){
		"tfe": func() (tfprotov5.ProviderServer, error) {
			ctx := context.Background()
			nextProvider := providerserver.NewProtocol5(NewFrameworkProvider())

			mux, err := tf5muxserver.NewMuxServer(
				ctx, nextProvider, PluginProviderServer, testAccProvider.GRPCProvider,
			)
			if err != nil {
				return nil, err
			}

			return mux.ProviderServer(), nil
		},
	}
}

func providerWithDefaultOrganization(defaultOrgName string) map[string]*schema.Provider {
	testAccProviderDefaultOrganization := Provider()
	testAccProviderDefaultOrganization.ConfigureContextFunc = func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client, err := getClientUsingEnv()
		return ConfiguredClient{
			Client:       client,
			Organization: defaultOrgName,
		}, diag.FromErr(err)
	}
	return map[string]*schema.Provider{
		"tfe": testAccProviderDefaultOrganization,
	}
}

func setupDefaultOrganization(t *testing.T) (string, int) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	defaultOrgName := fmt.Sprintf("tst-default-org-%d", rInt)

	client, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	_, cleanup := createOrganization(t, client, tfe.OrganizationCreateOptions{
		Name:  &defaultOrgName,
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})

	t.Cleanup(cleanup)
	return defaultOrgName, rInt
}

func getClientUsingEnv() (*tfe.Client, error) {
	hostname := defaultHostname
	if os.Getenv("TFE_HOSTNAME") != "" {
		hostname = os.Getenv("TFE_HOSTNAME")
	}
	token := os.Getenv("TFE_TOKEN")

	client, err := getClient(hostname, token, defaultSSLSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("Error getting client: %w", err)
	}
	return client, nil
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func TestProvider_versionConstraints(t *testing.T) {
	cases := map[string]struct {
		constraints *disco.Constraints
		version     string
		result      string
	}{
		"compatible version": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "tfe-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.6.0",
		},
		"version too old": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "tfe-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.3.0",
			result:  "upgrade the TFE provider to >= 0.4.0",
		},
		"version too new": {
			constraints: &disco.Constraints{
				Service: "tfe.v2.1",
				Product: "tfe-provider",
				Minimum: "0.4.0",
				Maximum: "0.7.0",
			},
			version: "0.8.0",
			result:  "downgrade the TFE provider to <= 0.7.0",
		},
	}

	// Save and restore the actual version.
	v := version.ProviderVersion
	defer func() {
		version.ProviderVersion = v
	}()

	for name, tc := range cases {
		// Set the version for this test.
		version.ProviderVersion = tc.version

		err := checkConstraints(tc.constraints)
		if err == nil && tc.result != "" {
			t.Fatalf("%s: expected error to contain %q, but got no error", name, tc.result)
		}
		if err != nil && tc.result == "" {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		if err != nil && !strings.Contains(err.Error(), tc.result) {
			t.Fatalf("%s: expected error to contain %q, got: %v", name, tc.result, err)
		}
	}
}

func TestProvider_locateConfigFile(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalTfCliConfigFile := os.Getenv("TF_CLI_CONFIG_FILE")
	originalTerraformConfig := os.Getenv("TERRAFORM_CONFIG")
	reset := func() {
		os.Setenv("HOME", originalHome)
		if originalTfCliConfigFile != "" {
			os.Setenv("TF_CLI_CONFIG_FILE", originalTfCliConfigFile)
		} else {
			os.Unsetenv("TF_CLI_CONFIG_FILE")
		}
		if originalTerraformConfig != "" {
			os.Setenv("TERRAFORM_CONFIG", originalTerraformConfig)
		} else {
			os.Unsetenv("TERRAFORM_CONFIG")
		}
	}
	defer reset()

	// Use a predictable value for $HOME
	os.Setenv("HOME", "/Users/someone")

	setup := func(tfCliConfigFile, terraformConfig string) {
		os.Setenv("TF_CLI_CONFIG_FILE", tfCliConfigFile)
		os.Setenv("TERRAFORM_CONFIG", terraformConfig)
	}

	cases := map[string]struct {
		tfCliConfigFile string
		terraformConfig string
		result          string
	}{
		"has TF_CLI_CONFIG_FILE": {
			tfCliConfigFile: "~/.terraform_alternate/terraformrc",
			terraformConfig: "",
			result:          "~/.terraform_alternate/terraformrc",
		},
		"has TERRAFORM_CONFIG": {
			tfCliConfigFile: "",
			terraformConfig: "~/.terraform_alternate_rc",
			result:          "~/.terraform_alternate_rc",
		},
		"has both env vars": {
			tfCliConfigFile: "~/.from_TF_CLI",
			terraformConfig: "~/.from_TERRAFORM_CONFIG",
			result:          "~/.from_TF_CLI",
		},
		"has neither env var": {
			tfCliConfigFile: "",
			terraformConfig: "",
			result:          "/Users/someone/.terraformrc", // expect tests run on unix
		},
	}

	for name, tc := range cases {
		setup(tc.tfCliConfigFile, tc.terraformConfig)

		fileResult := locateConfigFile()
		if tc.result != fileResult {
			t.Fatalf("%s: expected config file at %s, got %s", name, tc.result, fileResult)
		}
	}
}

func TestProvider_cliConfig(t *testing.T) {
	// This only tests the behavior of merging various combinations of
	// (credentials file, .terraformrc file, absent). Locating the .terraformrc
	// file is tested separately.
	originalHome := os.Getenv("HOME")
	originalTfCliConfigFile := os.Getenv("TF_CLI_CONFIG_FILE")
	reset := func() {
		os.Setenv("HOME", originalHome)
		if originalTfCliConfigFile != "" {
			os.Setenv("TF_CLI_CONFIG_FILE", originalTfCliConfigFile)
		} else {
			os.Unsetenv("TF_CLI_CONFIG_FILE")
		}
	}
	defer reset()

	// Summary of fixtures: the credentials file and terraformrc file each have
	// credentials for two hosts, they both have credentials for app.terraform.io,
	// and the terraformrc also has one service discovery override.
	hasCredentials := "test-fixtures/cli-config-files/home"
	noCredentials := "test-fixtures/cli-config-files/no-credentials"
	terraformrc := "test-fixtures/cli-config-files/terraformrc"
	noTerraformrc := "test-fixtures/cli-config-files/no-terraformrc"
	tokenFromTerraformrc := "something.atlasv1.prod_rc_file"
	tokenFromCredentials := "something.atlasv1.prod_credentials_file"

	cases := map[string]struct {
		home             string
		rcfile           string
		expectCount      int
		expectProdToken  string
		expectHostsCount int
	}{
		"both main config and credentials JSON": {
			home:             hasCredentials,
			rcfile:           terraformrc,
			expectCount:      3,
			expectProdToken:  tokenFromTerraformrc,
			expectHostsCount: 1,
		},
		"only main config": {
			home:             noCredentials,
			rcfile:           terraformrc,
			expectCount:      2,
			expectProdToken:  tokenFromTerraformrc,
			expectHostsCount: 1,
		},
		"only credentials JSON": {
			home:             hasCredentials,
			rcfile:           noTerraformrc,
			expectCount:      2,
			expectProdToken:  tokenFromCredentials,
			expectHostsCount: 0,
		},
		"neither file": {
			home:             noCredentials,
			rcfile:           noTerraformrc,
			expectCount:      0,
			expectProdToken:  "",
			expectHostsCount: 0,
		},
	}

	for name, tc := range cases {
		os.Setenv("HOME", tc.home)
		os.Setenv("TF_CLI_CONFIG_FILE", tc.rcfile)
		config := cliConfig()
		credentialsCount := len(config.Credentials)
		if credentialsCount != tc.expectCount {
			t.Fatalf("%s: expected %d credentials, got %d", name, tc.expectCount, credentialsCount)
		}
		prodToken := ""
		if config.Credentials["app.terraform.io"] != nil {
			prodToken = config.Credentials["app.terraform.io"]["token"].(string)
		}
		if prodToken != tc.expectProdToken {
			t.Fatalf("%s: expected %s as prod token, got %s", name, tc.expectProdToken, prodToken)
		}
		hostsCount := len(config.Hosts)
		if hostsCount != tc.expectHostsCount {
			t.Fatalf("%s: expected %d `host` blocks in the final config, got %d", name, tc.expectHostsCount, hostsCount)
		}
	}
}

func testAccPreCheck(t *testing.T) {
	// The credentials must be provided by the CLI config file for testing.
	if diags := Provider().Configure(context.Background(), &terraform.ResourceConfig{}); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}
}

func TestConfigureEnvOrganization(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	originalTFEOrganization := os.Getenv("TFE_ORGANIZATION")
	reset := func() {
		if originalTFEOrganization != "" {
			os.Setenv("TFE_ORGANIZATION", originalTFEOrganization)
		} else {
			os.Unsetenv("TFE_ORGANIZATION")
		}
	}
	defer reset()

	expectedOrganization := fmt.Sprintf("tst-organization-%d", rInt)
	os.Setenv("TFE_ORGANIZATION", expectedOrganization)

	provider := Provider()

	// The credentials must be provided by the CLI config file for testing.
	if diags := provider.Configure(context.Background(), &terraform.ResourceConfig{}); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}

	config := provider.Meta().(ConfiguredClient)
	if config.Organization != expectedOrganization {
		t.Fatalf("unexpected organization configuration: got %s, wanted %s", config.Organization, expectedOrganization)
	}
}

// The TFE Provider tests use these environment variables, which are set in the
// GitHub Action workflow file .github/workflows/ci.yml.
func testAccGithubPreCheck(t *testing.T) {
	if envGithubToken == "" {
		t.Skip("Please set GITHUB_TOKEN to run this test")
	}
	if envGithubWorkspaceIdentifier == "" {
		t.Skip("Please set GITHUB_WORKSPACE_IDENTIFIER to run this test")
	}
	if envGithubWorkspaceBranch == "" {
		t.Skip("Please set GITHUB_WORKSPACE_BRANCH to run this test")
	}
}

func testAccGHAInstallationPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if envGithubAppInstallationID == "" {
		t.Skip("Please set GITHUB_APP_INSTALLATION_ID to run this test")
	}
}

func init() {
	envGithubPolicySetIdentifier = os.Getenv("GITHUB_POLICY_SET_IDENTIFIER")
	envGithubPolicySetBranch = os.Getenv("GITHUB_POLICY_SET_BRANCH")
	envGithubPolicySetPath = os.Getenv("GITHUB_POLICY_SET_PATH")
	envGithubRegistryModuleIdentifer = os.Getenv("GITHUB_REGISTRY_MODULE_IDENTIFIER")
	envGithubToken = os.Getenv("GITHUB_TOKEN")
	envGithubAppInstallationID = os.Getenv("GITHUB_APP_INSTALLATION_ID")
	envGithubAppInstallationName = os.Getenv("GITHUB_APP_INSTALLATION_NAME")
	envGithubWorkspaceIdentifier = os.Getenv("GITHUB_WORKSPACE_IDENTIFIER")
	envGithubWorkspaceBranch = os.Getenv("GITHUB_WORKSPACE_BRANCH")
	envTFEUser1 = os.Getenv("TFE_USER1")
	envTFEUser2 = os.Getenv("TFE_USER2")
}

var envGithubPolicySetIdentifier string
var envGithubPolicySetBranch string
var envGithubPolicySetPath string
var envGithubRegistryModuleIdentifer string
var envGithubToken string
var envGithubAppInstallationID string
var envGithubAppInstallationName string
var envGithubWorkspaceIdentifier string
var envGithubWorkspaceBranch string
var envTFEUser1 string
var envTFEUser2 string
