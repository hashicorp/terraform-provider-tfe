package tfe

import (
	"fmt"
	"log"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceCurrentRun() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCurrentRunRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"workspace": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
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

									"oauth_token_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCurrentRunRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the run ID
	runID, exists := os.LookupEnv("TFE_RUN_ID")
	if !exists {
		runID, exists = os.LookupEnv("TFC_RUN_ID")
		if !exists {
			log.Printf("[DEBUG] No run ID is set")
			return nil
		}
	}

	run, err := tfeClient.Runs.Read(ctx, runID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find run ID: %s", runID)
		}
		return fmt.Errorf("Error retrieving run: %v", err)
	}

	ws, err := tfeClient.Workspaces.ReadByID(ctx, run.Workspace.ID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not find workspace ID: %s", run.Workspace.ID)
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	var workspace []interface{}
	workspaceConfig := map[string]interface{}{
		"id":   ws.ID,
		"name": ws.Name,
	}
	var vcsRepo []interface{}
	if ws.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier":     ws.VCSRepo.Identifier,
			"oauth_token_id": ws.VCSRepo.OAuthTokenID,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
		workspaceConfig["vcs_repo"] = vcsRepo
	}
	workspace = append(workspace, workspaceConfig)

	d.Set("workspace", workspace)
	d.SetId(run.ID)

	return nil
}
