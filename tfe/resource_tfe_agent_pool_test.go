package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEAgentPool_basic(t *testing.T) {
	skipIfFreeOnly(t)

	agentPool := &tfe.AgentPool{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributes(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-test"),
				),
			},
		},
	})
}

func TestAccTFEAgentPool_update(t *testing.T) {
	skipIfFreeOnly(t)

	agentPool := &tfe.AgentPool{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributes(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-test"),
				),
			},

			{
				Config: testAccTFEAgentPool_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentPoolExists(
						"tfe_agent_pool.foobar", agentPool),
					testAccCheckTFEAgentPoolAttributesUpdated(agentPool),
					resource.TestCheckResourceAttr(
						"tfe_agent_pool.foobar", "name", "agent-pool-updated"),
				),
			},
		},
	})
}

func TestAccTFEAgentPool_import(t *testing.T) {
	skipIfFreeOnly(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentPool_basic(rInt),
			},

			{
				ResourceName:      "tfe_agent_pool.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEAgentPoolExists(
	n string, agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		sk, err := tfeClient.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("agent pool not found")
		}

		*agentPool = *sk

		return nil
	}
}

func testAccCheckTFEAgentPoolAttributes(
	agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentPool.Name != "agent-pool-test" {
			return fmt.Errorf("Bad name: %s", agentPool.Name)
		}
		return nil
	}
}

func testAccCheckTFEAgentPoolAttributesUpdated(
	agentPool *tfe.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentPool.Name != "agent-pool-updated" {
			return fmt.Errorf("Bad name: %s", agentPool.Name)
		}
		return nil
	}
}

func testAccCheckTFEAgentPoolDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_agent_pool" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.AgentPools.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("agent pool %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEAgentPool_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = tfe_organization.foobar.id
}`, rInt)
}

func testAccTFEAgentPool_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-updated"
  organization = tfe_organization.foobar.id
}`, rInt)
}
