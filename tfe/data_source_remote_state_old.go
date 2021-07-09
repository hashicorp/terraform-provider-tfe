package tfe

import (
	"encoding/json"
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFERemoteState() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFERemoteStateRead,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},

			"download_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state_output": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type stateFile struct {
	Outputs map[string]outputValue `json:"outputs"`
}

//type outputValue struct {
//	Type  string      `json:"type"`
//	Value interface{} `json:"value"`
//}

func dataSourceTFERemoteStateRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	wsName := d.Get("workspace").(string)
	orgName := d.Get("organization").(string)
	log.Printf("[DEBUG] Read remote state for and Workspace %s", wsName)

	ws, err := tfeClient.Workspaces.Read(ctx, orgName, wsName)
	if err != nil {
		return fmt.Errorf("Error reading workspace: %v", err)
	}

	sv, err := tfeClient.StateVersions.Current(ctx, ws.ID)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return fmt.Errorf("Could not read  remote state for workspace '%s'", wsName)
		}
		return fmt.Errorf("Error remote state: %v", err)
	}

	log.Printf("[DEBUG] Setting Remote State Output")
	d.SetId(sv.ID)

	d.Set("download_url", sv.DownloadURL)
	stateData, err := tfeClient.StateVersions.Download(ctx, sv.DownloadURL)
	if err != nil {
		return fmt.Errorf("Error downloading remote state: %v", err)
	}
	stateOuptput := &stateFile{}
	if err := json.Unmarshal(stateData, stateOuptput); err != nil {
		return err
	}
	log.Printf("[DEBUG] STATE OUTPUT: %v", stateOuptput)

	for k, v := range stateOuptput.Outputs {
		log.Printf("[DEBUG] STATE KEY: %s", k)
		log.Printf("[DEBUG] STATE VALUE: %s", v.Value)
	}
	d.Set("state_output", "foo")

	return nil
}
