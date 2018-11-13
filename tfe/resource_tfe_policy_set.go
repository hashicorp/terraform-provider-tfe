package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEPolicySetCreate,
		Read:   resourceTFEPolicySetRead,
		Update: resourceTFEPolicySetUpdate,
		Delete: resourceTFEPolicySetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"global": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_external_ids"},
			},

			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"workspace_external_ids": &schema.Schema{
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"global"},
			},
		},
	}
}

func resourceTFEPolicySetCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.PolicySetCreateOptions{
		Name:   tfe.String(name),
		Global: tfe.Bool(d.Get("global").(bool)),
	}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = tfe.String(desc.(string))
	}

	// Set up the policies.
	for _, policyID := range d.Get("policy_ids").(*schema.Set).List() {
		options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
	}

	// Set up the workspaces.
	for _, workspaceID := range d.Get("workspace_external_ids").(*schema.Set).List() {
		options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
	}

	log.Printf("[DEBUG] Create policy set %s for organization: %s", name, organization)
	policySet, err := tfeClient.PolicySets.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating policy set %s for organization %s: %v", name, organization, err)
	}

	d.SetId(policySet.ID)

	return resourceTFEPolicySetRead(d, meta)
}

func resourceTFEPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read policy set: %s", d.Id())
	policySet, err := tfeClient.PolicySets.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Policy set %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading policy set %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", policySet.Name)
	d.Set("description", policySet.Description)
	d.Set("global", policySet.Global)

	if policySet.Organization != nil {
		d.Set("organization", policySet.Organization.Name)
	}

	// Update the policies.
	var policyIDs []interface{}
	for _, policy := range policySet.Policies {
		policyIDs = append(policyIDs, policy.ID)
	}
	d.Set("policy_ids", policyIDs)

	// Update the workspaces.
	var workspaceIDs []interface{}
	if !policySet.Global {
		for _, workspace := range policySet.Workspaces {
			workspaceIDs = append(workspaceIDs, workspace.ID)
		}
	}
	d.Set("workspace_external_ids", workspaceIDs)

	return nil
}

func resourceTFEPolicySetUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	name := d.Get("name").(string)
	global := d.Get("global").(bool)

	// If a user is setting the policy set to "global", make sure the workspaces
	// that _had_ been set are explicitly removed. This helps keep the policy
	// set's state in check
	if global && d.HasChange("global") {
		// The new set of workspaces will be an empty set, so we don't need it
		oldSet, _ := d.GetChange("workspace_external_ids")
		oldWorkspaceIDs := oldSet.(*schema.Set)

		if oldWorkspaceIDs.Len() > 0 {
			options := tfe.PolicySetRemoveWorkspacesOptions{}

			for _, workspaceID := range oldWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Removing previous workspaces from now-global policy set: %s", d.Id())
			err := tfeClient.PolicySets.RemoveWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error detaching policy set %s from workspaces: %v", d.Id(), err)
			}
		}
	}

	// Don't bother updating the policy set's attributes if they haven't changed
	if d.HasChange("name") || d.HasChange("description") || d.HasChange("global") {
		// Create a new options struct.
		options := tfe.PolicySetUpdateOptions{
			Name:   tfe.String(name),
			Global: tfe.Bool(global),
		}

		if desc, ok := d.GetOk("description"); ok {
			options.Description = tfe.String(desc.(string))
		}

		log.Printf("[DEBUG] Update configuration for policy set: %s", d.Id())
		_, err := tfeClient.PolicySets.Update(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for policy set %s: %v", d.Id(), err)
		}
	}

	if d.HasChange("policy_ids") {
		oldSet, newSet := d.GetChange("policy_ids")
		oldPolicyIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newPolicyIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		// First add the new policies.
		if newPolicyIDs.Len() > 0 {
			options := tfe.PolicySetAddPoliciesOptions{}

			for _, policyID := range newPolicyIDs.List() {
				options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
			}

			log.Printf("[DEBUG] Add policies to policy set: %s", d.Id())
			err := tfeClient.PolicySets.AddPolicies(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error adding policies to policy set %s: %v", d.Id(), err)
			}
		}

		// Then remove all the old policies.
		if oldPolicyIDs.Len() > 0 {
			options := tfe.PolicySetRemovePoliciesOptions{}

			for _, policyID := range oldPolicyIDs.List() {
				options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
			}

			log.Printf("[DEBUG] Remove policies from policy set: %s", d.Id())
			err := tfeClient.PolicySets.RemovePolicies(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error removing policies from policy set %s: %v", d.Id(), err)
			}
		}
	}

	if !global && d.HasChange("workspace_external_ids") {
		oldSet, newSet := d.GetChange("workspace_external_ids")
		oldWorkspaceIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newWorkspaceIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		// First add the new workspaces.
		if newWorkspaceIDs.Len() > 0 {
			options := tfe.PolicySetAddWorkspacesOptions{}

			for _, workspaceID := range newWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Attach policy set to workspaces: %s", d.Id())
			err := tfeClient.PolicySets.AddWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error attaching policy set %s to workspaces: %v", d.Id(), err)
			}
		}

		// Then remove all the old workspaces.
		if oldWorkspaceIDs.Len() > 0 {
			options := tfe.PolicySetRemoveWorkspacesOptions{}

			for _, workspaceID := range oldWorkspaceIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: workspaceID.(string)})
			}

			log.Printf("[DEBUG] Detach policy set from workspaces: %s", d.Id())
			err := tfeClient.PolicySets.RemoveWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error detaching policy set %s from workspaces: %v", d.Id(), err)
			}
		}
	}

	return resourceTFEPolicySetRead(d, meta)
}

func resourceTFEPolicySetDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete policy set: %s", d.Id())
	err := tfeClient.PolicySets.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting policy set %s: %v", d.Id(), err)
	}

	return nil
}
