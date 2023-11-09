// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOrganizationModuleSharing() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "the tfe_organization_module_sharing resource is deprecated, please use tfe_admin_organization_settings instead",
		Create:             resourceTFEOrganizationModuleSharingCreate,
		Read:               resourceTFEOrganizationModuleSharingRead,
		Update:             resourceTFEOrganizationModuleSharingUpdate,
		Delete:             resourceTFEOrganizationModuleSharingDelete,
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, current string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, current)
				},
			},

			"module_consumers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceTFEOrganizationModuleSharingCreate(d *schema.ResourceData, meta interface{}) error {
	// Get the organization name that will share "produce" modules
	config := meta.(ConfiguredClient)

	producer, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Create %s module consumers", producer)
	d.SetId(producer)

	return resourceTFEOrganizationModuleSharingUpdate(d, meta)
}

func resourceTFEOrganizationModuleSharingUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	var consumers []string
	for _, name := range d.Get("module_consumers").([]interface{}) {
		// ignore empty strings
		if name == nil {
			continue
		}
		consumers = append(consumers, name.(string))
	}

	log.Printf("[DEBUG] Update %s module consumers", d.Id())
	err := config.Client.Admin.Organizations.UpdateModuleConsumers(ctx, d.Id(), consumers)
	if err != nil {
		return fmt.Errorf("error updating module consumers to %s: %w", d.Id(), err)
	}

	return resourceTFEOrganizationModuleSharingRead(d, meta)
}

func resourceTFEOrganizationModuleSharingRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	options := &tfe.AdminOrganizationListModuleConsumersOptions{}

	log.Printf("[DEBUG] Read configuration of module sharing for organization: %s", d.Id())
	for {
		consumerList, err := config.Client.Admin.Organizations.ListModuleConsumers(ctx, d.Id(), options)
		if err != nil {
			if err == tfe.ErrResourceNotFound {
				log.Printf("[DEBUG] Organization %s does not longer exist", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error reading organization %s module consumer list: %w", d.Id(), err)
		}

		if consumerList.CurrentPage >= consumerList.TotalPages {
			break
		}

		options.PageNumber = consumerList.NextPage
	}

	return nil
}

func resourceTFEOrganizationModuleSharingDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Disable module sharing for organization: %s", d.Id())
	err := config.Client.Admin.Organizations.UpdateModuleConsumers(ctx, d.Id(), []string{})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete module sharing for organization %s: %w", d.Id(), err)
	}

	return nil
}
