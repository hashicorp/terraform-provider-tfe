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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"global": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"workspace_ids", "workspace_external_ids"},
			},

			"policy_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"workspace_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"global", "workspace_external_ids"},
			},

			"workspace_external_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"global", "workspace_ids"},
				Deprecated:    "please use the workspace_ids attribute",
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

	// Process all configured options.
	if desc, ok := d.GetOk("description"); ok {
		options.Description = tfe.String(desc.(string))
	}

	// Set up the policies.
	for _, policyID := range d.Get("policy_ids").(*schema.Set).List() {
		options.Policies = append(options.Policies, &tfe.Policy{ID: policyID.(string)})
	}

	// Set up the workspaces.
	if d.Get("workspace_ids").(*schema.Set).Len() > 0 {
		workspaces, err := getWorkspaces(d, meta, d.Get("workspace_ids").(*schema.Set).List())
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %v", err)
		}
		if len(workspaces) > 0 {
			options.Workspaces = workspaces
		}
	}

	// Add any workspaces configured using the deprecated workspace_external_ids.
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
	if _, ok := d.GetOk("workspace_ids"); ok {
		var workspaceIDs []interface{}
		if !policySet.Global {
			workspaceIDs, err = getWorkspaceIDs(d, meta, policySet.Workspaces)
			if err != nil {
				return fmt.Errorf("Error retrieving workspace names: %v", err)
			}
		}
		d.Set("workspace_ids", workspaceIDs)
	}

	// Update the workspaces.
	if _, ok := d.GetOk("workspace_external_ids"); ok {
		var workspaceExternalIDs []interface{}
		if !policySet.Global {
			for _, workspace := range policySet.Workspaces {
				workspaceExternalIDs = append(workspaceExternalIDs, workspace.ID)
			}
		}
		d.Set("workspace_external_ids", workspaceExternalIDs)
	}

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
		options := tfe.PolicySetRemoveWorkspacesOptions{}

		// The new set of workspace IDs will be an empty set, so we don't need it
		oldWorkspaceIDs, _ := d.GetChange("workspace_ids")
		if oldWorkspaceIDs.(*schema.Set).Len() > 0 {
			workspaces, err := getWorkspaces(d, meta, oldWorkspaceIDs.(*schema.Set).List())
			if err != nil {
				return fmt.Errorf("Error retrieving workspaces: %v", err)
			}
			options.Workspaces = workspaces
		}

		oldExternalIDs, _ := d.GetChange("workspace_external_ids")
		for _, externalID := range oldExternalIDs.(*schema.Set).List() {
			options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: externalID.(string)})
		}

		if len(options.Workspaces) > 0 {
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

	if !global && d.HasChange("workspace_ids") {
		oldSet, newSet := d.GetChange("workspace_ids")
		oldWorkspaceIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newWorkspaceIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		// First add the new workspaces.
		if newWorkspaceIDs.Len() > 0 {
			workspaces, err := getWorkspaces(d, meta, newWorkspaceIDs.List())
			if err != nil {
				return fmt.Errorf("Error retrieving workspaces: %v", err)
			}

			if len(workspaces) > 0 {
				options := tfe.PolicySetAddWorkspacesOptions{
					Workspaces: workspaces,
				}

				log.Printf("[DEBUG] Attach policy set to workspaces: %s", d.Id())
				err = tfeClient.PolicySets.AddWorkspaces(ctx, d.Id(), options)
				if err != nil {
					return fmt.Errorf("Error attaching policy set %s to workspaces: %v", d.Id(), err)
				}
			}
		}

		// Then remove all the old workspaces.
		if oldWorkspaceIDs.Len() > 0 {
			workspaces, err := getWorkspaces(d, meta, oldWorkspaceIDs.List())
			if err != nil {
				return fmt.Errorf("Error retrieving workspaces: %v", err)
			}

			if len(workspaces) > 0 {
				options := tfe.PolicySetRemoveWorkspacesOptions{
					Workspaces: workspaces,
				}

				log.Printf("[DEBUG] Detach policy set from workspaces: %s", d.Id())
				err = tfeClient.PolicySets.RemoveWorkspaces(ctx, d.Id(), options)
				if err != nil {
					return fmt.Errorf("Error detaching policy set %s from workspaces: %v", d.Id(), err)
				}
			}
		}
	}

	if !global && d.HasChange("workspace_external_ids") {
		oldSet, newSet := d.GetChange("workspace_external_ids")
		oldExternalIDs := oldSet.(*schema.Set).Difference(newSet.(*schema.Set))
		newExternalIDs := newSet.(*schema.Set).Difference(oldSet.(*schema.Set))

		// First add the new workspaces.
		if newExternalIDs.Len() > 0 {
			options := tfe.PolicySetAddWorkspacesOptions{}

			for _, externalID := range newExternalIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: externalID.(string)})
			}

			log.Printf("[DEBUG] Attach policy set to workspaces: %s", d.Id())
			err := tfeClient.PolicySets.AddWorkspaces(ctx, d.Id(), options)
			if err != nil {
				return fmt.Errorf("Error attaching policy set %s to workspaces: %v", d.Id(), err)
			}
		}

		// Then remove all the old workspaces.
		if oldExternalIDs.Len() > 0 {
			options := tfe.PolicySetRemoveWorkspacesOptions{}

			for _, externalID := range oldExternalIDs.List() {
				options.Workspaces = append(options.Workspaces, &tfe.Workspace{ID: externalID.(string)})
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

func getWorkspaces(d *schema.ResourceData, meta interface{}, ids []interface{}) ([]*tfe.Workspace, error) {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	organization := d.Get("organization").(string)

	// Create a map with workspace names.
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		org, name, err := unpackWorkspaceID(id.(string))
		if err != nil {
			return nil, err
		}
		if org != organization {
			return nil, fmt.Errorf("workspace %s does not belong to organization %s", id, organization)
		}
		m[name] = true
	}

	// Create a slice for any matching workspaces.
	var workspaces []*tfe.Workspace

	options := tfe.WorkspaceListOptions{}
	for {
		wl, err := tfeClient.Workspaces.List(ctx, organization, options)
		if err != nil {
			return nil, err
		}

		for _, w := range wl.Items {
			if m[w.Name] {
				workspaces = append(workspaces, w)
				if len(workspaces) == len(ids) {
					return workspaces, nil
				}
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	return workspaces, nil
}

func getWorkspaceIDs(d *schema.ResourceData, meta interface{}, workspaces []*tfe.Workspace) ([]interface{}, error) {
	tfeClient := meta.(*tfe.Client)

	// Get the organization name.
	organization := d.Get("organization").(string)

	// Create a map with workspace IDs.
	m := make(map[string]bool, len(workspaces))
	for _, w := range workspaces {
		m[w.ID] = true
	}

	// Create a slice for any matching workspaces.
	var ids []interface{}

	options := tfe.WorkspaceListOptions{}
	for {
		wl, err := tfeClient.Workspaces.List(ctx, organization, options)
		if err != nil {
			return nil, err
		}

		for _, w := range wl.Items {
			if m[w.ID] {
				ids = append(ids, organization+"/"+w.Name)
				if len(ids) == len(workspaces) {
					return ids, nil
				}
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	return ids, nil
}
