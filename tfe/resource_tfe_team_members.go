package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFETeamMembers() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamMembersCreate,
		Read:   resourceTFETeamMembersRead,
		Update: resourceTFETeamMembersUpdate,
		Delete: resourceTFETeamMembersDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamMembersImporter,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"usernames": {
				Type:         schema.TypeSet,
				Optional:     true,
				ExactlyOneOf: []string{"usernames", "organization_membership_ids"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},

			"organization_membership_ids": {
				Type:         schema.TypeSet,
				Optional:     true,
				ExactlyOneOf: []string{"organization_membership_ids", "usernames"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTFETeamMembersCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the team ID.
	teamID := d.Get("team_id").(string)

	// Create a new options struct.
	options := tfe.TeamMemberAddOptions{}

	// Add all the users that need to be added.
	for _, username := range d.Get("usernames").(*schema.Set).List() {
		options.Usernames = append(options.Usernames, username.(string))
	}

	// Add all the organization membership IDs that need to be added.
	for _, membership := range d.Get("organization_membership_ids").(*schema.Set).List() {
		options.OrganizationMembershipIDs = append(options.OrganizationMembershipIDs, membership.(string))
	}

	log.Printf("[DEBUG] Add users to team: %s", teamID)
	err := tfeClient.TeamMembers.Add(ctx, teamID, options)
	if err != nil {
		return fmt.Errorf("Error adding users to team %s: %w", teamID, err)
	}

	d.SetId(teamID)

	return nil
}

func resourceTFETeamMembersRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read users from team: %s", d.Id())
	users, err := tfeClient.TeamMembers.List(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Users do no longer exist")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading users from team %s: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Read organization memberships from team: %s", d.Id())
	organizationMemberships, err := tfeClient.TeamMembers.ListOrganizationMemberships(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Organization memberships do no longer exist")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading organization memberships from team %s: %w", d.Id(), err)
	}

	var usernames []interface{}
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}

	var organizationMembershipIDs []interface{}
	for _, membership := range organizationMemberships {
		organizationMembershipIDs = append(organizationMembershipIDs, membership.ID)
	}

	if len(usernames) > 0 {
		d.Set("usernames", usernames)
	} else {
		log.Printf("[DEBUG] Users do no longer exist")
		d.SetId("")
	}

	if len(organizationMembershipIDs) > 0 {
		d.Set("organization_membership_ids", organizationMembershipIDs)
	} else {
		log.Printf("[DEBUG] Organization memberships do no longer exist")
		d.SetId("")
	}

	return nil
}

func resourceTFETeamMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	var oldUsers, newUsers, oldMemberships, newMemberships *schema.Set
	if d.HasChange("usernames") {
		oldUsernames, newUsernames := d.GetChange("usernames")
		oldUsers = oldUsernames.(*schema.Set).Difference(newUsernames.(*schema.Set))
		newUsers = newUsernames.(*schema.Set).Difference(oldUsernames.(*schema.Set))
	}

	if d.HasChange("organization_membership_ids") {
		oldMembershipIDs, newMembershipIDs := d.GetChange("organization_membership_ids")
		oldMemberships = oldMembershipIDs.(*schema.Set).Difference(newMembershipIDs.(*schema.Set))
		newMemberships = newMembershipIDs.(*schema.Set).Difference(oldMembershipIDs.(*schema.Set))
	}

	// First add the new users and new memberships.
	if newUsers.Len() > 0 || newMemberships.Len() > 0 {
		// Create a new options struct.
		options := tfe.TeamMemberAddOptions{}

		// Add all the users that need to be added.
		for _, username := range newUsers.List() {
			options.Usernames = append(options.Usernames, username.(string))
		}

		// Add all the memberships that need to be added.
		for _, membership := range newMemberships.List() {
			options.OrganizationMembershipIDs = append(options.OrganizationMembershipIDs, membership.(string))
		}

		log.Printf("[DEBUG] Add users to team: %s", d.Id())
		err := tfeClient.TeamMembers.Add(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf("Error adding users to team %s: %w", d.Id(), err)
		}
	}

	// Then delete all the old users.
	if oldUsers.Len() > 0 || oldMemberships.Len() > 0 {
		// Create a new options struct.
		options := tfe.TeamMemberRemoveOptions{}

		// Add all the users that need to be added.
		for _, username := range oldUsers.List() {
			options.Usernames = append(options.Usernames, username.(string))
		}

		// Add all the memberships that need to be added.
		for _, membership := range newMemberships.List() {
			options.OrganizationMembershipIDs = append(options.OrganizationMembershipIDs, membership.(string))
		}

		log.Printf("[DEBUG] Remove user from team: %s", d.Id())
		err := tfeClient.TeamMembers.Remove(ctx, d.Id(), options)
		if err != nil {
			return fmt.Errorf("Error removing users to team %s: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTFETeamMembersDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Retrieve users to remove from team: %s", d.Id())
	users, err := tfeClient.TeamMembers.List(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error retrieving users to remove from team %s: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Retrieve organization memberships to remove from team: %s", d.Id())
	organizationMemberships, err := tfeClient.TeamMembers.ListOrganizationMemberships(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error retrieving organization memberships from team %s: %w", d.Id(), err)
	}

	// Create a new options struct.
	options := tfe.TeamMemberRemoveOptions{}

	// Add all the users that need to be removed.
	for _, user := range users {
		options.Usernames = append(options.Usernames, user.Username)
	}

	// Add all memberships that need to be removed.
	for _, membership := range organizationMemberships {
		options.OrganizationMembershipIDs = append(options.OrganizationMembershipIDs, membership.ID)
	}

	log.Printf("[DEBUG] Remove users from team: %s", d.Id())
	err = tfeClient.TeamMembers.Remove(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error removing users to team %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFETeamMembersImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Set the team ID field.
	d.Set("team_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
