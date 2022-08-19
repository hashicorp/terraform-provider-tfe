## What is terraform-plugin-docs (tfplugindocs)[https://github.com/hashicorp/terraform-plugin-docs]?
- CLI tool to generate and validate plugin documentation for the Registry website
- Currently used by only a few internal providers (one's I could find listed below)
- Questions in Slack Channel: #proj-tf-plugin-framework
- My [lightning talk](https://hashicorp.zoom.us/rec/play/Ev3yCVccPL1zaT5DihkgzjAz3hwsVPz4onfJOXdjy9zvUr3F96nOuT2AmBDczFIqkzEqUAG810kiqIxd.oyhE-vXIhZ5CrnSV?startTime=1659714349000) (3:00-10:00) explains what tfplugindocs is and the purpose of this project.

## What does tfplugindocs do? 
- Generates markdown in a "docs" directory based on customizable templates (in the "templates" directory)
- Parses and lists resource attributes and descriptions from schema in the "tfe" directory
- Pulls example code snippets from .tf files in the "examples" directory 

See [PR #595](https://github.com/hashicorp/terraform-provider-tfe/pull/595) for a summary of changes

See **info-project-docsparkle/1b-in-depth-tfplugindocs.md** for more details on how the tool functions and things to watch out for

## Providers Using tfplugindocs 
*internal*
basic: 
- [terraform-provider-random](https://github.com/hashicorp/terraform-provider-random)
- [terraform-provider-null](https://github.com/hashicorp/terraform-provider-null)

advanced: 
- [terraform-provider-tls](https://github.com/hashicorp/terraform-provider-tls)
- [terraform-provider-awscc](https://github.com/hashicorp/terraform-provider-awscc)

*external*
- [terraform-provider-fastly](https://github.com/hashicorp/terraform-provider-tls)

## Useful Links
[terraform-plugin-docs Github](https://github.com/hashicorp/terraform-plugin-docs)

[slack thread w/ engineer from tls provider](https://hashicorp.slack.com/archives/CTQBQ9G0Y/p1657141043514899)