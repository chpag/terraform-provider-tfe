---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_smtp_settings"
description: |-
  Manages SMTP settings for Terraform Enterprise.
---

# tfe_smtp_settings

Manages SMTP settings for Terraform Enterprise. This resource allows you to configure the SMTP server used for sending email notifications.

~> **NOTE:** This resource is only available in Terraform Enterprise and requires admin privileges.

## Example Usage

### Basic Configuration (No Authentication)

```hcl
resource "tfe_smtp_settings" "example" {
  enabled = true
  host    = "smtp.example.com"
  port    = 25
  sender  = "terraform@example.com"
  auth    = "none"
}
```

### With Plain Authentication

```hcl
resource "tfe_smtp_settings" "example" {
  enabled  = true
  host     = "smtp.example.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "plain"
  username = "smtp_user"
  password = var.smtp_password
}
```

### With Login Authentication

```hcl
resource "tfe_smtp_settings" "example" {
  enabled  = true
  host     = "smtp.gmail.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "login"
  username = "smtp_user@example.com"
  password = var.smtp_password
}
```

### With Test Email

```hcl
resource "tfe_smtp_settings" "example" {
  enabled           = true
  host              = "smtp.example.com"
  port              = 587
  sender            = "terraform@example.com"
  auth              = "login"
  username          = "smtp_user"
  password          = var.smtp_password
  test_email_address = "admin@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Required) Whether SMTP is enabled. When `true`, all other required attributes must have valid values.
* `host` - (Optional) The hostname of the SMTP server.
* `port` - (Optional) The port of the SMTP server. Defaults to `25`.
* `sender` - (Optional) The email address to use as the sender for outgoing emails.
* `auth` - (Optional) The authentication type. Valid values are `none`, `plain`, and `login`. Defaults to `none`.
* `username` - (Optional) The username used to authenticate to the SMTP server. Required if `auth` is `login` or `plain`.
* `password` - (Optional) The password used to authenticate to the SMTP server. Required if `auth` is `login` or `plain`. This value is write-only and will not be returned in state.
* `password_wo` - (Optional) **Deprecated** Use `password` instead. This attribute will be removed in a future version.
* `test_email_address` - (Optional) An email address to send a test message to. This value is not persisted and is only used during the update operation to verify SMTP configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the SMTP settings. Always `smtp`.

## Import

SMTP settings can be imported using the ID `smtp`:

```shell
terraform import tfe_smtp_settings.example smtp
```

## Notes

### Deleting SMTP Settings

When you destroy this resource, SMTP will be disabled (by setting `enabled = false`), but the configuration will remain in Terraform Enterprise. To re-enable SMTP, simply recreate the resource.

### Password Security

The `password` attribute is sensitive and write-only. It will not be stored in Terraform state after being set. If you need to update the password, you must provide the new value explicitly.

### Testing SMTP Configuration

Use the `test_email_address` attribute to send a test email when updating SMTP settings. This helps verify that your configuration is correct before relying on it for notifications.

```hcl
resource "tfe_smtp_settings" "example" {
  enabled            = true
  host               = "smtp.example.com"
  port               = 587
  sender             = "terraform@example.com"
  auth               = "login"
  username           = "smtp_user"
  password           = var.smtp_password
  test_email_address = "admin@example.com"  # Test email will be sent here
}
```

### Authentication Requirements

When using `auth = "plain"` or `auth = "login"`, both `username` and `password` must be provided. If `auth = "none"`, these fields should be omitted.