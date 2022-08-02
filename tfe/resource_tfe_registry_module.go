package tfe

import (
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
			State: resourceTFERegistryModuleImporter,
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
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
				ForceNew: false,
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

func resourceTFERegistryModuleCreateWithVCS(v interface{}, meta interface{}) (*tfe.RegistryModule, error) {
	tfeClient := meta.(*tfe.Client)
	// Create module with VCS repo configuration block.
	options := tfe.RegistryModuleCreateWithVCSConnectionOptions{}
	vcsRepo := v.([]interface{})[0].(map[string]interface{})

	options.VCSRepo = &tfe.RegistryModuleVCSRepoOptions{
		Identifier:        tfe.String(vcsRepo["identifier"].(string)),
		OAuthTokenID:      tfe.String(vcsRepo["oauth_token_id"].(string)),
		DisplayIdentifier: tfe.String(vcsRepo["display_identifier"].(string)),
	}

	log.Printf("[DEBUG] Create registry module from repository %s", *options.VCSRepo.Identifier)
	registryModule, err := tfeClient.RegistryModules.CreateWithVCSConnection(ctx, options)
	if err != nil {
		return nil, fmt.Errorf(
			"Error creating registry module from repository %s: %w", *options.VCSRepo.Identifier, err)
	}
	return registryModule, nil
}

func resourceTFERegistryModuleCreateWithoutVCS(meta interface{}, d *schema.ResourceData) (*tfe.RegistryModule, error) {
	tfeClient := meta.(*tfe.Client)

	options := tfe.RegistryModuleCreateOptions{
		Name:     tfe.String(d.Get("name").(string)),
		Provider: tfe.String(d.Get("module_provider").(string)),
		NoCode:   d.Get("no_code").(bool),
	}

	if registryName, ok := d.GetOk("registry_name"); ok {
		options.RegistryName = tfe.RegistryName(registryName.(string))

		if registryName.(string) == "public" {
			options.Namespace = d.Get("namespace").(string)
		}
	}

	orgName := d.Get("organization").(string)

	log.Printf("[DEBUG] Create registry module named %s", *options.Name)
	registryModule, err := tfeClient.RegistryModules.Create(ctx, orgName, options)

	if err != nil {
		return nil, fmt.Errorf("Error creating registry module %s: %w", *options.Name, err)
	}

	return registryModule, nil
}

func resourceTFERegistryModuleCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	var registryModule *tfe.RegistryModule
	var err error

	if v, ok := d.GetOk("vcs_repo"); ok {
		registryModule, err = resourceTFERegistryModuleCreateWithVCS(v, meta)
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
	d.Set("namespace", registryModule.Namespace)
	d.Set("registry_name", registryModule.RegistryName)

	return resourceTFERegistryModuleRead(d, meta)
}

func resourceTFERegistryModuleUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	options := tfe.RegistryModuleUpdateOptions{
		NoCode: tfe.Bool(d.Get("no_code").(bool)),
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
		registryModule, err = tfeClient.RegistryModules.Update(ctx, rmID, options)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error while waiting for module %s/%s to be ingested: %w", registryModule.Organization.Name, registryModule.Name, err)
	}

	d.SetId(registryModule.ID)
	d.Set("no_code", registryModule.NoCode)

	return nil
}

func resourceTFERegistryModuleRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read registry module: %s", d.Id())

	// Get the fields we need to read the registry module
	rmID := tfe.RegistryModuleID{
		Organization: d.Get("organization").(string),
		Name:         d.Get("name").(string),
		Provider:     d.Get("module_provider").(string),
		Namespace:    d.Get("namespace").(string),
		RegistryName: tfe.RegistryName(d.Get("registry_name").(string)),
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
	d.Set("namespace", registryModule.Namespace)
	d.Set("registry_name", registryModule.RegistryName)

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
		// see https://www.terraform.io/cloud-docs/api-docs/private-registry/modules#get-a-module
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
