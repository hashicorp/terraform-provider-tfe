package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Admin settings are a per-instance singleton, so testing them is inconvenient
// and hazardous -- can't isolate our actions to a per-test resource instance,
// can't test them in parallel, and testing them mutates global state that can
// affect the instance's behavior in unrelated acceptance tests.

// And the cost estimation settings consist mostly of credentials, for which
// reasonable "reset" values cannot be encoded in an acceptance test. That means
// we can't do much of anything interesting here; about all we can safely do is
// a smoke test to verify that the resource itself doesn't explode.

func TestAccTFEAdminSettingsCostEstimation(t *testing.T) {
	skipIfCloud(t)

	t.Cleanup(testAccResetAdminSettingsCostEstimation)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAdminSettingsCostEstimationConfig(false),
				Check:  testAccCheckTFEAdminSettingsCostEstimation(false),
			},
			{
				Config: testAccTFEAdminSettingsCostEstimationConfig(true),
				Check:  testAccCheckTFEAdminSettingsCostEstimation(true),
			},
		},
	})
}

// As this is only used in t.Cleanup, it's too late to use t.Fatalf, so just complain.
func testAccResetAdminSettingsCostEstimation() {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		fmt.Printf("failed to get client for resetting admin cost estimation settings: %s", err)
	}

	opts := tfe.AdminCostEstimationSettingOptions{
		Enabled: tfe.Bool(true),
	}

	_, err = tfeClient.Admin.Settings.CostEstimation.Update(ctx, opts)
	if err != nil {
		fmt.Printf("failed to reset admin cost estimation settings: %s", err)
	}
}

func testAccCheckTFEAdminSettingsCostEstimation(v bool) resource.TestCheckFunc {
	return func(_s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		settings, err := tfeClient.Admin.Settings.CostEstimation.Read(ctx)
		if err != nil {
			return err
		}

		if actual := settings.Enabled; v != actual {
			return fmt.Errorf("admin cost estimation settings enabled didn't match. expected %t, got %t", v, actual)
		}
		return nil
	}
}

func testAccTFEAdminSettingsCostEstimationConfig(v bool) string {
	return fmt.Sprintf(`
resource "tfe_admin_settings_cost_estimation" "stuff" {
  enabled = %t
}
`, v)
}
