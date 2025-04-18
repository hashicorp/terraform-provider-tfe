---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_token"
description: |-
  Generates an ephemeral agent token.
---


<!-- Please do not edit this file, it is generated. -->
# tfe_agent_token

Generates a new agent token as an ephemeral value. 

Each agent pool can have multiple tokens and they can be long-lived. For that reason, this ephemeral resource does not implement the Close method, which would tear the token down after the configuration is complete. 

Agent token strings are sensitive and only returned on creation, so making those strings ephemeral values is beneficial to avoid state exposure.

If you need to use this value in the future, make sure to capture the token and save it in a secure location. Any resource with write-only values can accept ephemeral resource attributes.

## Example Usage

Basic usage:

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformStack
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) Id for the Agent Pool.
* `description` - (Required) A brief description about the Agent Pool.

## Example Usage

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformOutput, TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.tfe.agent_pool import AgentPool
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        TerraformOutput(self, "my-agent-token",
            value=tfe_agent_token.this.token,
            description="Token for tfe agent."
        )
        AgentPool(self, "foobar",
            name="agent-pool-test",
            organization="my-org-name"
        )
```

## Attributes Reference

* `token` - The generated token.


<!-- cache-key: cdktf-0.20.8 input-081987bf1a7aae73eb98d94a8e172d4302e45db3a244040a070812672515dcdd -->