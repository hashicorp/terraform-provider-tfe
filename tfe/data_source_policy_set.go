package tfe

import (
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEPolicySetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"global": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(tfe.Sentinel),
			},

			"overridable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"policies_path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"policy_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"vcs_repo": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"branch": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ingress_submodules": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"oauth_token_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"workspace_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceTFEPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	listOptions := tfe.PolicySetListOptions{}

	for {
		policySetList, err := tfeClient.PolicySets.List(ctx, organization, &listOptions)

		if err != nil {
			if errors.Is(err, tfe.ErrResourceNotFound) {
				return fmt.Errorf("could not find policy set %s/%s", organization, name)
			}
			return fmt.Errorf("Error retrieving policy set %s: %w", name, err)
		}

		for _, policySet := range policySetList.Items {
			// nolint: nestif
			if policySet.Name == name {
				d.Set("name", policySet.Name)
				d.Set("description", policySet.Description)
				d.Set("global", policySet.Global)
				d.Set("policies_path", policySet.PoliciesPath)

				if policySet.Organization != nil {
					d.Set("organization", policySet.Organization.Name)
				}

				if policySet.Kind != "" {
					d.Set("kind", policySet.Kind)
				}

				if policySet.Overridable != nil {
					d.Set("overridable", policySet.Overridable)
				}

				var vcsRepo []interface{}
				if policySet.VCSRepo != nil {
					vcsRepo = append(vcsRepo, map[string]interface{}{
						"identifier":         policySet.VCSRepo.Identifier,
						"branch":             policySet.VCSRepo.Branch,
						"ingress_submodules": policySet.VCSRepo.IngressSubmodules,
						"oauth_token_id":     policySet.VCSRepo.OAuthTokenID,
					})
				}
				d.Set("vcs_repo", vcsRepo)

				var policyIDs []interface{}
				for _, policy := range policySet.Policies {
					policyIDs = append(policyIDs, policy.ID)
				}
				d.Set("policy_ids", policyIDs)

				var workspaceIDs []interface{}
				if !policySet.Global {
					for _, workspace := range policySet.Workspaces {
						workspaceIDs = append(workspaceIDs, workspace.ID)
					}
				}
				d.Set("workspace_ids", workspaceIDs)

				d.SetId(policySet.ID)

				return nil
			}
		}
		// Exit the loop when we've seen all pages.
		if policySetList.CurrentPage >= policySetList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		listOptions.PageNumber = policySetList.NextPage
	}
	return fmt.Errorf("Could not find policy set %s/%s", organization, name)
}
