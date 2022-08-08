package tfe

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	slug "github.com/hashicorp/go-slug"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFESlug() *schema.Resource {
	return &schema.Resource{
		Description: "This data source is used to represent configuration files on a local filesystem intended to be uploaded to Terraform Cloud/Enterprise, in lieu of those files being sourced from a configured VCS provider." +
			"\n\nA unique checksum is generated for the specified local directory, which allows resources such as `tfe_policy_set` track the files and upload a new gzip compressed tar file containing configuration files (a Terraform slug) when those files change.",
		Read: dataSourceTFESlugRead,

		Schema: map[string]*schema.Schema{
			"source_path": {
				Description: "The path to the directory where the files are located.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTFESlugRead(d *schema.ResourceData, meta interface{}) error {
	sourcePath := d.Get("source_path").(string)

	log.Printf("[DEBUG] Hashing the source path files: %s", sourcePath)
	chksum, err := hashPolicies(sourcePath)
	if err != nil {
		return fmt.Errorf("Error generating the checksum for the source path files: %w", err)
	}
	d.SetId(chksum)

	return nil
}

func hashPolicies(path string) (string, error) {
	body := bytes.NewBuffer(nil)
	file, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !file.Mode().IsDir() {
		return "", fmt.Errorf("The path is not a directory")
	}

	_, err = slug.Pack(path, body, true)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(body.Bytes())
	chksum := hex.EncodeToString(hash.Sum(nil))

	return chksum, nil
}
