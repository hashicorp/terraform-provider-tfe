package tfe

import (
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

const testResourceName = "tfe_saml_settings.foobar"

func TestAccTFESAMLSettings_basic(t *testing.T) {
	s := tfe.AdminSAMLSetting{
		IDPCert:        "testIDPCert",
		SLOEndpointURL: "https://foobar.com/slo_endpoint_url",
		SSOEndpointURL: "https://foobar.com/sso_endpoint_url",
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESAMLSettings_basic(s),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "debug", "false"),
					resource.TestCheckResourceAttr(testResourceName, "authn_requests_signed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "want_assertions_signed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "team_management_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "idp_cert", s.IDPCert),
					resource.TestCheckResourceAttr(testResourceName, "slo_endpoint_url", s.SLOEndpointURL),
					resource.TestCheckResourceAttr(testResourceName, "sso_endpoint_url", s.SSOEndpointURL),
					resource.TestCheckResourceAttr(testResourceName, "attr_username", samlDefaultAttrUsername),
					resource.TestCheckResourceAttr(testResourceName, "attr_site_admin", samlDefaultAttrSiteAdmin),
					resource.TestCheckResourceAttr(testResourceName, "attr_groups", samlDefaultAttrGroups),
					resource.TestCheckResourceAttr(testResourceName, "site_admin_role", samlDefaultSiteAdminRole),
					resource.TestCheckResourceAttr(testResourceName, "sso_api_token_session_timeout", strconv.Itoa(int(samlDefaultSSOAPITokenSessionTimeoutSeconds))),
					resource.TestCheckResourceAttrSet(testResourceName, "acs_consumer_url"),
					resource.TestCheckResourceAttrSet(testResourceName, "metadata_url"),
					resource.TestCheckResourceAttr(testResourceName, "signature_signing_method", samlSignatureMethodSHA256),
					resource.TestCheckResourceAttr(testResourceName, "signature_digest_method", samlSignatureMethodSHA256),
				),
			},
		},
	})
}

func testAccTFESAMLSettings_basic(s tfe.AdminSAMLSetting) string {
	return fmt.Sprintf(`
resource "tfe_saml_settings" "foobar" {
  idp_cert         = "%s"
  slo_endpoint_url = "%s"
  sso_endpoint_url = "%s"
}`, s.IDPCert, s.SLOEndpointURL, s.SSOEndpointURL)
}
