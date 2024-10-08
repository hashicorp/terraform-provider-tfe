// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAuditTrailToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAuditTrailTokenCreate,
		Read:   resourceTFEAuditTrailTokenRead,
		Delete: resourceTFEAuditTrailTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEAuditTrailTokenImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"force_regenerate": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"expired_at": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEAuditTrailTokenCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	auditTrailTokenType := tfe.AuditTrailToken

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	readOptions := tfe.OrganizationTokenReadOptions{
		TokenType: &auditTrailTokenType,
	}
	log.Printf("[DEBUG] Check if a token already exists for organization: %s", organization)
	_, err = config.Client.OrganizationTokens.ReadWithOptions(ctx, organization, readOptions)
	if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
		return fmt.Errorf("error checking if a token exists for organization %s: %w", organization, err)
	}

	// If error is nil, the token already exists.
	if err == nil {
		if !d.Get("force_regenerate").(bool) {
			return fmt.Errorf("a token already exists for organization: %s", organization)
		}
		log.Printf("[DEBUG] Regenerating existing token for organization: %s", organization)
	}

	// Get the token create options.
	createOptions := tfe.OrganizationTokenCreateOptions{
		TokenType: &auditTrailTokenType,
	}

	// Check whether the optional expiry was provided.
	expiredAt, expiredAtProvided := d.GetOk("expired_at")

	// If an expiry was provided, parse it and update the options struct.
	if expiredAtProvided {
		expiry, err := time.Parse(time.RFC3339, expiredAt.(string))

		createOptions.ExpiredAt = &expiry

		if err != nil {
			return fmt.Errorf("%s must be a valid date or time, provided in iso8601 format", expiredAt)
		}
	}

	token, err := config.Client.OrganizationTokens.CreateWithOptions(ctx, organization, createOptions)
	if err != nil {
		return fmt.Errorf(
			"error creating new token for organization %s: %w", organization, err)
	}

	d.SetId(organization)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	d.Set("token", token.Token)

	return resourceTFEOrganizationTokenRead(d, meta)
}

func resourceTFEAuditTrailTokenRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	auditTrailTokenType := tfe.AuditTrailToken
	readOptions := tfe.OrganizationTokenReadOptions{
		TokenType: &auditTrailTokenType,
	}
	log.Printf("[DEBUG] Read the token from organization: %s", d.Id())
	_, err := config.Client.OrganizationTokens.ReadWithOptions(ctx, d.Id(), readOptions)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Token for organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading token from organization %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEAuditTrailTokenDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}
	auditTrailTokenType := tfe.AuditTrailToken
	deleteOptions := tfe.OrganizationTokenDeleteOptions{
		TokenType: &auditTrailTokenType,
	}
	log.Printf("[DEBUG] Delete token from organization: %s", organization)
	err = config.Client.OrganizationTokens.DeleteWithOptions(ctx, organization, deleteOptions)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("error deleting token from organization %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEAuditTrailTokenImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the organization field.
	d.Set("organization", d.Id())

	return []*schema.ResourceData{d}, nil
}
