package tfe

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTFEAgentPoolDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	// HACK: Agent pools do not have a full resource lifecycle at the
	// moment, so a matching resource does not exist.  This makes testing
	// with the provider SDK...difficult. Here we hackily init the provider
	// with a dummy resource config just to get access to the API client to
	// then set up an organization and agent pool out-of-band to test...for now.
	testAccProvider.Configure(&terraform.ResourceConfig{})
	tfeClient := testAccProvider.Meta().(*tfe.Client)
	testAccTFEAgentPoolDataSourceSetup(tfeClient, orgName)
	defer testAccTFEAgentPoolDataSourceCleanup(tfeClient, orgName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPoolDataSourceConfig(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "name", "Default"),
					resource.TestCheckResourceAttr(
						"data.tfe_agent_pool.foobar", "organization", orgName),
					resource.TestCheckResourceAttrSet("data.tfe_agent_pool.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFEAgentPoolDataSourceSetup(tfeClient *tfe.Client, orgName string) {
	org, err := tfeClient.Organizations.Create(
		context.Background(),
		tfe.OrganizationCreateOptions{Name: tfe.String(orgName), Email: tfe.String("admin@company.com")},
	)
	if err != nil {
		log.Fatalf("Failed to create organization out of band: %v", err)
	}

	_, err = tfeClient.AgentPools.Create(context.Background(), org.Name, tfe.AgentPoolCreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create agent pool out of band: %v", err)
	}
}

func testAccTFEAgentPoolDataSourceCleanup(tfeClient *tfe.Client, orgName string) {
	err := tfeClient.Organizations.Delete(context.Background(), orgName)
	if err != nil {
		log.Fatalf("WARNING: Failed to clean up organization out of band: %v", err)
	}
}

func testAccTFEAgentPoolDataSourceConfig(orgName string) string {
	return fmt.Sprintf(`
data "tfe_agent_pool" "foobar" {
  name         = "Default"
  organization = "%s"
}`, orgName)
}
