package provider

import (
	"os"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

type adminRoleType string

const (
	siteAdmin                adminRoleType = "site-admin"
	configurationAdmin       adminRoleType = "configuration"
	provisionLicensesAdmin   adminRoleType = "provision-licenses"
	subscriptionAdmin        adminRoleType = "subscription"
	supportAdmin             adminRoleType = "support"
	securityMaintenanceAdmin adminRoleType = "security-maintenance"
	versionMaintenanceAdmin  adminRoleType = "version-maintenance"
)

func getTokenForAdminRole(adminRole adminRoleType) string {
	token := ""

	switch adminRole {
	case siteAdmin:
		token = os.Getenv("TFE_ADMIN_SITE_ADMIN_TOKEN")
	case configurationAdmin:
		token = os.Getenv("TFE_ADMIN_CONFIGURATION_TOKEN")
	case provisionLicensesAdmin:
		token = os.Getenv("TFE_ADMIN_PROVISION_LICENSES_TOKEN")
	case subscriptionAdmin:
		token = os.Getenv("TFE_ADMIN_SUBSCRIPTION_TOKEN")
	case supportAdmin:
		token = os.Getenv("TFE_ADMIN_SUPPORT_TOKEN")
	case securityMaintenanceAdmin:
		token = os.Getenv("TFE_ADMIN_SECURITY_MAINTENANCE_TOKEN")
	case versionMaintenanceAdmin:
		token = os.Getenv("TFE_ADMIN_VERSION_MAINTENANCE_TOKEN")
	}

	return token
}

func testAdminClient(t *testing.T, adminRole adminRoleType) *tfe.Client {
	token := getTokenForAdminRole(adminRole)
	if token == "" {
		t.Fatal("missing API token for admin role " + adminRole)
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token: token,
	})
	if err != nil {
		t.Fatal(err)
	}

	return client
}
