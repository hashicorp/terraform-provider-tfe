package tfe

import (
	"fmt"
	"log"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFERegistryModule() *schema.Resource {
	return &schema.Resource{
		Description: "Terraform Cloud's private module registry helps you share Terraform modules across your organization.",

		Create: resourceTFERegistryModuleCreate,
		Read:   resourceTFERegistryModuleRead,
		Delete: resourceTFERegistryModuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFERegistryModuleImporter,
		},

		Schema: map[string]*schema.Schema{
			"organization": {
				Description: "The name of the organization associated with the registry module.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"module_provider": {
				Description: "The provider of the registry module.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of registry module.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcs_repo": {
				Description: "Settings for the registry module's VCS repository. Forces a new resource if changed.",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_identifier": {
							Description: "The display identifier for your VCS repository. For most VCS providers outside of BitBucket Cloud, this will match the `identifier` string.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"identifier": {
							Description: "A reference to your VCS repository in the format",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"oauth_token_id": {
							Description: "Token ID of the VCS Connection (OAuth Connection Token to use.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceTFERegistryModuleCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.RegistryModuleCreateWithVCSConnectionOptions{}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &tfe.RegistryModuleVCSRepoOptions{
			Identifier:        tfe.String(vcsRepo["identifier"].(string)),
			OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
			DisplayIdentifier: tfe.String(vcsRepo["display_identifier"].(string)),
		}
	}

	log.Printf("[DEBUG] Create registry module from repository %s", *options.VCSRepo.Identifier)
	registryModule, err := tfeClient.RegistryModules.CreateWithVCSConnection(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating registry module from repository %s: %w", *options.VCSRepo.Identifier, err)
	}

	err = resource.Retry(time.Duration(5)*time.Minute, func() *resource.RetryError {
		rmID := tfe.RegistryModuleID{
			Organization: registryModule.Organization.Name,
			Name:         registryModule.Name,
			Provider:     registryModule.Provider,
		}
		_, err := tfeClient.RegistryModules.Read(ctx, rmID)
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

	return resourceTFERegistryModuleRead(d, meta)
}

func resourceTFERegistryModuleRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read registry module: %s", d.Id())

	// Get the fields we need to read the registry module
	rmID := tfe.RegistryModuleID{
		Organization: d.Get("organization").(string),
		Name:         d.Get("name").(string),
		Provider:     d.Get("module_provider").(string),
	}

	registryModule, err := tfeClient.RegistryModules.Read(ctx, rmID)
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

	// Set VCS repo options.
	var vcsRepo []interface{}
	if registryModule.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":         registryModule.VCSRepo.Identifier,
			"oauth_token_id":     registryModule.VCSRepo.OAuthTokenID,
			"display_identifier": registryModule.VCSRepo.DisplayIdentifier,
		}
		vcsRepo = append(vcsRepo, vcsConfig)

		d.Set("vcs_repo", vcsRepo)
	}

	return nil
}

func resourceTFERegistryModuleDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete registry module: %s", d.Id())
	organization := d.Get("organization").(string)
	name := d.Get("name").(string)
	err := tfeClient.RegistryModules.Delete(ctx, organization, name)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting registry module %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFERegistryModuleImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	registryModuleInfo := strings.SplitN(d.Id(), "/", 4)
	if len(registryModuleInfo) != 4 {
		return nil, fmt.Errorf(
			"invalid registry module import format: %s (expected <ORGANIZATION>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("name", registryModuleInfo[1])
	d.Set("module_provider", registryModuleInfo[2])
	d.Set("organization", registryModuleInfo[0])
	d.SetId(registryModuleInfo[3])

	return []*schema.ResourceData{d}, nil
}
