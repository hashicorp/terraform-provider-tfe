// Copyright IBM Corp. 2018, 2026
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

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOrganizationToken() *schema.Resource {
	return &schema.Resource{
		Description: "Generates a new organization token, replacing any existing token, which can be used to act as the organization service account.",

		Create: resourceTFEOrganizationTokenCreate,
		Read:   resourceTFEOrganizationTokenRead,
		Delete: resourceTFEOrganizationTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEOrganizationTokenImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the token.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"force_regenerate": {
				Description: "If set to `true`, a new token will be generated even if a token already exists. This will invalidate the existing token!",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},

			"token": {
				Description: "The generated token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},

			"expired_at": {
				Description: "The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., \"2024-12-31T23:59:59Z\"). If no expiration date is supplied, the token will expire 24 months from creation.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

// newOrganizationTokenEnvelope builds the request body for creating an organization token.
func newOrganizationTokenEnvelope(expiry *time.Time) *models.AuthenticationTokensEnvelope {
	attributes := models.NewAuthenticationTokens_attributes()
	attributes.SetExpiredAt(expiry)

	tokenType := models.AUTHENTICATIONTOKENS_AUTHENTICATIONTOKENS_TYPE
	data := models.NewAuthenticationTokens()
	data.SetTypeEscaped(&tokenType)
	data.SetAttributes(attributes)

	envelope := models.NewAuthenticationTokensEnvelope()
	envelope.SetData(data)
	return envelope
}

func resourceTFEOrganizationTokenCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the organization name.
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Issue warning if expired_at is not provided
	if _, ok := d.GetOk("expired_at"); !ok {
		log.Printf("[WARN] The 'expired_at' attribute is not set for organization token. The token will default to an expiration of 24 months from creation.")
	}

	log.Printf("[DEBUG] Check if a token already exists for organization: %s", organization)
	_, err = config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Get(ctx, nil)
	if err != nil && !errors.Is(err, tfe.ErrNotFound) {
		return fmt.Errorf("error checking if a token exists for organization %s: %w", organization, err)
	}

	// If error is nil, the token already exists.
	if err == nil {
		if !d.Get("force_regenerate").(bool) {
			return fmt.Errorf("a token already exists for organization: %s", organization)
		}
		log.Printf("[DEBUG] Regenerating existing token for organization: %s", organization)
	}

	// Check whether the optional expiry was provided.
	expiredAt, expiredAtProvided := d.GetOk("expired_at")

	// If an expiry was provided, parse it and update the options struct.
	var expiry *time.Time
	if expiredAtProvided {
		parsed, err := time.Parse(time.RFC3339, expiredAt.(string))
		if err != nil {
			return fmt.Errorf("%s must be a valid date or time, provided in iso8601 format", expiredAt)
		}
		expiry = &parsed
	}

	envelope := newOrganizationTokenEnvelope(expiry)

	tokenEnvelope, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Post(ctx, envelope, nil)
	if err != nil {
		return fmt.Errorf(
			"error creating new token for organization %s: %w", organization, err)
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		return fmt.Errorf("error creating new token for organization %s: no data was returned by the API", organization)
	}
	token := tokenEnvelope.GetData()

	d.SetId(organization)

	// We need to set this here in the create function as this value will
	// only be returned once during the creation of the token.
	if attrs := token.GetAttributes(); attrs != nil {
		d.Set("token", valueOrZero(attrs.GetToken()))
		if expiredAt := attrs.GetExpiredAt(); expiredAt != nil {
			d.Set("expired_at", expiredAt.Format(time.RFC3339))
		}
	}
	return resourceTFEOrganizationTokenRead(d, meta)
}

func resourceTFEOrganizationTokenRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read the token from organization: %s", d.Id())
	tokenEnvelope, err := config.ClientV2.API.Organizations().ByOrganization_name(d.Id()).AuthenticationToken().Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
			log.Printf("[DEBUG] Token for organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading token from organization %s: %w", d.Id(), err)
	}
	if tokenEnvelope == nil || tokenEnvelope.GetData() == nil {
		log.Printf("[DEBUG] Token for organization %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	// if expired_at was set to null at creation, the API returns a default value of 24 months from the creation date.
	if attrs := tokenEnvelope.GetData().GetAttributes(); attrs != nil {
		if expiredAt := attrs.GetExpiredAt(); expiredAt != nil {
			d.Set("expired_at", expiredAt.Format(time.RFC3339))
		}
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
	err = config.ClientV2.API.Organizations().ByOrganization_name(organization).AuthenticationToken().Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfe.ErrNotFound) {
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
