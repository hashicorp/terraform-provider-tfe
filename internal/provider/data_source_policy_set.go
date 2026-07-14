// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a policy set defined in a specified organization.",

		Read: dataSourceTFEPolicySetRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the policy set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "Name of the policy set.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"description": {
				Description: "Description of the policy set.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"global": {
				Description: "Whether or not the policy set applies to all workspaces in the organization.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"kind": {
				Description: "The policy-as-code framework for the policy. Valid values are `sentinel` and `opa`.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"overridable": {
				Description: "Whether users can override this policy when it fails during a run. Only valid for OPA policies.",
				Type:        schema.TypeBool,
				Optional:    true,
			},

			"agent_enabled": {
				Description: "Whether the policy set is executed in the HCP Terraform agent. true by default for OPA policies.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"policy_tool_version": {
				Description: "The policy tool version to run the policy evaluation against. For `opa` policy sets, `latest` will not be a valid input.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"policies_path": {
				Description: "The sub-path within the attached VCS repository when using `vcs_repo`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"policy_update_patterns": {
				Description: "Glob patterns specifying which file changes trigger policy set updates. Patterns are relative to the repository root, and a maximum of 100 patterns can be returned. This attribute is only valid when the policy set specifies a VCS repository.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},

			"policy_ids": {
				Description: "IDs of the policies attached to the policy set.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},

			"vcs_repo": {
				Description: "Settings for the workspace's VCS repository.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"branch": {
							Description: "The repository branch that Terraform will execute from.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"ingress_submodules": {
							Description: "Indicates whether submodules should be fetched when cloning the VCS repository.",
							Type:        schema.TypeBool,
							Computed:    true,
						},

						"oauth_token_id": {
							Description: "OAuth token ID of the configured VCS connection.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"github_app_installation_id": {
							Description: "The installation ID of the GitHub App.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},

			"workspace_ids": {
				Description: "IDs of the workspaces that use the policy set.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},

			"excluded_workspace_ids": {
				Description: "IDs of the workspaces that do not use the policy set.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},

			"project_ids": {
				Description: "IDs of the projects attached to the policy set.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataSourceTFEPolicySetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	listOptions := tfe.PolicySetListOptions{}

	for {
		policySetList, err := config.Client.PolicySets.List(ctx, organization, &listOptions)

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
				d.Set("policy_update_patterns", policySet.PolicyUpdatePatterns)
				d.Set("agent_enabled", policySet.AgentEnabled)

				if policySet.Kind != "" {
					d.Set("kind", policySet.Kind)
				}

				if policySet.Overridable != nil {
					d.Set("overridable", policySet.Overridable)
				}

				if policySet.PolicyToolVersion != "" {
					d.Set("policy_tool_version", policySet.PolicyToolVersion)
				}

				var vcsRepo []interface{}
				if policySet.VCSRepo != nil {
					vcsRepo = append(vcsRepo, map[string]interface{}{
						"identifier":                 policySet.VCSRepo.Identifier,
						"branch":                     policySet.VCSRepo.Branch,
						"ingress_submodules":         policySet.VCSRepo.IngressSubmodules,
						"oauth_token_id":             policySet.VCSRepo.OAuthTokenID,
						"github_app_installation_id": policySet.VCSRepo.GHAInstallationID,
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

				var excludedWorkspaceIDs []interface{}
				for _, excludedWorkspace := range policySet.WorkspaceExclusions {
					excludedWorkspaceIDs = append(excludedWorkspaceIDs, excludedWorkspace.ID)
				}
				d.Set("excluded_workspace_ids", excludedWorkspaceIDs)

				var projectIDs []interface{}
				if !policySet.Global {
					for _, project := range policySet.Projects {
						projectIDs = append(projectIDs, project.ID)
					}
				}
				d.Set("project_ids", projectIDs)

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
	return fmt.Errorf("could not find policy set %s/%s", organization, name)
}
