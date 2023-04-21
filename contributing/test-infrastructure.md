# Test Infrastructure

We rely on acceptance tests that test the provider against infrastructure that simulates both Terraform Cloud and Terraform Enterprise. The only exception is features that depend on a feature flag evaluation: In this case, we rely on running the tests locally and reporting the results in the PR description.

Within the Pull Request process, test checks are executed against infrastructure that simulates Terraform Cloud, which is rebuilt nightly and uses the latest nightly TFE build image. Changes to the underlying Terraform Cloud platform may take 24-48 hours to be reflected in this infrastructure. Tests that use the helper `skipIfCloud` are skipped during this check.

In addition, we test the main branch of the provider once each night against the same image but configured to simulate Terraform Enterprise. Tests that use the helper `skipIfEnterprise` are skipped during this nightly job.

In both cases, all feature flags are disabled and not evaluated. Features that are behind a feature flag should use the `skipUnlessBeta` flag to avoid failing, even if that feature flag is enabled for all users in production. See our [beta feature policy](beta.md) for more details.
