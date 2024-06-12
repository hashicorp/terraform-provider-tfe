package provider

import (
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"fmt"
)

func TestAccTFEDataRetentionPolicy_basic(t *testing.T) {
	policy := &tfe.DataRetentionPolicyChoice{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEDataRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEDataRetentionPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEDataRetentionPolicyExists("tfe_data_retention_policy.foobar", policy),
					testAccCheckTFEDataRetentionPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"tfe_data_retention_policy.foobar", "delete_older_than.days", "42"),
				),
			},
		},
	})
}

func testAccTFEDataRetentionPolicy_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.foobar.id
	
  delete_older_than {
    days = 42
  }
}`, rInt)
}

func testAccCheckTFEDataRetentionPolicyExists(
	n string, policy *tfe.DataRetentionPolicyChoice) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(ConfiguredClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]
		ws, err := config.Client.Workspaces.ReadByID(ctx, wsID)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving workspace %s: %w", wsID, err)
		}

		drp, err := config.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, ws.ID)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving data retention policy for workspace %s: %w", ws.ID, err)
		}

		*policy = *drp

		return nil
	}
}

func testAccCheckTFEDataRetentionPolicyAttributes(
	policy *tfe.DataRetentionPolicyChoice) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.DataRetentionPolicyDeleteOlder == nil {
			return fmt.Errorf("policy wasn't of type 'delete_older'")
		}

		if policy.DataRetentionPolicyDeleteOlder.DeleteOlderThanNDays != 42 {
			return fmt.Errorf("bad delete_older_than_n_days: %d", policy.DataRetentionPolicyDeleteOlder.DeleteOlderThanNDays)
		}

		return nil
	}
}

func testAccCheckTFEDataRetentionPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(ConfiguredClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_data_retention_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		dataRetentionPolicy, err := config.Client.Workspaces.ReadDataRetentionPolicyChoice(ctx, rs.Primary.Attributes["workspace_id"])
		if err == nil {
			if dataRetentionPolicy.DataRetentionPolicyDeleteOlder != nil {
				return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDeleteOlder.ID)
			}
			if dataRetentionPolicy.DataRetentionPolicyDontDelete != nil {
				return fmt.Errorf("data retention policy %s still exists", dataRetentionPolicy.DataRetentionPolicyDontDelete.ID)
			}
			return fmt.Errorf("data retention policy likely exists but couldn't be serialized")
		}
	}

	return nil
}
