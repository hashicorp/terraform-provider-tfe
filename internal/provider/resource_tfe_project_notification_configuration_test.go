// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
		Name: tfe.String("test-project"),
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
		Name: tfe.String("test-project"),
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
		Name: tfe.String("test-project"),
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
		Name: tfe.String("test-project"),
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

func testAccCheckTFEProjectNotificationConfigurationExists(n string, notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(ConfiguredClient)

		nc, err := config.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
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

		if notificationConfiguration.Token != "1234567890" {
			return fmt.Errorf("bad token: %s", notificationConfiguration.Token)
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

		if notificationConfiguration.Token != "1234567890_update" {
			return fmt.Errorf("bad token: %s", notificationConfiguration.Token)
		}

		if len(notificationConfiguration.Triggers) != 1 {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{"change_request:created"}) {
			return fmt.Errorf("bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFEProjectNotificationConfigurationDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project_notification_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := config.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
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
  triggers         = ["change_request:created"]
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
