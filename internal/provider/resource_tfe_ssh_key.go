// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESSHKeyCreate,
		Read:   resourceTFESSHKeyRead,
		Update: resourceTFESSHKeyUpdate,
		Delete: resourceTFESSHKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceTFESSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := tfe.SSHKeyCreateOptions{
		Name:  tfe.String(name),
		Value: tfe.String(d.Get("key").(string)),
	}

	log.Printf("[DEBUG] Create new SSH key for organization: %s", organization)
	sshKey, err := config.Client.SSHKeys.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating SSH key %s for organization %s: %w", name, organization, err)
	}

	d.SetId(sshKey.ID)

	return resourceTFESSHKeyUpdate(d, meta)
}

func resourceTFESSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of SSH key: %s", d.Id())
	sshKey, err := config.Client.SSHKeys.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] SSH key %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of SSH key %s: %w", d.Id(), err)
	}

	// Update the config.
	d.Set("name", sshKey.Name)

	return nil
}

func resourceTFESSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Create a new options struct.
	options := tfe.SSHKeyUpdateOptions{
		Name: tfe.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Update SSH key: %s", d.Id())
	_, err := config.Client.SSHKeys.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating SSH key %s: %w", d.Id(), err)
	}

	return resourceTFESSHKeyRead(d, meta)
}

func resourceTFESSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete SSH key: %s", d.Id())
	err := config.Client.SSHKeys.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting SSH key %s: %w", d.Id(), err)
	}

	return nil
}
