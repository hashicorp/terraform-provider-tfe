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
	"log"

	"regexp"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTFEProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTFEProjectCreate,
		ReadContext:   resourceTFEProjectRead,
		UpdateContext: resourceTFEProjectUpdate,
		DeleteContext: resourceTFEProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 40),
					validation.StringMatch(regexp.MustCompile(`\A[\w\-][\w\- ]+[\w\-]\z`),
						"can only include letters, numbers, spaces, -, and _."),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"auto_destroy_activity_duration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{1,4}[dh]$`), "must be 1-4 digits followed by d or h"),
			},
		},
	}
}

func resourceTFEProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)

	options := tfe.ProjectCreateOptions{
		Name:        name,
		Description: tfe.String(d.Get("description").(string)),
	}

	if v, ok := d.GetOk("auto_destroy_activity_duration"); ok {
		options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(v.(string))
	}

	log.Printf("[DEBUG] Create new project: %s", name)
	project, err := config.Client.Projects.Create(ctx, organization, options)
	if err != nil {
		return diag.Errorf("Error creating the new project %s: %v", name, err)
	}

	d.SetId(project.ID)

	return resourceTFEProjectUpdate(ctx, d, meta)
}

func resourceTFEProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read configuration of project: %s", d.Id())
	project, err := config.Client.Projects.Read(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			log.Printf("[DEBUG] Project %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", project.Name)
	d.Set("description", project.Description)
	d.Set("organization", project.Organization.Name)

	if project.AutoDestroyActivityDuration.IsSpecified() {
		v, err := project.AutoDestroyActivityDuration.Get()
		if err != nil {
			return diag.Errorf("Error reading auto destroy activity duration: %v", err)
		}

		d.Set("auto_destroy_activity_duration", v)
		workspaces, err := project.Workspaces.Get()

	}

	return nil
}

func resourceTFEProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	options := tfe.ProjectUpdateOptions{
		Name:        tfe.String(d.Get("name").(string)),
		Description: tfe.String(d.Get("description").(string)),
	}

	if d.HasChange("auto_destroy_activity_duration") {
		duration, ok := d.GetOk("auto_destroy_activity_duration")
		if !ok {
			options.AutoDestroyActivityDuration = jsonapi.NewNullNullableAttr[string]()
		} else {
			options.AutoDestroyActivityDuration = jsonapi.NewNullableAttrWithValue(duration.(string))
		}
	}

	log.Printf("[DEBUG] Update configuration of project: %s", d.Id())
	project, err := config.Client.Projects.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating project %s: %v", d.Id(), err)
	}

	d.SetId(project.ID)

	return resourceTFEProjectRead(ctx, d, meta)
}

func resourceTFEProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete project: %s", d.Id())
	err := config.Client.Projects.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting project %s: %v", d.Id(), err)
	}

	return nil
}
