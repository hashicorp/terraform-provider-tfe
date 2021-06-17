package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEPolicySetVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetVersionCreate,
		Read:   resourceTFEPolicySetVersionRead,
		Delete: resourceTFEPolicySetVersionDelete,

		Schema: map[string]*schema.Schema{
			"policy_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policies_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policies_path_contents_checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"error_message": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceTFEPolicySetVersionRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] =====OMAR DEBUG READ==========")
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read policy set version: %s", d.Id())
	policySetVersion, err := tfeClient.PolicySetVersions.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Policy set version %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading policy set version %s: %v", d.Id(), err)
	}

	policiesPath := d.Get("policies_path").(string)
	currentHash := d.Get("policies_path_contents_checksum").(string)
	newHash, err := hashPolicies(policiesPath)
	if err != nil {
		return fmt.Errorf("Error hashing the policies contents %v", err)
	}
	if currentHash != newHash {
		d.Set("policies_path_contents_checksum", newHash)
	}

	d.Set("status", policySetVersion.Status)
	d.Set("error_message", policySetVersion.ErrorMessage)

	return nil
}

func resourceTFEPolicySetVersionCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] =====OMAR DEBUG CREATE==========")
	tfeClient := meta.(*tfe.Client)

	policySetID := d.Get("policy_set_id").(string)
	policiesPath := d.Get("policies_path").(string)

	psv, err := tfeClient.PolicySetVersions.Create(ctx, policySetID)
	if err != nil {
		return fmt.Errorf("Error creating policy set version for policy set %s: %s", policySetID, err.Error())
	}

	err = tfeClient.PolicySetVersions.Upload(ctx, *psv, policiesPath)
	if err != nil {
		return fmt.Errorf("Error uploading policies for policy set version %s: %s", psv.ID, err.Error())
	}

	d.SetId(psv.ID)

	return resourceTFEPolicySetVersionRead(d, meta)
}

func resourceTFEPolicySetVersionDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO: explain why delete must be here. ForceNew?
	return nil
}
