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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationRunTask_validateSchemaAttributeUrl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationRunTask_basic("org", 1, "", ""),
				ExpectError: regexp.MustCompile(`url to not be empty`),
			},
			{
				Config:      testAccTFEOrganizationRunTask_basic("org", 1, "https://", ""),
				ExpectError: regexp.MustCompile(`to have a host`),
			},
			{
				Config:      testAccTFEOrganizationRunTask_basic("org", 1, "ftp://a.valid.url/path", ""),
				ExpectError: regexp.MustCompile(`to have a url with schema of: "http,https"`),
			},
		},
	})
}

func TestAccTFEOrganizationRunTask_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	runTask := &tfe.RunTask{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Note - We cannot easily test updating the HMAC Key as that would require coordination between this test suite
	// and the external Run Task service to "magically" allow a different Key. Instead we "update" with the same key
	// and manually test HMAC Key changes.
	hmacKey := runTasksHMACKey()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTask_basic(org.Name, rInt, runTasksURL(), hmacKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskExists("tfe_organization_run_task.foobar", runTask),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "url", runTasksURL()),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", hmacKey),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "enabled", "false"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "description", ""),
				),
			},
			{
				Config: testAccTFEOrganizationRunTask_update(org.Name, rInt, runTasksURL(), hmacKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d-new", rInt)),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "url", runTasksURL()),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", hmacKey),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "description", "a description"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationRunTask_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTask_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
			},
			{
				ResourceName:      "tfe_organization_run_task.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/foobar-task-%d", org.Name, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationRunTaskExists(n string, runTask *tfe.RunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		rt, err := config.Client.RunTasks.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Run Task: %w", err)
		}

		if rt == nil {
			return fmt.Errorf("Organization Run Task not found")
		}

		*runTask = *rt

		return nil
	}
}

func testAccCheckTFEOrganizationRunTaskDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_run_task" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.RunTasks.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Organization Run Task %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOrganizationRunTask_basic(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
	organization = "%s"
	url          = "%s"
	name         = "foobar-task-%d"
	enabled      = false
	hmac_key     = "%s"
}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFEOrganizationRunTask_update(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
	resource "tfe_organization_run_task" "foobar" {
		organization = "%s"
		url          = "%s"
		name         = "foobar-task-%d-new"
		enabled      = true
		hmac_key     = "%s"
		description  = "a description"
	}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}
