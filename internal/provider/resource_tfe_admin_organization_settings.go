// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEAdminOrganizationSettingsCreate,
		Read:   resourceTFEAdminOrganizationSettingsRead,
		Update: resourceTFEAdminOrganizationSettingsUpdate,
		Delete: resourceTFEAdminOrganizationSettingsDelete,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"access_beta_tools": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"global_module_sharing": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sso_enabled": {
				Computed: true,
				Type:     schema.TypeBool,
			},
			"workspace_limit": {
				Optional: true,
				Type:     schema.TypeInt,
			},
			"module_sharing_consumer_organizations": {
				Optional: true,
				Computed: true,

				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTFEAdminOrganizationSettingsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name.
	name, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Read configuration of admin organization: %s", name)
	org, err := config.Client.Admin.Organizations.Read(ctx, name)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("failed to read admin organization %s: %w", name, err)
	}

	// Update the config.
	d.Set("organization", org.Name)
	d.Set("access_beta_tools", org.AccessBetaTools)
	d.Set("global_module_sharing", org.GlobalModuleSharing)
	d.Set("sso_enabled", org.SsoEnabled)
	d.Set("workspace_limit", org.WorkspaceLimit)
	d.SetId(org.Name)

	consumerOrgNames := make([]string, 0, 20)
	if org.GlobalModuleSharing != nil && !*org.GlobalModuleSharing {
		options := &tfe.AdminOrganizationListModuleConsumersOptions{}

		log.Printf("[DEBUG] Read configuration of module sharing for organization: %s", d.Id())
		for {
			consumerList, err := config.Client.Admin.Organizations.ListModuleConsumers(ctx, d.Id(), options)
			if err != nil {
				if errors.Is(err, tfe.ErrResourceNotFound) {
					log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
					d.SetId("")
					return nil
				}
				return fmt.Errorf("Error reading organization %s module consumer list: %w", d.Id(), err)
			}

			for _, c := range consumerList.Items {
				consumerOrgNames = append(consumerOrgNames, c.Name)
			}

			if consumerList.CurrentPage >= consumerList.TotalPages {
				break
			}

			options.PageNumber = consumerList.NextPage
		}
	}

	d.Set("module_sharing_consumer_organizations", consumerOrgNames)

	return nil
}

func resourceTFEAdminOrganizationSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceTFEAdminOrganizationSettingsUpdate(d, meta)
}

func resourceTFEAdminOrganizationSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func resourceTFEAdminOrganizationSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	name, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}
	globalModuleSharing := d.Get("global_module_sharing").(bool)

	_, err = config.Client.Admin.Organizations.Update(ctx, name, tfe.AdminOrganizationUpdateOptions{
		AccessBetaTools:     tfe.Bool(d.Get("access_beta_tools").(bool)),
		GlobalModuleSharing: tfe.Bool(globalModuleSharing),
		WorkspaceLimit:      tfe.Int(d.Get("workspace_limit").(int)),
	})

	if err != nil {
		return fmt.Errorf("failed to update admin organization settings: %w", err)
	}

	set := d.Get("module_sharing_consumer_organizations").(*schema.Set)
	if globalModuleSharing && set != nil {
		if set.Len() > 0 {
			return fmt.Errorf("global_module_sharing cannot be true if module_sharing_consumer_organizations are set")
		}
	}

	if !globalModuleSharing && set != nil && set.Len() > 0 {
		if err != nil {
			return fmt.Errorf("failed to fetch admin organizations for module consumer ids: %w", err)
		}

		// Copy set to list of string
		consumerOrgNames := make([]string, set.Len())
		for i, v := range set.List() {
			consumerOrgNames[i] = v.(string)
		}

		err = config.Client.Admin.Organizations.UpdateModuleConsumers(ctx, name, consumerOrgNames)
		if err != nil {
			return fmt.Errorf("failed to update organization module consumers: %w", err)
		}
	}

	return resourceTFEAdminOrganizationSettingsRead(d, meta)
}
