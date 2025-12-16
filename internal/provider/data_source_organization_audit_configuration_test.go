// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEOrganizationAuditConfigurationDataSource_basic(t *testing.T) {
	skipUnlessBeta(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationAuditConfigDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_organization_audit_configuration.foobar", "organization", org.Name),
					resource.TestCheckResourceAttrSet("data.tfe_organization_audit_configuration.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_audit_configuration.foobar", "audit_trails_enabled"),
					resource.TestCheckResourceAttrSet("data.tfe_organization_audit_configuration.foobar", "hcp_log_streaming_enabled"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationAuditConfigurationDataSource_disallowed(t *testing.T) {
	skipUnlessBeta(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createTrialOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationAuditConfigDataSourceConfig(org.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_organization_audit_configuration.foobar", "organization", org.Name),
					resource.TestCheckNoResourceAttr("data.tfe_organization_audit_configuration.foobar", "id"),
					resource.TestCheckNoResourceAttr("data.tfe_organization_audit_configuration.foobar", "audit_trails_enabled"),
					resource.TestCheckNoResourceAttr("data.tfe_organization_audit_configuration.foobar", "hcp_log_streaming_enabled"),
				),
			},
		},
	})
}

func testAccTFEOrganizationAuditConfigDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

data "tfe_organization_audit_configuration" "foobar" {
	organization = local.organization_name
}`, orgName)
}
