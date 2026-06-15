---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_verify_registry_provider_version"
description: |-
  Verifies if a provider version configuration in the private registry is complete.
---

# Action: tfe_verify_registry_provider_version

Verifies that a provider version in the private registry has all required components properly configured. This action checks for the presence of GPG keys, protocols, SHA256SUMS files, signatures, and platform binaries to ensure the provider version is complete and ready for use.

## Example Usage

### Basic Verification

```terraform
resource "tfe_registry_provider" "example" {
  organization = "my-organization"
  name         = "example-provider"
}

resource "tfe_registry_provider_version" "example" {
  organization = tfe_registry_provider.example.organization
  registry_name = tfe_registry_provider.example.registry_name
  namespace = tfe_registry_provider.example.namespace
  provider_name = tfe_registry_provider.example.name
  
  version              = "1.0.0"
  key_id               = tfe_registry_gpg_key.example.id
  protocols            = ["5.0"]
  shasums_filename     = "./terraform-provider-example_1.0.0_SHA256SUMS"
  shasums_sig_filename = "./terraform-provider-example_1.0.0_SHA256SUMS.sig"

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.tfe_verify_registry_provider_version.verify]
    }
  }
}

action "tfe_verify_registry_provider_version" "verify" {
  config {
    organization  = "my-organization"
    provider_name = "example-provider"
    version       = "1.0.0"
  }
}
```

### Verification with Required Platforms

```terraform
resource "tfe_registry_provider_version" "example" {
  organization = "my-organization"
  registry_name = "private"
  namespace = "my-organization"
  provider_name = "example-provider"
  
  version              = "1.0.0"
  key_id               = tfe_registry_gpg_key.example.id
  protocols            = ["5.0"]
  shasums_filename     = "./terraform-provider-example_1.0.0_SHA256SUMS"
  shasums_sig_filename = "./terraform-provider-example_1.0.0_SHA256SUMS.sig"

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.tfe_verify_registry_provider_version.verify]
    }
  }
}

action "tfe_verify_registry_provider_version" "verify" {
  config {
    organization  = "my-organization"
    provider_name = "example-provider"
    version       = "1.0.0"
    required_platforms = [
      "linux_amd64",
      "darwin_amd64",
      "darwin_arm64",
      "windows_amd64"
    ]
  }
}
```

### Verification for Public Registry Provider

```terraform
action "tfe_verify_registry_provider_version" "verify_public" {
  config {
    organization  = "my-organization"
    registry_name = "public"
    namespace     = "hashicorp"
    provider_name = "aws"
    version       = "5.0.0"
  }
}
```

### Invoking the action directly

```sh
terraform apply -invoke=action.tfe_verify_registry_provider_version.verify
```

## Argument Reference

This action supports the following arguments within the config block:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.
* `namespace` - (Optional) The namespace of the provider. For private providers this is the same as the organization. Required when `registry_name` is `public`.
* `provider_name` - (Required) Name of the provider to verify.
* `version` - (Required) The version of the provider to verify (e.g., "1.0.0").
* `required_platforms` - (Optional) A list of required platform identifiers (e.g., ["linux_amd64", "darwin_amd64"]). If not specified, the action only checks that at least one platform exists. Platform identifiers can use either underscore format (linux_amd64) or slash format (linux/amd64).

## Verification Checks

The action performs the following verification checks:

1. **Provider Existence** - Verifies the provider exists in the registry
2. **Version Existence** - Confirms the specified version exists
3. **GPG Key Configuration** - Checks that a GPG key ID is configured
4. **GPG Key Availability** - Verifies the GPG key actually exists in the registry
5. **Protocols** - Verifies that at least one protocol version is defined
6. **SHA256SUMS** - Confirms the SHA256SUMS file has been uploaded
7. **SHA256SUMS Signature** - Verifies the SHA256SUMS.sig file has been uploaded
8. **Platform Binaries** - Checks that at least one platform binary exists
9. **Required Platforms** - If specified, verifies all required platforms are present
10. **Binary Upload Status** - Confirms all platforms have their binaries uploaded

## Behavior

The action will:
- Send progress updates during the verification process
- Report all issues found in a single error message if the configuration is incomplete
- Complete successfully if all checks pass

If any verification check fails, the action will fail with a detailed error message listing all issues found.

## Common Use Cases

1. **Post-Upload Validation** - Verify a provider version is complete after uploading all components
2. **CI/CD Integration** - Ensure provider versions are properly configured before marking them as ready
3. **Compliance Checks** - Validate that all required platforms are available for a provider version
4. **Troubleshooting** - Quickly identify missing components in a provider version configuration