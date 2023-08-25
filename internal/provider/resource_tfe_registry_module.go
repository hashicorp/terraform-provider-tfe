// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFERegistryModule() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFERegistryModuleCreate,
		Read:   resourceTFERegistryModuleRead,
		Update: resourceTFERegistryModuleUpdate,
		Delete: resourceTFERegistryModuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFERegistryModuleImporter,
		},

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"module_provider": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"vcs_repo"},
				RequiredWith: []string{"organization", "name"},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"vcs_repo": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_identifier": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"identifier": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"oauth_token_id": {
							Type:          schema.TypeString,
							ForceNew:      true,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.github_app_installation_id"},
						},
						"github_app_installation_id": {
							Type:          schema.TypeString,
							ForceNew:      true,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.oauth_token_id"},
							AtLeastOneOf:  []string{"vcs_repo.0.oauth_token_id", "vcs_repo.0.github_app_installation_id"},
						},
					},
				},
			},
			"namespace": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"registry_name"},
			},
			"no_code": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"registry_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"module_provider"},
				ValidateFunc: validation.StringInSlice(
					[]string{"private", "public"},
					true,
				),
			},
		},
	}
}

func resourceTFERegistryModuleCreateWithVCS(v interface{}, meta interface{}, d *schema.ResourceData) (*tfe.RegistryModule, error) {
	config := meta.(ConfiguredClient)
	// Create module with VCS repo configuration block.
	options := tfe.RegistryModuleCreateWithVCSConnectionOptions{}
	vcsRepo := v.([]interface{})[0].(map[string]interface{})

	orgName, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		log.Printf("[WARN] Error getting organization name: %s", err)
	}

	options.VCSRepo = &tfe.RegistryModuleVCSRepoOptions{
		Identifier:        tfe.String(vcsRepo["identifier"].(string)),
		GHAInstallationID: tfe.String(vcsRepo["github_app_installation_id"].(string)),
		DisplayIdentifier: tfe.String(vcsRepo["display_identifier"].(string)),
		OrganizationName:  tfe.String(orgName),
	}

	if vcsRepo["oauth_token_id"] != nil && vcsRepo["oauth_token_id"].(string) != "" {
		options.VCSRepo.OAuthTokenID = tfe.String(vcsRepo["oauth_token_id"].(string))
	}

	log.Printf("[DEBUG] Create registry module from repository %s", *options.VCSRepo.Identifier)
	registryModule, err := config.Client.RegistryModules.CreateWithVCSConnection(ctx, options)
	if err != nil {
		return nil, fmt.Errorf(
			"Error creating registry module from repository %s: %w", *options.VCSRepo.Identifier, err)
	}
	return registryModule, nil
}

func resourceTFERegistryModuleCreateWithoutVCS(meta interface{}, d *schema.ResourceData) (*tfe.RegistryModule, error) {
	config := meta.(ConfiguredClient)

	options := tfe.RegistryModuleCreateOptions{
		Name:     tfe.String(d.Get("name").(string)),
		Provider: tfe.String(d.Get("module_provider").(string)),
	}

	if v, ok := d.GetOk("no_code"); ok {
		log.Println("[WARN] The attribute no_code is deprecated as of release 0.44.0 and may be removed in a future version. The preferred way to create a no-code registry module is to use the tfe_no_code_module resource.")
		options.NoCode = tfe.Bool(v.(bool))
	}

	if registryName, ok := d.GetOk("registry_name"); ok {
		options.RegistryName = tfe.RegistryName(registryName.(string))

		if registryName.(string) == "public" {
			options.Namespace = d.Get("namespace").(string)
		}
	}

	orgName := d.Get("organization").(string)

	log.Printf("[DEBUG] Create registry module named %s", *options.Name)
	registryModule, err := config.Client.RegistryModules.Create(ctx, orgName, options)

	if err != nil {
		return nil, fmt.Errorf("Error creating registry module %s: %w", *options.Name, err)
	}

	return registryModule, nil
}

func resourceTFERegistryModuleCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	var registryModule *tfe.RegistryModule
	var err error

	if v, ok := d.GetOk("vcs_repo"); ok {
		registryModule, err = resourceTFERegistryModuleCreateWithVCS(v, meta, d)
	} else {
		registryModule, err = resourceTFERegistryModuleCreateWithoutVCS(meta, d)
	}

	if err != nil {
		return err
	}

	err = resource.Retry(time.Duration(5)*time.Minute, func() *resource.RetryError {
		rmID := tfe.RegistryModuleID{
			Organization: registryModule.Organization.Name,
			Name:         registryModule.Name,
			Provider:     registryModule.Provider,
			Namespace:    registryModule.Namespace,
			RegistryName: registryModule.RegistryName,
		}
		_, err := config.Client.RegistryModules.Read(ctx, rmID)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "not found") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error while waiting for module %s/%s to be ingested: %w", registryModule.Organization.Name, registryModule.Name, err)
	}

	d.SetId(registryModule.ID)

	// Set these fields so we have the information needed to read the registry module
	d.Set("name", registryModule.Name)
	d.Set("module_provider", registryModule.Provider)
	d.Set("organization", registryModule.Organization.Name)
	d.Set("namespace", registryModule.Namespace)
	d.Set("registry_name", registryModule.RegistryName)

	return resourceTFERegistryModuleRead(d, meta)
}

func resourceTFERegistryModuleUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	options := tfe.RegistryModuleUpdateOptions{}
	if v, ok := d.GetOk("no_code"); ok {
		log.Println("[WARN] The attribute no_code is deprecated as of release 0.44.0 and may be removed in a future version. The preferred way to create a no-code registry module is to use the tfe_no_code_module resource.")
		options.NoCode = tfe.Bool(v.(bool))
	}

	var registryModule *tfe.RegistryModule
	var err error

	rmID := tfe.RegistryModuleID{
		Organization: d.Get("organization").(string),
		Name:         d.Get("name").(string),
		Provider:     d.Get("module_provider").(string),
		Namespace:    d.Get("namespace").(string),
		RegistryName: tfe.RegistryName(d.Get("registry_name").(string)),
	}

	err = resource.Retry(time.Duration(5)*time.Minute, func() *resource.RetryError {
		registryModule, err = config.Client.RegistryModules.Update(ctx, rmID, options)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error while waiting for module %s/%s to be updated: %w", registryModule.Organization.Name, registryModule.Name, err)
	}

	d.SetId(registryModule.ID)

	return resourceTFERegistryModuleRead(d, meta)
}

func resourceTFERegistryModuleRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read registry module: %s", d.Id())

	// Get the fields we need to read the registry module
	rmID := tfe.RegistryModuleID{
		Organization: d.Get("organization").(string),
		Name:         d.Get("name").(string),
		Provider:     d.Get("module_provider").(string),
		Namespace:    d.Get("namespace").(string),
		RegistryName: tfe.RegistryName(d.Get("registry_name").(string)),
	}

	registryModule, err := config.Client.RegistryModules.Read(ctx, rmID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Registry module %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading registry module %s: %w", d.Id(), err)
	}

	// Update the config
	log.Printf("[DEBUG] Update config for registry module: %s", d.Id())
	d.Set("name", registryModule.Name)
	d.Set("module_provider", registryModule.Provider)
	d.Set("organization", registryModule.Organization.Name)
	d.Set("namespace", registryModule.Namespace)
	d.Set("registry_name", registryModule.RegistryName)
	d.Set("no_code", registryModule.NoCode)

	// Set VCS repo options.
	var vcsRepo []interface{}
	if registryModule.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":                 registryModule.VCSRepo.Identifier,
			"oauth_token_id":             registryModule.VCSRepo.OAuthTokenID,
			"github_app_installation_id": registryModule.VCSRepo.GHAInstallationID,
			"display_identifier":         registryModule.VCSRepo.DisplayIdentifier,
		}
		vcsRepo = append(vcsRepo, vcsConfig)

		d.Set("vcs_repo", vcsRepo)
	}

	return nil
}

func resourceTFERegistryModuleDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete registry module: %s", d.Id())
	organization := d.Get("organization").(string)
	name := d.Get("name").(string)
	err := config.Client.RegistryModules.Delete(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting registry module %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFERegistryModuleImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	registryModuleInfo := strings.SplitN(d.Id(), "/", 6)
	if len(registryModuleInfo) == 4 {
		// for format: <ORGANIZATION>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>
		log.Printf("[WARN] The import format <ORGANIZATION>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID> is deprecated as of release 0.33.0 and may be removed in a future version. The preferred format is <ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>.")
		d.Set("organization", registryModuleInfo[0])
		d.Set("name", registryModuleInfo[1])
		d.Set("module_provider", registryModuleInfo[2])
		d.SetId(registryModuleInfo[3])

		return []*schema.ResourceData{d}, nil
	} else if len(registryModuleInfo) == 6 {
		// for format: <ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>
		// see https://developer.hashicorp.com/terraform/cloud-docs/api-docs/private-registry/modules#get-a-module
		d.Set("organization", registryModuleInfo[0])
		d.Set("registry_name", registryModuleInfo[1])
		d.Set("namespace", registryModuleInfo[2])
		d.Set("name", registryModuleInfo[3])
		d.Set("module_provider", registryModuleInfo[4])
		d.SetId(registryModuleInfo[5])

		return []*schema.ResourceData{d}, nil
	}

	return nil, fmt.Errorf(
		"invalid registry module import format: %s (expected <ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>)",
		d.Id(),
	)
}
