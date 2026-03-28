variable "smtp_username" {
  description = "SMTP username for authentication"
  type        = string
  sensitive   = true
}

variable "smtp_password" {
  description = "SMTP password for authentication"
  type        = string
  sensitive   = true
}