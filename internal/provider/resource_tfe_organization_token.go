// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func resourceTFEOrganizationToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationTokenCreate,
		Read:   resourceTFEOrganizationTokenRead,
		Delete: resourceTFEOrganizationTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOrganizationTokenImporter,
		},

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

func resourceTFEOrganizationTokenCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Check if a token already exists for organization: %s", organization)
	_, err = config.Client.OrganizationTokens.Read(ctx, organization)
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
	options := tfe.OrganizationTokenCreateOptions{}

	// Check whether the optional expiry was provided.
	expiredAt, expiredAtProvided := d.GetOk("expired_at")

	// If an expiry was provided, parse it and update the options struct.
	if expiredAtProvided {
		expiry, err := time.Parse(time.RFC3339, expiredAt.(string))

		options.ExpiredAt = &expiry

		if err != nil {
			return fmt.Errorf("%s must be a valid date or time, provided in iso8601 format", expiredAt)
		}
	}

	token, err := config.Client.OrganizationTokens.CreateWithOptions(ctx, organization, options)
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

func resourceTFEOrganizationTokenRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read the token from organization: %s", d.Id())
	_, err := config.Client.OrganizationTokens.Read(ctx, d.Id())
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

func resourceTFEOrganizationTokenDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Delete token from organization: %s", organization)
	err = config.Client.OrganizationTokens.Delete(ctx, organization)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("error deleting token from organization %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationTokenImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the organization field.
	d.Set("organization", d.Id())

	return []*schema.ResourceData{d}, nil
}
