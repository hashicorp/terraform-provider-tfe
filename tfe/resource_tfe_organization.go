package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceTFEOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationCreate,
		Read:   resourceTFEOrganizationRead,
		Update: resourceTFEOrganizationUpdate,
		Delete: resourceTFEOrganizationDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"session_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  20160,
			},

			"session_remember": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  20160,
			},

			"collaborator_auth_policy": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(tfe.AuthPolicyPassword),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.AuthPolicyPassword),
						string(tfe.AuthPolicyTwoFactor),
					},
					false,
				),
			},
		},
	}
}

func resourceTFEOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	name := d.Get("name").(string)

	// Create a new options struct.
	options := tfe.OrganizationCreateOptions{
		Name:  tfe.String(name),
		Email: tfe.String(d.Get("email").(string)),
	}

	log.Printf("[DEBUG] Create new organization: %s", name)
	org, err := tfeClient.Organizations.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating the new organization %s: %v", name, err)
	}

	d.SetId(org.Name)

	return resourceTFEOrganizationUpdate(d, meta)
}

func resourceTFEOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of organization: %s", d.Id())
	org, err := tfeClient.Organizations.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Organization %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	// Update the config.
	d.Set("name", org.Name)
	d.Set("email", org.Email)
	d.Set("session_timeout", org.SessionTimeout)
	d.Set("session_remember", org.SessionRemember)
	d.Set("collaborator_auth_policy", string(org.CollaboratorAuthPolicy))

	return nil
}

func resourceTFEOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.OrganizationUpdateOptions{
		Name:  tfe.String(d.Get("name").(string)),
		Email: tfe.String(d.Get("email").(string)),
	}

	// If session_timeout is supplied, set it using the options struct.
	if sessionTimeout, ok := d.GetOk("session_timeout"); ok {
		options.SessionTimeout = tfe.Int(sessionTimeout.(int))
	}

	// If session_remember is supplied, set it using the options struct.
	if sessionRemember, ok := d.GetOk("session_remember"); ok {
		options.SessionRemember = tfe.Int(sessionRemember.(int))
	}

	// If collaborator_auth_policy is supplied, set it using the options struct.
	if authPolicy, ok := d.GetOk("collaborator_auth_policy"); ok {
		options.CollaboratorAuthPolicy = tfe.AuthPolicy(tfe.AuthPolicyType(authPolicy.(string)))
	}

	log.Printf("[DEBUG] Update configuration of organization: %s", d.Id())
	org, err := tfeClient.Organizations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating organization %s: %v", d.Id(), err)
	}

	d.SetId(org.Name)

	return resourceTFEOrganizationRead(d, meta)
}

func resourceTFEOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete organization: %s", d.Id())
	err := tfeClient.Organizations.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting organization %s: %v", d.Id(), err)
	}

	return nil
}
