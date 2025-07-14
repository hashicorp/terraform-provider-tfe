---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_audit_configuration"
description: |-
  Get information on an organization's audit configuration.
---

# Data Source: tfe_organization_audit_configuration

Use this data source to get information about the audit configuration for a given organization.

~> **NOTE:** This data source requires using the provider with HCP Terraform.

## Example Usage

```hcl
data "tfe_organization_audit_configuration" "example" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the audit configuration

* `audit_trails_enabled` - Whether or not Audit Trails is enabled as an auditing method.

* `hcp_log_streaming_enabled` - Whether or not HCP Audit Log Streaming is enabled as an auditing method.

* `hcp_organization` - The destination HCP Organization for HCP Audit Log Streaming.
