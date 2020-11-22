package tfe

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-tfe/version"
	"github.com/hashicorp/terraform-svchost/disco"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"tfe": testAccProvider,
	}
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

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
var GITHUB_WORKSPACE_IDENTIFIER = os.Getenv("GITHUB_WORKSPACE_IDENTIFIER")
var GITHUB_WORKSPACE_BRANCH = os.Getenv("GITHUB_WORKSPACE_BRANCH")
var GITHUB_POLICY_SET_IDENTIFIER = os.Getenv("GITHUB_POLICY_SET_IDENTIFIER")
var GITHUB_POLICY_SET_BRANCH = os.Getenv("GITHUB_POLICY_SET_BRANCH")
var GITHUB_POLICY_SET_PATH = os.Getenv("GITHUB_POLICY_SET_PATH")
var GITHUB_REGISTRY_MODULE_IDENTIFIER = os.Getenv("GITHUB_REGISTRY_MODULE_IDENTIFIER")
var TFE_USER1 = os.Getenv("TFE_USER1")
var TFE_USER2 = os.Getenv("TFE_USER2")
