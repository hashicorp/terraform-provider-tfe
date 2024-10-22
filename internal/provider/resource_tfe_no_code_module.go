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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFENoCodeModule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTFENoCodeModuleCreate,
		ReadContext:   resourceTFENoCodeModuleRead,
		UpdateContext: resourceTFENoCodeModuleUpdate,
		DeleteContext: resourceTFENoCodeModuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"registry_module": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version_pin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"variable_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				// The version_pin needs to be set when variable_options are set
				RequiredWith: []string{"version_pin"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"options": {
							Type:     schema.TypeList,
							ForceNew: false,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceTFENoCodeModuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	options := tfe.RegistryNoCodeModuleCreateOptions{
		RegistryModule: &tfe.RegistryModule{
			ID: d.Get("registry_module").(string),
		},
	}

	if enabled, ok := d.GetOk("enabled"); ok {
		options.Enabled = tfe.Bool(enabled.(bool))
	}

	versionPin, ok := d.GetOk("version_pin")
	if ok {
		options.VersionPin = versionPin.(string)
	}

	orgName, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// If variable_options are set, we need to ensure the pinned version is available
	// before creating the no-code module with the variable options.
	// We check for the version_pin and wait for the pinned version to be available.
	if variableOptions, ok := d.GetOk("variable_options"); ok {
		if options.VersionPin == "" {
			return diag.Errorf("version_pin must be set when variable_options are set")
		}
		moduleID, err := getFullModuleID(ctx, config.Client, orgName, options.RegistryModule.ID)
		if err != nil {
			return diag.Errorf("Error getting full module ID for registry module %s: %s", options.RegistryModule.ID, err)
		}
		if err := waitForModuleVersion(ctx, config.Client, moduleID, options.VersionPin); err != nil {
			return diag.Errorf("Error reading registry module version %s: %s", options.VersionPin, err)
		}
		options.VariableOptions = variableOptionsMaptoStruct(variableOptions.([]interface{}))
	}

	log.Printf("[DEBUG] Create no-code module for registry module %s", options.RegistryModule.ID)
	noCodeModule, err := config.Client.RegistryNoCodeModules.Create(ctx, orgName, options)

	if err != nil {
		return diag.Errorf("Error creating no-code module for registry module %s: %s", options.RegistryModule.ID, err)
	}

	d.SetId(noCodeModule.ID)
	return resourceTFENoCodeModuleRead(ctx, d, meta)
}

func variableOptionsMaptoStruct(variableOptions []interface{}) []*tfe.NoCodeVariableOption {
	var variableOptionsRes []*tfe.NoCodeVariableOption
	for _, v := range variableOptions {
		vOpt := v.(map[string]interface{})
		option := &tfe.NoCodeVariableOption{
			VariableName: vOpt["name"].(string),
			VariableType: vOpt["type"].(string),
		}
		if vOpt["options"] != nil {
			for _, o := range vOpt["options"].([]interface{}) {
				option.Options = append(option.Options, o.(string))
			}
		}
		variableOptionsRes = append(variableOptionsRes, option)
	}
	return variableOptionsRes
}

func getFullModuleID(ctx context.Context, client *tfe.Client, orgName, ID string) (tfe.RegistryModuleID, error) {
	module, err := client.RegistryModules.Read(ctx, tfe.RegistryModuleID{ID: ID})
	if err != nil {
		return tfe.RegistryModuleID{}, err
	}
	return tfe.RegistryModuleID{
		Organization: orgName,
		Namespace:    module.Namespace,
		Name:         module.Name,
		Provider:     module.Provider,
		RegistryName: module.RegistryName,
	}, nil
}

func waitForModuleVersion(ctx context.Context, client *tfe.Client, moduleID tfe.RegistryModuleID, versionPin string) error {
	timeout := time.Duration(5) * time.Minute
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err := client.RegistryModules.ReadVersion(ctx, moduleID, versionPin)
		if err == tfe.ErrResourceNotFound {
			return retry.RetryableError(fmt.Errorf("Version %s not found for module %s", versionPin, moduleID))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

func resourceTFENoCodeModuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	readOpts := &tfe.RegistryNoCodeModuleReadOptions{
		Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
	}
	noCodeModule, err := config.Client.RegistryNoCodeModules.Read(ctx, d.Id(), readOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	options := tfe.RegistryNoCodeModuleUpdateOptions{
		Enabled:        tfe.Bool(d.Get("enabled").(bool)),
		RegistryModule: &tfe.RegistryModule{ID: d.Get("registry_module").(string)},
	}

	if versionPin, ok := d.GetOk("version_pin"); ok {
		options.VersionPin = versionPin.(string)
	}
	if variableOptions, ok := d.GetOk("variable_options"); ok {
		options.VariableOptions = variableOptionsMaptoStruct(variableOptions.([]interface{}))
	}

	err = retry.RetryContext(ctx, time.Duration(5)*time.Minute, func() *retry.RetryError {
		noCodeModule, err = config.Client.RegistryNoCodeModules.Update(ctx, d.Id(), options)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Error while waiting for no-code module %s to be updated: %s", noCodeModule.ID, err)
	}

	d.SetId(noCodeModule.ID)

	return resourceTFENoCodeModuleRead(ctx, d, meta)
}

func resourceTFENoCodeModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read no-code module: %s", d.Id())
	options := &tfe.RegistryNoCodeModuleReadOptions{
		Include: []tfe.RegistryNoCodeModuleIncludeOpt{tfe.RegistryNoCodeIncludeVariableOptions},
	}

	noCodeModule, err := config.Client.RegistryNoCodeModules.Read(ctx, d.Id(), options)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] no-code module %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading no-code module %s: %s", d.Id(), err)
	}

	// Update the config
	d.Set("enabled", noCodeModule.Enabled)
	d.Set("registry_module", noCodeModule.RegistryModule.ID)
	d.Set("organization", noCodeModule.Organization.Name)
	d.Set("version_pin", noCodeModule.VersionPin)

	mp := make([]map[string]interface{}, 0)
	for _, v := range noCodeModule.VariableOptions {
		m := make(map[string]interface{})
		m["name"] = v.VariableName
		m["type"] = v.VariableType
		m["options"] = v.Options
		mp = append(mp, m)
	}
	d.Set("variable_options", mp)

	return nil
}

func resourceTFENoCodeModuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete no-code module: %s", d.Id())
	err := config.Client.RegistryNoCodeModules.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting no-code module %s: %s", d.Id(), err)
	}
	return nil
}
