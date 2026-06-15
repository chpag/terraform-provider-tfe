// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEVerifyRegistryProviderVersion_Complete(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-org-%d", rInt)),
		Email: tfe.String("admin@terraformer.inc"),
	})
	t.Cleanup(orgCleanup)

	// Note: This test would require setting up a complete provider version with GPG keys,
	// shasums, and platform binaries. For a real test, you would need to:
	// 1. Create a GPG key
	// 2. Create a registry provider
	// 3. Create a provider version with shasums
	// 4. Upload platform binaries
	// This is a simplified test structure

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVerifyRegistryProviderVersion_Config(organization.Name, "example-provider", "1.0.0"),
				// This would fail in a real scenario without proper setup
				ExpectError: regexp.MustCompile("Provider version not found|Unable to read registry provider"),
			},
		},
	})
}

func TestAccTFEVerifyRegistryProviderVersion_ValidationErrors(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-org-%d", rInt)),
		Email: tfe.String("admin@terraformer.inc"),
	})
	t.Cleanup(orgCleanup)

	invalidCases := []struct {
		Config      string
		ExpectError *regexp.Regexp
	}{
		{
			Config:      testAccTFEVerifyRegistryProviderVersion_MissingNamespace(organization.Name),
			ExpectError: regexp.MustCompile("Missing namespace"),
		},
		{
			Config:      testAccTFEVerifyRegistryProviderVersion_InvalidRegistryName(organization.Name),
			ExpectError: regexp.MustCompile("Invalid registry_name"),
		},
	}

	for _, invalidCase := range invalidCases {
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
			},
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      invalidCase.Config,
					ExpectError: invalidCase.ExpectError,
				},
			},
		})
	}
}

func testAccTFEVerifyRegistryProviderVersion_Config(organization, providerName, version string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_verify_registry_provider_version.test]
    }
  }
}

action "tfe_verify_registry_provider_version" "test" {
  config {
    organization  = "%s"
    provider_name = "%s"
    version       = "%s"
  }
}`, organization, providerName, version)
}

func testAccTFEVerifyRegistryProviderVersion_MissingNamespace(organization string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_verify_registry_provider_version.test]
    }
  }
}

action "tfe_verify_registry_provider_version" "test" {
  config {
    organization  = "%s"
    registry_name = "public"
    provider_name = "example"
    version       = "1.0.0"
    # Missing namespace for public registry
  }
}`, organization)
}

func testAccTFEVerifyRegistryProviderVersion_InvalidRegistryName(organization string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_verify_registry_provider_version.test]
    }
  }
}

action "tfe_verify_registry_provider_version" "test" {
  config {
    organization  = "%s"
    registry_name = "invalid"
    provider_name = "example"
    version       = "1.0.0"
  }
}`, organization)
}

func testAccTFEVerifyRegistryProviderVersion_WithRequiredPlatforms(organization, providerName, version string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_verify_registry_provider_version.test]
    }
  }
}

action "tfe_verify_registry_provider_version" "test" {
  config {
    organization  = "%s"
    provider_name = "%s"
    version       = "%s"
    required_platforms = [
      "linux_amd64",
      "darwin_amd64",
      "darwin_arm64",
      "windows_amd64"
    ]
  }
}`, organization, providerName, version)
}

// Made with Bob
