// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFESMTPSettingsDataSource_basic(t *testing.T) {
	skipIfCloud(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettingsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "id", "smtp"),
					resource.TestCheckResourceAttrSet("data.tfe_smtp_settings.foobar", "enabled"),
					resource.TestCheckResourceAttrSet("data.tfe_smtp_settings.foobar", "host"),
					resource.TestCheckResourceAttrSet("data.tfe_smtp_settings.foobar", "port"),
					resource.TestCheckResourceAttrSet("data.tfe_smtp_settings.foobar", "sender"),
					resource.TestCheckResourceAttrSet("data.tfe_smtp_settings.foobar", "auth"),
					// Password should never be returned
					resource.TestCheckNoResourceAttr("data.tfe_smtp_settings.foobar", "password"),
				),
			},
		},
	})
}

func TestAccTFESMTPSettingsDataSource_withResource(t *testing.T) {
	skipIfCloud(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESMTPSettingsDataSourceConfig_withResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check resource attributes
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "host", "smtp.example.com"),
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "port", "587"),
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "sender", "terraform@example.com"),
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "auth", "plain"),
					resource.TestCheckResourceAttr("tfe_smtp_settings.test", "username", "smtp_user"),
					
					// Check data source attributes match resource
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "id", "smtp"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "enabled", "true"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "host", "smtp.example.com"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "port", "587"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "sender", "terraform@example.com"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "auth", "plain"),
					resource.TestCheckResourceAttr("data.tfe_smtp_settings.foobar", "username", "smtp_user"),
					
					// Password should never be returned in data source
					resource.TestCheckNoResourceAttr("data.tfe_smtp_settings.foobar", "password"),
				),
			},
		},
	})
}

func testAccTFESMTPSettingsDataSourceConfig_basic() string {
	return `
data "tfe_smtp_settings" "foobar" {
}`
}

func testAccTFESMTPSettingsDataSourceConfig_withResource() string {
	return `
resource "tfe_smtp_settings" "test" {
  enabled  = true
  host     = "smtp.example.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "plain"
  username = "smtp_user"
  password = "test_password"
}

data "tfe_smtp_settings" "foobar" {
  depends_on = [tfe_smtp_settings.test]
}`
}

// Made with Bob
