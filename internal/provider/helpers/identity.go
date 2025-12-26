package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Writes a standard resource identity for an SDKv2 resource in the TFE provider
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
