package tfe

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/svchost/disco"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-tfe/version"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"tfe": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
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
	if v := os.Getenv("GITHUB_TOKEN"); v == "" {
		t.Fatal("GITHUB_TOKEN must be set for acceptance tests")
	}
	// The credentials must be provided by the CLI config file for testing.
	if err := Provider().Configure(&terraform.ResourceConfig{}); err != nil {
		t.Fatalf("err: %s", err)
	}
}

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
