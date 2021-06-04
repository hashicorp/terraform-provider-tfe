---
name: Bug report
about: Let us know about an unexpected error, a crash, or an incorrect behavior.
labels: bug

---

<!--
Hi there,

Thank you for opening an issue! Please note that we try to keep the this issue tracker reserved for
bug reports and feature requests related to the Terraform Cloud/Enterprise provider. If you know
your issue relates to the Terraform Cloud/Enterprise platform itself, please contact
tf-cloud@hashicorp.support. For general usage questions, please post to our community forum:
https://discuss.hashicorp.com.
-->

#### Terraform Cloud/Enterprise version

<!---
Are you using Terraform Cloud or Terraform Enterprise? If Terraform Enterprise, please include the version
Example: Terraform Enterprise v202104-1
-->


#### Terraform version
<!---
If you are using Terraform CLI, run `terraform version` to show the version.

If you are using Terraform Cloud or Terraform Enterprise, see the Terraform version being used by a
workspace in the workspace's Overview tab.

Paste the result between the ``` marks below.
-->

```plaintext
...
```

#### Terraform Configuration Files
<!--
Paste the relevant parts of your Terraform configuration between the ``` marks below.
-->

```terraform
...
```

#### Debug Output
<!--
Full debug output can be obtained by running Terraform with the environment variable `TF_LOG=trace`. Please create a GitHub Gist containing the debug output. Please do _not_ paste the debug output in the issue, since debug output is long.

Debug output may contain sensitive information. Please review it before posting publicly, and if you are concerned feel free to encrypt the files using the HashiCorp security public key.
-->

```plaintext
...
```

#### Expected Behavior
<!--
What should have happened?
-->

#### Actual Behavior
<!--
What actually happened?
-->

#### Additional Context
<!--
Are there anything atypical about your situation that we should know? For example: is Terraform running in a wrapper script or in a CI system? Are you passing any unusual command line options or environment variables to opt-in to non-default behavior?
-->
