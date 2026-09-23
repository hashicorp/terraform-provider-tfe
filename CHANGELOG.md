## 0.81.0 (September 15, 2026)


FEATURES: 
* `r/tfe_workspace_hyok_enabled`: Adds a resource to enable HYOK (Hold Your Own Key) on a workspace. Destroying the resource leaves HYOK enabled on the workspace (non-destructive). By @danieldnedialkov [#2192](https://github.com/hashicorp/terraform-provider-tfe/pull/2192)

ENHANCEMENTS:
* `r/tfe_oauth_client`, `d/tfe_oauth_client`: Add `ado_org_name` support for Azure DevOps Services connections that use organization-scoped personal access tokens.
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
