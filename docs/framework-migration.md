# Migrating Resources from SDKv2 to the Plugin Framework

The provider still serves some resources and data sources with
[terraform-plugin-sdk/v2](https://developer.hashicorp.com/terraform/plugin/sdkv2).
We are moving them to the
[Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
one at a time. This page is the process for migrating one of them. It is
written to be followed step by step, by a person or by a coding agent.

The goal of every migration is that **nobody using the provider can tell it
happened**: the same configuration is valid, plans for existing state are
empty, state values are unchanged, import IDs and resource identity still
work.

## How the provider is put together

- `main.go` serves the provider over plugin protocol v6. The SDKv2 provider
  (`internal/provider/provider.go`) is upgraded from protocol v5 with
  `tf5to6server` and combined with the framework provider
  (`internal/provider/provider_next.go`) by `tf6muxserver`.
- A resource lives in exactly one of them. Migrating moves it from
  `ResourcesMap` / `DataSourcesMap` in `provider.go` to `Resources()` /
  `DataSources()` in `provider_next.go`.
- Acceptance tests use `testAccMuxedProviders`
  (`internal/provider/provider_test.go`), which serves both, so existing tests
  run against whichever implementation is registered.

## What is left

`internal/provider/sdkv2_inventory_test.go` lists every resource and data
source still on SDKv2. `TestSDKv2Inventory` fails if one is added to the SDKv2
provider, and fails if one is removed without removing it from the list.

## Rules

These come from HashiCorp's
[migration guide](https://developer.hashicorp.com/terraform/plugin/framework/migrating)
and this repository's history.

1. **Configuration must not change.** Breaking changes are only allowed in a
   major release. If a faithful migration is not possible, stop and raise it
   with the team instead of changing behavior.
2. **Tests first.** Make sure the SDKv2 implementation is well tested, and
   write the migration tests before writing framework code.
3. **Nested blocks stay blocks.** Use `ListNestedBlock` / `SetNestedBlock`,
   never nested attributes. Nested attributes change the HCL syntax users
   write (`foo { }` becomes `foo = { }`).
4. **Keep import IDs, resource identity, and the schema version.**
5. **Don't change the API client in the same PR.** If the SDKv2 code uses
   go-tfe v1 (`config.Client`), the framework code does too; moving to
   go-tfe v2 (`config.ClientV2`) is a separate change.
6. **One resource or data source per PR.** Keep the file name.

## Before your first migration

Read these once:

- [Migrating overview](https://developer.hashicorp.com/terraform/plugin/framework/migrating),
  [Benefits](https://developer.hashicorp.com/terraform/plugin/framework/migrating/benefits)
  (feature comparison), and
  [Testing](https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing)
- [Plan modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
  — the order in which defaults, unknown values, and modifiers are applied
  explains most unexpected plans.
- [Types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types)
  — null and unknown values, which SDKv2 did not have.

And look at these in this repository:

| Example | What it shows |
|---|---|
| `resource_tfe_variable.go`, `TestAccTFEVariable_rewrite` | A completed SDKv2 migration with a state upgrader, resource identity, and import by legacy ID or identity |
| commit `aa0e599b` (tfe_project) | A full migration PR |
| `resource_tfe_registry_provider.go` | Provider default organization with `ModifyPlan` and `modifyPlanForDefaultOrganizationChange` |
| `resource_tfe_ssh_key.go` | A small framework resource; write-only attributes |
| `identity_upgrade_test.go` | Unit tests for identity upgraders |

Look up other topics as you need them:
[import](https://developer.hashicorp.com/terraform/plugin/framework/resources/import),
[identity](https://developer.hashicorp.com/terraform/plugin/framework/resources/identity),
[state upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade),
[defaults](https://developer.hashicorp.com/terraform/plugin/framework/resources/default),
[validation](https://developer.hashicorp.com/terraform/plugin/framework/validation),
[timeouts](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts),
[write-only arguments](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments),
[custom types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/custom-types).
Older versions of the migration guide had side-by-side SDKv2 / framework
examples for each topic, e.g.
[v1.15.x](https://developer.hashicorp.com/terraform/plugin/framework/v1.15.x/migrating).

## Process

Each step ends with something to record in the pull request description.
Run only the tests for the resource you are migrating; the full suite is
slow.

```sh
TF_ACC=1 go test ./internal/provider -run 'TestAccTFERunTrigger' -v -timeout 60m
```

### Step 1: Gather facts about the resource

Start the PR description with what the SDKv2 resource does. For
`tfe_run_trigger` in `resource_tfe_run_trigger.go`:

```sh
f=internal/provider/resource_tfe_run_trigger.go
grep -n 'SchemaVersion\|StateUpgraders' $f                # state upgrades
grep -n 'Identity:' $f                                     # resource identity
grep -n 'CustomizeDiff' $f                                 # plan customization
grep -n 'Importer' -A3 $f                                  # import
grep -n 'Timeouts' $f                                      # timeouts
grep -n 'organization' $f                                  # provider default organization?
grep -n 'ForceNew\|Default:\|DiffSuppressFunc\|StateFunc\|ValidateFunc\|ConflictsWith\|ExactlyOneOf\|MaxItems\|Sensitive\|Deprecated' $f
grep -c '^func TestAcc' ${f%.go}_test.go                   # existing tests
```

Then describe, attribute by attribute: type, required / optional / computed,
default, forces replacement, sensitive, deprecated, validation, diff
suppression, normalization. For the resource: every accepted import ID
format, what Read does when the object no longer exists, what
`CustomizeDiff` does, and any state upgraders.

**Record:** the facts and the attribute table.

### Step 2: Fill gaps in the existing tests

With the SDKv2 code still in place, add acceptance tests for anything the
existing tests don't cover:

- every optional attribute both set and unset
- updates that happen in place and updates that replace the resource
- import with each accepted ID format
- the object being deleted outside Terraform
- `organization` taken from the provider configuration, if supported

**Record:** the targeted test run passing against the SDKv2 implementation.

### Step 3: Write the migration tests

`testAccFrameworkMigrationSteps`
(`internal/provider/framework_migration_helpers_test.go`) returns two steps:

1. Apply the configuration with a released provider version, where the
   resource is still SDKv2, and record the state of every resource of the
   given type.
2. Apply the same configuration with the code in this repository. The plan
   must be empty, and the state of those resources (attribute values, schema
   version, identity) must equal step 1's. The comparison uses
   `terraform show -json`, so `""` and `null` are different.

```go
func TestAccTFERunTrigger_frameworkMigration(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps:    testAccFrameworkMigrationSteps(t, "tfe_run_trigger", "0.82.0", testAccTFERunTrigger_basic(rInt)),
	})
}
```

The version is written directly in each test. Use the latest release at the
time of the migration, which is the last release where the resource was
SDKv2, and never change it afterward; otherwise the test stops proving
anything.

Write one test per meaningfully different configuration:

- only required arguments
- every argument set, nested blocks populated
- optional arguments with defaults left out
- empty collections versus omitted collections
- `organization` from the provider configuration, if supported
- each state version, if the resource has state upgraders: add a test pinned
  to an older release that wrote that version, with a configuration valid
  for that release

`withIgnoredAttributes` leaves attributes out of the state comparison. Use it
only for values that change between two reads of the same object, or for
attributes added after an older pinned release. Comment every use.

Append more steps where useful: an update after the migration, or an import.

**Record:** the new tests passing (they pass trivially before Step 4).

### Step 4: Implement

1. Replace the contents of the resource file with a framework resource. Use
   the [mapping](#sdkv2-to-framework-mapping) below and the examples above.
   Typical structure: `New<Name>Resource()`, a resource struct holding
   `config ConfiguredClient`, a model struct with `tfsdk` tags, then
   `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`,
   `ImportState`, and as needed `ModifyPlan`, `ConfigValidators`,
   `UpgradeState`, `IdentitySchema`, `UpgradeIdentity`.
2. Add a `var _ resource.ResourceWith... = &resourceTFE<Name>{}` assertion
   for every interface the resource implements.
3. Remove the resource from `ResourcesMap` in `provider.go`, register it in
   `provider_next.go`, and remove it from the list in
   `sdkv2_inventory_test.go`.
4. Unit-test logic that doesn't need the API: mapping API objects to the
   model, `ModifyPlan` decisions, state and identity upgraders.
5. Remove SDKv2 helpers that nothing else uses anymore.

```sh
go build ./... && make fmt && make lint
go test ./internal/provider -run 'TestSDKv2Inventory|<your unit tests>' -count=1
```

**Record:** build, lint, and unit tests passing.

### Step 5: Verify

1. **Acceptance tests.** Run the targeted tests. All existing tests and the
   migration tests must pass.
2. **Schema comparison.**

    ```sh
    scripts/schema-diff.sh tfe_run_trigger
    ```

    This compares the schema Terraform sees at the commit your branch started
    from with your working tree: attribute types and flags, descriptions,
    blocks, schema version, and resource identity. It must report no
    unexpected differences. The one known framework difference is that
    blocks no longer report `max_items` / `min_items` (a size validator
    enforces the limit instead); allow it explicitly:

    ```sh
    scripts/schema-diff.sh --allow 'resource.tfe_team.block.organization_access.max_items=1->0' tfe_team
    ```

    The schema doesn't include replacement behavior, defaults, validation,
    or diff suppression. The migration tests and review cover those.
3. **Manual check.** In a scratch directory, apply a configuration with the
   released provider, then switch to your build (`eval $(make devoverride)`,
   see [development environment](development-environment.md)), and run
   `terraform plan` (expect no changes), `terraform import` into a fresh
   state (expect no changes after), and `terraform destroy`.
4. **Docs.** Regenerate the docs and confirm nothing changed unexpectedly.

**Record:** test output, `schema-diff.sh` output, and what you checked
manually.

### Step 6: Open the pull request

- Add a `NOTES` changelog entry with `npx changie new`, e.g. "`r/tfe_run_trigger`:
  migrated to the plugin framework. No changes in behavior are expected."
  ([changelog process](changelog-process.md))
- Fill in the PR description from the steps above.
- Ask for review against the [review checklist](#review-checklist).

## SDKv2 to framework mapping

### Schema

| SDKv2 | Framework | Watch out for |
|---|---|---|
| `TypeString`, `TypeBool`, `TypeInt`, `TypeFloat` | `StringAttribute`, `BoolAttribute`, `Int64Attribute`, `Float64Attribute` | |
| `TypeList` / `TypeSet` / `TypeMap` of primitives | `ListAttribute` / `SetAttribute` / `MapAttribute` with `ElementType` | |
| `TypeList` / `TypeSet` with `Elem: &schema.Resource{}` | `ListNestedBlock` / `SetNestedBlock` under `Blocks` | Must stay a block |
| `MaxItems` / `MinItems` on a block | `listvalidator.SizeAtMost` / `SizeAtLeast` | Shows in `schema-diff.sh` as `max_items 1 -> 0`; allow it |
| `Default` | `Default: stringdefault.StaticString(...)` etc. | The attribute must also be `Computed` |
| `ForceNew` | `RequiresReplace()` plan modifier | For optional + computed, `RequiresReplaceIfConfigured()` may match SDKv2 better |
| Optional + Computed | Optional + Computed, plus `UseStateForUnknown()` if the value doesn't change on update | Without it, every update shows `(known after apply)` |
| `ValidateFunc` | `Validators` (terraform-plugin-framework-validators) | Error messages may change; check tests that match them |
| `ConflictsWith`, `ExactlyOneOf`, `AtLeastOneOf`, `RequiredWith` | Attribute validators or `ConfigValidators()` | |
| `DiffSuppressFunc` | Custom type with semantic equality, or a plan modifier | No direct equivalent; test the suppressed case |
| `StateFunc` | Normalize before writing state, or a custom type | State must hold the same normalized value |
| `Sensitive`, `Deprecated` | `Sensitive`, `DeprecationMessage` | |
| `Description` | `Description` | Keep the text identical; docs are generated from it |
| `Timeouts` | `terraform-plugin-framework-timeouts`, as a block | |
| `Identity` | `IdentitySchema`, `UpgradeIdentity` | Names, types, import flags, and version must match |
| `SchemaVersion` + `StateUpgraders` | `Version` + `UpgradeState` | See below |

### Behavior

| SDKv2 | Framework |
|---|---|
| `meta.(ConfiguredClient)` | Save `req.ProviderData.(ConfiguredClient)` in `Configure` (it can be nil) |
| `d.Get`, `d.GetOk` | Read the plan or state into the model; check `IsNull()` / `IsUnknown()` |
| `d.HasChange` | Compare plan and state values in `Update` |
| `d.Set`, `d.SetId` | Fill the model and `resp.State.Set` |
| `d.SetId("")` in Read | `resp.State.RemoveResource(ctx)` |
| Returning an error | `resp.Diagnostics.AddError(summary, detail)`, then return |
| `CustomizeDiff` | `ModifyPlan`; handle create (`req.State.Raw.IsNull()`) and destroy (`req.Plan.Raw.IsNull()`) |
| `customdiff.ForceNewIf` | `RequiresReplaceIf`, or `resp.RequiresReplace` in `ModifyPlan` |
| `SetNewComputed` | Set the attribute to unknown in `resp.Plan` |
| Custom `Importer` | `ImportState`; accept every ID format the SDKv2 version accepted |

### Null and zero values

SDKv2 has no null for simple types: an unset optional string is stored as
`""`, a bool as `false`, a number as `0`. The framework stores null. If Read
now writes null where SDKv2 wrote `""`, the migration test's state comparison
fails, and users' expressions that use the value can break. Make Read write
what SDKv2 wrote for the same API response. The state comparison failure
tells you which attribute to look at.

### State upgrades

- Set `Version` to the SDKv2 `SchemaVersion`.
- Add an upgrader for **every** older version. Each one must produce the
  current version directly; the framework doesn't chain them.
- Describe the old schema with `PriorSchema` and read it into a model, as in
  `resource_tfe_variable.go`. Port the logic from the existing SDKv2
  upgraders (`resource_tfe_workspace_migrate.go`,
  `resource_tfe_team_access_migrate.go`).
- Write the whole new state; nothing is copied for you, and unknown values
  are errors.

### Data sources

The process is the same, with `datasource.DataSource` and `DataSources()` in
`provider_next.go`. Data sources have no prior state, so for the migration
test, pass the data source's value through a managed resource and pass
`"terraform_data"` as the resource type:

```hcl
data "tfe_slug" "test" { ... }
resource "terraform_data" "test" { input = data.tfe_slug.test }
```

Computed-only nested lists may become `ListNestedAttribute{Computed: true}`,
since users never write them in configuration; anything configurable stays a
block.

## When a migration test shows a difference

| Symptom | Likely cause |
|---|---|
| Plan shows `(known after apply)` on update | Optional + computed attribute missing `UseStateForUnknown()` |
| State comparison shows `"" -> <nil>`, `false -> <nil>`, `[] -> <nil>` | Read writes null where SDKv2 wrote a zero value |
| Plan shows a change SDKv2 suppressed | `DiffSuppressFunc` or `StateFunc` not reproduced |
| Plan modifier inside a list or set block sees the wrong prior value | The framework matches list/set elements to prior state by position, not by content; removing or reordering elements shifts them |
| Plan wants to replace the resource | `RequiresReplace` on an attribute that wasn't `ForceNew`, or a default that differs |
| "Provider produced inconsistent result" | `Create` / `Update` wrote a value different from the plan |
| Identity differs | Identity schema or the values written in Read/Create changed |

## Review checklist

- [ ] Facts and attribute table in the PR description match the SDKv2 code.
- [ ] Every SDKv2 attribute exists with the same type and flags; blocks are
      still blocks.
- [ ] Defaults, replacement behavior, sensitive, and deprecation match.
- [ ] Optional + computed attributes use `UseStateForUnknown()` only where
      the value doesn't change on update.
- [ ] Read writes the same values as SDKv2 for the same API response.
- [ ] Update writes state equal to the plan.
- [ ] `CustomizeDiff`, diff suppression, and normalization are reproduced
      and tested.
- [ ] Read removes the resource from state in the same cases as SDKv2.
- [ ] Every import ID format still works; identity is unchanged.
- [ ] Every older state version has an upgrader.
- [ ] Migration tests cover the configurations in Step 3; every
      `withIgnoredAttributes` is justified.
- [ ] No test was loosened to make it pass.
- [ ] `schema-diff.sh` output is clean; every `--allow` is a block size
      limit replaced by a validator.
- [ ] Removed from `provider.go` and `sdkv2_inventory_test.go`, registered in
      `provider_next.go`.
- [ ] The API client (go-tfe v1 / v2) didn't change.
- [ ] Changelog entry; docs regenerated without unexpected changes.
