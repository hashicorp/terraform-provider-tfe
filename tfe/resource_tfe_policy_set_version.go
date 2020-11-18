package tfe

import (
	"fmt"
	"log"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTFEPolicySetVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetVersionCreate,
		Read:   resourceTFEPolicySetVersionRead,
		Delete: resourceTFEPolicySetVersionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFEPolicySetVersionImporter,
		},

		Schema: map[string]*schema.Schema{
			"policy_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"directory": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEPolicySetVersionCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	ps := d.Get("policy_set_id").(string)
	policySet, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %v", ps, err)
	}

	// Create a new options struct.
	options := tfe.PolicySetVersionCreateOptions{}

	// Create the policy set version
	log.Printf("[DEBUG] Create policy set version for policy set: %s", ps)
	psv, err := tfeClient.PolicySetVersions.Create(ctx, policySet.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating policy set version for policy set: %s", policySet.ID)
	}

	d.SetId(psv.Data.ID)

	// Upload policy set version files
	log.Printf("[DEBUG] Upload policy set version files for policy set version: %s", d.Id())
	uploadLink := psv.Data.Links.Upload
	version := d.Get("version").(string)
	directory := d.Get("directory").(string)
	err = tfeClient.PolicySetVersions.Upload(ctx, uploadLink, directory)
	if err != nil {
		return fmt.Errorf("Error uploading version %s from %s against upload link %s for policy set version: %s", version, directory, uploadLink, d.Id())
	}

	// Read the resource until status is ready (up to 1 minute)
	for i := 1; i < 12; i++ {
		log.Printf("[DEBUG] Reading policy set version %s after upload to get ready status", d.Id())
		psv, err = tfeClient.PolicySetVersions.Read(ctx, d.Id())
		if psv.Data.Attributes.Status == "ready" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return resourceTFEPolicySetVersionRead(d, meta)
}

func resourceTFEPolicySetVersionRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Read policy set. This is not needed to read the policy set version
	// but if the policy set was deleted, the policy set version will no longer
	// be available, and the debug statements would be useful to indicate that.
	ps := d.Get("policy_set_id").(string)
	log.Printf("[DEBUG] Read policy set: %s", ps)
	_, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] policy set %s no longer exists", ps)
		} else {
			log.Printf("[DEBUG] Error retrieving policy set %s: %v", ps, err)
		}
	}

	// Read policy set version
	log.Printf("[DEBUG] Read policy set version: %s", d.Id())
	psv, err := tfeClient.PolicySetVersions.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] policy set version %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading policy set version %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("status", psv.Data.Attributes.Status)
	d.Set("created_at", psv.Data.Attributes.CreatedAt.String())
	d.Set("updated_at", psv.Data.Attributes.UpdatedAt.String())

	return nil
}

func resourceTFEPolicySetVersionDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] policy set versions cannot be deleted")

	return nil
}

func resourceTFEPolicySetVersionImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid parameter import format: %s (expected <POLICY SET ID>/<POLICY SET VERSION ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("policy_set_id", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
