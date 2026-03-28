# SMTP Settings Example

This example demonstrates how to configure SMTP settings for Terraform Enterprise using the `tfe_smtp_settings` resource.

## Prerequisites

- Terraform Enterprise (self-hosted) installation
- Admin access to Terraform Enterprise
- SMTP server details (host, port, credentials)

## Usage

1. Set your SMTP credentials as environment variables:

```bash
export TF_VAR_smtp_username="your-smtp-username"
export TF_VAR_smtp_password="your-smtp-password"
```

2. Initialize Terraform:

```bash
terraform init
```

3. Review the plan:

```bash
terraform plan
```

4. Apply the configuration:

```bash
terraform apply
```

## Examples Included

### Basic Configuration (No Authentication)

Configures SMTP without authentication - suitable for internal mail servers that don't require credentials.

### With Login Authentication

Configures SMTP with login authentication - suitable for most external SMTP providers like Gmail, SendGrid, etc.

### With Test Email

Includes a test email address to verify the SMTP configuration works correctly.

## Configuration Options

### Authentication Types

- `none`: No authentication required
- `plain`: PLAIN authentication mechanism
- `login`: LOGIN authentication mechanism

### Common SMTP Ports

- `25`: Standard SMTP (usually unencrypted)
- `587`: SMTP with STARTTLS (recommended)
- `465`: SMTP over SSL (legacy)

## Security Notes

- Always use environment variables or a secure secret management system for SMTP credentials
- The password is write-only and will not be stored in Terraform state
- Consider using TLS/STARTTLS (port 587) for secure communication

## Outputs

The example includes outputs to verify the SMTP configuration:

- `smtp_enabled`: Whether SMTP is currently enabled
- `smtp_host`: The configured SMTP host
- `smtp_sender`: The configured sender email address