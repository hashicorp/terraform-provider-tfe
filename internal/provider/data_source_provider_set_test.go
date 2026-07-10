package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEProviderSetDataSource_read(t *testing.T) {
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

	providerSetName := "tst-provider-set-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSetDataSource_basic_without_data(providerSetName, org.Name),
			},
			{
				Config: testAccTFEProviderSetDataSource_basic_with_data(providerSetName, org.Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "name", providerSetName,
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "description", "Provider Set description",
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "project_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "workspace_ids.#", "1",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSetDataSource_read_uninitialized_provider_set(t *testing.T) {
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

	providerSetName := "tst-provider-set-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProviderSetDataSource_minimal_with_data(providerSetName, &org.Name),
				ExpectError: regexp.MustCompile("resource not found"),
			},
		},
	})
}

func TestAccTFEProviderSetDataSource_read_provider_set_deleted(t *testing.T) {
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

	providerSetName := "tst-provider-set-" + randomString(t)
	providerSetId := "uninitialized"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSetDataSource_minimal_without_data(providerSetName, &org.Name),
			},
			{
				Config: testAccTFEProviderSetDataSource_minimal_with_data(providerSetName, &org.Name),
				PreConfig: func() {
					providerSetFromApi, err := tfeClient.ProviderSets.ReadByName(context.Background(), org.Name, providerSetName)
					if err != nil {
						panic(err)
					}
					providerSetId = providerSetFromApi.ID
				},
			},
			{
				Config: testAccTFEProviderSetDataSource_minimal_with_data(providerSetName, &org.Name),
				Check: func(s *terraform.State) error {
					return resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "id", providerSetId,
					)(s)
				},
			},
			{
				PreConfig: func() {
					err := tfeClient.ProviderSets.Delete(context.Background(), providerSetId)
					if err != nil {
						panic(err)
					}
				},
				Config:      testAccTFEProviderSetDataSource_minimal_with_data(providerSetName, &org.Name),
				ExpectError: regexp.MustCompile("resource not found"),
			},
		},
	})
}

func TestAccTFEProviderSetDataSource_read_with_default_organization(t *testing.T) {
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

	providerSetName := "tst-provider-set-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// when organization is not specified in the config, the provider should use
		// the default organization from the provider configuration, so we need to mux
		// the providers to include that default organization
		ProtoV6ProviderFactories: muxedProvidersWithDefaultOrganization(org.Name),
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSetDataSource_minimal_without_data(providerSetName, nil),
			},
			{
				Config: testAccTFEProviderSetDataSource_minimal_with_data(providerSetName, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "name", providerSetName,
					),
					resource.TestCheckResourceAttr(
						"data.tfe_provider_set.foobar", "organization", org.Name,
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSetDataSource_validation(t *testing.T) {
	skipUnlessBeta(t)

	providerSetName := "tst-provider-set"
	organizationName := "tst-organization"
	invalidOrganization := "invalid/org"
	invalidName := "invalid/name"
	emptyString := ""

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProviderSetDataSource_only_data(&providerSetName, nil),
				ExpectError: regexp.MustCompile("No organization was specified on the resource or provider"),
			},
			{
				Config:      testAccTFEProviderSetDataSource_only_data(&providerSetName, &invalidOrganization),
				ExpectError: regexp.MustCompile("invalid value for organization"),
			},
			{
				Config:      testAccTFEProviderSetDataSource_only_data(&providerSetName, &emptyString),
				ExpectError: regexp.MustCompile("invalid value for organization"),
			},
			{
				Config:      testAccTFEProviderSetDataSource_only_data(nil, &organizationName),
				ExpectError: regexp.MustCompile("The argument \"name\" is required, but no definition was found"),
			},
			{
				Config:      testAccTFEProviderSetDataSource_only_data(&invalidName, &organizationName),
				ExpectError: regexp.MustCompile("invalid value for name"),
			},
			{
				Config:      testAccTFEProviderSetDataSource_only_data(&emptyString, &organizationName),
				ExpectError: regexp.MustCompile("name is required"),
			},
		},
	})
}

func testAccTFEProviderSetDataSource_basic_without_data(name string, organization string) string {
	return fmt.Sprintf(`
locals {
  name         = %q
  organization = %q
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization
}

resource "tfe_project" "foo" {
  name         = "project-foo"
  organization = local.organization
}

resource "tfe_provider_set" "foobar" {
  name                = local.name
  description         = "Provider Set description"
  organization        = local.organization
  provider_source     = "registry.terraform.io/hashicorp/aws"
  global              = false
  provider_config_hcl = <<-EOT
provider "aws" {
  region = "us-east-1"
}
EOT

  project_ids =   [ tfe_project.foo.id ]
  workspace_ids = [ tfe_workspace.foo.id ]
}`, name, organization)
}

func testAccTFEProviderSetDataSource_basic_with_data(name string, organization string) string {
	return providerSetDataSource_add_data(
		testAccTFEProviderSetDataSource_basic_without_data(name, organization),
		&name,
		&organization,
	)
}

func testAccTFEProviderSetDataSource_minimal_without_data(name string, organization *string) string {
	organizationStr := "// omitted organization since value is nil"
	if organization != nil {
		organizationStr = fmt.Sprintf("organization = %q", *organization)
	}

	return fmt.Sprintf(`
resource "tfe_provider_set" "foobar" {
  name                = %q
  description         = "Provider Set description"
  %s
  provider_source     = "registry.terraform.io/hashicorp/aws"
  global              = true
  provider_config_hcl = <<-EOT
provider "aws" {
  region = "us-east-1"
}
EOT
}`, name, organizationStr)
}

func testAccTFEProviderSetDataSource_minimal_with_data(name string, organization *string) string {
	return providerSetDataSource_add_data(
		testAccTFEProviderSetDataSource_minimal_without_data(name, organization),
		&name,
		organization,
	)
}

func testAccTFEProviderSetDataSource_only_data(name *string, organization *string) string {
	return providerSetDataSource_add_data("", name, organization)
}

func providerSetDataSource_add_data(template string, name, organization *string) string {
	nameStr := "// omitted name since value is nil"
	organizationStr := "// omitted organization since value is nil"

	if name != nil {
		nameStr = fmt.Sprintf("name = %q", *name)
	}
	if organization != nil {
		organizationStr = fmt.Sprintf("organization = %q", *organization)
	}

	return fmt.Sprintf(`
%s

data "tfe_provider_set" "foobar" {
  %s
  %s
}
`, template, nameStr, organizationStr)
}
