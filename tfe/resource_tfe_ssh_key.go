package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFESSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFESSHKeyCreate,
		Read:   resourceTFESSHKeyRead,
		Update: resourceTFESSHKeyUpdate,
		Delete: resourceTFESSHKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceTFESSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.SSHKeyCreateOptions{
		Name:  tfe.String(name),
		Value: tfe.String(d.Get("key").(string)),
	}

	log.Printf("[DEBUG] Create new SSH key for organization: %s", organization)
	sshKey, err := tfeClient.SSHKeys.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating SSH key %s for organization %s: %v", name, organization, err)
	}

	d.SetId(sshKey.ID)

	return resourceTFESSHKeyUpdate(d, meta)
}

func resourceTFESSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of SSH key: %s", d.Id())
	sshKey, err := tfeClient.SSHKeys.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] SSH key %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of SSH key %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", sshKey.Name)

	return nil
}

func resourceTFESSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Create a new options struct.
	options := tfe.SSHKeyUpdateOptions{
		Name:  tfe.String(d.Get("name").(string)),
		Value: tfe.String(d.Get("key").(string)),
	}

	log.Printf("[DEBUG] Update SSH key: %s", d.Id())
	_, err := tfeClient.SSHKeys.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating SSH key %s: %v", d.Id(), err)
	}

	return resourceTFESSHKeyRead(d, meta)
}

func resourceTFESSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete SSH key: %s", d.Id())
	err := tfeClient.SSHKeys.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting SSH key %s: %v", d.Id(), err)
	}

	return nil
}
