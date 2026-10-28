# AI Assistant Context for hashicorp/tfe Terraform Provider

This file provides guidance for AI coding assistants (Claude Code, GitHub Copilot, etc.) when working with code in this repository.

## About this Project

terraform-provider-tfe implements the **official** HCP Terraform / Terraform Enterprise Terraform provider plugin for terraform >= 1.0. It consists of 120+ resources and data sources and is a top 10 non-utility provider, with over 200 million downloads across 125+ releases.

## Provider Notes

#### Upgrades in Flight

The provider is in the midst of two ongoing upgrades: from the Provider SDK to the Provider Framework, with  resources being provided by each in a muxed configuration, as well as of the HCP Terraform SDK, go-tfe, from v1.x to v2.x.

#### Configuration

Many resources allow an optional `organization` attribute, reflecting the tenant slug the resource belongs to. Typically, this attribute is required information, but the provider block itself also supports an `organization` attribute, which is applied in the absence of the resource attribute.

Because config can either be run locally or in the HCP Terraform platform, it can create ambiguous configuration scenarios, for example, the provider can reuse `terraform login` credentials but, when run in the context of the HCP Terraform platform, those same credentials are much more restricted and apply to the run itself.

#### Testing

Automated tests are mostly acceptance level tests, integrated with a real working backend this is pre-configured by the developer. The entire test suite is too time consuming to run, so don't try. To run tests relevant to a single resource, use this command, where "TestAccTFEXXX" is the prefix of the test function(s) you want to run:

```bash
TESTARGS="-run TestAccTFEXXX -v" make testacc
```
