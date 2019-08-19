package tfe

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFENotificationConfiguration_basic(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_basic"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_update(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_basic"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com"),
				),
			},
			{
				Config: testAccTFENotificationConfiguration_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesUpdate(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_update"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token", "1234567890_update"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributesUpdate
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com/?update=true"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_slackWithToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_slackWithToken,
				ExpectError: regexp.MustCompile(`Token cannot be set with destination_type of slack`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_duplicateTriggers(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_duplicateTriggers,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesDuplicateTriggers(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_duplicate_triggers"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfigurationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_update,
			},

			{
				ResourceName:            "tfe_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "workspace_external_id"},
			},
		},
	})
}

func testAccCheckTFENotificationConfigurationExists(n string, notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		nc, err := tfeClient.NotificationConfigurations.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*notificationConfiguration = *nc

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributes(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_basic" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeGeneric {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != false {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		// Token is write only, can't read it

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com" {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributesUpdate(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_update" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeGeneric {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != true {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		// Token is write only, can't read it

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{tfe.NotificationTriggerCreated, tfe.NotificationTriggerNeedsAttention}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com/?update=true" {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributesDuplicateTriggers(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_duplicate_triggers" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeGeneric {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != false {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		// Token is write only, can't read it

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{tfe.NotificationTriggerCreated}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com" {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_notification_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.NotificationConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Notification configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFENotificationConfiguration_basic = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_basic"
  destination_type      = "generic"
  url                   = "http://example.com"
  workspace_external_id = "${tfe_workspace.foobar.external_id}"
}`

const testAccTFENotificationConfiguration_update = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_update"
  destination_type      = "generic"
  enabled               = true
  token                 = "1234567890_update"
  triggers              = ["run:created", "run:needs_attention"]
  url                   = "http://example.com/?update=true"
  workspace_external_id = "${tfe_workspace.foobar.external_id}"
}`

const testAccTFENotificationConfiguration_slackWithToken = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_slack_with_token"
  destination_type      = "slack"
  token                 = "1234567890"
  url                   = "http://example.com"
  workspace_external_id = "${tfe_workspace.foobar.external_id}"
}`

const testAccTFENotificationConfiguration_duplicateTriggers = `
resource "tfe_organization" "foobar" {
  name  = "terraform-test"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_duplicate_triggers"
  destination_type      = "generic"
  triggers              = ["run:created", "run:created", "run:created"]
  url                   = "http://example.com"
  workspace_external_id = "${tfe_workspace.foobar.external_id}"
}`
