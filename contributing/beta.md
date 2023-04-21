# Introducing Beta Features

This guide discusses how to introduce features that are not yet generally available in Terraform Cloud.

In general, beta features should not be merged/released until generally available (GA). However, the maintainers recognize almost any reason to release beta features on a case-by-case basis. These could include: partial customer availability, software dependency, or any reason short of feature completeness.

If planning to release a limited beta feature, each resource should be clearly noted as such in the website documentation and CHANGELOG.

```markdown
~> **NOTE:** This resource is currently in beta and isn't generally
available to all users. It is subject to change or be removed.
```

When adding test cases, understand that feature flags are not evaluated in our [automated test infrastructure](test-infrastructure.md). Features that are behind a feature flag will probably fail. You can temporarily use the `skipUnlessBeta` test helper to omit beta features from running in CI.

```go
func TestAccTFEMyNewResource_basic(t *testing.T) {
  skipUnlessBeta(t)
}
```

When the feature reaches general availability _and_ the feature flag is removed, you should create a new PR to remove the `skipUnlessBeta` flags, beta notes, and re-announce the feature in the CHANGELOG as being generally available.
