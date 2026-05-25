---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_scim_token"
description: |-
  Manages SCIM authentication tokens.
---

# tfe_scim_token

Use this resource to create and manage SCIM authentication tokens. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.

SCIM must be enabled before a token can be created. SCIM in turn requires SAML, so the examples below depend on both a `tfe_saml_settings` and a `tfe_scim_settings` resource.

The token value is only returned when the token is first created and cannot be read afterwards. Treat the `token` attribute as a sensitive secret and store it accordingly (for example, in a secrets manager).

~> **Note:** Changing `description` or `expired_at` (including removing `expired_at` after it has been set) will force the resource to be replaced, which generates a new token value and invalidates the previous one.

## Example Usage

### Basic usage

Create a SCIM token with the default expiration (365 days):

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

resource "tfe_scim_token" "this" {
  description = "scim-token-for-okta"
  depends_on  = [tfe_scim_settings.this]
}
```

### With an explicit expiration

`expired_at` accepts an RFC3339/iso8601 timestamp and must be no more than 365 days in the future. You can use `time_rotating` to generate this dynamically:

```hcl
resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_scim_token" "this" {
  description = "scim-token-30-day"
  expired_at  = time_rotating.example.rotation_rfc3339
  depends_on  = [tfe_scim_settings.this]
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) A human-readable description of the SCIM token. Changing this forces the resource to be replaced.
* `expired_at` - (Optional) The time when the SCIM token expires, in RFC3339/iso8601 format. Defaults to 365 days from creation if unset. Changing this — or removing it after it has been set — forces the resource to be replaced.

## Attributes Reference

* `id` - The ID of the SCIM token.
* `token` - The SCIM token value. Only set immediately after creation; this attribute is `null` after an import and cannot be read back from the API afterwards.
* `created_at` - The time when the SCIM token was created.
* `last_used_at` - The time when the SCIM token was last used.

## Import

SCIM tokens can be imported by their token ID.

```shell
terraform import tfe_scim_token.this at-XXXXXXXXXXXXXXXX
```

The `token` attribute is only returned when the token is first created, so it will be `null` after import.
