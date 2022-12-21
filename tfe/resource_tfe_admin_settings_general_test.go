package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEAdminSettingsGeneral(t *testing.T) {
	// Put admin settings in a known state before checking changes to them
	err := testAccResetAdminSettingsGeneral()
	if err != nil {
		t.Fatal(err)
	}
	// Remember to reset global singleton resources when done!
	t.Cleanup(testAccCleanupAdminSettings)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Set all settings and note the ones that changed.
			{
				Config: testAccTFEAdminSettingsGeneral_all,
				Check: testAccCheckAdminSettingsGeneral(testAccAdminSettingsGeneralExpectation{
					// all changed from their defaults:
					DefaultRemoteStateAccess:          false,
					SendPassingStatusUntriggeredPlans: true,
					APIRateLimit:                      40,
				}),
			},
			// Different config: set minimal settings, observe both changed and unchanged values.
			{
				Config: testAccTFEAdminSettingsGeneral_minimal,
				Check: testAccCheckAdminSettingsGeneral(testAccAdminSettingsGeneralExpectation{
					// changed back explicitly:
					DefaultRemoteStateAccess: true,
					// NOT changed back to default when omitted:
					SendPassingStatusUntriggeredPlans: true,
					// NOT changed back to default when omitted:
					APIRateLimit: 40,
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
			DefaultRemoteStateAccess:          settings.DefaultRemoteStateAccess,
			SendPassingStatusUntriggeredPlans: settings.SendPassingStatusesEnabled,
			APIRateLimit:                      settings.APIRateLimit,
		}
		if actual != expected {
			return fmt.Errorf("Admin settings didn't match: expected %+v, got %+v", expected, actual)
		}
		return nil
	}
}

// testAccAdminSettingsGeneralExpectation represents expected values of a few
// choice admin settings, for legible/printable all-at-once comparisons.
type testAccAdminSettingsGeneralExpectation struct {
	DefaultRemoteStateAccess          bool
	SendPassingStatusUntriggeredPlans bool
	APIRateLimit                      int
}

// Admin settings are a per-instance singleton, so they're dicey to test --
// there's no good way to isolate our actions to a per-test instance, and they
// mutate a shared resource. So heed ye these rules: Don't mess with any
// settings that would interfere with other tests, and reset stuff to a known
// baseline before *and* after testing. Thou shalt not flake.
func testAccResetAdminSettingsGeneral() error {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		return fmt.Errorf("failed to get client for resetting admin settings: %s", err)
	}

	// These happen to be the default values for stock TFE instances:
	opts := tfe.AdminGeneralSettingsUpdateOptions{
		LimitUserOrgCreation:              tfe.Bool(true),
		APIRateLimitingEnabled:            tfe.Bool(true),
		APIRateLimit:                      tfe.Int(30),
		SendPassingStatusUntriggeredPlans: tfe.Bool(false),
		AllowSpeculativePlansOnPR:         tfe.Bool(false),
		DefaultRemoteStateAccess:          tfe.Bool(true),
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
  default_remote_state_access                             = false # non-default
  limit_user_organization_creation                        = true
  api_rate_limiting_enabled                               = true
  api_rate_limit                                          = 40    # non-default
  send_passing_statuses_for_untriggered_speculative_plans = true  # non-default
  allow_speculative_plans_on_pull_requests_from_forks     = false
}
`

const testAccTFEAdminSettingsGeneral_minimal = `
resource "tfe_admin_settings_general" "stuff" {
  default_remote_state_access                             = true
  limit_user_organization_creation                        = true
  api_rate_limiting_enabled                               = false
}
`
