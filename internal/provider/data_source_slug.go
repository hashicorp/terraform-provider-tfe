// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

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
		Read: dataSourceTFESlugRead,

		Schema: map[string]*schema.Schema{
			"source_path": {
				Type:     schema.TypeString,
				Required: true,
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
		return "", fmt.Errorf("the path is not a directory")
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
