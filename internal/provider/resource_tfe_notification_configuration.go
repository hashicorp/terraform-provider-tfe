// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFENotificationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFENotificationConfigurationCreate,
		Read:   resourceTFENotificationConfigurationRead,
		Update: resourceTFENotificationConfigurationUpdate,
		Delete: resourceTFENotificationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"destination_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.NotificationDestinationTypeEmail),
						string(tfe.NotificationDestinationTypeGeneric),
						string(tfe.NotificationDestinationTypeSlack),
						string(tfe.NotificationDestinationTypeMicrosoftTeams),
					},
					false,
				),
			},

			"email_addresses": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"token", "url"},
			},

			"email_user_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"token", "url"},
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"triggers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{
							string(tfe.NotificationTriggerCreated),
							string(tfe.NotificationTriggerPlanning),
							string(tfe.NotificationTriggerNeedsAttention),
							string(tfe.NotificationTriggerApplying),
							string(tfe.NotificationTriggerCompleted),
							string(tfe.NotificationTriggerErrored),
							string(tfe.NotificationTriggerAssessmentCheckFailed),
							string(tfe.NotificationTriggerAssessmentDrifted),
							string(tfe.NotificationTriggerAssessmentFailed),
						},
						false,
					),
				},
			},

			"url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"email_addresses", "email_user_ids"},
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFENotificationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get workspace
	workspaceID := d.Get("workspace_id").(string)

	// Get attributes
	destinationType := tfe.NotificationDestinationType(d.Get("destination_type").(string))
	enabled := d.Get("enabled").(bool)
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	url := d.Get("url").(string)

	// Make sure only the correct schema attributes are set
	if destinationType == tfe.NotificationDestinationTypeEmail {
		// When destination_type is 'email':
		// 1. url and token cannot be set
		err := validateSchemaAttributesForDestinationTypeEmail(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeGeneric {
		// When destination_type is 'generic':
		// 1. email_addresses and email_user_ids cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeGeneric(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeSlack {
		// When destination_type is 'slack':
		// 1. email_addresses, email_user_ids, and token cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeSlack(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeMicrosoftTeams {
		// When destination_type is 'microsoft-teams':
		// 1. email_addresses, email_user_ids, and token cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeMicrosoftTeams(d)
		if err != nil {
			return err
		}
	}

	// Create a new options struct
	options := tfe.NotificationConfigurationCreateOptions{
		DestinationType: tfe.NotificationDestination(destinationType),
		Enabled:         tfe.Bool(enabled),
		Name:            tfe.String(name),
		Token:           tfe.String(token),
		URL:             tfe.String(url),
	}

	// Add triggers set to the options struct
	for _, trigger := range d.Get("triggers").(*schema.Set).List() {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.(string)))
	}

	// Add email_addresses set to the options struct
	if emailAddresses, ok := d.GetOk("email_addresses"); ok {
		for _, emailAddress := range emailAddresses.(*schema.Set).List() {
			options.EmailAddresses = append(options.EmailAddresses, emailAddress.(string))
		}
	}

	// Add email_user_ids set to the options struct
	if emailUserIDs, ok := d.GetOk("email_user_ids"); ok {
		for _, emailUserID := range emailUserIDs.(*schema.Set).List() {
			options.EmailUsers = append(options.EmailUsers, &tfe.User{ID: emailUserID.(string)})
		}
	}

	log.Printf("[DEBUG] Create notification configuration: %s", name)
	notificationConfiguration, err := config.Client.NotificationConfigurations.Create(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("Error creating notification configuration %s: %w", name, err)
	}

	d.SetId(notificationConfiguration.ID)

	return resourceTFENotificationConfigurationRead(d, meta)
}

func resourceTFENotificationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read notification configuration: %s", d.Id())
	notificationConfiguration, err := config.Client.NotificationConfigurations.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Notification configuration %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading notification configuration %s: %w", d.Id(), err)
	}

	// Update config
	d.Set("destination_type", notificationConfiguration.DestinationType)
	d.Set("enabled", notificationConfiguration.Enabled)

	// Update the email addresses
	var emailAddresses []interface{}
	for _, emailAddress := range notificationConfiguration.EmailAddresses {
		emailAddresses = append(emailAddresses, emailAddress)
	}
	d.Set("email_addresses", emailAddresses)

	// Update the email user ids
	var emailUserIDs []interface{}
	for _, emailUser := range notificationConfiguration.EmailUsers {
		emailUserIDs = append(emailUserIDs, emailUser.ID)
	}
	d.Set("email_user_ids", emailUserIDs)

	d.Set("name", notificationConfiguration.Name)
	// Don't set token here, as it is write only
	// and setting it here would make it blank
	d.Set("triggers", notificationConfiguration.Triggers)

	if notificationConfiguration.URL != "" {
		d.Set("url", notificationConfiguration.URL)
	}

	d.Set("workspace_id", notificationConfiguration.Subscribable.ID)

	return nil
}

func resourceTFENotificationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get attributes
	enabled := d.Get("enabled").(bool)
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	url := d.Get("url").(string)

	// Make sure only the correct schema attributes are set
	destinationType := tfe.NotificationDestinationType(d.Get("destination_type").(string))
	if destinationType == tfe.NotificationDestinationTypeEmail {
		// When destination_type is 'email':
		// 1. url and token cannot be set
		err := validateSchemaAttributesForDestinationTypeEmail(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeGeneric {
		// When destination_type is 'generic':
		// 1. email_addresses and email_user_ids cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeGeneric(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeSlack {
		// When destination_type is 'slack':
		// 1. email_addresses, email_user_ids, and token cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeSlack(d)
		if err != nil {
			return err
		}
	} else if destinationType == tfe.NotificationDestinationTypeMicrosoftTeams {
		// When destination_type is 'microsoft-teams':
		// 1. email_addresses, email_user_ids, and token cannot be set
		// 2. url must be set
		err := validateSchemaAttributesForDestinationTypeMicrosoftTeams(d)
		if err != nil {
			return err
		}
	}

	// Create a new options struct
	options := tfe.NotificationConfigurationUpdateOptions{
		Enabled: tfe.Bool(enabled),
		Name:    tfe.String(name),
		Token:   tfe.String(token),
		URL:     tfe.String(url),
	}

	// Add triggers set to the options struct
	for _, trigger := range d.Get("triggers").(*schema.Set).List() {
		options.Triggers = append(options.Triggers, tfe.NotificationTriggerType(trigger.(string)))
	}

	// Add email_addresses set to the options struct
	if emailAddresses, ok := d.GetOk("email_addresses"); ok {
		for _, emailAddress := range emailAddresses.(*schema.Set).List() {
			options.EmailAddresses = append(options.EmailAddresses, emailAddress.(string))
		}
	}

	// Add email_user_ids set to the options struct
	if emailUserIDs, ok := d.GetOk("email_user_ids"); ok {
		for _, emailUserID := range emailUserIDs.(*schema.Set).List() {
			options.EmailUsers = append(options.EmailUsers, &tfe.User{ID: emailUserID.(string)})
		}
	}

	log.Printf("[DEBUG] Update notification configuration: %s", d.Id())
	_, err := config.Client.NotificationConfigurations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating notification configuration %s: %w", d.Id(), err)
	}

	return resourceTFENotificationConfigurationRead(d, meta)
}

func resourceTFENotificationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete notification configuration: %s", d.Id())
	err := config.Client.NotificationConfigurations.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting notification configuration %s: %w", d.Id(), err)
	}

	return nil
}

// Custom CustomizeDiff functions and helpers
func validateSchemaAttributesForDestinationTypeEmail(d *schema.ResourceData) error {
	// Make sure url and token are not set when destination_type is 'email'
	_, urlIsSet := d.GetOk("url")
	if urlIsSet {
		return fmt.Errorf("URL cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeEmail))
	}
	token, tokenIsSet := d.GetOk("token")
	if tokenIsSet && token != "" {
		return fmt.Errorf("token cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeEmail))
	}

	return nil
}

func validateSchemaAttributesForDestinationTypeGeneric(d *schema.ResourceData) error {
	// Make sure email_addresses and email_user_ids are not set when destination_type is 'generic'
	_, emailAddressesIsSet := d.GetOk("email_addresses")
	if emailAddressesIsSet {
		return fmt.Errorf("email addresses cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeGeneric))
	}
	_, emailUserIDsIsSet := d.GetOk("email_user_ids")
	if emailUserIDsIsSet {
		return fmt.Errorf("email user IDs cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeGeneric))
	}

	// Make sure url is set when destination_type is 'generic'
	_, urlIsSet := d.GetOk("url")
	if !urlIsSet {
		return fmt.Errorf("URL is required with destination type of %s", string(tfe.NotificationDestinationTypeGeneric))
	}

	return nil
}

func validateSchemaAttributesForDestinationTypeSlack(d *schema.ResourceData) error {
	// Make sure email_addresses, email_user_ids, and token are not set when destination_type is 'slack'
	_, emailAddressesIsSet := d.GetOk("email_addresses")
	if emailAddressesIsSet {
		return fmt.Errorf("email addresses cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeSlack))
	}
	_, emailUserIDsIsSet := d.GetOk("email_user_ids")
	if emailUserIDsIsSet {
		return fmt.Errorf("email user IDs cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeSlack))
	}
	token, tokenIsSet := d.GetOk("token")
	if tokenIsSet && token != "" {
		return fmt.Errorf("token cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeSlack))
	}

	// Make sure url is set when destination_type is 'slack'
	_, urlIsSet := d.GetOk("url")
	if !urlIsSet {
		return fmt.Errorf("URL is required with destination type of %s", string(tfe.NotificationDestinationTypeSlack))
	}

	return nil
}

func validateSchemaAttributesForDestinationTypeMicrosoftTeams(d *schema.ResourceData) error {
	// Make sure email_addresses, email_user_ids, and token are not set when destination_type is 'microsoft-teams'
	_, emailAddressesIsSet := d.GetOk("email_addresses")
	if emailAddressesIsSet {
		return fmt.Errorf("email addresses cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeMicrosoftTeams))
	}
	_, emailUserIDsIsSet := d.GetOk("email_user_ids")
	if emailUserIDsIsSet {
		return fmt.Errorf("email user IDs cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeMicrosoftTeams))
	}
	token, tokenIsSet := d.GetOk("token")
	if tokenIsSet && token != "" {
		return fmt.Errorf("token cannot be set with destination type of %s", string(tfe.NotificationDestinationTypeMicrosoftTeams))
	}

	// Make sure url is set when destination_type is 'microsoft-teams'
	_, urlIsSet := d.GetOk("url")
	if !urlIsSet {
		return fmt.Errorf("URL is required with destination type of %s", string(tfe.NotificationDestinationTypeMicrosoftTeams))
	}

	return nil
}
