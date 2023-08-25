// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFENotificationConfiguration_basic(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_basic(rInt),
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

func TestAccTFENotificationConfiguration_emailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_update(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_basic(rInt),
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
				Config: testAccTFENotificationConfiguration_update(rInt),
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

func TestAccTFENotificationConfiguration_updateEmailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
			{
				Config: testAccTFENotificationConfiguration_updateEmailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesUpdateEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_email_update"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributesUpdateEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "email_user_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesEmail(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_emailWithURL(rInt),
				ExpectError: regexp.MustCompile(`URL cannot be set with destination type of email`),
			},
			{
				Config:      testAccTFENotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of email`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesGeneric(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of generic`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of generic`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of generic`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesSlack(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of slack`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesMicrosoftTeams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of microsoft-teams`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesEmail(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_emailWithURL(rInt),
				ExpectError: regexp.MustCompile(`URL cannot be set with destination type of email`),
			},
			{
				Config:      testAccTFENotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of email`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesGeneric(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_basic(rInt),
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
				Config:      testAccTFENotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of generic`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of generic`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of generic`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesSlack(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_slack(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesSlack(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "slack"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_slack"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFENotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com"),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of slack`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of slack`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesMicrosoftTeams(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_microsoftTeams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributesMicrosoftTeams(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "microsoft-teams"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_msteams"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", "http://example.com"),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`email addresses cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`email user IDs cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`token cannot be set with destination type of microsoft-teams`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`URL is required with destination type of microsoft-teams`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_duplicateTriggers(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_duplicateTriggers(rInt),
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

func TestAccTFENotificationConfigurationImport_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_update(rInt),
			},

			{
				ResourceName:            "tfe_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccTFENotificationConfigurationImport_emailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_updateEmailUserIDs(rInt),
			},

			{
				ResourceName:            "tfe_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccTFENotificationConfigurationImport_emptyEmailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_emailUserIDs(rInt),
			},

			{
				ResourceName:            "tfe_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFENotificationConfigurationExists(n string, notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		nc, err := config.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerCreated), string(tfe.NotificationTriggerNeedsAttention)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com/?update=true" {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributesEmailUserIDs(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_email" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeEmail {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != false {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		// Token is write only, can't read it

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if len(notificationConfiguration.EmailUsers) != 0 {
			return fmt.Errorf("Wrong number of email users: %v", len(notificationConfiguration.EmailUsers))
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributesUpdateEmailUserIDs(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_email_update" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeEmail {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != true {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		// Token is write only, can't read it

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerCreated), string(tfe.NotificationTriggerNeedsAttention)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if len(notificationConfiguration.EmailUsers) != 1 {
			return fmt.Errorf("Wrong number of email users: %v", len(notificationConfiguration.EmailUsers))
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationAttributesSlack(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_slack" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeSlack {
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

func testAccCheckTFENotificationConfigurationAttributesMicrosoftTeams(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if notificationConfiguration.Name != "notification_msteams" {
			return fmt.Errorf("Bad name: %s", notificationConfiguration.Name)
		}

		if notificationConfiguration.DestinationType != tfe.NotificationDestinationTypeMicrosoftTeams {
			return fmt.Errorf("Bad destination type: %s", notificationConfiguration.DestinationType)
		}

		if notificationConfiguration.Enabled != false {
			return fmt.Errorf("Bad enabled value: %t", notificationConfiguration.Enabled)
		}

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com" {
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerCreated)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != "http://example.com" {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_notification_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := config.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Notification configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFENotificationConfiguration_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_emailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_email"
  destination_type = "email"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_slack(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_slack"
  destination_type = "slack"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_microsoftTeams(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_msteams"
  destination_type = "microsoft-teams"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_update"
  destination_type = "generic"
  enabled          = true
  token            = "1234567890_update"
  triggers         = ["run:created", "run:needs_attention"]
  url              = "http://example.com/?update=true"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_updateEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_email_update"
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.foobar.user_id]
  enabled          = true
  triggers         = ["run:created", "run:needs_attention"]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_emailWithURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_email_with_url"
  destination_type = "email"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_emailWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_email_with_token"
  destination_type = "email"
  token            = "1234567890"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_genericWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_addresses"
  destination_type = "generic"
  email_addresses  = ["test@example.com", "test2@example.com"]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_genericWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_user_ids"
  destination_type = "generic"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_genericWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_generic_without_url"
  destination_type = "generic"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_slackWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_addresses"
  destination_type = "slack"
  email_addresses  = ["test@example.com", "test2@example.com"]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_slackWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_user_ids"
  destination_type = "slack"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_slackWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_slack_with_token"
  destination_type = "slack"
  token            = "1234567890"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_slackWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_slack_without_url"
  destination_type = "slack"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_addresses"
  destination_type = "microsoft-teams"
  email_addresses  = ["test@example.com", "test2@example.com"]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_user_ids"
  destination_type = "microsoft-teams"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_microsoftTeamsWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_msteams_with_token"
  destination_type = "microsoft-teams"
  token            = "1234567890"
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_microsoftTeamsWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_msteams_without_url"
  destination_type = "microsoft-teams"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFENotificationConfiguration_duplicateTriggers(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_duplicate_triggers"
  destination_type = "generic"
  triggers         = ["run:created", "run:created", "run:created"]
  url              = "http://example.com"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}
