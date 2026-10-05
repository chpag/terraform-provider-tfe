# Create a public provider version

resource "tfe_registry_provider" "public_example" {
  organization  = "my-org-name"
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}

resource "tfe_registry_provider_version" "public_example" {
  organization  = "my-org-name"
  registry_name = "public"
  namespace     = "hashicorp"
  name          = tfe_registry_provider.public_example.name
  version       = "5.0.0"
  key_id        = tfe_registry_gpg_key.example.id
  protocols     = ["5.0", "6.0"]
}
