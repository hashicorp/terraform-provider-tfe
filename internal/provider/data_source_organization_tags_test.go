// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEOrganizationTagsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationTagsDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_organization_tags.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_tags.foobar", "tags.0.name", "modules"),
					resource.TestCheckResourceAttr(
						"data.tfe_organization_tags.foobar", "tags.0.workspace_count", "1"),
				),
			},
		},
	})
}

func testAccTFEOrganizationTagsDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.id
  description           = "provider-testing"
  allow_destroy_plan    = false
  auto_apply            = true
  file_triggers_enabled = true
  queue_all_runs        = false
  speculative_enabled   = true
  assessments_enabled       = false
  tag_names             = ["modules"]
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "terraform/test"
  global_remote_state   = true
}

data "tfe_organization_tags" "foobar" {
  organization = tfe_workspace.foobar.organization
  depends_on   = [tfe_workspace.foobar]
}`, rInt, rInt)
}
