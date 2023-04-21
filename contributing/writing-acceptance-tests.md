# Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimizes the amount of boilerplate code necessary to use common testing patterns. This guide is meant to augment the general SDKv2 documentation with Terraform AWS Provider specific conventions and helpers.

## Resource Acceptance Testing

Most resources that implement standard Create, Read, Update, and Delete functionality should follow the pattern below. Each test type has a section that describes them in more detail:

    basic: This represents the bare minimum verification that the resource can be created, read, deleted, and optionally imported.
    Per Attribute: A test that verifies the resource with a single additional argument can be created, read, optionally updated (or force resource recreation), deleted, and optionally imported.

## Smoke Testing Tips

After creating new schema, it's important to test your changes beyond the automated testing provided by the framework. Use these tips to ensure your provider resources behave as expected.

- Is the resource replaced when non-updatable attributes are changed?
- Is the resource unchanged after successive plans with no config changes?
- What happens when you configure export-only computed arguments?
- Are mutually exclusive config arguments constrained by an error?
- If adding a new argument to an existing resource: is it required? (This would be a breaking change)
- If adding a new attribute to an existing resource: is new or unexpected API authorization required?
- Is the new resource argument updated when it is the _only_ change in a plan?
- Does Terraform warn about abnormalities when TF_LOG=debug is used?
