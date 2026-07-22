// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccTFEProjectNotificationConfiguration_basic(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectNotificationConfiguration_basic(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFEProjectNotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "name", "notification_basic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "token", "1234567890"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFEProjectNotificationConfiguration_emailUserIDs(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectNotificationConfiguration_emailUserIDs(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFEProjectNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "name", "notification_email"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFEProjectNotificationConfiguration_update(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectNotificationConfiguration_basic(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFEProjectNotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "name", "notification_basic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "token", "1234567890"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config: testAccTFEProjectNotificationConfiguration_update(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFEProjectNotificationConfigurationAttributesUpdate(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "name", "notification_update"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "token", "1234567890_update"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "triggers.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFEProjectNotificationConfiguration_validateSchemaAttributesSlack(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProjectNotificationConfiguration_slackWithEmailAddresses(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFEProjectNotificationConfiguration_slackWithEmailUserIDs(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFEProjectNotificationConfiguration_slackWithToken(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFEProjectNotificationConfiguration_slackWithoutURL(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
		},
	})
}

func TestAccTFEProjectNotificationConfiguration_slack(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectNotificationConfiguration_slack(org.Name, project.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "slack"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "name", "notification_slack"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url", runTasksURL()),
					resource.TestCheckNoResourceAttr(
						"tfe_project_notification_configuration.foobar", "token"),
				),
			},
		},
	})
}

func testAccCheckTFEProjectNotificationConfigurationExists(n string, notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		nc, err := testAccConfiguredClient.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if nc == nil {
			return fmt.Errorf("project notification configuration not found")
		}

		*notificationConfiguration = *nc

		return nil
	}
}

func testAccCheckTFEProjectNotificationConfigurationAttributes(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_basic" {
			return fmt.Errorf("bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeGeneric {
			return fmt.Errorf("bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled {
			return fmt.Errorf("bad enabled: %t", notificationConfiguration.Enabled)
		}

		if len(notificationConfiguration.Triggers) != 0 {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFEProjectNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_email" {
			return fmt.Errorf("bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeEmail {
			return fmt.Errorf("bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled {
			return fmt.Errorf("bad enabled: %t", notificationConfiguration.Enabled)
		}

		if len(notificationConfiguration.Triggers) != 0 {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if len(notificationConfiguration.EmailUsers) != 0 {
			return fmt.Errorf("bad email users: %v", notificationConfiguration.EmailUsers)
		}

		return nil
	}
}

func testAccCheckTFEProjectNotificationConfigurationAttributesUpdate(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_update" {
			return fmt.Errorf("bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeGeneric {
			return fmt.Errorf("bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if !notificationConfiguration.Enabled {
			return fmt.Errorf("bad enabled: %t", notificationConfiguration.Enabled)
		}

		if len(notificationConfiguration.Triggers) != 1 {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{"run:applying"}) {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFEProjectNotificationConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project_notification_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("project notification configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEProjectNotificationConfiguration_basic(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  enabled          = false
  token            = "1234567890"
  triggers         = []
  url              = "%s"
  project_id       = "%s"
}`, orgName, runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_emailUserIDs(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_email"
  destination_type = "email"
  email_user_ids   = []
  project_id       = "%s"
}`, orgName, projectID)
}

func testAccTFEProjectNotificationConfiguration_update(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_update"
  destination_type = "generic"
  enabled          = true
  token            = "1234567890_update"
  triggers         = ["run:applying"]
  url              = "%s"
  project_id       = "%s"
}`, orgName, runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_slack(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack"
  destination_type = "slack"
  url              = "%s"
  project_id       = "%s"
}`, orgName, runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_slackWithEmailAddresses(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_addresses"
  destination_type = "slack"
  email_addresses  = ["test@example.com", "test2@example.com"]
  project_id       = "%s"
}`, orgName, projectID)
}

func testAccTFEProjectNotificationConfiguration_slackWithEmailUserIDs(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_user_ids"
  destination_type = "slack"
  email_user_ids   = ["user-abc123"]
  project_id       = "%s"
}`, orgName, projectID)
}

func testAccTFEProjectNotificationConfiguration_slackWithToken(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack_with_token"
  destination_type = "slack"
  token            = "1234567890"
  url              = "%s"
  project_id       = "%s"
}`, orgName, runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_slackWithoutURL(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack_without_url"
  destination_type = "slack"
  project_id       = "%s"
}`, orgName, projectID)
}

func preCheckTFEProjectNotificationConfiguration(t *testing.T) {
	testAccPreCheck(t)
	skipIfEnterprise(t)
}

// TestAccTFEProjectNotificationConfiguration_urlWriteOnly tests auto-managed url_wo:
// - create with url_wo (version auto-set to 1)
// - update with changed url value (version auto-increments to 2)
// - same url again (version stays at 2)
func TestAccTFEProjectNotificationConfiguration_urlWriteOnly(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	notificationConfiguration := &tfe.NotificationConfiguration{}
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				// Create with url_wo — version should be auto-set to 1
				Config: testAccTFEProjectNotificationConfiguration_urlWriteOnly(org.Name, project.ID, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(
						"tfe_project_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Update with a different URL — version should auto-increment to 2
				Config: testAccTFEProjectNotificationConfiguration_urlWriteOnly(org.Name, project.ID, runTasksURL()+"?updated=true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "2"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Same resource, not recreated
					compareValuesSame.AddStateValue(
						"tfe_project_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Same URL again — version should stay at 2 (no hash change)
				Config: testAccTFEProjectNotificationConfiguration_urlWriteOnly(org.Name, project.ID, runTasksURL()+"?updated=true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "2"),
				),
			},
			{
				// Attempting to switch from url_wo to plaintext url should be blocked
				Config:      testAccTFEProjectNotificationConfiguration_basic(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`Cannot switch from write-only to plaintext`),
			},
		},
	})
}

// TestAccTFEProjectNotificationConfiguration_urlWriteOnlyManualVersion tests manual url_wo_version mode:
// explicitly setting url_wo_version disables hash auto-detection.
func TestAccTFEProjectNotificationConfiguration_urlWriteOnlyManualVersion(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProjectNotificationConfiguration_urlWriteOnlyManual(org.Name, project.ID, runTasksURL(), 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_project_notification_configuration.foobar", "url"),
				),
			},
			{
				// Increment version manually to trigger URL update
				Config: testAccTFEProjectNotificationConfiguration_urlWriteOnlyManual(org.Name, project.ID, runTasksURL()+"?v2=true", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "2"),
				),
			},
		},
	})
}

// TestAccTFEProjectNotificationConfiguration_urlWriteOnlyValidation tests that schema
// validators reject invalid combinations for url_wo.
func TestAccTFEProjectNotificationConfiguration_urlWriteOnlyValidation(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createStandardOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	project := createProject(t, tfeClient, org.Name, tfe.ProjectCreateOptions{
		Name: "test-project",
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFEProjectNotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProjectNotificationConfiguration_urlAndUrlWriteOnly(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`Attribute "url_wo" cannot be specified when "url" is specified`),
			},
			{
				Config:      testAccTFEProjectNotificationConfiguration_urlWriteOnlyVersionWithoutURL(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`Attribute "url_wo" must be specified when "url_wo_version" is specified`),
			},
		},
	})
}

func testAccTFEProjectNotificationConfiguration_urlWriteOnly(orgName, projectID, url string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url_wo           = "%s"
  project_id       = "%s"
}`, orgName, url, projectID)
}

func testAccTFEProjectNotificationConfiguration_urlWriteOnlyManual(orgName, projectID, url string, version int64) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url_wo           = "%s"
  url_wo_version   = %d
  project_id       = "%s"
}`, orgName, url, version, projectID)
}

func testAccTFEProjectNotificationConfiguration_urlAndUrlWriteOnly(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url              = "%s"
  url_wo           = "%s"
  project_id       = "%s"
}`, orgName, runTasksURL(), runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_urlWriteOnlyVersionWithoutURL(orgName, projectID string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url_wo_version   = 1
  project_id       = "%s"
}`, orgName, projectID)
}
