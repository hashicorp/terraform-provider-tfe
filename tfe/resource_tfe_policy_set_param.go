package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceTFEPolicySetParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetParameterCreate,
		Read:   resourceTFEPolicySetParameterRead,
		Update: resourceTFEPolicySetParameterUpdate,
		Delete: resourceTFEPolicySetParameterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFEPolicySetParameterImporter,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"value": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"category": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.CategoryPolicySet),
					},
					false,
				),
			},

			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"policy_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFEPolicySetParameterCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get key and category.
	key := d.Get("key").(string)
	category := d.Get("category").(string)

	ps := d.Get("policy_set_id").(string)
	policySet, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %v", ps, err)
	}

	// Create a new options struct.
	options := tfe.PolicySetParameterCreateOptions{
		Key:       tfe.String(key),
		Value:     tfe.String(d.Get("value").(string)),
		Category:  tfe.Category(tfe.CategoryType(category)),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
	}

	log.Printf("[DEBUG] Create %s variable: %s", category, key)
	variable, err := tfeClient.PolicySetParameters.Create(ctx, policySet.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating %s variable %s: %v", category, key, err)
	}

	d.SetId(variable.ID)

	return resourceTFEPolicySetParameterRead(d, meta)
}

func resourceTFEPolicySetParameterRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	ps := d.Get("policy_set_id").(string)
	policySet, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %v", ps, err)
	}

	log.Printf("[DEBUG] Read variable: %s", d.Id())
	variable, err := tfeClient.PolicySetParameters.Read(ctx, policySet.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] PolicySetParameter %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading variable %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("key", variable.Key)
	d.Set("category", string(variable.Category))
	d.Set("sensitive", variable.Sensitive)

	// Only set the value if its not sensitive, as otherwise it will be empty.
	if !variable.Sensitive {
		d.Set("value", variable.Value)
	}

	return nil
}

func resourceTFEPolicySetParameterUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	ps := d.Get("policy_set_id").(string)
	policySet, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %v", ps, err)
	}

	// Create a new options struct.
	options := tfe.PolicySetParameterUpdateOptions{
		Key:       tfe.String(d.Get("key").(string)),
		Value:     tfe.String(d.Get("value").(string)),
		Sensitive: tfe.Bool(d.Get("sensitive").(bool)),
	}

	log.Printf("[DEBUG] Update variable: %s", d.Id())
	_, err = tfeClient.PolicySetParameters.Update(ctx, policySet.ID, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating variable %s: %v", d.Id(), err)
	}

	return resourceTFEPolicySetParameterRead(d, meta)
}

func resourceTFEPolicySetParameterDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	ps := d.Get("policy_set_id").(string)
	policySet, err := tfeClient.PolicySets.Read(ctx, ps)
	if err != nil {
		return fmt.Errorf("Error retrieving policy set %s: %v", ps, err)
	}

	log.Printf("[DEBUG] Delete variable: %s", d.Id())
	err = tfeClient.PolicySetParameters.Delete(ctx, policySet.ID, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting variable%s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEPolicySetParameterImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid variable import format: %s (expected <POLICY SET>/<VARIABLE ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("policy_set_id", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
