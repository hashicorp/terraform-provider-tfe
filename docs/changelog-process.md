# Updating the Changelog

The Changelog is for the benefit of users who need to quickly and easily understand what changes will impact them (& how) if they upgrade. The best way to help folks do that is to keep the signal-to-noise ratio high, so we try to keep the Changelog entries limited to user-facing changes.

Only update the `Unreleased` section. Make sure you change the unreleased tag to an appropriate version, using [Semantic Versioning](https://semver.org/) as a guideline.

Please use the template below when updating the changelog:
```
<change category>:
* **New Resource:** `name_of_new_resource` ([#123](link-to-PR))
* r/tfe_resource: description of change or bug fix ([#124](link-to-PR))
```

### Change categories

- BREAKING CHANGES: Use this for any changes that aren't backwards compatible. Include details on how to handle these changes.
- FEATURES: Use this for any large new features added.
- ENHANCEMENTS: Use this for smaller new features added.
- BUG FIXES: Use this for any bugs that were fixed.
- NOTES: Use this section if you need to include any additional notes on things like upgrading, upcoming deprecations, or any other information you might want to highlight.


### Updating the documentation

For pull requests that update provider documentation, please help us verify that the
markdown will display correctly on the Registry:

- Copy the new markdown and paste it here to preview: https://registry.terraform.io/tools/doc-preview
- Paste a screenshot of that preview in your pull request.
