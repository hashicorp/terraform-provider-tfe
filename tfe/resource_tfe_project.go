package tfe

import (
	"context"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tfeClient := meta.(*tfe.Client)

	organizationName := d.Get("organization").(string)
	name := d.Get("name").(string)

	options := tfe.ProjectCreateOptions{
		Name: name,
	}

	log.Printf("[DEBUG] Create new project: %s", name)
	project, err := tfeClient.Projects.Create(ctx, organizationName, options)
	if err != nil {
		return diag.Errorf("Error creating the new project %s: %v", name, err)
	}

	d.SetId(project.ID)

	return resourceTFEProjectUpdate(ctx, d, meta)
}

func resourceTFEProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of project: %s", d.Id())
	project, err := tfeClient.Projects.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Project %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", project.Name)
	d.Set("organization", project.Organization.Name)

	return nil
}

func resourceTFEProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tfeClient := meta.(*tfe.Client)

	options := tfe.ProjectUpdateOptions{
		Name: tfe.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Update configuration of project: %s", d.Id())
	project, err := tfeClient.Projects.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating project %s: %v", d.Id(), err)
	}

	d.SetId(project.ID)

	return resourceTFEProjectRead(ctx, d, meta)
}

func resourceTFEProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete project: %s", d.Id())
	err := tfeClient.Projects.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return diag.Errorf("Error deleting project %s: %v", d.Id(), err)
	}

	return nil
}
