// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"
	"log"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/admin"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEAdminOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		Description: "(Only for Terraform Enterprise) Manages admin settings for an organization." +
			"\n\nThis resource requires the use of an admin token. See example usage for incorporating an admin token in your provider config.",

		Create: resourceTFEAdminOrganizationSettingsCreate,
		Read:   resourceTFEAdminOrganizationSettingsRead,
		Update: resourceTFEAdminOrganizationSettingsUpdate,
		Delete: resourceTFEAdminOrganizationSettingsDelete,

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this resource. Do not rely on this value — use `organization` instead.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"organization": {
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"access_beta_tools": {
				Description: "If true, the organization has access to beta tool versions.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"global_module_sharing": {
				Description: "If true, modules in the organization's private module repository will be available to all other organizations. Enabling this will disable any previously configured `module_sharing_consumer_organizations`. Cannot be true if `module_sharing_consumer_organizations` is set.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"sso_enabled": {
				Description: "If true, SSO is enabled in this organization.",
				Computed:    true,
				Type:        schema.TypeBool,
			},
			"workspace_limit": {
				Description: "Maximum number of workspaces for this organization. If this number is set to a value lower than the number of workspaces the organization has, it will prevent additional workspaces from being created, but existing workspaces will not be affected. If set to 0, this limit will have no effect.",
				Optional:    true,
				Type:        schema.TypeInt,
			},
			"module_sharing_consumer_organizations": {
				Description: "A list of organization names with which to share modules in the organization's private module repository. Cannot be set if `global_module_sharing` is true.",
				Optional:    true,
				Computed:    true,

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
	env, err := config.ClientV2.API.Admin().Organizations().ByOrganization_name(name).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfev2.ErrNotFound) {
			log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("failed to read admin organization %s: %w", name, err)
	}

	org := env.GetData()
	if org == nil {
		log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	attrs := org.GetAttributes()
	if attrs == nil {
		return fmt.Errorf("failed to read admin organization %s: API returned no attributes", name)
	}

	// Update the config.
	orgName := valueOrZero(org.GetId())
	d.Set("organization", orgName)
	d.Set("access_beta_tools", valueOrZero(attrs.GetAccessBetaTools()))
	d.Set("global_module_sharing", valueOrZero(attrs.GetGlobalModuleSharing()))
	d.Set("sso_enabled", valueOrZero(attrs.GetSsoEnabled()))
	d.Set("workspace_limit", int(valueOrZero(attrs.GetWorkspaceLimit())))
	d.SetId(orgName)

	consumerOrgNames := make([]string, 0, 20)
	if globalModuleSharing := attrs.GetGlobalModuleSharing(); globalModuleSharing != nil && !*globalModuleSharing {
		names, err := listAdminOrganizationModuleConsumers(config, d.Id())
		if err != nil {
			if errors.Is(err, tfev2.ErrNotFound) {
				log.Printf("[DEBUG] Organization %s no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error reading organization %s module consumer list: %w", d.Id(), err)
		}
		consumerOrgNames = names
	}

	d.Set("module_sharing_consumer_organizations", consumerOrgNames)

	return nil
}

// listAdminOrganizationModuleConsumers returns the names of all organizations
// that are permitted to consume modules from orgName's private module
// registry, following pagination until it is exhausted.
func listAdminOrganizationModuleConsumers(config ConfiguredClient, orgName string) ([]string, error) {
	consumerOrgNames := make([]string, 0, 20)

	log.Printf("[DEBUG] Read configuration of module sharing for organization: %s", orgName)
	queryParams := &admin.OrganizationsItemRelationshipsModuleConsumersRequestBuilderGetQueryParameters{}
	for {
		consumerList, err := config.ClientV2.API.Admin().Organizations().ByOrganization_name(orgName).Relationships().ModuleConsumers().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, err
		}

		for _, c := range consumerList.GetData() {
			consumerOrgNames = append(consumerOrgNames, valueOrZero(c.GetId()))
		}

		nextPage := nextPageFromMeta(consumerList.GetMeta())
		if nextPage == nil {
			break
		}

		queryParams = &admin.OrganizationsItemRelationshipsModuleConsumersRequestBuilderGetQueryParameters{
			Pagenumber: nextPage,
		}
	}

	return consumerOrgNames, nil
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
	accessBetaTools := d.Get("access_beta_tools").(bool)
	workspaceLimit := int32(d.Get("workspace_limit").(int)) //nolint:gosec // workspace_limit is a small user-supplied count, never near int32 overflow

	attrs := models.NewAdminOrganizations_attributes()
	attrs.SetAccessBetaTools(&accessBetaTools)
	attrs.SetGlobalModuleSharing(&globalModuleSharing)
	attrs.SetWorkspaceLimit(&workspaceLimit)

	orgType := models.ORGANIZATIONS_ADMINORGANIZATIONS_TYPE
	org := models.NewAdminOrganizations()
	org.SetTypeEscaped(&orgType)
	org.SetAttributes(attrs)
	body := models.NewAdminOrganizationsEnvelope()
	body.SetData(org)

	_, err = config.ClientV2.API.Admin().Organizations().ByOrganization_name(name).Patch(ctx, body, nil)
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
		// Copy set to list of string
		consumerOrgNames := make([]string, set.Len())
		for i, v := range set.List() {
			consumerOrgNames[i] = v.(string)
		}

		err = updateAdminOrganizationModuleConsumers(config, name, consumerOrgNames)
		if err != nil {
			return fmt.Errorf("failed to update organization module consumers: %w", err)
		}
	}

	return resourceTFEAdminOrganizationSettingsRead(d, meta)
}

// updateAdminOrganizationModuleConsumers replaces the full set of
// organizations permitted to consume modules from orgName's private module
// registry.
func updateAdminOrganizationModuleConsumers(config ConfiguredClient, orgName string, consumerOrgNames []string) error {
	identifiers := make([]models.JsonapiResourceIdentifierable, 0, len(consumerOrgNames))
	for _, name := range consumerOrgNames {
		identifier := models.NewJsonapiResourceIdentifier()
		identifier.SetId(&name)
		identifier.SetTypeEscaped(ptr("organizations"))
		identifiers = append(identifiers, identifier)
	}

	doc := models.NewJsonapiIdentifierArrayDocument()
	doc.SetData(identifiers)

	return config.ClientV2.API.Admin().Organizations().ByOrganization_name(orgName).Relationships().ModuleConsumers().Patch(ctx, doc, nil)
}
