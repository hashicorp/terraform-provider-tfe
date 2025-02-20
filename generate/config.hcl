ephemeral_resource "organization_token" {
  upcase_name = "OrganizationToken"
  description = "An organization token is a unique identifier that can be used to authenticate and authorize requests to the organization's API. Organization tokens are generated and managed by the organization's administrators."

  field "organization" {
      upcase_name = "Organization"
      description = "Name of the organization. If omitted, organization must be defined in the provider config."
      type =       "String"
      suppress_test_check = true
  }

  field "token" {
      upcase_name = "Token"
      description = "The generated token."
      type =       "String"
      model_attr =  "Token"
      computed =    true
  }

  field "force_generate" {
      upcase_name = "ForceRegenerate"
      description = "If set to true, a new token will be generated even if a token already exists. This will invalidate the existing token!"
      type =        "Bool"
      suppress_test_check = true
  }

  field "expired_at" {
      upcase_name = "ExpiredAt"
      description = "The token's expiration date. The expiration date must be a date/time string in RFC3339 format (e.g., \"2024-12-31T23:59:59Z\"). If no expiration date is supplied, the expiration date will default to null and never expire."
      type =        "String"
      model_attr =  "ExpiredAt.String()"
      suppress_test_check = true
  }
}
