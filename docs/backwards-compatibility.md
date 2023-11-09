# Maintaining Backwards Compatibility (Terraform Enterprise)

## Terraform Enterprise Version Checking

**Note:** If you are introducing a new resource that is only available at a certain TFE release, you do not need to perform any sort of checks.

### Implied Check (Recommended Approach)

The simplest solution when a particular attribute is not supported by a given TFE release is a `nil` check. These checks are TFE release agnostic and assume that if a field was not returned by the API, the TFE release does not support it. They are particularly important for fields that are structs, where attempting to dereference a nil pointer causes a panic.

You can use this implied checks when:

- The field is a pointer
- When a `null` value is ignored by the API or by go-tfe (see if the struct tag has `omitempty`)

**Example:**

```go
if tmAccess.ProjectAccess != nil {
    projectAccess := []map[string]interface{}{{
    		"settings": tmAccess.ProjectAccess.ProjectSettingsPermission,
    		"teams":    tmAccess.ProjectAccess.ProjectTeamsPermission,
    	}}
    // Write project access to state
    err := d.Set("project_access", projectAccess)
}
```

### Explicit Enterprise Checks

If a resource or attribute is **only** available in Terraform Enterprise, use the go-tfe helper [IsEnterprise()](https://pkg.go.dev/github.com/hashicorp/go-tfe#Client.IsEnterprise) to ensure the client is configured against a TFE instance. This check is derived from the `TFP-AppName` header that Terraform Cloud emits, of which if not present, indicates a Terraform Enterprise installation.

```go
config := meta.(ConfiguredClient)

if config.Client.IsEnterprise() {
    // do something with TFE only behavior
}
```

### Documentation

It is important to communicate with practitioners which resources and fields are supported for a particular TFE release.

For a new resource, add the minimum release required to the top level documentation.

**Example:**

```md
# my_new_resource

Provides a my new resource.

~> **NOTE:** Using this resource requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202302-1.
```


If an attribute has a TFE release constraint, add a second sentence to the attribute's description:

```md
## Argument Reference

The following arguments are supported:

* `foo` - (Required) Foo is bar.
* `bar` - (Optional) Bar is foo.
* `foobar` - (Optional) Foobar is barfoo. This attribute requires Terraform Cloud or an instance of Terraform Enterprise at least as recent as `v202302-1`.
```
