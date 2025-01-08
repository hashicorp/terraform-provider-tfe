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

func TestAccTFETeamNotificationConfiguration_basic(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_basic"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_emailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_update(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_basic"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesUpdate(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_update"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "token", "1234567890_update"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributesUpdate
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", fmt.Sprintf("%s?update=true", runTasksURL())),
				),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateEmailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_updateEmailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesUpdateEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_email_update"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributesUpdateEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "email_user_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesEmail(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesGeneric(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesSlack(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesMicrosoftTeams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesBadDestinationType(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_badDestinationType(rInt),
				ExpectError: regexp.MustCompile(`.*Invalid Attribute Value Match.*`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesEmail(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "email"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_email"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributesEmailUserIDs
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "email_user_ids.#", "0"),
				),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(rInt),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesGeneric(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_basic"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_basic(rInt),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesSlack(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_slack(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesSlack(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "slack"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_slack"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_slack(rInt),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesMicrosoftTeams(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_microsoftTeams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesMicrosoftTeams(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "microsoft-teams"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_msteams"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_microsoftTeams(rInt),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_duplicateTriggers(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_duplicateTriggers(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamNotificationConfigurationExists(
						"tfe_team_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFETeamNotificationConfigurationAttributesDuplicateTriggers(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "name", "notification_duplicate_triggers"),
					// Just test the number of items in triggers
					// Values in triggers attribute are tested by testCheckTFETeamNotificationConfigurationAttributes
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "triggers.#", "1"),
					resource.TestCheckResourceAttr(
						"tfe_team_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFETeamNotificationConfigurationImport_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_update(rInt),
			},

			{
				ResourceName:            "tfe_team_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccTFETeamNotificationConfigurationImport_emailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_updateEmailUserIDs(rInt),
			},

			{
				ResourceName:            "tfe_team_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccTFETeamNotificationConfigurationImport_emptyEmailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(rInt),
			},

			{
				ResourceName:            "tfe_team_notification_configuration.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFETeamNotificationConfigurationExists(n string, notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

func testAccCheckTFETeamNotificationConfigurationAttributes(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationAttributesUpdate(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerChangeRequestCreated)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != fmt.Sprintf("%s?update=true", runTasksURL()) {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationAttributesEmailUserIDs(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

func testAccCheckTFETeamNotificationConfigurationAttributesUpdateEmailUserIDs(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerChangeRequestCreated)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if len(notificationConfiguration.EmailUsers) != 1 {
			return fmt.Errorf("Wrong number of email users: %v", len(notificationConfiguration.EmailUsers))
		}

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationAttributesSlack(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationAttributesMicrosoftTeams(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationAttributesDuplicateTriggers(notificationConfiguration *tfe.NotificationConfiguration) resource.TestCheckFunc {
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{string(tfe.NotificationTriggerChangeRequestCreated)}) {
			return fmt.Errorf("Bad triggers: %v", notificationConfiguration.Triggers)
		}

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFETeamNotificationConfigurationDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_team_notification_configuration" {
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

func testAccTFETeamNotificationConfiguration_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
	url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_emailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email"
  destination_type = "email"
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_slack(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack"
  destination_type = "slack"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_microsoftTeams(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams"
  destination_type = "microsoft-teams"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_badDestinationType(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "bad_type"
	url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_update"
  destination_type = "generic"
  enabled          = true
  token            = "1234567890_update"
  triggers         = ["change_request:created"]
  url              = "%s?update=true"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_updateEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_team_organization_member" "foobar" {
  team_id = tfe_team.foobar.id
  organization_membership_id = tfe_organization_membership.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email_update"
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.foobar.user_id]
  enabled          = true
  triggers         = ["change_request:created"]
  team_id          = tfe_team.foobar.id

	depends_on = [tfe_team_organization_member.foobar]
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_emailWithURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email_with_url"
  destination_type = "email"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_emailWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email_with_token"
  destination_type = "email"
  token            = "1234567890"
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_addresses"
  destination_type = "generic"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_user_ids"
  destination_type = "generic"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_genericWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_without_url"
  destination_type = "generic"
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_addresses"
  destination_type = "slack"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_user_ids"
  destination_type = "slack"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_slackWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_token"
  destination_type = "slack"
  token            = "1234567890"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_slackWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_without_url"
  destination_type = "slack"
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_addresses"
  destination_type = "microsoft-teams"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_organization_membership" "foobar" {
  organization = tfe_organization.foobar.id
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_user_ids"
  destination_type = "microsoft-teams"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_token"
  destination_type = "microsoft-teams"
  token            = "1234567890"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_without_url"
  destination_type = "microsoft-teams"
  team_id          = tfe_team.foobar.id
}`, rInt)
}

func testAccTFETeamNotificationConfiguration_duplicateTriggers(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_duplicate_triggers"
  destination_type = "generic"
  triggers         = ["change_request:created", "change_request:created", "change_request:created"]
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, rInt, runTasksURL())
}

func preCheckTFETeamNotificationConfiguration(t *testing.T) {
	testAccPreCheck(t)

	if runTasksURL() == "" {
		t.Skip("RUN_TASKS_URL must be set for team notification configuration acceptance tests")
	}
}
