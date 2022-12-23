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

// The Twilio settings are a worst-case scenario: they consist mostly of
// credentials, for which neither reasonable "reset" values nor reasonable test
// values can be encoded in an acceptance test, AND attempting to enable the
// service performs a live validation request to Twilio, which of course fails
// (blocking the update) if your account and key aren't functional. That means
// we can't do much of anything interesting here; about all we can safely do is
// a smoke test to verify that the resource itself doesn't explode.

func TestAccTFEAdminSettingsTwilio(t *testing.T) {
	skipIfCloud(t)

	t.Cleanup(testAccResetAdminSettingsTwilio)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAdminSettingsTwilioConfig(false),
				Check:  testAccCheckTFEAdminSettingsTwilio(false),
			},
		},
	})
}

// As this is only used in t.Cleanup, it's too late to use t.Fatalf, so just complain.
func testAccResetAdminSettingsTwilio() {
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		fmt.Printf("failed to get client for resetting admin Twilio settings: %s", err)
	}

	opts := tfe.AdminTwilioSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	}

	_, err = tfeClient.Admin.Settings.Twilio.Update(ctx, opts)
	if err != nil {
		fmt.Printf("failed to reset admin Twilio settings: %s", err)
	}
}

func testAccCheckTFEAdminSettingsTwilio(v bool) resource.TestCheckFunc {
	return func(_s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		settings, err := tfeClient.Admin.Settings.Twilio.Read(ctx)
		if err != nil {
			return err
		}

		if actual := settings.Enabled; v != actual {
			return fmt.Errorf("admin Twilio settings enabled didn't match. expected %t, got %t", v, actual)
		}
		return nil
	}
}

func testAccTFEAdminSettingsTwilioConfig(v bool) string {
	return fmt.Sprintf(`
resource "tfe_admin_settings_twilio" "stuff" {
  enabled = %t
}
`, v)
}
