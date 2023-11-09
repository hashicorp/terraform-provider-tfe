// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEIPRanges() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFEIPRangesRead,

		Schema: map[string]*schema.Schema{
			"api": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"notifications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sentinel": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vcs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEIPRangesRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Reading IP Ranges")
	ipRanges, err := config.Client.Meta.IPRanges.Read(ctx, "")
	if err != nil {
		return fmt.Errorf("Error retrieving IP ranges: %w", err)
	}

	d.SetId("ip-ranges")
	d.Set("api", ipRanges.API)
	d.Set("notifications", ipRanges.Notifications)
	d.Set("sentinel", ipRanges.Sentinel)
	d.Set("vcs", ipRanges.VCS)

	return nil
}
