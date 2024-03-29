---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_ids"
description: |-
  Get information on workspace IDs.
---

# Data Source: tfe_workspace_ids

Use this data source to get a map of workspace IDs.

## Example Usage

```java
import software.constructs.*;
import com.hashicorp.cdktf.*;
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import gen.providers.tfe.dataTfeWorkspaceIds.*;
public class MyConvertedCode extends TerraformStack {
    public MyConvertedCode(Construct scope, String name) {
        super(scope, name);
        new DataTfeWorkspaceIds(this, "all", new DataTfeWorkspaceIdsConfig()
                .names(List.of("*"))
                .organization("my-org-name")
                );
        new DataTfeWorkspaceIds(this, "app-frontend", new DataTfeWorkspaceIdsConfig()
                .names(List.of("app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"))
                .organization("my-org-name")
                );
        new DataTfeWorkspaceIds(this, "prod-apps", new DataTfeWorkspaceIdsConfig()
                .organization("my-org-name")
                .tagNames(List.of("prod", "app", "aws"))
                );
        new DataTfeWorkspaceIds(this, "prod-only", new DataTfeWorkspaceIdsConfig()
                .excludeTags(List.of("app"))
                .organization("my-org-name")
                .tagNames(List.of("prod"))
                );
    }
}
```

## Argument Reference

The following arguments are supported. At least one of `names` or `tagNames` must be present. Both can be used together.

* `names` - (Optional) A list of workspace names to search for. Names that don't
  match a valid workspace will be omitted from the results, but are not an error.

    To select _all_ workspaces for an organization, provide a list with a single
    asterisk, like `["*"]`. The asterisk also supports partial matching on prefix and/or suffix, like `[*Prod]`, `[test-*]`, `[*dev*]`.
* `tagNames` - (Optional) A list of tag names to search for.
* `excludeTags` - (Optional) A list of tag names to exclude when searching.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `fullNames` - A map of workspace names and their full names, which look like `<ORGANIZATION>/<WORKSPACE>`.
* `ids` - A map of workspace names and their opaque, immutable IDs, which look like `ws-<RANDOM STRING>`.

<!-- cache-key: cdktf-0.17.0-pre.15 input-a50ddfd1d990de8d1cbdba1a7182f9b5d086fbc397439bdd1d0bd057263938e3 -->