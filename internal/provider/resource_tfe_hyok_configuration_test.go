// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

const (
	hyokConfigurationStatusTestFailed = "test_failed"
	hyokConfigurationStatusRevoked    = "revoked"
)

func TestAccTFEHYOKConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createPremiumOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	var hyokConfigID string

	// With AWS OIDC configuration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAWSHYOKConfigurationConfig(org.Name, "apple", "arn:aws:kms:us-east-1:123456789012:key/key1", "us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", &hyokConfigID),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "apple"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "aws"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "arn:aws:kms:us-east-1:123456789012:key/key1"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_region", "us-east-1"),
				),
			},
			// Import
			{
				ResourceName:      "tfe_hyok_configuration.hyok",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEAWSHYOKConfigurationConfig(org.Name, "orange", "arn:aws:kms:us-east-1:123456789012:key/key2", "us-east-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "orange"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "aws"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "arn:aws:kms:us-east-1:123456789012:key/key2"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_region", "us-east-2"),
				),
			},
			// Delete - must first revoke configuration to avoid dangling resources
			{
				PreConfig: func() { revokeHYOKConfiguration(t, hyokConfigID) },
				Config:    testAccTFEHYOKConfigurationDestroyConfig(org.Name),
			},
		},
	})

	// With Vault OIDC configuration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVaultHYOKConfigurationConfig(org.Name, "peach", "key1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", &hyokConfigID),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "peach"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "vault"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "key1"),
				),
			},
			// Import
			{
				ResourceName:      "tfe_hyok_configuration.hyok",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEVaultHYOKConfigurationConfig(org.Name, "strawberry", "key2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "strawberry"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "vault"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "key2"),
				),
			},
			// Delete - must first revoke configuration to avoid dangling resources
			{
				PreConfig: func() { revokeHYOKConfiguration(t, hyokConfigID) },
				Config:    testAccTFEHYOKConfigurationDestroyConfig(org.Name),
			},
		},
	})

	// With GCP OIDC configuration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGCPHYOKConfigurationConfig(org.Name, "cucumber", "key1", "global", "key-ring-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", &hyokConfigID),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "cucumber"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "gcp"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "key1"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_location", "global"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_ring_id", "key-ring-1"),
				),
			},
			// Import
			{
				ResourceName:      "tfe_hyok_configuration.hyok",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEGCPHYOKConfigurationConfig(org.Name, "tomato", "key2", "global", "key-ring-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", &hyokConfigID),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "tomato"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "gcp"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "key2"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_location", "global"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kms_options.key_ring_id", "key-ring-2"),
				),
			},
			// Delete - must first revoke configuration to avoid dangling resources
			{
				PreConfig: func() { revokeHYOKConfiguration(t, hyokConfigID) },
				Config:    testAccTFEHYOKConfigurationDestroyConfig(org.Name),
			},
		},
	})

	// With Azure OIDC configuration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAzureHYOKConfigurationConfig(org.Name, "banana", "https://random.vault.azure.net/keys/key1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", &hyokConfigID),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "banana"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "azure"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "https://random.vault.azure.net/keys/key1"),
				),
			},
			// Import
			{
				ResourceName:      "tfe_hyok_configuration.hyok",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEAzureHYOKConfigurationConfig(org.Name, "blueberry", "https://random.vault.azure.net/keys/key2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "name", "blueberry"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "oidc_configuration_type", "azure"),
					resource.TestCheckResourceAttrSet("tfe_hyok_configuration.hyok", "oidc_configuration_id"),
					resource.TestCheckResourceAttr("tfe_hyok_configuration.hyok", "kek_id", "https://random.vault.azure.net/keys/key2"),
				),
			},
			// Delete - must first revoke configuration to avoid dangling resources
			{
				PreConfig: func() { revokeHYOKConfiguration(t, hyokConfigID) },
				Config:    testAccTFEHYOKConfigurationDestroyConfig(org.Name),
			},
		},
	})
}

func waitForHYOKConfigurationStatus(t *testing.T, id string, status string) error {
	// Wait for configuration to reach the given status
	_, err := retryFn(10, 1, func() (any, error) {
		env, err := testAccConfiguredClient.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(id).Get(ctx, nil)
		if err != nil {
			t.Fatalf("failed to read HYOK configuration: %v", err)
		}

		current := hyokConfigurationStatusFromEnvelope(env)
		if current != status {
			return nil, fmt.Errorf("expected HYOK configuration to be %s, got %s", status, current)
		}
		return nil, nil
	})

	return err
}

// hyokConfigurationStatusFromEnvelope extracts the status attribute from a v2
// HYOK configuration envelope. The v2 model does not surface a typed status
// field, so it is read from the untyped additional data map.
func hyokConfigurationStatusFromEnvelope(env models.HyokConfigurationsEnvelopeable) string {
	if env == nil || env.GetData() == nil {
		return ""
	}
	attributes := env.GetData().GetAttributes()
	if attributes == nil {
		return ""
	}
	switch v := attributes.GetAdditionalData()["status"].(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
	}
	return ""
}

func revokeHYOKConfiguration(t *testing.T, id string) {
	// Wait for configuration to be in the test_failed status before revoking
	err := waitForHYOKConfigurationStatus(t, id, hyokConfigurationStatusTestFailed)
	if err != nil {
		t.Fatal(err)
	}

	revokeConfig := &abstractions.RequestConfiguration[abstractions.DefaultQueryParameters]{
		Headers: abstractions.NewRequestHeaders(),
	}
	revokeConfig.Headers.TryAdd("Content-Type", "application/vnd.api+json")

	err = testAccConfiguredClient.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(id).Actions().Revoke().Post(ctx, revokeConfig)
	if err != nil {
		t.Fatalf("failed to revoke HYOK configuration: %v", err)
	}

	// Wait for configuration to be in the revoked status
	err = waitForHYOKConfigurationStatus(t, id, hyokConfigurationStatusRevoked)
	if err != nil {
		t.Fatal(err)
	}
}

func testAccCheckTFEHYOKConfigurationExists(n string, hyokConfigID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		result, err := testAccConfiguredClient.ClientV2.API.HyokConfigurations().ByHyok_configuration_id(rs.Primary.ID).Get(ctx, nil)
		if err != nil {
			return err
		}

		if result.GetData() == nil || result.GetData().GetId() == nil {
			return fmt.Errorf("no HYOK configuration ID returned for %s", rs.Primary.ID)
		}

		*hyokConfigID = *result.GetData().GetId()

		return nil
	}
}

func testAccTFEAWSHYOKConfigurationConfig(orgName string, name string, kekID string, keyRegion string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "pool" {
	name            = "hyok-pool"
	organization    = "%s"
}

resource "tfe_aws_oidc_configuration" "aws_oidc_config" {
	role_arn        = "arn:aws:iam::111111111111:role/example-role-arn"
	organization    = "%s"
}

resource "tfe_hyok_configuration" "hyok" {
	organization                = "%s"
	name                        = "%s"
	kek_id                      = "%s"
	agent_pool_id               = resource.tfe_agent_pool.pool.id
	oidc_configuration_id       = resource.tfe_aws_oidc_configuration.aws_oidc_config.id
	oidc_configuration_type     = "aws"
	kms_options {
		key_region = "%s"
	}
}
`, orgName, orgName, orgName, name, kekID, keyRegion)
}

func testAccTFEVaultHYOKConfigurationConfig(orgName string, name string, kekID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "pool" {
	name            = "hyok-pool"
	organization    = "%s"
}

resource "tfe_vault_oidc_configuration" "vault_oidc_config" {
	address          = "https://vault.example.com"
	role_name        = "vault-role-name"
	namespace        = "admin"
	organization     = "%s"
}

resource "tfe_hyok_configuration" "hyok" {
	organization                    = "%s"
	name                            = "%s"
	kek_id                          = "%s"
	agent_pool_id                   = resource.tfe_agent_pool.pool.id
	oidc_configuration_id           = resource.tfe_vault_oidc_configuration.vault_oidc_config.id
	oidc_configuration_type         = "vault"
}
`, orgName, orgName, orgName, name, kekID)
}

func testAccTFEGCPHYOKConfigurationConfig(orgName string, name string, kekID string, keyLocation string, keyRingID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "pool" {
	name            = "hyok-pool"
	organization    = "%s"
}

resource "tfe_gcp_oidc_configuration" "gcp_oidc_config" {
	service_account_email   = "myemail@gmail.com"
	project_number          = "11111111"
	workload_provider_name  = "projects/1/locations/global/workloadIdentityPools/1/providers/1"
	organization            = "%s"
}

resource "tfe_hyok_configuration" "hyok" {
	organization                = "%s"
	name                        = "%s"
	kek_id                      = "%s"
	agent_pool_id               = resource.tfe_agent_pool.pool.id
	oidc_configuration_id       = resource.tfe_gcp_oidc_configuration.gcp_oidc_config.id
	oidc_configuration_type     = "gcp"
	kms_options {
		key_location    = "%s"
		key_ring_id     = "%s"
	}
}
`, orgName, orgName, orgName, name, kekID, keyLocation, keyRingID)
}

func testAccTFEAzureHYOKConfigurationConfig(orgName string, name string, kekID string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "pool" {
	name            = "hyok-pool"
	organization    = "%s"
}

resource "tfe_azure_oidc_configuration" "azure_oidc_config" {
	client_id           = "application-id-1"
	subscription_id     = "subscription-id-1"
	tenant_id           = "tenant-id1"
	organization        = "%s"
}

resource "tfe_hyok_configuration" "hyok" {
	organization                    = "%s"
	name                            = "%s"
	kek_id                          = "%s"
	agent_pool_id                   = resource.tfe_agent_pool.pool.id
	oidc_configuration_id           = resource.tfe_azure_oidc_configuration.azure_oidc_config.id
	oidc_configuration_type         = "azure"
}
`, orgName, orgName, orgName, name, kekID)
}

func testAccTFEHYOKConfigurationDestroyConfig(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_agent_pool" "pool" {
	name            = "hyok-pool"
	organization    = "%s"
}
`, orgName)
}
