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

// TestAccTFEProjectNotificationConfiguration_slack is a regression test for
// https://github.com/hashicorp/terraform-provider-tfe/issues/2102.
//
// For destination types that forbid setting the `token` attribute (slack,
// microsoft-teams, email), Create previously returned state with
// token = "" (empty string) while the plan had token = null. Terraform
// core's post-apply consistency check then failed with "inconsistent values
// for sensitive attribute".
//
// This test creates a slack notification configuration with no token in
// config, applies it, then runs a second step with the same config to
// confirm no spurious drift on subsequent plans. Both steps must succeed
// and the `token` attribute must remain null in state.
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
					// State must reflect the user's config: no token attribute set.
					// Initializing Token: types.StringValue("") in
					// modelFromTFEProjectNotificationConfiguration regressed this and
					// caused "inconsistent values for sensitive attribute" on apply.
					resource.TestCheckNoResourceAttr(
						"tfe_project_notification_configuration.foobar", "token"),
				),
			},
			{
				// Re-apply the same config to confirm no drift / no Update planned.
				// With the bug present, the second plan would detect "drift" between
				// config (token = null) and state (token = "") and attempt to update.
				Config:   testAccTFEProjectNotificationConfiguration_slack(org.Name, project.ID),
				PlanOnly: true,
			},
		},
	})
}

// TestAccTFEProjectNotificationConfiguration_urlWO exercises the write-only
// URL attribute pair (`url_wo` + `url_wo_version`). It mirrors the equivalent
// test on the workspace-scoped sibling and verifies that:
//
//  1. Create accepts `url_wo` without storing the URL in state (state's `url`
//     remains null).
//  2. The auto-managed `url_wo_version` is set to 1 after Create.
//  3. Re-applying the same config (PlanOnly) shows no drift.
//  4. Changing `url_wo` triggers an auto-increment of `url_wo_version` and a
//     successful Update.
//  5. Switching from `url_wo` back to plaintext `url` is blocked with a clear
//     error (preventing accidental exposure of a previously secret URL).
func TestAccTFEProjectNotificationConfiguration_urlWO(t *testing.T) {
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
				Config: testAccTFEProjectNotificationConfiguration_urlWO(org.Name, project.ID, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectNotificationConfigurationExists(
						"tfe_project_notification_configuration.foobar", notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "destination_type", "slack"),
					// The URL is never persisted in state when using url_wo.
					resource.TestCheckNoResourceAttr(
						"tfe_project_notification_configuration.foobar", "url"),
					// url_wo_version auto-initializes to 1 on Create.
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "1"),
				),
			},
			{
				// Re-apply identical config — no drift expected.
				Config:   testAccTFEProjectNotificationConfiguration_urlWO(org.Name, project.ID, runTasksURL()),
				PlanOnly: true,
			},
			{
				// Change url_wo — provider must detect via hash and increment url_wo_version.
				Config: testAccTFEProjectNotificationConfiguration_urlWO(org.Name, project.ID, runTasksURL()+"/rotated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_project_notification_configuration.foobar", "url_wo_version", "2"),
					resource.TestCheckNoResourceAttr(
						"tfe_project_notification_configuration.foobar", "url"),
				),
			},
			{
				// Attempting to switch from url_wo back to plaintext url must be blocked.
				Config:      testAccTFEProjectNotificationConfiguration_slack(org.Name, project.ID),
				ExpectError: regexp.MustCompile(`Cannot switch from write-only to plaintext`),
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
  triggers         = ["change_request:created"]
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
  enabled          = true
  url              = "%s"
  triggers         = ["run:errored", "run:needs_attention"]
  project_id       = "%s"
}`, orgName, runTasksURL(), projectID)
}

func testAccTFEProjectNotificationConfiguration_urlWO(orgName, projectID, urlValue string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_project_notification_configuration" "foobar" {
  name             = "notification_slack_url_wo"
  destination_type = "slack"
  enabled          = true
  url_wo           = "%s"
  triggers         = ["run:errored", "run:needs_attention"]
  project_id       = "%s"
}`, orgName, urlValue, projectID)
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
