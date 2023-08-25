// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	testFixtureVersionFiles = "test-fixtures/policy-set-version"
)

func TestAccTFEVersionFiles_basic(t *testing.T) {
	expectedChecksum, err := hashPolicies(testFixtureVersionFiles)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVersionFilesConfig_basic(testFixtureVersionFiles),
				Check: resource.ComposeAggregateTestCheckFunc(
					// check data attrs
					resource.TestCheckResourceAttr("data.tfe_slug.policy", "source_path", testFixtureVersionFiles),
					resource.TestCheckResourceAttr("data.tfe_slug.policy", "id", expectedChecksum),
				),
			},
		},
	})
}

func testAccTFEVersionFilesConfig_basic(sourcePath string) string {
	return fmt.Sprintf(`
data "tfe_slug" "policy" {
  source_path = "%s"
}
`, sourcePath)
}
