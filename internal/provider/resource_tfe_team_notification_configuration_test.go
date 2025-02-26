// Copyright (c) HashiCorp, Inc.
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

func TestAccTFETeamNotificationConfiguration_basic(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(org.Name),
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
				Config: testAccTFETeamNotificationConfiguration_update(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(org.Name),
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
				Config: testAccTFETeamNotificationConfiguration_updateEmailUserIDs(org.Name),
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
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithToken(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesGeneric(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesSlack(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithToken(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesMicrosoftTeams(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_validateSchemaAttributesBadDestinationType(t *testing.T) {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETeamNotificationConfiguration_badDestinationType(org.Name),
				ExpectError: regexp.MustCompile(`.*Invalid Attribute Value Match.*`),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesEmail(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(org.Name),
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
				Config:      testAccTFETeamNotificationConfiguration_emailWithURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_emailWithToken(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(org.Name),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesGeneric(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_basic(org.Name),
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
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_genericWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_basic(org.Name),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesSlack(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_slack(org.Name),
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
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithToken(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_slackWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_slack(org.Name),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_updateValidateSchemaAttributesMicrosoftTeams(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_microsoftTeams(org.Name),
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
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(org.Name),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(org.Name),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
			{
				Config: testAccTFETeamNotificationConfiguration_microsoftTeams(org.Name),
			},
		},
	})
}

func TestAccTFETeamNotificationConfiguration_duplicateTriggers(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	notificationConfiguration := &tfe.NotificationConfiguration{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_duplicateTriggers(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_update(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_updateEmailUserIDs(org.Name),
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
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, cleanupOrg := createPlusOrganization(t, tfeClient)
	t.Cleanup(cleanupOrg)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFETeamNotificationConfiguration(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETeamNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamNotificationConfiguration_emailUserIDs(org.Name),
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

func testAccTFETeamNotificationConfiguration_basic(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
	url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_emailUserIDs(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_organization_membership" "foobar" {
  organization = data.tfe_organization.foobar.name
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email"
  destination_type = "email"
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_slack(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack"
  destination_type = "slack"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_microsoftTeams(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams"
  destination_type = "microsoft-teams"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_badDestinationType(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "bad_type"
	url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_update(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_update"
  destination_type = "generic"
  enabled          = true
  token            = "1234567890_update"
  triggers         = ["change_request:created"]
  url              = "%s?update=true"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_updateEmailUserIDs(orgName string) string {
	return fmt.Sprintf(`data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_organization_membership" "foobar" {
  organization = data.tfe_organization.foobar.name
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
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_emailWithURL(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email_with_url"
  destination_type = "email"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_emailWithToken(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_email_with_token"
  destination_type = "email"
  token            = "1234567890"
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_genericWithEmailAddresses(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_addresses"
  destination_type = "generic"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_genericWithEmailUserIDs(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_organization_membership" "foobar" {
  organization = data.tfe_organization.foobar.name
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_with_email_user_ids"
  destination_type = "generic"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_genericWithoutURL(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_generic_without_url"
  destination_type = "generic"
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_slackWithEmailAddresses(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_addresses"
  destination_type = "slack"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_slackWithEmailUserIDs(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_organization_membership" "foobar" {
  organization = data.tfe_organization.foobar.name
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_email_user_ids"
  destination_type = "slack"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_slackWithToken(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_with_token"
  destination_type = "slack"
  token            = "1234567890"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_slackWithoutURL(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_slack_without_url"
  destination_type = "slack"
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailAddresses(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_addresses"
  destination_type = "microsoft-teams"
  email_addresses  = ["test@example.com", "test2@example.com"]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithEmailUserIDs(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_organization_membership" "foobar" {
  organization = data.tfe_organization.foobar.name
  email        = "foo@foobar.com"
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_email_user_ids"
  destination_type = "microsoft-teams"
  email_user_ids   = [tfe_organization_membership.foobar.id]
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithToken(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_with_token"
  destination_type = "microsoft-teams"
  token            = "1234567890"
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func testAccTFETeamNotificationConfiguration_microsoftTeamsWithoutURL(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_msteams_without_url"
  destination_type = "microsoft-teams"
  team_id          = tfe_team.foobar.id
}`, orgName)
}

func testAccTFETeamNotificationConfiguration_duplicateTriggers(orgName string) string {
	return fmt.Sprintf(`
data "tfe_organization" "foobar" {
  name = "%s"
}

resource "tfe_team" "foobar" {
  name         = "team-test"
  organization = data.tfe_organization.foobar.name
}

resource "tfe_team_notification_configuration" "foobar" {
  name             = "notification_duplicate_triggers"
  destination_type = "generic"
  triggers         = ["change_request:created", "change_request:created", "change_request:created"]
  url              = "%s"
  team_id          = tfe_team.foobar.id
}`, orgName, runTasksURL())
}

func preCheckTFETeamNotificationConfiguration(t *testing.T) {
	testAccPreCheck(t)

	if runTasksURL() == "" {
		t.Skip("RUN_TASKS_URL must be set for team notification configuration acceptance tests")
	}
}
