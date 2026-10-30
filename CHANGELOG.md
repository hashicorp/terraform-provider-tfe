## 0.83.0 (October 30, 2026)


NEW FEATURES:

* `d/tfe_saml_idp_certificate`, `d/tfe_saml_idp_certificates`: Adds data sources to read a single trusted SAML IdP certificate by ID or list them, optionally filtered by display name. Requires Terraform Enterprise v2.1.0 or later. by @skj-skj

* `r/tfe_admin_customization_settings`: Adds a resource to manage the admin customization settings, including the support email and support URL address, internal support/documentation/tutorials URL overrides, avatar sources, and login/footer/error/new-user messages. Requires Terraform Enterprise v2.1.0 or later. by @sanjay-kansyakar-hashicorp

* `r/tfe_admin_general_settings`: Adds a resource to manage the admin general settings, including `enable_public_registration`. Requires Terraform Enterprise v2.1.0 or later. by @sanjay-kansyakar-hashicorp

* `r/tfe_stack_policy_set`, `r/tfe_stack_policy_set_exclusion`: Adds resources to scope Terraform policy (`tfpolicy`) policy sets to Stacks and to exclude Stacks from them. On Terraform Enterprise, requires v2.1.0 or later. by @shrishyamb

* `r/tfe_admin_banner`: Adds a `tfe_admin_banner` resource to manage the site-wide announcement banner in Terraform Enterprise. Requires Terraform Enterprise v2.1.0 or later. by @sanjay-kansyakar-hashicorp

* `r/tfe_tfpolicy_version`: Adds a resource for Terraform Enterprise admins to manage Terraform policy plugin versions, with architecture-specific binaries configured using the `archs` attribute. Requires Terraform Enterprise v2.1.0 or later. by @shrishyamb

* `r/tfe_policy_set`, `d/tfe_policy_set`, `r/tfe_policy_set_parameter`, `r/tfe_workspace_policy_set`, `r/tfe_workspace_policy_set_exclusion`, `r/tfe_project_policy_set`, `r/tfe_project_policy_set_exclusion`, `r/tfe_tag_policy_set`, `r/tfe_tag_policy_set_exclusion`: Terraform policy (`tfpolicy`) policy sets are now generally available. Documents the `latest` and `managed` values of `policy_tool_version`. On Terraform Enterprise, requires v2.1.0 or later. by @shrishyamb


ENHANCEMENTS:

* `r/tfe_saml_settings`: `idp_cert` is now optional and deprecated in favor of the `tfe_saml_idp_certificate` resource. `idp_cert` is still required on Terraform Enterprise releases earlier than v2.1.0, and omitting it there fails at plan time. Removing `idp_cert` from configuration only stops Terraform from managing it. Changing only the formatting of `idp_cert` (line wrapping or PEM armor) is no longer a diff. by @skj-skj

* `r/tfe_policy_set`: Rejects `agent_enabled` and `overridable` at plan time when `kind` is `tfpolicy`, because Terraform policy (`tfpolicy`) policy sets do not support them. by @shrishyamb

* `a/tfe_query_run`: Include the query run ID and the end of the query log in the error when a query run errors. by @shwetamurali


BUG FIXES:

* `r/tfe_variable`: Conditionally emit the "Available Write-Only Attribute Alternative" warning when `value` is set and sensitive is true. by @hirokimatsueda ([#2265](https://github.com/hashicorp/terraform-provider-tfe/issues/2265))

* `r/tfe_team_access`: apply errors on some resources now contain more detail when triggered by a backend validation issue by @brandonc


NOTES:

* `r/tfe_saml_settings`, `r/tfe_saml_idp_certificate`: Add a migration guide from tfe_saml_settings.idp_cert to tfe_saml_idp_certificate. by @skj-skj


BREAKING CHANGES:

* `r/tfe_saml_settings`: `idp_cert` can no longer be set to an empty string, whitespace or PEM armor without a certificate body; such values now fail at plan time. Omit `idp_cert` instead. On Terraform Enterprise v2.1.0 or later, `terraform destroy` no longer clears `idp_cert`; it only disables SAML and leaves every IdP certificate on the instance. Use `tfe_saml_idp_certificate` to manage certificate removal. by @skj-skj

## 0.82.0 (October 8, 2026)


NEW FEATURES:

* `r/tfe_registry_artifact_tag`: Adds a resource to manage the tags bound to a registry artifact (module, provider, or component). Tags use replace-all semantics. by @mrinalirao ([#2222](https://github.com/hashicorp/terraform-provider-tfe/pull/2222))

* `d/tfe_registry_artifact_tags`: Adds a data source to read the tags currently associated with a registry artifact. by @mrinalirao ([#2222](https://github.com/hashicorp/terraform-provider-tfe/pull/2222))

* `r/tfe_saml_idp_certificate`: Adds a resource to manage a trusted SAML IdP certificate. Requires Terraform Enterprise v2.1.0 or later. by @skj-skj ([#2240](https://github.com/hashicorp/terraform-provider-tfe/pull/2240))


ENHANCEMENTS:

* `r/tfe_variable`: value_wo_version no longer needs to be incremented manually by @danquack ([#2195](https://github.com/hashicorp/terraform-provider-tfe/pull/2195))

* `r/tfe_policy_set`, `d/tfe_policy_set`, `r/tfe_tag_policy_set`, `r/tfe_tag_policy_set_exclusion`: Add `tag_match_logic` attribute to control whether a policy set applies to workspaces matching any or all configured tags. by @anubhav-goel ([#2197](https://github.com/hashicorp/terraform-provider-tfe/pull/2197), [#2223](https://github.com/hashicorp/terraform-provider-tfe/issues/2223))

* `d/tfe_organization_members`: added `user_email` to members by @kaefferlein ([#2231](https://github.com/hashicorp/terraform-provider-tfe/pull/2231))

* `r/tfe_oauth_client`, `d/tfe_oauth_client`: Add `ado_org_name` support for Azure DevOps Services connections that use organization-scoped personal access tokens. by @zainq11 ([#2218](https://github.com/hashicorp/terraform-provider-tfe/pull/2218))

* `r/tfe_notification_configuration`, `r/tfe_project_notification_configuration`: Expand the set of valid notification configuration `triggers` values by @jillirami ([#2238](https://github.com/hashicorp/terraform-provider-tfe/pull/2238))

* `r/tfe_project_notification_configuration`, `r/tfe_team_notification_configuration`: Make `token_wo_version` Computed with auto-managed hash-based change detection by @jillirami ([#2123](https://github.com/hashicorp/terraform-provider-tfe/pull/2123))


BUG FIXES:

* `r/tfe_project`: Fixes unable to create/update tfe_project when using tag_bindings attribute with Terraform Enterprise by @brandonc ([#2239](https://github.com/hashicorp/terraform-provider-tfe/pull/2239))

* `r/tfe_project`, `r/tfe_stack`, `r/tfe_variable`, `r/tfe_registry_provider`: Fix `Unexpected Identity Change` errors during resource refresh. The identity `hostname` now uses the configured hostname, and `localterraform.com` resolves to the real hostname. Existing identities are upgraded automatically. by @Uk1288 ([#2245](https://github.com/hashicorp/terraform-provider-tfe/pull/2245))
## 0.81.0 (September 15, 2026)


FEATURES: 
* `r/tfe_workspace_hyok_enabled`: Adds a resource to enable HYOK (Hold Your Own Key) on a workspace. Destroying the resource leaves HYOK enabled on the workspace (non-destructive). By @danieldnedialkov [#2192](https://github.com/hashicorp/terraform-provider-tfe/pull/2192)

ENHANCEMENTS:
* `r/tfe_policy_set`: Add `tfpolicy` as a valid value for the `kind` attribute. **NOTE:** This policy kind is currently in beta and not yet available to all users. By @subhro-acharjee-ibm [#2109](https://github.com/hashicorp/terraform-provider-tfe/pull/2109)

* **New Resource List:** `tfe_provider_set` By @kierramarie [#2171](https://github.com/hashicorp/terraform-provider-tfe/pull/2171)

* `r/tfe_hyok_configuration`: Added multi-region key support for AWS HYOK. By @helenjw [#2187](https://github.com/hashicorp/terraform-provider-tfe/pull/2187)

* `r/tfe_saml_settings`, `d/tfe_saml_settings`: Add `attr_site_auditor` and `site_auditor_role` attributes for provisioning the Site Auditor role through SAML. Requires Terraform Enterprise v2.1.0 or later; on earlier releases the attributes are ignored unless set explicitly, in which case a minimum-version error is returned. By @tanushreegorai [#2201](https://github.com/hashicorp/terraform-provider-tfe/pull/2201)

* `r/tfe_scim_settings`, `d/tfe_scim_settings`: Add `site_auditor_group_scim_id` and `site_auditor_group_display_name` attributes for granting the Site Auditor role to the members of a SCIM group, mirroring the existing site admin group mapping. Setting `site_auditor_group_scim_id` to `""` (or omitting it) unlinks the group and revokes the site auditor role from every member of it. Requires Terraform Enterprise v2.1.0 or later; on earlier releases a minimum-version error is returned when `site_auditor_group_scim_id` is set. By @tanushreegorai [#2209](https://github.com/hashicorp/terraform-provider-tfe/pull/2209)

BUG FIXES:
* `r/tfe_saml_settings`: Fix `Provider produced inconsistent result after apply` error on the sensitive `private_key` attribute when updating any other attribute without changing the private key. By @tanushreegorai [#2201](https://github.com/hashicorp/terraform-provider-tfe/pull/2201)

* `r/tfe_saml_settings`: Fix `Provider produced inconsistent result after apply` on `idp_cert`, and a plan that kept showing the same change on every run. Terraform Enterprise stores the certificate wrapped its own way, and the provider was treating that as a different certificate. Affects Terraform Enterprise v2.1.0 and later. By @skj-skj [#2213](https://github.com/hashicorp/terraform-provider-tfe/pull/2213)

* `r/tfe_vault_oidc_configuration`: Fixed method call casing (`SetEncodedCaCert` → `SetEncodedCacert`) to match updated `go-tfe/v2` SDK. By @danielnedialkov [#2207](https://github.com/hashicorp/terraform-provider-tfe/pull/2207)

* `d/tfe_variables`: Fix a nil pointer dereference panic when reading a workspace or variable set containing more than one page (20) of variables. By @justinclayton [#2186](https://github.com/hashicorp/terraform-provider-tfe/pull/2186)

* `d/tfe_project`, `d/tfe_projects`: Fix a nil pointer dereference panic when following pagination links for large result sets. By @ctrombley [#2219](https://github.com/hashicorp/terraform-provider-tfe/pull/2219)

## Older Releases:

* [versions 0.1.0 ~ 0.80.0](CHANGELOG-0.1.0-0.80.0.md)