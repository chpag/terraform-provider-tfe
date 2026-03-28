// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testSMTPResourceName = "tfe_smtp_settings.foobar"

// FLAKE ALERT: SMTP settings are a singleton resource shared by the entire TFE
// instance, and any test touching them is at high risk to flake.
// In order for these tests to be safe, the following requirements MUST be met:
//  1. All test cases for this resource must run within a SINGLE test func, using
//     t.Run to separate the individual test cases.
//  2. The inner sub-tests must not call t.Parallel.
//
// If these tests are split into multiple test funcs and they get allocated to
// different test runner partitions in CI, then they will inevitably flake, as
// tests running concurrently in different containers will be competing to set
// the same shared global state in the TFE instance.

func TestAccTFESMTPSettings_omnibus(t *testing.T) {
	skipIfCloud(t)

	t.Run("basic SMTP settings without authentication", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_basic(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", "smtp.example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", "25"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "sender", "terraform@example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "auth", "none"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "username", ""),
						resource.TestCheckResourceAttr(testSMTPResourceName, "id", "smtp"),
					),
				},
			},
		})
	})

	t.Run("SMTP settings with plain authentication", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_withPlainAuth(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", "smtp.example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", "587"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "sender", "terraform@example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "auth", "plain"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "username", "smtp_user"),
						// Password should not be in state
						resource.TestCheckNoResourceAttr(testSMTPResourceName, "password"),
					),
				},
			},
		})
	})

	t.Run("SMTP settings with login authentication", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_withLoginAuth(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", "smtp.gmail.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", "587"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "sender", "terraform@example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "auth", "login"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "username", "smtp_user@example.com"),
					),
				},
			},
		})
	})

	t.Run("update SMTP settings", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_basic(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", "smtp.example.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", "25"),
					),
				},
				{
					Config: testAccTFESMTPSettings_updated(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "host", "smtp.updated.com"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "port", "587"),
						resource.TestCheckResourceAttr(testSMTPResourceName, "sender", "updated@example.com"),
					),
				},
			},
		})
	})

	t.Run("disable SMTP settings", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_basic(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "true"),
					),
				},
				{
					Config: testAccTFESMTPSettings_disabled(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(testSMTPResourceName, "enabled", "false"),
					),
				},
			},
		})
	})

	t.Run("import SMTP settings", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccMuxedProviders,
			CheckDestroy:             testAccTFESMTPSettingsDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccTFESMTPSettings_basic(),
				},
				{
					ResourceName:      testSMTPResourceName,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateId:     "smtp",
					// Password is not returned by the API
					ImportStateVerifyIgnore: []string{"password", "password_wo", "test_email_address"},
				},
			},
		})
	})
}

func testAccTFESMTPSettingsDestroy(_ *terraform.State) error {
	settings, err := testAccConfiguredClient.Client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read SMTP Settings: %w", err)
	}
	
	// SMTP settings cannot be deleted, only disabled
	// So we check if they are disabled after destroy
	if settings.Enabled {
		return errors.New("SMTP settings are still enabled")
	}
	
	return nil
}

func testAccTFESMTPSettings_basic() string {
	return `
resource "tfe_smtp_settings" "foobar" {
  enabled = true
  host    = "smtp.example.com"
  port    = 25
  sender  = "terraform@example.com"
  auth    = "none"
}`
}

func testAccTFESMTPSettings_withPlainAuth() string {
	return `
resource "tfe_smtp_settings" "foobar" {
  enabled  = true
  host     = "smtp.example.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "plain"
  username = "smtp_user"
  password = "test_password_plain"
}`
}

func testAccTFESMTPSettings_withLoginAuth() string {
	return `
resource "tfe_smtp_settings" "foobar" {
  enabled  = true
  host     = "smtp.gmail.com"
  port     = 587
  sender   = "terraform@example.com"
  auth     = "login"
  username = "smtp_user@example.com"
  password = "test_password_login"
}`
}

func testAccTFESMTPSettings_updated() string {
	return `
resource "tfe_smtp_settings" "foobar" {
  enabled = true
  host    = "smtp.updated.com"
  port    = 587
  sender  = "updated@example.com"
  auth    = "none"
}`
}

func testAccTFESMTPSettings_disabled() string {
	return `
resource "tfe_smtp_settings" "foobar" {
  enabled = false
}`
}

// Made with Bob
