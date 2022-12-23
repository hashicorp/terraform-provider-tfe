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
// affect the instance's behavior in unrelated acceptance tests. So heed ye
// these warnings: Don't mess with any settings that would interfere with other
// tests, and reset stuff to a known baseline before *and* after testing.

func TestAccTFEAdminSettingsGeneral(t *testing.T) {
	skipIfCloud(t)

	// Put admin settings in a known state before checking changes to them
	err := testAccResetAdminSettingsGeneral()
	if err != nil {
		t.Fatal(err)
	}
	// Reset when done
	t.Cleanup(testAccCleanupAdminSettings)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Set all settings and check a few select values.
			{
				Config: testAccTFEAdminSettingsGeneral_all,
				Check: testAccCheckAdminSettingsGeneral(testAccAdminSettingsGeneralExpectation{
					// unchanged from default:
					LimitUserOrgCreation: true,
					// changed from their defaults:
					APIRateLimit:                      40,
					DefaultRemoteStateAccess:          false,
					SendPassingStatusUntriggeredPlans: true,
				}),
			},
			// Different config that specifies only a few settings; any omitted
			// ones will have their default values re-enforced
			{
				Config: testAccTFEAdminSettingsGeneral_minimal,
				Check: testAccCheckAdminSettingsGeneral(testAccAdminSettingsGeneralExpectation{
					// changed explicitly by new config:
					DefaultRemoteStateAccess: true,
					LimitUserOrgCreation:     false,
					// arbitrary default re-enforced when omitted from config:
					SendPassingStatusUntriggeredPlans: false,
					APIRateLimit:                      30,
				}),
			},
		},
	})

}

// testAccCheckAdminSettingsGeneral returns a check function that tests whether
// a limited number of admin settings have their expected values.
func testAccCheckAdminSettingsGeneral(expected testAccAdminSettingsGeneralExpectation) resource.TestCheckFunc {
	return func(_s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		settings, err := tfeClient.Admin.Settings.General.Read(ctx)
		if err != nil {
			return err
		}

		actual := testAccAdminSettingsGeneralExpectation{
			APIRateLimit:                      settings.APIRateLimit,
			DefaultRemoteStateAccess:          settings.DefaultRemoteStateAccess,
			LimitUserOrgCreation:              settings.LimitUserOrganizationCreation,
			SendPassingStatusUntriggeredPlans: settings.SendPassingStatusesEnabled,
		}
		if actual != expected {
			return fmt.Errorf("admin general settings didn't match. expected:\n%+v, got:\n%+v", expected, actual)
		}
		return nil
	}
}

// testAccAdminSettingsGeneralExpectation represents expected values of a few
// choice admin settings, for easy-to-print all-at-once comparisons.
type testAccAdminSettingsGeneralExpectation struct {
	APIRateLimit                      int
	DefaultRemoteStateAccess          bool
	LimitUserOrgCreation              bool
	SendPassingStatusUntriggeredPlans bool
}

// Use the client directly to reset all admin settings to their default values.
func testAccResetAdminSettingsGeneral() error {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		return fmt.Errorf("failed to get client for resetting admin settings: %s", err)
	}

	// These happen to be the default values for stock TFE instances:
	opts := tfe.AdminGeneralSettingsUpdateOptions{
		AllowSpeculativePlansOnPR:         tfe.Bool(false),
		APIRateLimit:                      tfe.Int(30),
		APIRateLimitingEnabled:            tfe.Bool(true),
		DefaultRemoteStateAccess:          tfe.Bool(true),
		LimitUserOrgCreation:              tfe.Bool(true),
		SendPassingStatusUntriggeredPlans: tfe.Bool(false),
	}

	_, err = tfeClient.Admin.Settings.General.Update(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to reset general settings: %s", err)
	}
	return nil
}

// Wrapper for reset function to be used in t.Cleanup
func testAccCleanupAdminSettings() {
	err := testAccResetAdminSettingsGeneral()
	if err != nil {
		// The test is already over at this point, so just complain:
		fmt.Printf("error during test cleanup: %s", err)
	}
}

const testAccTFEAdminSettingsGeneral_all = `
resource "tfe_admin_settings_general" "stuff" {
  api_rate_limit                                          = 40    # non-default
  api_rate_limiting_enabled                               = false
  allow_speculative_plans_on_pull_requests_from_forks     = false
  default_remote_state_access                             = false # non-default
  limit_user_organization_creation                        = true
  send_passing_statuses_for_untriggered_speculative_plans = true  # non-default
}
`

const testAccTFEAdminSettingsGeneral_minimal = `
resource "tfe_admin_settings_general" "stuff" {
  default_remote_state_access                             = true # different from "_all"
  limit_user_organization_creation                        = false # different from "_all"
}
`
