package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
			{
				// Query configuration with no org set
				Config: testAccTFEProviderSetList_query_no_org(),
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

// TestAccTFEProviderSetList_ExactLength verifies the exact count of provider sets in a
// freshly created, isolated organization. Using ExpectLength (not AtLeast) makes the
// assertion tight and independent of any pre-existing provider sets.
func TestAccTFEProviderSetList_ExactLength(t *testing.T) {
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
				Config: testAccTFEProviderSetList_setup(org.Name),
			},
			{
				Config: testAccTFEProviderSetList_query_with_org(org.Name),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					// Only the three sets created by setup exist in this fresh org.
					querycheck.ExpectLength("tfe_provider_set.test", 3),
				},
			},
		},
	})
}

// TestAccTFEProviderSetList_Empty verifies that querying an organization that has no
// provider sets returns an empty result set rather than an error.
func TestAccTFEProviderSetList_Empty(t *testing.T) {
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
				// No setup step — the org is empty.
				Config: testAccTFEProviderSetList_query_with_org(org.Name),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("tfe_provider_set.test", 0),
				},
			},
		},
	})
}

// TestAccTFEProviderSetList_InvalidOrg verifies that listing provider sets for an
// organization that does not exist surfaces a diagnostic error rather than silently
// returning an empty list.
func TestAccTFEProviderSetList_InvalidOrg(t *testing.T) {
	t.Parallel()
	skipUnlessBeta(t)

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
				Config:      testAccTFEProviderSetList_query_with_org("org-does-not-exist-" + randomString(t)),
				Query:       true,
				ExpectError: regexp.MustCompile("Error Retrieving Provider Sets"),
			},
		},
	})
}

// TestAccTFEProviderSetList_IncludeResource verifies the IncludeResource code path in
// the List function. When include_resource = true, each result must carry a fully
// populated resource model, including provider_source and the global flag.
func TestAccTFEProviderSetList_IncludeResource(t *testing.T) {
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
				Config: testAccTFEProviderSetList_setup(org.Name),
			},
			{
				Config: testAccTFEProviderSetList_query_include_resource(org.Name),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("tfe_provider_set.test", 3),
					// provider-set-one: aws, not global.
					querycheck.ExpectResourceKnownValues(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-one")),
						[]querycheck.KnownValueCheck{
							{Path: tfjsonpath.New("provider_source"), KnownValue: knownvalue.StringExact("registry.terraform.io/hashicorp/aws")},
							{Path: tfjsonpath.New("global"), KnownValue: knownvalue.Bool(false)},
						},
					),
					// provider-set-two: google, not global.
					querycheck.ExpectResourceKnownValues(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-two")),
						[]querycheck.KnownValueCheck{
							{Path: tfjsonpath.New("provider_source"), KnownValue: knownvalue.StringExact("registry.terraform.io/hashicorp/google")},
							{Path: tfjsonpath.New("global"), KnownValue: knownvalue.Bool(false)},
						},
					),
					// provider-set-three: azurerm and global=true.
					querycheck.ExpectResourceKnownValues(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-three")),
						[]querycheck.KnownValueCheck{
							{Path: tfjsonpath.New("provider_source"), KnownValue: knownvalue.StringExact("registry.terraform.io/hashicorp/azurerm")},
							{Path: tfjsonpath.New("global"), KnownValue: knownvalue.Bool(true)},
						},
					),
				},
			},
		},
	})
}

// TestAccTFEProviderSetList_IncludeResource_Description verifies that an optional
// description field is correctly surfaced through the list's resource model.
func TestAccTFEProviderSetList_IncludeResource_Description(t *testing.T) {
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
				Config: testAccTFEProviderSetList_setup_with_description(org.Name),
			},
			{
				Config: testAccTFEProviderSetList_query_include_resource(org.Name),
				Query:  true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectResourceKnownValues(
						"tfe_provider_set.test",
						queryfilter.ByDisplayName(knownvalue.StringExact("provider-set-described")),
						[]querycheck.KnownValueCheck{
							{Path: tfjsonpath.New("description"), KnownValue: knownvalue.StringExact("A described provider set")},
						},
					),
				},
			},
		},
	})
}

// --- config helpers ---
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

func testAccTFEProviderSetList_query_no_org() string {
	return `
list "tfe_provider_set" "test" {
	provider = tfe

	config {}
}
`
}

func testAccTFEProviderSetList_query_with_org(org string) string {
	return fmt.Sprintf(`
list "tfe_provider_set" "test" {
	provider = tfe

	config {
		organization_name = %q
	}
}
`, org)
}

func testAccTFEProviderSetList_query_include_resource(org string) string {
	return fmt.Sprintf(`
list "tfe_provider_set" "test" {
	provider = tfe

	include_resource = true

	config {
		organization_name = %q
	}
}
`, org)
}

func testAccTFEProviderSetList_setup_with_description(organization string) string {
	return fmt.Sprintf(`
locals {
	organization_name = "%s"
}

resource "tfe_provider_set" "described" {
	name                = "provider-set-described"
	organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/aws"
	description         = "A described provider set"
	global              = false
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}
`, organization)
}
