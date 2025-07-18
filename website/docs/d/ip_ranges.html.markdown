---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ip_ranges"
description: |-
  Get HCP Terraform and Terraform Enterprise's IP ranges of its services
---

# Data Source: tfe_ip_ranges

Use this data source to retrieve a list of HCP Terraform's IP ranges. For more information about these IP ranges, view our [documentation about HCP Terraform IP Ranges](https://developer.hashicorp.com/terraform/cloud-docs/architectural-details/ip-ranges).

## Example Usage

```hcl
data "tfe_ip_ranges" "addresses" {}

output "notifications_ips" {
  value = data.tfe_ip_ranges.addresses.notifications
}
```

## Argument Reference

No arguments are required for this datasource.

## Attributes Reference

The following attributes are exported:

* `api` - The list of IP ranges in CIDR notation used for connections from user site to HCP Terraform APIs.
* `notifications` - The list of IP ranges in CIDR notation used for notifications.
* `sentinel` - The list of IP ranges in CIDR notation used for outbound requests from Sentinel policies. Applicable for [Policy Checks](https://developer.hashicorp.com/terraform/cloud-docs/policy-enforcement/manage-policy-sets#policy-checks) mode only (also known as Legacy mode). 
* `vcs` - The list of IP ranges in CIDR notation used for connecting to VCS providers.
