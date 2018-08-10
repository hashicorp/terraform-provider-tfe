---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_token"
sidebar_current: "docs-tfe-datasource-oauth-token"
description: |-
  Use this datasource to get the ID of the OAuth token of an assocaited VCS configuration.
---

# tfe_oauth_token

Use this datasource to get the ID of the OAuth token of an assocaited VCS
configuration. This token is used to identify which VCS connection to use.

### Example Usage

```hcl
data "tfe_oauth_token" "oauth_token" {
  template_filter = "featured"
}
```

### Argument Reference

* `template_filter` - (Required) The template filter. Possible values are `featured`, `self`, `selfexecutable`, `sharedexecutable`, `executable` and `community` (see the Cloudstack API *listTemplate* command documentation).

* `filter` - (Required) One or more name/value pairs to filter off of. You can apply filters on any exported attributes.

## Attributes Reference

The following attributes are exported:

* `id` - The template ID.
* `account` - The account name to which the template belongs.
* `created` - The date this template was created.
* `display_text` - The template display text.
* `format` - The format of the template.
* `hypervisor` - The hypervisor on which the templates runs.
* `name` - The template name.
* `size` - The size of the template.
* `tags` - The tags associated with this template.

