# Configure SMTP settings for Terraform Enterprise

terraform {
  required_providers {
    tfe = {
      source  = "hashicorp/tfe"
      version = "~> 0.58"
    }
  }
}

# Basic SMTP configuration without authentication
resource "tfe_smtp_settings" "basic" {
  enabled = true
  host    = "smtp.example.com"
  port    = 25
  sender  = "terraform@example.com"
  auth    = "none"
}

# SMTP configuration with login authentication
resource "tfe_smtp_settings" "with_auth" {
  enabled  = true
  host     = "smtp.gmail.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "login"
  username = var.smtp_username
  password = var.smtp_password
}

# SMTP configuration with test email
resource "tfe_smtp_settings" "with_test" {
  enabled            = true
  host               = "smtp.example.com"
  port               = 587
  sender             = "terraform@example.com"
  auth               = "plain"
  username           = var.smtp_username
  password           = var.smtp_password
  test_email_address = "admin@example.com"
}

# Read current SMTP settings
data "tfe_smtp_settings" "current" {}

# Output SMTP configuration status
output "smtp_enabled" {
  description = "Whether SMTP is currently enabled"
  value       = data.tfe_smtp_settings.current.enabled
}

output "smtp_host" {
  description = "The configured SMTP host"
  value       = data.tfe_smtp_settings.current.host
}

output "smtp_sender" {
  description = "The configured sender email address"
  value       = data.tfe_smtp_settings.current.sender
}