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

	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEPolicySet() *schema.Resource {
	return &schema.Resource{
		Description: "Gets a policy set defined in a specified organization.",

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
				Description: "The policy-as-code framework for the policy. Valid values are \"sentinel\" and \"opa\".",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"overridable": {
				Description: "Whether users can override this policy when it fails during a run. Only valid for \"opa\" policies.",
				Type:        schema.TypeBool,
				Optional:    true,
			},

			"agent_enabled": {
				Description: "Whether the policy set is executed in the HCP Terraform agent. `true` by default for all \"opa\" policy sets.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"policy_tool_version": {
				Description: "The policy tool version to run the policy evaluation against. For \"opa\" policy sets, 'latest' will not be a valid input.",
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
				Description: "IDs of the projects that use the policy set.",
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

	pageSize := int32(100)
	queryParams := &organizations.ItemPolicySetsRequestBuilderGetQueryParameters{
		Pagesize:   &pageSize,
		Searchname: &name,
	}

	for {
		policySetList, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).PolicySets().Get(ctx, withQueryParams(queryParams))
		if err != nil {
			if errors.Is(err, tfeV2.ErrNotFound) {
				return fmt.Errorf("could not find policy set %s/%s", organization, name)
			}
			return fmt.Errorf("Error retrieving policy set %s: %w", name, err)
		}

		for _, policySet := range policySetList.GetData() {
			attrs := policySet.GetAttributes()
			if attrs == nil || valueOrZero(attrs.GetName()) != name {
				continue
			}
			rels := policySet.GetRelationships()

			d.Set("name", valueOrZero(attrs.GetName()))
			d.Set("description", valueOrZero(attrs.GetDescription()))
			d.Set("global", attrs.GetGlobal() != nil && *attrs.GetGlobal())
			d.Set("policies_path", valueOrZero(attrs.GetPoliciesPath()))
			d.Set("policy_update_patterns", attrs.GetPolicyUpdatePatterns())
			d.Set("agent_enabled", attrs.GetAgentEnabled() != nil && *attrs.GetAgentEnabled())

			if attrs.GetKind() != nil {
				d.Set("kind", enumStringOrEmpty(attrs.GetKind()))
			}
			if attrs.GetOverridable() != nil {
				d.Set("overridable", *attrs.GetOverridable())
			}
			if attrs.GetPolicyToolVersion() != nil {
				d.Set("policy_tool_version", *attrs.GetPolicyToolVersion())
			}

			var vcsRepo []interface{}
			if attrs.GetVcsRepo() != nil {
				vr := attrs.GetVcsRepo()
				vcsRepo = append(vcsRepo, map[string]interface{}{
					"identifier":                 valueOrZero(vr.GetIdentifier()),
					"branch":                     valueOrZero(vr.GetBranch()),
					"ingress_submodules":         vr.GetIngressSubmodules() != nil && *vr.GetIngressSubmodules(),
					"oauth_token_id":             valueOrZero(vr.GetOauthTokenId()),
					"github_app_installation_id": valueOrZero(vr.GetGithubAppInstallationId()),
				})
			}
			d.Set("vcs_repo", vcsRepo)

			if rels != nil {
				var policyIDs []interface{}
				if rels.GetPolicies() != nil {
					for _, p := range rels.GetPolicies().GetData() {
						policyIDs = append(policyIDs, valueOrZero(p.GetId()))
					}
				}
				d.Set("policy_ids", policyIDs)

				isGlobal := attrs.GetGlobal() != nil && *attrs.GetGlobal()

				var workspaceIDs []interface{}
				if !isGlobal && rels.GetWorkspaces() != nil {
					for _, ws := range rels.GetWorkspaces().GetData() {
						workspaceIDs = append(workspaceIDs, valueOrZero(ws.GetId()))
					}
				}
				d.Set("workspace_ids", workspaceIDs)

				var excludedWorkspaceIDs []interface{}
				if rels.GetWorkspaceExclusions() != nil {
					for _, ws := range rels.GetWorkspaceExclusions().GetData() {
						excludedWorkspaceIDs = append(excludedWorkspaceIDs, valueOrZero(ws.GetId()))
					}
				}
				d.Set("excluded_workspace_ids", excludedWorkspaceIDs)

				var projectIDs []interface{}
				if !isGlobal && rels.GetProjects() != nil {
					for _, proj := range rels.GetProjects().GetData() {
						projectIDs = append(projectIDs, valueOrZero(proj.GetId()))
					}
				}
				d.Set("project_ids", projectIDs)
			}

			d.SetId(valueOrZero(policySet.GetId()))
			return nil
		}

		nextPage := nextPageFromMeta(policySetList.GetMeta())
		if nextPage == nil {
			break
		}
		queryParams.Pagenumber = nextPage
	}
	return fmt.Errorf("could not find policy set %s/%s", organization, name)
}
