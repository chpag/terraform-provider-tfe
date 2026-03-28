---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_smtp_settings"
description: |-
  Get information about SMTP settings.
---

# Data Source: tfe_smtp_settings

Use this data source to get information about SMTP settings in Terraform Enterprise.

~> **NOTE:** This data source is only available in Terraform Enterprise and requires admin privileges.

## Example Usage

```hcl
data "tfe_smtp_settings" "current" {}

output "smtp_enabled" {
  value = data.tfe_smtp_settings.current.enabled
}

output "smtp_host" {
  value = data.tfe_smtp_settings.current.host
}

output "smtp_sender" {
  value = data.tfe_smtp_settings.current.sender
}
```

### Using with Conditional Logic

```hcl
data "tfe_smtp_settings" "current" {}

resource "tfe_notification_configuration" "example" {
  # Only create notification if SMTP is enabled
  count = data.tfe_smtp_settings.current.enabled ? 1 : 0

  name             = "smtp-notification"
  destination_type = "email"
  email_addresses  = ["admin@example.com"]
  workspace_id     = tfe_workspace.example.id
}
```

## Argument Reference

This data source does not require any arguments.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the SMTP settings. Always `smtp`.
* `enabled` - Whether SMTP is enabled.
* `host` - The hostname of the SMTP server.
* `port` - The port of the SMTP server.
* `sender` - The email address used as the sender for outgoing emails.
* `auth` - The authentication type. Possible values are `none`, `plain`, and `login`.
* `username` - The username used to authenticate to the SMTP server.

~> **NOTE:** The `password` field is not returned by this data source for security reasons.