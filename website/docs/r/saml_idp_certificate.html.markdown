---
layout: "tfe"
page_title: "Terraform Enterprise: Resource tfe_saml_idp_certificate"
description: |-
  (Only for Terraform Enterprise) Manages a trusted SAML Identity Provider (IdP) certificate. Terraform Enterprise trusts every certificate in the list when it verifies a SAML response, so you can trust several IdP certificates at once, such as one per IdP region, or old and new certificates during a rotation.
  Requires admin token configuration. See example usage for incorporating an admin token in your provider config.
  ~> Note: This resource requires Terraform Enterprise v2.1.0 or later. On earlier releases the provider returns a minimum-version error.
  ~> Note: Terraform Enterprise does not allow deleting the last trusted certificate while SAML is enabled. To remove every certificate, disable SAML first.
  ~> Note: Terraform Enterprise does not allow enabling SAML without a trusted certificate. When tfe_saml_settings relies on these certificates instead of idp_cert, add depends_on pointing at them, as in the example below. This creates the certificates before SAML is enabled and disables SAML before the certificates are destroyed.
  -> Migrating from idp_cert? See Migrating from idp_cert.
---

# Resource: tfe_saml_idp_certificate

(Only for Terraform Enterprise) Manages a trusted SAML Identity Provider (IdP) certificate. Terraform Enterprise trusts every certificate in the list when it verifies a SAML response, so you can trust several IdP certificates at once, such as one per IdP region, or old and new certificates during a rotation.

Requires admin token configuration. See example usage for incorporating an admin token in your provider config.

~> **Note:** This resource requires Terraform Enterprise v2.1.0 or later. On earlier releases the provider returns a minimum-version error.

~> **Note:** Terraform Enterprise does not allow deleting the last trusted certificate while SAML is enabled. To remove every certificate, disable SAML first.

~> **Note:** Terraform Enterprise does not allow enabling SAML without a trusted certificate. When `tfe_saml_settings` relies on these certificates instead of `idp_cert`, add `depends_on` pointing at them, as in the example below. This creates the certificates before SAML is enabled and disables SAML before the certificates are destroyed.

-> **Migrating from `idp_cert`?** See [Migrating from `idp_cert`](#migrating-from-idp_cert).

## Example Usage

```terraform
variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The Terraform Enterprise hostname."
  default     = "tfe.example.com"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_idp_certificate" "primary" {
  display_name = "fooidp-us-east"
  cert         = file("${path.module}/fooidp-us-east.pem")
}

resource "tfe_saml_idp_certificate" "failover" {
  display_name = "fooidp-eu-west"
  cert         = file("${path.module}/fooidp-eu-west.pem")
}

resource "tfe_saml_settings" "this" {
  sso_endpoint_url = "https://example.com/sso"
  slo_endpoint_url = "https://example.com/slo"

  depends_on = [
    tfe_saml_idp_certificate.primary,
    tfe_saml_idp_certificate.failover,
  ]
}
```

### Rotating a certificate

To rotate without SSO downtime, add the new certificate as another `tfe_saml_idp_certificate` and apply. Terraform Enterprise trusts both certificates while your IdP switches over. Once your IdP signs with the new certificate, remove the old resource and apply again. This also ensures the old certificate is not the last trusted certificate, which cannot be deleted while SAML is enabled.

You can also replace `cert` in place. Terraform updates the existing certificate without changing its ID, but the old certificate is overwritten and hence no longer trusted, so SSO logins fail until your IdP signs with the new certificate.

### Migrating from `idp_cert`

The `idp_cert` attribute of [`tfe_saml_settings`](saml_settings) is deprecated in favor of this resource.

-> **Note:** Removing `idp_cert` does not delete the certificate. Terraform Enterprise keeps it as a trusted legacy certificate, so SSO keeps working and you do not need to issue a new certificate. Follow the steps below only to replace it with a new, fully managed certificate.

Certificates are unique, so you cannot add the certificate currently set in `idp_cert` as a `tfe_saml_idp_certificate`; Terraform Enterprise rejects it with `Fingerprint is already trusted`. You also cannot remove the `idp_cert` certificate first, because it is the only trusted certificate keeping SAML working. The migration therefore starts with a **new** certificate from your IdP.

Starting configuration:

```terraform
resource "tfe_saml_settings" "this" {
  sso_endpoint_url = "https://example.com/sso"
  slo_endpoint_url = "https://example.com/slo"
  idp_cert         = file("${path.module}/idp-cert-old.pem")
}
```

**Step 1: Add a new certificate.** Issue a new signing certificate from your IdP and add it as a `tfe_saml_idp_certificate`. Keep `idp_cert` for now, then run `terraform apply`. Terraform Enterprise trusts both certificates.

```terraform
resource "tfe_saml_idp_certificate" "primary" {
  display_name = "my-idp-2026"
  cert         = file("${path.module}/idp-cert-new.pem")
}

resource "tfe_saml_settings" "this" {
  sso_endpoint_url = "https://example.com/sso"
  slo_endpoint_url = "https://example.com/slo"
  idp_cert         = file("${path.module}/idp-cert-old.pem")

  depends_on = [tfe_saml_idp_certificate.primary]
}
```

**Step 2: Confirm logins work.** Switch your IdP to sign SAML responses with the new certificate and confirm users can log in through SSO. Do not continue until they can.

**Step 3: Remove `idp_cert`.** Remove `idp_cert` from `tfe_saml_settings`, then run `terraform apply`. Nothing is deleted: the old certificate stays trusted as a legacy certificate (`cert_role` of `legacy_primary` or `legacy_old`), and `idp_cert` shows as known after apply on updates.

```terraform
resource "tfe_saml_idp_certificate" "primary" {
  display_name = "my-idp-2026"
  cert         = file("${path.module}/idp-cert-new.pem")
}

resource "tfe_saml_settings" "this" {
  sso_endpoint_url = "https://example.com/sso"
  slo_endpoint_url = "https://example.com/slo"

  depends_on = [tfe_saml_idp_certificate.primary]
}
```

**Step 4: Remove the legacy certificate.** Either delete it from the Terraform Enterprise UI or API, or import it into a `tfe_saml_idp_certificate` resource (see [Import](#import)), then remove that resource from your configuration and run `terraform apply`.

~> **Note:** Remove `idp_cert` from `tfe_saml_settings` before importing a legacy certificate. Otherwise both resources manage the same certificate.

To use Terraform for removal, add this temporary resource alongside the configuration from step 3:

```terraform
resource "tfe_saml_idp_certificate" "legacy" {
  display_name = "legacy-idp"
  cert         = file("${path.module}/idp-cert-old.pem")
}
```

Import the existing legacy certificate before running `terraform apply`. Replace `idpc-AbC123XyZ456` with its ID from the Terraform Enterprise API:

```shell
terraform import tfe_saml_idp_certificate.legacy idpc-AbC123XyZ456
```

After the import succeeds, remove only the temporary `tfe_saml_idp_certificate.legacy` block and run `terraform apply`. Terraform deletes the legacy certificate while keeping the new `primary` certificate trusted.

~> **Note:** Importing a legacy certificate is not a full migration on its own. Legacy certificates only allow changing `display_name`, so follow steps 1 to 3 to move to a fully managed certificate.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cert` (String) The PEM encoded X.509 certificate as provided by the IdP. Legacy certificates cannot have their body changed.
- `display_name` (String) A human-readable label for the certificate.

### Read-Only

- `cert_role` (String) The role of the certificate: `managed`, `legacy_primary` or `legacy_old`.
- `created_at` (String) The time the certificate was added, in RFC 3339 format.
- `expires_at` (String) The time the certificate expires, in RFC 3339 format.
- `fingerprint` (String) Colon-separated uppercase SHA-256 fingerprint of the certificate.
- `id` (String) The ID of the IdP certificate (idpc-*).
- `issuer` (String) The issuer of the certificate, derived from its O and CN fields.



## Import

Certificates can be imported by ID:

```shell
terraform import tfe_saml_idp_certificate.this idpc-AbC123XyZ456
```

Certificates set through the `idp_cert` attribute of `tfe_saml_settings` are legacy certificates (`cert_role` of `legacy_primary` or `legacy_old`). You can import them, but you can only change their `display_name`. Changing `cert` on a legacy certificate returns an error.
