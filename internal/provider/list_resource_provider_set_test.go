package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEProviderSetList_QueryCheck(t *testing.T) {
	t.Parallel()
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"tfe": providerserver.NewProtocol6WithError(NewFrameworkProvider()),
		},
		Steps: []resource.TestStep{
			{
				// Create three provider sets for querying
				Config: testAccTFEProviderSetList_setup(org.Name),
			},
			{
				// Query configuration to list provider sets
				Config: testAccTFEProviderSetList_query(),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("tfe_provider_set.test", 3),
					querycheck.ExpectResourceDisplayName(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-one")),
						knownvalue.StringExact("provider-set-one"),
					),
					querycheck.ExpectResourceDisplayName(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-two")),
						knownvalue.StringExact("provider-set-two"),
					),
					querycheck.ExpectResourceDisplayName(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-three")),
						knownvalue.StringExact("provider-set-three"),
					),
				},
			},
		},
	})
}

func testAccTFEProviderSetList_setup(organization string) string {
	return fmt.Sprintf(`
locals {
	organization_name = "%s"
}

resource "tfe_provider_set" "one" {
	name                = "provider-set-one"
	organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/aws"
	global              = false
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}

resource "tfe_provider_set" "two" {
	name                = "provider-set-two"
	organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/google"
	global              = false
	provider_config_hcl = <<-EOT
provider "google" {
	region = "us-central1"
}
EOT
}

resource "tfe_provider_set" "three" {
	name                = "provider-set-three"
	organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/azurerm"
	global              = true
	provider_config_hcl = <<-EOT
provider "azurerm" {
	features {}
}
EOT
}
`, organization)
}

func testAccTFEProviderSetList_query() string {
	return `
list "tfe_provider_set" "test" {
	provider = tfe

	config {
		organization_name = local.organization_name
	}
}
`
}
