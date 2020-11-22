package tfe

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func getOrgSweeper(name string) *resource.Sweeper {
	return &resource.Sweeper{
		Name: name,
		F: func(ununsed string) error {
			if os.Getenv("TFE_HOSTNAME") == "" && (os.Getenv("TFE_TOKEN") == "") {
				return fmt.Errorf("must provide environment variables TFE_HOSTNAME and TFE_TOKEN")
			}
			hostname := os.Getenv("TFE_HOSTNAME")
			token := os.Getenv("TFE_TOKEN")
			insecure := os.Getenv("TFE_INSECURE") != ""

			client, err := getClient(hostname, token, insecure)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}

			ctx := context.TODO()
			orgList, err := client.Organizations.List(ctx, tfe.OrganizationListOptions{})
			if err != nil {
				return fmt.Errorf("Error listing organizations: %s", err)
			}
			for _, org := range orgList.Items {
				log.Printf("[DEBUG] Testing if org %s starts with tst-terraform or named terraform-updated", org.Name)
				if strings.HasPrefix(org.Name, "tst-terraform") {
					log.Printf("[DEBUG] deleting org %s", org.Name)
					err = client.Organizations.Delete(ctx, org.Name)
					if err != nil {
						return fmt.Errorf("Error deleting organization %q %s", org.Name, err)
					}
				}
			}
			return nil
		},
	}
}

// Sweepers usually go along with the tests. In TF[CE]'s case everything depends on the organization,
// which means that if we delete it then all the other entities will  be deleted automatically.
func init() {
	resource.AddTestSweepers("org_sweeper", getOrgSweeper("org_sweeper"))
}
