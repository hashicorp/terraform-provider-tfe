package tfe

import (
	"fmt"
	"log"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEIPRanges() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a list of Terraform Cloud's IP ranges. For more information about these IP ranges, view our [documentation about Terraform Cloud IP Ranges](https://www.terraform.io/docs/cloud/architectural-details/ip-ranges.html).",
		Read:        dataSourceTFEIPRangesRead,

		Schema: map[string]*schema.Schema{
			"api": {
				Description: "The list of IP ranges in CIDR notation used for connections from user site to Terraform Cloud APIs.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"notifications": {
				Description: "The list of IP ranges in CIDR notation used for notifications.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"sentinel": {
				Description: "The list of IP ranges in CIDR notation used for outbound requests from Sentinel policies.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"vcs": {
				Description: "The list of IP ranges in CIDR notation used for connecting to VCS providers.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTFEIPRangesRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Reading IP Ranges")
	ipRanges, err := tfeClient.Meta.IPRanges.Read(ctx, "")
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
