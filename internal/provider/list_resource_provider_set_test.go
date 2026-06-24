package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
)

func TestAccTFEProviderSetList_QueryCheck(t *testing.T) {
	t.Parallel()
	//skipUnlessBeta(t)
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				// Create three provider sets for querying
				Config: testAccTFEProviderSetList_setup(org.Name),
			},
			{
				// Query configuration to list provider sets
				Config: testAccTFEProviderSetList_query(org.Name),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity(
						"tfe_provider_set_list.test",
						map[string]knownvalue.Check{
							"name":            knownvalue.StringExact("provider-set-one"),
							"provider_source": knownvalue.StringExact("registry.terraform.io/hashicorp/aws"),
							"global":          knownvalue.Bool(false),
						},
					),
					querycheck.ExpectIdentity(
						"tfe_provider_set_list.test",
						map[string]knownvalue.Check{
							"name":            knownvalue.StringExact("provider-set-two"),
							"provider_source": knownvalue.StringExact("registry.terraform.io/hashicorp/google"),
							"global":          knownvalue.Bool(false),
						},
					),
					querycheck.ExpectIdentity(
						"tfe_provider_set_list.test",
						map[string]knownvalue.Check{
							"name":            knownvalue.StringExact("provider-set-three"),
							"provider_source": knownvalue.StringExact("registry.terraform.io/hashicorp/azurerm"),
							"global":          knownvalue.Bool(true),
						},
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

func testAccTFEProviderSetList_query(organization string) string {
	return fmt.Sprintf(`
		locals {
			organization_name = "%s"
		}

		list "tfe_provider_set_list" "test" {
			provider = tfe

			config {
				organization_name = local.organization_name
			}
		}
`, organization)
}
