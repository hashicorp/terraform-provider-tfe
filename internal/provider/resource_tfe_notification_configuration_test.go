// Copyright IBM Corp. 2018, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccTFENotificationConfiguration_basic(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_WriteOnly(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())
	var tokenOne, tokenTwo string
	var versionOne, versionTwo int64
	tokenOne, tokenTwo, versionOne, versionTwo = "tokenone", "tokentwo", 1, 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_tokenWriteOnly(rInt, tokenOne, versionOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					testAccCheckTFENotificationConfigurationAttributes(notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "name", "notification_basic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "triggers.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue(
						"tfe_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				Config: testAccTFENotificationConfiguration_tokenWriteOnly(rInt, tokenTwo, versionTwo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "2"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
				),
			},
			{
				Config: testAccTFENotificationConfiguration_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo_version"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_tokenWriteOnlyValidation(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_tokenWriteOnlyVersionWithoutToken(rInt),
				ExpectError: regexp.MustCompile(`Attribute "token_wo" must be specified when "token_wo_version" is specified`),
			},
			{
				Config:      testAccTFENotificationConfiguration_tokenAndTokenWriteOnly(rInt),
				ExpectError: regexp.MustCompile(`Attribute "token_wo" cannot be specified when "token" is specified`),
			},
			{
				// Create using token_wo, then attempt to switch to plaintext token — should be blocked
				Config: testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt, "secret-token"),
			},
			{
				Config:      testAccTFENotificationConfiguration_update(rInt),
				ExpectError: regexp.MustCompile(`Cannot switch from write-only to plaintext`),
			},
		},
	})
}

// TestAccTFENotificationConfiguration_tokenWriteOnlyAutoDetect tests auto-managed token_wo:
// - create with token_wo (version auto-set to 1)
// - update with changed token value (version auto-increments to 2)
// - remove token_wo entirely (token_wo_version cleared, no token set)
// - re-add the same token value (version must increment to 1, not stay null)
func TestAccTFENotificationConfiguration_tokenWriteOnlyAutoDetect(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				// Create with token_wo — version should be auto-set to 1
				Config: testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt, "token-v1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(
						"tfe_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Update with a different token — version should auto-increment to 2
				Config: testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt, "token-v2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "2"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(
						"tfe_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Same token again — version should stay at 2 (no hash change)
				Config: testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt, "token-v2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "2"),
				),
			},
			{
				// Remove token_wo entirely (no token set) — token_wo_version should be cleared
				Config: testAccTFENotificationConfiguration_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token_wo_version"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "token"),
				),
			},
			{
				// Re-add the same token value that was previously used — the stale hash must have
				// been cleared on removal, so this is treated as a new value and version increments.
				Config: testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt, "token-v2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "token_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_emailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
	t.Skip("temporarily skipped due to flakiness")

	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
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
						"tfe_notification_configuration.foobar", "url", fmt.Sprintf("%s?update=true", runTasksURL())),
				),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateEmailUserIDs(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_emailWithURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesGeneric(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesSlack(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_validateSchemaAttributesMicrosoftTeams(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesEmail(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
				ExpectError: regexp.MustCompile(`The attribute 'url' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_emailWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'email'`),
			},
			{
				Config: testAccTFENotificationConfiguration_emailUserIDs(rInt),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesGeneric(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'generic'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_genericWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'generic'`),
			},
			{
				Config: testAccTFENotificationConfiguration_basic(rInt),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesSlack(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithToken(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'token' cannot be set when 'destination_type' is 'slack'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_slackWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'slack'`),
			},
			{
				Config: testAccTFENotificationConfiguration_slack(rInt),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_updateValidateSchemaAttributesMicrosoftTeams(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailAddresses(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_addresses' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithEmailUserIDs(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'email_user_ids' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithToken(rInt),
				ExpectError: regexp.MustCompile(`(?s).*The attribute 'token' cannot be set when 'destination_type' is.*'microsoft-teams'`),
			},
			{
				Config:      testAccTFENotificationConfiguration_microsoftTeamsWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`The attribute 'url' is required when 'destination_type' is 'microsoft-teams'`),
			},
			{
				Config: testAccTFENotificationConfiguration_microsoftTeams(rInt),
			},
		},
	})
}

func TestAccTFENotificationConfiguration_duplicateTriggers(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
						"tfe_notification_configuration.foobar", "url", runTasksURL()),
				),
			},
		},
	})
}

func TestAccTFENotificationConfigurationImport_basic(t *testing.T) {
	t.Skip("temporarily skipped due to flakiness")

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	fmt.Printf("Config for testAccTFENotificationConfigurationImport_basic:\n %s\n", testAccTFENotificationConfiguration_basic(rInt))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
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
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		nc, err := testAccConfiguredClient.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
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

		if notificationConfiguration.URL != runTasksURL() {
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

		if notificationConfiguration.URL != fmt.Sprintf("%s?update=true", runTasksURL()) {
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

		if notificationConfiguration.URL != runTasksURL() {
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

		if notificationConfiguration.URL != runTasksURL() {
			return fmt.Errorf("Bad URL: %s", notificationConfiguration.URL)
		}

		return nil
	}
}

func testAccCheckTFENotificationConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_notification_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.NotificationConfigurations.Read(ctx, rs.Primary.ID)
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

variable "url" {
  default = "%s"
}

resource "tfe_notification_configuration" "foobar" {
  name             = "notification_basic"
  destination_type = "generic"
  url              = var.url
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFENotificationConfiguration_tokenWriteOnly(rInt int, wo string, version int64) string {
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
	token_wo         = "%s"
	token_wo_version = %d
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, wo, version, runTasksURL())
}

func testAccTFENotificationConfiguration_tokenWriteOnlyAuto(rInt int, token string) string {
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
  token_wo         = "%s"
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, token, runTasksURL())
}

func testAccTFENotificationConfiguration_tokenWriteOnlyVersionWithoutToken(rInt int) string {
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
	token_wo_version = 1
	url              = "%s"
	workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
}

func testAccTFENotificationConfiguration_tokenAndTokenWriteOnly(rInt int) string {
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
	token            = "some-token"
	token_wo         = "some-token"
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s?update=true"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
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
  url              = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL())
}

func preCheckTFENotificationConfiguration(t *testing.T) {
	testAccPreCheck(t)

	if runTasksURL() == "" {
		t.Skip("RUN_TASKS_URL must be set for notification configuration acceptance tests")
	}
}

// TestAccTFENotificationConfiguration_urlWriteOnly tests auto-managed url_wo:
// - create with url_wo (version auto-set to 1)
// - update with changed url value (version auto-increments to 2)
// - same url again (version stays at 2)
func TestAccTFENotificationConfiguration_urlWriteOnly(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				// Create with url_wo — version should be auto-set to 1
				Config: testAccTFENotificationConfiguration_urlWriteOnly(rInt, runTasksURL()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFENotificationConfigurationExists(
						"tfe_notification_configuration.foobar", notificationConfiguration),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "destination_type", "generic"),
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(
						"tfe_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Update with a different URL — version should auto-increment to 2
				Config: testAccTFENotificationConfiguration_urlWriteOnly(rInt, runTasksURL()+"?updated=true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url_wo_version", "2"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Same resource, not recreated
					compareValuesSame.AddStateValue(
						"tfe_notification_configuration.foobar", tfjsonpath.New("id"),
					),
				},
			},
			{
				// Same URL again — version should stay at 2 (no hash change)
				Config: testAccTFENotificationConfiguration_urlWriteOnly(rInt, runTasksURL()+"?updated=true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url_wo_version", "2"),
				),
			},
			{
				// Attempting to switch from url_wo to plaintext url should be blocked
				Config:      testAccTFENotificationConfiguration_basic(rInt),
				ExpectError: regexp.MustCompile(`Cannot switch from write-only to plaintext`),
			},
		},
	})
}

// TestAccTFENotificationConfiguration_urlWriteOnlyManualVersion tests manual url_wo_version mode:
// explicitly setting url_wo_version disables hash auto-detection.
func TestAccTFENotificationConfiguration_urlWriteOnlyManualVersion(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFENotificationConfiguration_urlWriteOnlyManual(rInt, runTasksURL(), 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url_wo_version", "1"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url_wo"),
					resource.TestCheckNoResourceAttr("tfe_notification_configuration.foobar", "url"),
				),
			},
			{
				// Increment version manually to trigger URL update
				Config: testAccTFENotificationConfiguration_urlWriteOnlyManual(rInt, runTasksURL()+"?v2=true", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_notification_configuration.foobar", "url_wo_version", "2"),
				),
			},
		},
	})
}

// TestAccTFENotificationConfiguration_urlWriteOnlyValidation tests that schema
// validators reject invalid combinations for url_wo.
func TestAccTFENotificationConfiguration_urlWriteOnlyValidation(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckTFENotificationConfiguration(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFENotificationConfiguration_urlAndUrlWriteOnly(rInt),
				ExpectError: regexp.MustCompile(`Attribute "url_wo" cannot be specified when "url" is specified`),
			},
			{
				Config:      testAccTFENotificationConfiguration_urlWriteOnlyVersionWithoutURL(rInt),
				ExpectError: regexp.MustCompile(`Attribute "url_wo" must be specified when "url_wo_version" is specified`),
			},
		},
	})
}

func testAccTFENotificationConfiguration_urlWriteOnly(rInt int, url string) string {
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
  url_wo           = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, url)
}

func testAccTFENotificationConfiguration_urlWriteOnlyManual(rInt int, url string, version int64) string {
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
  url_wo           = "%s"
  url_wo_version   = %d
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, url, version)
}

func testAccTFENotificationConfiguration_urlAndUrlWriteOnly(rInt int) string {
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
  url              = "%s"
  url_wo           = "%s"
  workspace_id     = tfe_workspace.foobar.id
}`, rInt, runTasksURL(), runTasksURL())
}

func testAccTFENotificationConfiguration_urlWriteOnlyVersionWithoutURL(rInt int) string {
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
  url_wo_version   = 1
  workspace_id     = tfe_workspace.foobar.id
}`, rInt)
}
