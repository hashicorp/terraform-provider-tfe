---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_scim_group_mapping"
description: |-
  Maps a SCIM group to a team.
---

# tfe_scim_group_mapping

Use this resource to map a team in Terraform Enterprise to a SCIM group. When a
team is mapped to a SCIM group, its membership is managed by your Identity
Provider (IdP) through SCIM provisioning. A team can be mapped to at most one
SCIM group.

This resource applies only to Terraform Enterprise and requires admin token
configuration. SCIM (and SAML) must be enabled first; see
[`tfe_scim_settings`](scim_settings.html) for details. The SCIM group you map to
must already be provisioned into Terraform Enterprise by your IdP before the
mapping can be created.

## Example Usage

The SCIM group you map to must already exist in Terraform Enterprise. Groups are
created by your IdP after SCIM provisioning is enabled, so this typically follows
a workflow where SCIM is enabled first, the IdP pushes the group, and then the
mapping is created. Use the [`tfe_scim_group`](../d/scim_group.html) data source
to look up the group's ID by name:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
  provider_type    = "okta"
}

resource "tfe_scim_settings" "this" {
  depends_on = [tfe_saml_settings.this]
}

data "tfe_organization" "this" {
  name = "my-org-name"
}

resource "tfe_team" "engineering" {
  name         = "engineering"
  organization = data.tfe_organization.this.name
}

# Look up the SCIM group provisioned by your IdP.
data "tfe_scim_group" "engineering" {
  name       = "engineering-scim-group"
  depends_on = [tfe_scim_settings.this]
}

resource "tfe_scim_group_mapping" "engineering" {
  team_id       = tfe_team.engineering.id
  scim_group_id = data.tfe_scim_group.engineering.id
}
```

You can pause provisioning for a single mapping without removing it:

```hcl
resource "tfe_scim_group_mapping" "engineering" {
  team_id       = tfe_team.engineering.id
  scim_group_id = data.tfe_scim_group.engineering.id
  paused        = true
}
```

~> **Note:** Creating a mapping with `paused = true` does not skip the initial
sync. The mapping API does not support creating a mapping in a paused state, so
the mapping is created active and then paused in a follow-up request. This means
an initial sync of the SCIM group's members onto the team runs before the pause
takes effect, and any changes made to the SCIM group during that brief window may
be reflected on the team.

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) The ID of the team to map the SCIM group to. Changing
  this forces a new mapping to be created.
* `scim_group_id` - (Required) The ID of the SCIM group to map to the team.
  Changing this forces a new mapping to be created, since the mapping API only
  supports updating the paused state.
* `paused` - (Optional) Whether provisioning for this mapping is paused. Defaults
  to `false`. Setting this to `true` on creation does not skip the initial sync,
  since the mapping is created active and then paused in a follow-up request.

## Attributes Reference

* `id` - The ID of the SCIM group mapping. Since a team can only be mapped to one
  SCIM group, this is the same as `team_id`.

## Import

SCIM group mappings can be imported using the team ID.

```shell
terraform import tfe_scim_group_mapping.engineering team-xxxxxxxxxxxxxxxx
```
