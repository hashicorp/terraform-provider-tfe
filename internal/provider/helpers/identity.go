package helpers

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Writes a standard resource identity for an SDKv2 resource in the TFE provider.
// Note: Use this when a resource does not require the organization for import
func WriteTFEIdentity(d *schema.ResourceData, externalID string, hostname string) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	// Set the identity if it has not been set
	if id := identity.Get("id"); id == nil || id == "" {
		err = identity.Set("id", externalID)
		if err != nil {
			return fmt.Errorf("Failed writing resource identity id %s: %w", externalID, err)
		}

		err = identity.Set("hostname", hostname)
		if err != nil {
			return fmt.Errorf("Failed writing resource identity hostname %s: %w", hostname, err)
		}
	}

	return nil
}

// Writes a standard resource identity for an SDKv2 resource in the TFE provider.
// Note: Use this when a resource does requires the organization for import or reading, and the organization cannot be fetched
// using the externalID of a resource alone.
//
// Example: tfe_team requires an organization to be specified when importing.
func WriteTFEIdentityWithOrg(d *schema.ResourceData, externalID string, organization string, hostname string) error {
	// First write the basic identity
	if err := WriteTFEIdentity(d, externalID, hostname); err != nil {
		return err
	}

	identity, err := d.Identity()
	if err != nil {
		return err
	}

	if err = identity.Set("organization", organization); err != nil {
		return fmt.Errorf("failed writing resource identity organization %s: %w", organization, err)
	}

	return nil
}

// WriteRegistryIdentity writes a standard resource identity using the SDKv2
// for a Registry module or provider resource in the TFE provider.
// Note: It expands the complex tfe.RegistryModuleID into individual state fields
// (namespace, name, module_provider, etc.) required for import and reading.
func WriteRegistryIdentity(d *schema.ResourceData, externalID string, rmID tfe.RegistryModuleID, hostname string) error {
	// First write the basic identity
	if err := WriteTFEIdentity(d, externalID, hostname); err != nil {
		return err
	}

	identity, err := d.Identity()
	if err != nil {
		return err
	}

	if err = identity.Set("organization", rmID.Organization); err != nil {
		return fmt.Errorf("failed writing registry identity organization %s: %w", rmID.Organization, err)
	}

	if err = identity.Set("registry_name", string(rmID.RegistryName)); err != nil {
		return fmt.Errorf("failed writing registry identity registry name %s: %w", rmID.RegistryName, err)
	}

	if err = identity.Set("namespace", rmID.Namespace); err != nil {
		return fmt.Errorf("failed writing registry identity namespace %s: %w", rmID.Namespace, err)
	}

	if err = identity.Set("name", rmID.Name); err != nil {
		return fmt.Errorf("failed writing registry identity name %s: %w", rmID.Name, err)
	}

	if err = identity.Set("module_provider", rmID.Provider); err != nil {
		return fmt.Errorf("failed writing registry identity module_provider %s: %w", rmID.Provider, err)
	}

	return nil
}
