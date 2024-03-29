---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_providers"
description: |-
  Get information on public and private providers in the private registry.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_registry_providers

Use this data source to get information about public and private providers in the private registry.

## Example Usage

All providers:

```typescript
import * as constructs from "constructs";
import * as cdktf from "cdktf";
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import * as tfe from "./.gen/providers/tfe";
class MyConvertedCode extends cdktf.TerraformStack {
  constructor(scope: constructs.Construct, name: string) {
    super(scope, name);
    new tfe.dataTfeRegistryProviders.DataTfeRegistryProviders(this, "all", {
      organization: "my-org-name",
    });
  }
}

```

All private providers:

```typescript
import * as constructs from "constructs";
import * as cdktf from "cdktf";
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import * as tfe from "./.gen/providers/tfe";
class MyConvertedCode extends cdktf.TerraformStack {
  constructor(scope: constructs.Construct, name: string) {
    super(scope, name);
    new tfe.dataTfeRegistryProviders.DataTfeRegistryProviders(this, "private", {
      organization: "my-org-name",
      registryName: "private",
    });
  }
}

```

Providers with "hashicorp" in their namespace or name:

```typescript
import * as constructs from "constructs";
import * as cdktf from "cdktf";
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import * as tfe from "./.gen/providers/tfe";
class MyConvertedCode extends cdktf.TerraformStack {
  constructor(scope: constructs.Construct, name: string) {
    super(scope, name);
    new tfe.dataTfeRegistryProviders.DataTfeRegistryProviders(
      this,
      "hashicorp",
      {
        organization: "my-org-name",
        search: "hashicorp",
      }
    );
  }
}

```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registryName` - (Optional) Whether to list only public or private providers. Must be either `public` or `private`.
* `search` - (Optional) A query string to do a fuzzy search on provider name and namespace.

## Attributes Reference

* `providers` - List of the providers. Each element contains the following attributes:
  * `id` - ID of the provider.
  * `organization` - Name of the organization.
  * `registryName` - Whether this is a publicly maintained provider or private.
  * `namespace` - Namespace of the provider.
  * `name` -  Name of the provider.
  * `createdAt` - Time when the provider was created.
  * `updatedAt` - Time when the provider was last updated.

<!-- cache-key: cdktf-0.17.0-pre.15 input-d5c2827100f6bd66c3891a0b03d513fbe1455639407e4d6cdc0ea28851e10d78 -->