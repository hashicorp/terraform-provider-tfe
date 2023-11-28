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

func TestAccTFEPolicy_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicy_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},
		},
	})
}

func TestAccTFEPolicy_basicWithDefaults(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicy_basicWithDefaults(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEDefaultPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "soft-mandatory"),
				),
			},
		},
	})
}

func TestAccTFEPolicyOPA_basic(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicyOPA_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEOPAPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "package example rule[\"not allowed\"] { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "query", "data.example.rule"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "mandatory"),
				),
			},
		},
	})
}

func TestAccTFEPolicy_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicy_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEPolicyAttributes(policy),
					testAccCheckTFEPolicyContent(policy, "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},

			{
				Config: testAccTFEPolicy_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEPolicyAttributesUpdated(policy),
					testAccCheckTFEPolicyContent(policy, "main = rule { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "An updated test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "soft-mandatory"),
				),
			},
		},
	})
}

func TestAccTFEPolicy_unsetEnforce(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicy_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "sentinel"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},

			{
				Config: testAccTFEPolicy_emptyEnforce(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "An updated test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "main = rule { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},
		},
	})
}

func TestAccTFEPolicyOPA_update(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicyOPA_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEOPAPolicyAttributes(policy),
					testAccCheckTFEPolicyContent(policy, "package example rule[\"not allowed\"] { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "kind", "opa"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "package example rule[\"not allowed\"] { false }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "query", "data.example.rule"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "mandatory"),
				),
			},

			{
				Config: testAccTFEPolicyOPA_update(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEPolicyExists(
						"tfe_policy.foobar", policy),
					testAccCheckTFEOPAPolicyAttributesUpdated(policy),
					testAccCheckTFEPolicyContent(policy, "package example ruler[\"not allowed\"] { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "description", "An updated test policy"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "policy", "package example ruler[\"not allowed\"] { true }"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "query", "data.example.ruler"),
					resource.TestCheckResourceAttr(
						"tfe_policy.foobar", "enforce_mode", "advisory"),
				),
			},
		},
	})
}

func TestAccTFEPolicy_import(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEPolicy_basic(org.Name),
			},

			{
				ResourceName:        "tfe_policy.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("%s/", org.Name),
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFEPolicyExists(
	n string, policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := config.Client.Policies.Read(ctx, rs.Primary.ID)
		if err != nil {
			// nolint: wrapcheck
			return err
		}

		if p.ID != rs.Primary.ID {
			return fmt.Errorf("Policy not found")
		}

		*policy = *p

		return nil
	}
}

func testAccCheckTFEPolicyContent(policy *tfe.Policy, content string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		b, err := config.Client.Policies.Download(ctx, policy.ID)
		if err != nil {
			return fmt.Errorf("Problem downloading policy content: %w", err)
		}
		s := string(b)
		if s != content {
			return fmt.Errorf("Policy content didn't match. Expected: %q; got: %q", content, s)
		}
		return nil
	}
}

func testAccCheckTFEPolicyAttributes(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "hard-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFEOPAPolicyAttributes(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		if *policy.Query != "data.example.rule" {
			return fmt.Errorf("Bad OPA query string: %s", *policy.Query)
		}

		return nil
	}
}

func testAccCheckTFEDefaultPolicyAttributes(policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		switch policy.Kind {
		case tfe.Sentinel:
			if policy.Enforce[0].Mode != "soft-mandatory" {
				return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
			}
		case tfe.OPA:
			if policy.Enforce[0].Mode != "advisory" {
				return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
			}
		}
		return nil
	}
}

func testAccCheckTFEPolicyAttributesUpdated(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "soft-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFEOPAPolicyAttributesUpdated(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "advisory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		if *policy.Query != "data.example.ruler" {
			return fmt.Errorf("Bad OPA query string: %s", *policy.Query)
		}

		return nil
	}
}

func testAccCheckTFEPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.Policies.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf(" policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEPolicy_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "A test policy"
  organization = "%s"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}`, organization)
}

func testAccTFEPolicy_basicWithDefaults(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "A test policy"
  organization = "%s"
  policy       = "main = rule { true }"
}`, organization)
}

func testAccTFEPolicyOPA_basic(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "A test policy"
  organization = "%s"
  kind         = "opa"
  policy       = "package example rule[\"not allowed\"] { false }"
  query        = "data.example.rule"
  enforce_mode = "mandatory"
}`, organization)
}

func testAccTFEPolicy_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "An updated test policy"
  organization = "%s"
  policy       = "main = rule { false }"
  enforce_mode = "soft-mandatory"
}`, organization)
}

func testAccTFEPolicy_emptyEnforce(organization string) string {
	return fmt.Sprintf(`
  resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "An updated test policy"
  organization = "%s"
  policy       = "main = rule { false }"
}`, organization)
}

func testAccTFEPolicyOPA_update(organization string) string {
	return fmt.Sprintf(`
resource "tfe_policy" "foobar" {
  name         = "policy-test"
  description  = "An updated test policy"
  organization = "%s"
  kind         = "opa"
  policy       = "package example ruler[\"not allowed\"] { true }"
  query        = "data.example.ruler"
  enforce_mode = "advisory"
}`, organization)
}
