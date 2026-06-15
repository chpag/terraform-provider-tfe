// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ action.Action                   = &actionTFEVerifyRegistryProviderVersion{}
	_ action.ActionWithConfigure      = &actionTFEVerifyRegistryProviderVersion{}
	_ action.ActionWithValidateConfig = &actionTFEVerifyRegistryProviderVersion{}
)

func NewVerifyRegistryProviderVersionAction() action.Action {
	return &actionTFEVerifyRegistryProviderVersion{}
}

type actionTFEVerifyRegistryProviderVersion struct {
	config ConfiguredClient
}

type actionTFEVerifyRegistryProviderVersionModel struct {
	Organization      types.String `tfsdk:"organization"`
	RegistryName      types.String `tfsdk:"registry_name"`
	Namespace         types.String `tfsdk:"namespace"`
	ProviderName      types.String `tfsdk:"provider_name"`
	Version           types.String `tfsdk:"version"`
	RequiredPlatforms types.Set    `tfsdk:"required_platforms"`
}

func (a *actionTFEVerifyRegistryProviderVersion) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected action Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on Github.", req.ProviderData),
		)
	}
	a.config = client
}

func (a *actionTFEVerifyRegistryProviderVersion) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verify_registry_provider_version"
}

func (a *actionTFEVerifyRegistryProviderVersion) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "Name of the organization. If omitted, organization must be defined in the provider config.",
				Optional:    true,
			},
			"registry_name": schema.StringAttribute{
				Description: "Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.",
				Optional:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the provider. For private providers this is the same as the organization.",
				Optional:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Name of the provider.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "The version of the provider to verify (e.g., '1.0.0').",
				Required:    true,
			},
			"required_platforms": schema.SetAttribute{
				Description: "Optional list of required platforms (e.g., ['linux_amd64', 'darwin_amd64']). If not specified, only checks that at least one platform exists.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (a *actionTFEVerifyRegistryProviderVersion) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	var data actionTFEVerifyRegistryProviderVersionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate registry_name if provided
	if !data.RegistryName.IsNull() && !data.RegistryName.IsUnknown() {
		registryName := data.RegistryName.ValueString()
		if registryName != "public" && registryName != "private" {
			resp.Diagnostics.AddAttributeError(
				path.Root("registry_name"),
				"Invalid registry_name",
				"registry_name must be either 'public' or 'private'",
			)
		}

		// Validate namespace requirement for public registry
		if registryName == "public" && data.Namespace.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("namespace"),
				"Missing namespace",
				"namespace is required when registry_name is 'public'",
			)
		}
	}
}

func (a *actionTFEVerifyRegistryProviderVersion) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data actionTFEVerifyRegistryProviderVersionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := a.config.Client

	// Determine organization
	var organization string
	if data.Organization.IsNull() {
		organization = a.config.Organization
		if organization == "" {
			resp.Diagnostics.AddError(
				"Missing organization",
				"organization must be specified either in the action config or provider config",
			)
			return
		}
	} else {
		organization = data.Organization.ValueString()
	}

	// Determine registry name
	registryName := "private"
	if !data.RegistryName.IsNull() {
		registryName = data.RegistryName.ValueString()
	}

	// Determine namespace
	var namespace string
	if registryName == "private" {
		namespace = organization
	} else {
		namespace = data.Namespace.ValueString()
	}

	providerName := data.ProviderName.ValueString()
	version := data.Version.ValueString()

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Verifying provider version %s/%s/%s@%s...", namespace, providerName, registryName, version),
	})

	// Read the provider to ensure it exists
	providerID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     tfe.RegistryName(registryName),
		Namespace:        namespace,
		Name:             providerName,
	}

	provider, err := client.RegistryProviders.Read(ctx, providerID, &tfe.RegistryProviderReadOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read registry provider",
			fmt.Sprintf("Error reading provider %s/%s: %s", namespace, providerName, err.Error()),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Provider %s found. Checking version %s...", provider.Name, version),
	})

	// List all versions for this provider
	versionListOpts := &tfe.RegistryProviderVersionListOptions{}
	versions, err := client.RegistryProviderVersions.List(ctx, providerID, versionListOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list provider versions",
			fmt.Sprintf("Error listing versions for provider %s/%s: %s", namespace, providerName, err.Error()),
		)
		return
	}

	// Find the specific version
	var targetVersion *tfe.RegistryProviderVersion
	for _, v := range versions.Items {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		resp.Diagnostics.AddError(
			"Provider version not found",
			fmt.Sprintf("Version %s not found for provider %s/%s", version, namespace, providerName),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Version %s found. Verifying configuration...", version),
	})

	// Verify the version has required metadata
	issues := []string{}

	// Check if key_id is set and verify the GPG key exists in the registry
	if targetVersion.KeyID == "" {
		issues = append(issues, "Missing GPG key ID")
	} else {
		// Verify the GPG key actually exists in the registry
		_, err := client.GPGKeys.Read(ctx, tfe.GPGKeyID{
			RegistryName: tfe.RegistryName(registryName),
			Namespace:    namespace,
			KeyID:        targetVersion.KeyID,
		})
		if err != nil {
			issues = append(issues, fmt.Sprintf("GPG key ID '%s' not found in registry: %s", targetVersion.KeyID, err.Error()))
		}
	}

	// Check if protocols are set
	if len(targetVersion.Protocols) == 0 {
		issues = append(issues, "No protocols defined")
	}

	// Check if shasums are uploaded
	if !targetVersion.ShasumsUploaded {
		issues = append(issues, "SHA256SUMS file not uploaded")
	}

	// Check if shasums signature is uploaded
	if !targetVersion.ShasumsSigUploaded {
		issues = append(issues, "SHA256SUMS.sig file not uploaded")
	}

	// List platforms for this version
	versionID := tfe.RegistryProviderVersionID{
		RegistryProviderID: providerID,
		Version:            version,
	}

	platformListOpts := &tfe.RegistryProviderPlatformListOptions{}
	platforms, err := client.RegistryProviderPlatforms.List(ctx, versionID, platformListOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list provider platforms",
			fmt.Sprintf("Error listing platforms for version %s: %s", version, err.Error()),
		)
		return
	}

	if len(platforms.Items) == 0 {
		issues = append(issues, "No platform binaries uploaded")
	} else {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Found %d platform(s)", len(platforms.Items)),
		})

		// Check if required platforms are present
		if !data.RequiredPlatforms.IsNull() {
			var requiredPlatforms []string
			resp.Diagnostics.Append(data.RequiredPlatforms.ElementsAs(ctx, &requiredPlatforms, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Build a map of available platforms
			availablePlatforms := make(map[string]bool)
			for _, p := range platforms.Items {
				availablePlatforms[p.OS+"/"+p.Arch] = true
			}

			// Check each required platform
			missingPlatforms := []string{}
			for _, required := range requiredPlatforms {
				// Handle both formats: "linux_amd64" and "linux/amd64"
				normalized := strings.ReplaceAll(required, "_", "/")
				if !availablePlatforms[normalized] {
					missingPlatforms = append(missingPlatforms, required)
				}
			}

			if len(missingPlatforms) > 0 {
				issues = append(issues, fmt.Sprintf("Missing required platforms: %s", strings.Join(missingPlatforms, ", ")))
			}
		}

		// Check if all platforms have binaries uploaded
		for _, platform := range platforms.Items {
			if platform.Filename == "" {
				issues = append(issues, fmt.Sprintf("Platform %s/%s has no binary uploaded", platform.OS, platform.Arch))
			}
		}
	}

	// Report results
	if len(issues) > 0 {
		resp.Diagnostics.AddError(
			"Provider version configuration incomplete",
			fmt.Sprintf("Version %s has the following issues:\n- %s", version, strings.Join(issues, "\n- ")),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("✓ Provider version %s is complete and ready to use", version),
	})
}

// Made with Bob
