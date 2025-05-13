## Description

_Describe why you're making this change._

_Remember to:_

- [ ] _Update the [Change Log](https://github.com/hashicorp/terraform-provider-tfe/blob/main/docs/changelog-process.md)_
- [ ] _Update the [Documentation](https://github.com/hashicorp/terraform-provider-tfe/blob/main/docs/changelog-process.md#updating-the-documentation)_

## Testing plan

1.  _Describe how to replicate_
1.  _the conditions_
1.  _under which your code performs its purpose,_
1.  _including example Terraform configs where necessary._

## External links

_Include any links here that might be helpful for people reviewing your PR. If there are none, feel free to delete this section._

- [go-tfe documentation](https://pkg.go.dev/github.com/hashicorp/go-tfe?tab=doc#xxxx)
- [API documentation](https://developer.hashicorp.com/terraform/cloud-docs/xxxx)
- [Related PR](https://github.com/hashicorp/terraform-provider-tfe/pull/xxxx)

## Output from acceptance tests

_Please run applicable acceptance tests locally and include the output here. See [testing.md](https://github.com/hashicorp/terraform-provider-tfe/blob/main/docs/testing.md) to learn how to run acceptance tests._

_If you are an external contributor, your contribution(s) will first be reviewed before running them against the project's CI pipeline._

```
$ TESTARGS="-run TestAccTFEWorkspace" make testacc

...
```

## Rollback Plan

<!--
Please outline a plan in the event changes need to be rolled back

Example: If a change needs to be reverted, we will roll out an update to the code within 7 days.
-->

## Changes to Security Controls

<!--
Are there any changes to security controls (access controls, encryption, logging) in this pull request? If so, explain.
-->
