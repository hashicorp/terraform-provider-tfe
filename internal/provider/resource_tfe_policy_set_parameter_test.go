// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetParameterDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetParameterDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicySetParameterDestroy,
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
