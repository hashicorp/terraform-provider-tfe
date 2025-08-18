// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOutputsEphemeralResource_basic(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	assertPathValues := tfjsonpath.New("data").AtMapKey("values")

	waitForOutputs(t, tfeClient, orgName, wsName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOutputsEphemeralResource(rInt, orgName, wsName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("tfe_organization.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("tst-%d", rInt))),

					statecheck.ExpectKnownValue("tfe_workspace.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("workspace-test-%d", rInt))),

					// These outputs rely on the values in test-fixtures/state-versions/terraform.tfstate
					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("us-west-1a"),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_string"),
						knownvalue.StringExact("9023256633839603543")),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_tuple_number"),
						knownvalue.TupleExact([]knownvalue.Check{
							knownvalue.Int32Exact(1),
							knownvalue.Int32Exact(2),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_tuple_string"),
						knownvalue.TupleExact([]knownvalue.Check{
							knownvalue.StringExact("one"),
							knownvalue.StringExact("two"),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_object"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"foo":         knownvalue.StringExact("bar"),
							"environment": knownvalue.Null(),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_number"),
						knownvalue.Int32Exact(5)),

					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_bool"),
						knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccOutputsEphemeralResource_readAllNonSensitiveValues(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	assertPathValues := tfjsonpath.New("data").AtMapKey("values")
	assertPathNonsensitiveValues := tfjsonpath.New("data").AtMapKey("nonsensitive_values")

	waitForOutputs(t, tfeClient, orgName, wsName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOutputsEphemeralResource(rInt, orgName, wsName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("tfe_organization.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("tst-%d", rInt))),

					statecheck.ExpectKnownValue("tfe_workspace.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("workspace-test-%d", rInt))),

					// These outputs rely on the values in test-fixtures/state-versions/terraform.tfstate
					statecheck.ExpectKnownValue("echo.this",
						assertPathValues.AtMapKey("test_output_string"),
						knownvalue.StringExact("9023256633839603543")),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_list_string"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("us-west-1a"),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_tuple_number"),
						knownvalue.TupleExact([]knownvalue.Check{
							knownvalue.Int32Exact(1),
							knownvalue.Int32Exact(2),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_tuple_string"),
						knownvalue.TupleExact([]knownvalue.Check{
							knownvalue.StringExact("one"),
							knownvalue.StringExact("two"),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_object"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"foo":         knownvalue.StringExact("bar"),
							"environment": knownvalue.Null(),
						})),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_number"),
						knownvalue.Int32Exact(5)),

					statecheck.ExpectKnownValue("echo.this",
						assertPathNonsensitiveValues.AtMapKey("test_output_bool"),
						knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccOutputsEphemeralResource_emptyOutputs(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform-empty-outputs.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	assertPathValues := tfjsonpath.New("data").AtMapKey("values")
	assertPathNonsensitiveValues := tfjsonpath.New("data").AtMapKey("nonsensitive_values")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOutputsEphemeralResource(rInt, orgName, wsName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("tfe_organization.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("tst-%d", rInt))),

					statecheck.ExpectKnownValue("tfe_workspace.this", tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("workspace-test-%d", rInt))),

					// This relies on test-fixtures/state-versions/terraform-empty-outputs.tfstate
					statecheck.ExpectKnownValue("echo.this", assertPathValues,
						knownvalue.ObjectExact(map[string]knownvalue.Check{})),

					statecheck.ExpectKnownValue("echo.this", assertPathNonsensitiveValues,
						knownvalue.ObjectExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccOutputsEphemeralResource(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "this" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "this" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.this.name
}

ephemeral "tfe_outputs" "this" {
  organization = "%s"
  workspace = "%s"
}

provider "echo" {
  data = ephemeral.tfe_outputs.this
}

resource "echo" "this" {}
`, rInt, rInt, org, workspace)
}
