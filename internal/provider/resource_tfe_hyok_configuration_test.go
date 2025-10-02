package provider

import (
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

func TestAccTFEHYOKConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createPremiumOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	state := &tfe.HYOKConfiguration{}

	// With AWS OIDC configuration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAWSHYOKConfigurationConfig(org.Name, "apple", "arn:aws:kms:us-east-1:123456789012:key/key1", "us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", state),
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
				PreConfig: func() { revokeHYOKConfiguration(t, state.ID) },
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
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", state),
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
				PreConfig: func() { revokeHYOKConfiguration(t, state.ID) },
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
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", state),
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
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", state),
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
				PreConfig: func() { revokeHYOKConfiguration(t, state.ID) },
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
					testAccCheckTFEHYOKConfigurationExists("tfe_hyok_configuration.hyok", state),
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
				PreConfig: func() { revokeHYOKConfiguration(t, state.ID) },
				Config:    testAccTFEHYOKConfigurationDestroyConfig(org.Name),
			},
		},
	})
}

func revokeHYOKConfiguration(t *testing.T, id string) {
	err := testAccConfiguredClient.Client.HYOKConfigurations.Revoke(ctx, id)
	if err != nil {
		t.Fatalf("failed to revoke HYOK configuration: %v", err)
	}

	// Wait for configuration to be in the revoked status
	_, err = retryFn(10, 1, func() (any, error) {
		hyok, err := testAccConfiguredClient.Client.HYOKConfigurations.Read(ctx, id, nil)
		if err != nil {
			t.Fatalf("failed to read HYOK configuration: %v", err)
		}

		if hyok.Status != tfe.HYOKConfigurationRevoked {
			return nil, fmt.Errorf("expected HYOK configuration to be revoked, got %s", hyok.Status)
		}
		return nil, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testAccCheckTFEHYOKConfigurationExists(n string, hyokConfig *tfe.HYOKConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		result, err := testAccConfiguredClient.Client.HYOKConfigurations.Read(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		*hyokConfig = *result

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
