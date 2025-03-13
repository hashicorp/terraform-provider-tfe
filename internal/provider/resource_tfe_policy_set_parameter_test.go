// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEPolicySetParameter_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parameter := &tfe.PolicySetParameter{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetParameter_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					testAccCheckTFEPolicySetParameterAttributes(parameter),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetParameter_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parameter := &tfe.PolicySetParameter{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetParameter_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					testAccCheckTFEPolicySetParameterAttributes(parameter),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "false"),
				),
			},

			{
				Config: testAccTFEPolicySetParameter_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					testAccCheckTFEPolicySetParameterAttributesUpdate(parameter),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

func TestAccTFEPolicySetParameter_import(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicySetParameter_basic(org.Name),
			},

			{
				ResourceName: "tfe_policy_set_parameter.foobar",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resources := s.RootModule().Resources
					policySet := resources["tfe_policy_set.foobar"]
					param := resources["tfe_policy_set_parameter.foobar"]

					return fmt.Sprintf("%s/%s", policySet.Primary.ID, param.Primary.ID), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEPolicySetParameter_valueWO(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	parameter := &tfe.PolicySetParameter{}

	paramValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	paramValue2 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Create the value comparer so we can add state values to it during the test steps
	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEPolicySetParameterDestroy,
		Steps: []resource.TestStep{
			{ // Should not be able to set both `value` and `value_wo` simultaneously
				Config:      testAccTFEPolicySetParameter_valueAndValueWO(org.Name, paramValue1),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo" is specified`),
			},
			{ // Provision a sensitive parameter with a write-only value
				Config: testAccTFEPolicySetParameter_valueWO(org.Name, paramValue1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					resource.TestCheckNoResourceAttr(
						"tfe_policy_set_parameter.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "true"),
				),
				// Register the id with the value comparer so we can assert that the
				// resource has been replaced in the next step.
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_policy_set_parameter.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{ // Update the value of the write-only parameter and set sensitive: false
				Config: testAccTFEPolicySetParameter_valueWO(org.Name, paramValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					resource.TestCheckNoResourceAttr(
						"tfe_policy_set_parameter.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "false"),
				),
				// Ensure that the resource has been replaced
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_policy_set_parameter.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				Config: testAccTFEPolicySetParameter_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicySetParameterExists(
						"tfe_policy_set_parameter.foobar", parameter),
					resource.TestCheckNoResourceAttr(
						"tfe_policy_set_parameter.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_policy_set_parameter.foobar", "sensitive", "false"),
				),
				// Ensure that the resource has been replaced
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_policy_set_parameter.foobar", tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

func testAccCheckTFEPolicySetParameterExists(
	n string, parameter *tfe.PolicySetParameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := config.Client.PolicySetParameters.Read(ctx, rs.Primary.Attributes["policy_set_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		*parameter = *v

		return nil
	}
}

func testAccCheckTFEPolicySetParameterAttributes(
	parameter *tfe.PolicySetParameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if parameter.Key != "key_test" {
			return fmt.Errorf("Bad key: %s", parameter.Key)
		}

		if parameter.Value != "value_test" {
			return fmt.Errorf("Bad value: %s", parameter.Value)
		}

		if parameter.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", parameter.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetParameterAttributesUpdate(
	parameter *tfe.PolicySetParameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if parameter.Key != "key_updated" {
			return fmt.Errorf("Bad key: %s", parameter.Key)
		}

		if parameter.Value != "" {
			return fmt.Errorf("Bad value: %s", parameter.Value)
		}

		if parameter.Sensitive != true {
			return fmt.Errorf("Bad sensitive: %t", parameter.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEPolicySetParameterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy_set_parameter" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.PolicySetParameters.Read(ctx, rs.Primary.Attributes["policy_set_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("PolicySetParameter %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicySetParameter_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  organization = "%s"
}

resource "tfe_policy_set_parameter" "foobar" {
  key          = "key_test"
  value        = "value_test"
  policy_set_id = tfe_policy_set.foobar.id
}`, organization)
}

func testAccTFEPolicySetParameter_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  organization = "%s"
}

resource "tfe_policy_set_parameter" "foobar" {
  key          = "key_updated"
  value        = "value_updated"
  sensitive    = true
  policy_set_id = tfe_policy_set.foobar.id
}`, organization)
}

func testAccTFEPolicySetParameter_valueWO(organization string, value int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  organization = "%s"
}

resource "tfe_policy_set_parameter" "foobar" {
  key          = "key_updated"
  value_wo        = "%d"
  sensitive    = %t
  policy_set_id = tfe_policy_set.foobar.id
}`, organization, value, sensitive)
}

func testAccTFEPolicySetParameter_valueAndValueWO(organization string, value int) string {
	return fmt.Sprintf(`
resource "tfe_policy_set" "foobar" {
  name         = "policy-set-test"
  organization = "%[1]s"
}

resource "tfe_policy_set_parameter" "foobar" {
  key          = "key_updated"
  value        = "%[2]d"
  value_wo      = "%[2]d"
  sensitive    = true
  policy_set_id = tfe_policy_set.foobar.id
}`, organization, value)
}
