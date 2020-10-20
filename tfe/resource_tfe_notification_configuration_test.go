package tfe

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basic, rInt),
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

func TestAccTFENotificationConfiguration_basicWorkspaceID(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basicWorkspaceID, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_emailUserIDs, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basic, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_update, rInt),
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

func TestAccTFENotificationConfiguration_updateWorkspaceID(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basicWorkspaceID, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_updateWorkspaceID, rInt),
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

func TestAccTFENotificationConfiguration_updateWorkspaceExternalIDToWorkspaceID(t *testing.T) {
	notificationConfiguration := &tfe.NotificationConfiguration{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basic, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basicWorkspaceID, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_updateWorkspaceID, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_emailUserIDs, rInt),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_updateEmailUserIDs, rInt),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_emailWithURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL cannot be set with destination type of email`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_emailWithToken, rInt),
				ExpectError: regexp.MustCompile(`^.*Token cannot be set with destination type of email`),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithEmailAddresses, rInt),
				ExpectError: regexp.MustCompile(`^.*Email addresses cannot be set with destination type of generic`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithEmailUserIDs, rInt),
				ExpectError: regexp.MustCompile(`^.*Email user IDs cannot be set with destination type of generic`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithoutURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL is required with destination type of generic`),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithEmailAddresses, rInt),
				ExpectError: regexp.MustCompile(`^.*Email addresses cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithEmailUserIDs, rInt),
				ExpectError: regexp.MustCompile(`^.*Email user IDs cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithToken, rInt),
				ExpectError: regexp.MustCompile(`^.*Token cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithoutURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL is required with destination type of slack`),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_emailUserIDs, rInt),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_emailWithURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL cannot be set with destination type of email`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_emailWithToken, rInt),
				ExpectError: regexp.MustCompile(`^.*Token cannot be set with destination type of email`),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_basic, rInt),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithEmailAddresses, rInt),
				ExpectError: regexp.MustCompile(`^.*Email addresses cannot be set with destination type of generic`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithEmailUserIDs, rInt),
				ExpectError: regexp.MustCompile(`^.*Email user IDs cannot be set with destination type of generic`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_genericWithoutURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL is required with destination type of generic`),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_slack, rInt),
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
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithEmailAddresses, rInt),
				ExpectError: regexp.MustCompile(`^.*Email addresses cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithEmailUserIDs, rInt),
				ExpectError: regexp.MustCompile(`^.*Email user IDs cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithToken, rInt),
				ExpectError: regexp.MustCompile(`^.*Token cannot be set with destination type of slack`),
			},
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_slackWithoutURL, rInt),
				ExpectError: regexp.MustCompile(`^.*URL is required with destination type of slack`),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_duplicateTriggers, rInt),
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

func TestAccTFENotificationConfiguration_noWorkspaceID(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testAccTFENotificationConfiguration_noWorkspaceID, rInt),
				ExpectError: regexp.MustCompile(`One of workspace_id or workspace_external_id must be set`),
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
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_update, rInt),
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

func TestAccTFENotificationConfigurationImport_emailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_updateEmailUserIDs, rInt),
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

func TestAccTFENotificationConfigurationImport_emptyEmailUserIDs(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFENotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccTFENotificationConfiguration_emailUserIDs, rInt),
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

		if !reflect.DeepEqual(notificationConfiguration.Triggers, []string{tfe.NotificationTriggerCreated, tfe.NotificationTriggerNeedsAttention}) {
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

var testAccTFENotificationConfiguration_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_basic"
  destination_type      = "generic"
  url                   = "http://example.com"
  workspace_external_id = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_emailUserIDs = `
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
}`

const testAccTFENotificationConfiguration_slack = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_slack"
  destination_type      = "slack"
  url                   = "http://example.com"
  workspace_external_id = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_basicWorkspaceID = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_basic"
  destination_type      = "generic"
  url                   = "http://example.com"
  workspace_id          = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_update = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_update"
  destination_type      = "generic"
  enabled               = true
  token                 = "1234567890_update"
  triggers              = ["run:created", "run:needs_attention"]
  url                   = "http://example.com/?update=true"
  workspace_external_id = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_updateWorkspaceID = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_update"
  destination_type      = "generic"
  enabled               = true
  token                 = "1234567890_update"
  triggers              = ["run:created", "run:needs_attention"]
  url                   = "http://example.com/?update=true"
  workspace_id          = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_updateEmailUserIDs = `
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
  name             = "notification_email_update"
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.foobar.user_id]
  enabled          = true
  triggers         = ["run:created", "run:needs_attention"]
  workspace_id     = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_emailWithURL = `
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
}`

const testAccTFENotificationConfiguration_emailWithToken = `
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
}`

const testAccTFENotificationConfiguration_genericWithEmailAddresses = `
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
}`

const testAccTFENotificationConfiguration_genericWithEmailUserIDs = `
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
  email_user_ids   = ["${tfe_organization_membership.foobar.id}"]
  workspace_id     = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_genericWithoutURL = `
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
}`

const testAccTFENotificationConfiguration_slackWithEmailAddresses = `
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
}`

const testAccTFENotificationConfiguration_slackWithEmailUserIDs = `
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
  email_user_ids   = ["${tfe_organization_membership.foobar.id}"]
  workspace_id     = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_slackWithToken = `
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
}`

const testAccTFENotificationConfiguration_slackWithoutURL = `
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
}`

const testAccTFENotificationConfiguration_duplicateTriggers = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_duplicate_triggers"
  destination_type      = "generic"
  triggers              = ["run:created", "run:created", "run:created"]
  url                   = "http://example.com"
  workspace_external_id = tfe_workspace.foobar.id
}`

const testAccTFENotificationConfiguration_noWorkspaceID = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_notification_configuration" "foobar" {
  name                  = "notification_basic"
  destination_type      = "generic"
  url                   = "http://example.com"
}`
