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
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccTFEOrganizationRunTask_validateSchemaAttributeUrl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
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

func TestAccTFEOrganizationRunTask_basic(t *testing.T) {
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamAccessDestroy,
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

func TestAccTFEOrganizationRunTask_Read(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	hmacKey := runTasksHMACKey()

	org_tf := fmt.Sprintf(`data "tfe_organization" "orgtask" { name = %q }`, org.Name)

	create_task_tf := fmt.Sprintf(`
		%s
		%s
		`, org_tf, testAccTFEOrganizationRunTask_basic(org.Name, rInt, runTasksURL(), hmacKey))

	delete_tasks := func() {
		tasks, err := tfeClient.RunTasks.List(ctx, org.Name, nil)
		if err != nil || tasks == nil {
			t.Fatalf("Error listing tasks: %s", err)
			return
		}
		// There shouldn't be more that 25 run tasks so we don't need to worry about pagination
		for _, task := range tasks.Items {
			if task != nil {
				if err := tfeClient.RunTasks.Delete(ctx, task.ID); err != nil {
					t.Fatalf("Error deleting task: %s", err)
				}
			}
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: create_task_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
				),
			},
			{
				// Delete the created run task and ensure we can re-create it
				PreConfig: delete_tasks,
				Config:    create_task_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
				),
			},
			{
				// Delete the created run task and ensure we can ignore it if we no longer need to manage it
				PreConfig: delete_tasks,
				Config:    org_tf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskDestroy,
				),
			},
		},
	})
}

func TestAccTFEOrganizationRunTask_HMACWriteOnly(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	runTask := &tfe.RunTask{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	// Create the value comparer so we can add state values to it during the test steps
	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	// Note - We cannot easily test updating the HMAC Key as that would require coordination between this test suite
	// and the external Run Task service to "magically" allow a different Key. Instead we "update" with the same key
	// and manually test HMAC Key changes.
	hmacKey := runTasksHMACKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationRunTask_hmacAndHMACWriteOnly(org.Name, rInt, runTasksURL()),
				ExpectError: regexp.MustCompile(`Attribute "hmac_key_wo" cannot be specified when "hmac_key" is specified`),
			},
			{
				Config: testAccTFEOrganizationRunTask_hmacWriteOnly(org.Name, rInt, runTasksURL(), hmacKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskExists("tfe_organization_run_task.foobar", runTask),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", ""),
					resource.TestCheckNoResourceAttr("tfe_organization_run_task.foobar", "hmac_key_wo"),
				),
				// Register the id with the value comparer so we can assert that the
				// resource has been replaced in the next step.
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_organization_run_task.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				Config: testAccTFEOrganizationRunTask_basic(org.Name, rInt, runTasksURL(), hmacKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskExists("tfe_organization_run_task.foobar", runTask),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", hmacKey),
					resource.TestCheckNoResourceAttr("tfe_organization_run_task.foobar", "hmac_key_wo"),
				),
				// Ensure that the resource has been replaced
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_organization_run_task.foobar", tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

func testAccCheckTFEOrganizationRunTaskExists(n string, runTask *tfe.RunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		rt, err := testAccConfiguredClient.Client.RunTasks.Read(ctx, rs.Primary.ID)
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
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_run_task" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.RunTasks.Read(ctx, rs.Primary.ID)
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

func testAccTFEOrganizationRunTask_hmacWriteOnly(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
	resource "tfe_organization_run_task" "foobar" {
		organization = "%s"
		url          = "%s"
		name         = "foobar-task-%d"
		enabled      = false
		hmac_key_wo  = "%s"
	}
	`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFEOrganizationRunTask_hmacAndHMACWriteOnly(orgName string, rInt int, runTaskURL string) string {
	return fmt.Sprintf(`
	resource "tfe_organization_run_task" "foobar" {
		organization = "%s"
		url          = "%s"
		name         = "foobar-task-%d"
		enabled      = false
		hmac_key     = "foo"
		hmac_key_wo  = "foo"
	}
	`, orgName, runTaskURL, rInt)
}
