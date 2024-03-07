# Creating New Resources

As we work to migrate older resources from the provider SDK v2 to the [Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework), the hashicorp/tfe provider uses plugin protocol v5 along with three separate provider servers [muxed together](https://github.com/hashicorp/terraform-provider-tfe/blob/20448c7293b2e116b633eef4bc73881b060aa32e/main.go#L51).

For all new resources, we ask that you add them to the [provider_next](https://github.com/hashicorp/terraform-provider-tfe/blob/20448c7293b2e116b633eef4bc73881b060aa32e/internal/provider/provider_next.go) Framework Provider to ensure that they don't need to be migrated in the future. For Hashicorp employees: It can be helpful to include a section in a relevant RFC defining your proposed new resource schema. Be sure to notify #team-tf-cli.

There are a few conventions to observe when authoring new resources:

1. Provider default organization: `organization` should typically be an optional argument and is allowed to be configured at the provider block. [Implement](https://github.com/hashicorp/terraform-provider-tfe/blob/20448c7293b2e116b633eef4bc73881b060aa32e/internal/provider/resource_tfe_registry_provider.go#L191-L196) dataOrDefaultOrganization to help resolve the resource organization. In addition, your resource should [implement](https://github.com/hashicorp/terraform-provider-tfe/blob/20448c7293b2e116b633eef4bc73881b060aa32e/internal/provider/resource_tfe_registry_provider.go#L177-L179) framework interface `resource.ResourceWithModifyPlan` in order to detect changes in the provider default organization.

2. Use [resource interfaces](https://github.com/hashicorp/terraform-provider-tfe/blob/20448c7293b2e116b633eef4bc73881b060aa32e/internal/provider/resource_tfe_registry_provider.go#L25-L29) to ensure your new resource implements all necessary behaviors.

3. Make ImportState arguments convenient and using the fewest arguments possible.
