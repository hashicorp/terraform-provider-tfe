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
	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-tfe/internal/client"
	"github.com/hashicorp/terraform-provider-tfe/version"
	"github.com/hashicorp/terraform-svchost/disco"
)

var (
	testAccMuxedProviders   map[string]func() (tfprotov5.ProviderServer, error)
	testAccConfiguredClient *ConfiguredClient
)

func init() {
	testAccMuxedProviders = map[string]func() (tfprotov5.ProviderServer, error){
		"tfe": func() (tfprotov5.ProviderServer, error) {
			ctx := context.Background()
			nextProvider := providerserver.NewProtocol5(NewFrameworkProvider())

			sdkProvider := Provider()
			sdkProvider.ConfigureContextFunc = func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
				client, err := getClientUsingEnv()
				cc := ConfiguredClient{
					Client: client,
				}

				// Save a reference to the configured client instance for use in tests.
				testAccConfiguredClient = &cc

				return cc, diag.FromErr(err)
			}

			mux, err := tf5muxserver.NewMuxServer(
				ctx, nextProvider, sdkProvider.GRPCProvider,
			)
			if err != nil {
				return nil, err
			}

			return mux.ProviderServer(), nil
		},
	}
}

func muxedProvidersWithDefaultOrganization(defaultOrgName string) map[string]func() (tfprotov5.ProviderServer, error) {
	sdkProvider := Provider()
	sdkProvider.ConfigureContextFunc = func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client, err := getClientUsingEnv()
		cc := ConfiguredClient{
			Client:       client,
			Organization: defaultOrgName,
		}

		// Save a reference to the configured client instance for use in tests.
		testAccConfiguredClient = &cc

		return cc, diag.FromErr(err)
	}
	return map[string]func() (tfprotov5.ProviderServer, error){
		"tfe": func() (tfprotov5.ProviderServer, error) {
			ctx := context.Background()

			nextProvider := providerserver.NewProtocol5(
				NewFrameworkProviderWithDefaultOrg(defaultOrgName),
			)

			mux, err := tf5muxserver.NewMuxServer(
				ctx, nextProvider, sdkProvider.GRPCProvider,
			)
			if err != nil {
				return nil, err
			}

			return mux.ProviderServer(), nil
		},
	}
}

func setupDefaultOrganization(t *testing.T) (string, int) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	defaultOrgName := fmt.Sprintf("tst-default-org-%d", rInt)

	testClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	_, cleanup := createOrganization(t, testClient, tfe.OrganizationCreateOptions{
		Name:  &defaultOrgName,
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})

	t.Cleanup(cleanup)
	return defaultOrgName, rInt
}

func getClientUsingEnv() (*tfe.Client, error) {
	hostname := client.DefaultHostname
	if os.Getenv("TFE_HOSTNAME") != "" {
		hostname = os.Getenv("TFE_HOSTNAME")
	}
	token := os.Getenv("TFE_TOKEN")

	tfeClient, err := client.GetClient(hostname, token, defaultSSLSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("Error getting client: %w", err)
	}
	return tfeClient, nil
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

		err := client.CheckConstraints(tc.constraints)
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

func testAccPreCheck(t *testing.T) {
	// This is currently a no-op.
}

func TestSkipUnlessAfterDate(t *testing.T) {
	skipUnlessAfterDate(t, time.Date(2199, 1, 1, 0, 0, 0, 0, time.UTC))
	t.Fatal("This test should have been skipped (Unless it's 2199!)")
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
	if diags := provider.Configure(context.Background(), &sdkTerraform.ResourceConfig{}); diags.HasError() {
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
