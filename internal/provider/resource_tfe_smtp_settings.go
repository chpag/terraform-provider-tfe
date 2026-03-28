// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	smtpDefaultPort int64 = 25
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &resourceTFESMTPSettings{}
	_ resource.ResourceWithConfigure   = &resourceTFESMTPSettings{}
	_ resource.ResourceWithImportState = &resourceTFESMTPSettings{}
)

// NewSMTPSettingsResource is a helper function to simplify the provider implementation.
func NewSMTPSettingsResource() resource.Resource {
	return &resourceTFESMTPSettings{}
}

type modelTFESMTPSettings struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	Sender           types.String `tfsdk:"sender"`
	Auth             types.String `tfsdk:"auth"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	PasswordWO       types.String `tfsdk:"password_wo"`
	TestEmailAddress types.String `tfsdk:"test_email_address"`
}

// resourceTFESMTPSettings implements the tfe_smtp_settings resource type
type resourceTFESMTPSettings struct {
	client *tfe.Client
}

// modelFromTFEAdminSMTPSettings builds a modelTFESMTPSettings struct from a tfe.AdminSMTPSetting value
func modelFromTFEAdminSMTPSettings(v *tfe.AdminSMTPSetting, password types.String, isWriteOnly bool) modelTFESMTPSettings {
	m := modelTFESMTPSettings{
		ID:       types.StringValue(v.ID),
		Enabled:  types.BoolValue(v.Enabled),
		Host:     types.StringValue(v.Host),
		Port:     types.Int64Value(int64(v.Port)),
		Sender:   types.StringValue(v.Sender),
		Auth:     types.StringValue(string(v.Auth)),
		Username: types.StringValue(v.Username),
		Password: types.StringValue(""),
	}

	if len(password.ValueString()) > 0 {
		m.Password = password
	}

	// Don't retrieve values if write-only is being used. Unset the password field before updating the state.
	if isWriteOnly {
		m.Password = types.StringValue("")
	}

	return m
}

// Configure implements resource.ResourceWithConfigure
func (r *resourceTFESMTPSettings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Early exit if provider is not properly configured (i.e. we're only validating config or something)
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource Configure type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
	}
	r.client = client.Client
}

// Metadata implements resource.Resource
func (r *resourceTFESMTPSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_settings"
}

// Schema implements resource.Resource
func (r *resourceTFESMTPSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages SMTP settings for Terraform Enterprise.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SMTP settings. Always 'smtp'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SMTP is enabled. When enabled, all other attributes must have valid values.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "The hostname of the SMTP server.",
				Optional:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port of the SMTP server.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(smtpDefaultPort),
			},
			"sender": schema.StringAttribute{
				Description: "The desired sender email address.",
				Optional:    true,
			},
			"auth": schema.StringAttribute{
				Description: "The authentication type. Valid values are 'none', 'plain', and 'login'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(tfe.SMTPAuthNone)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfe.SMTPAuthNone),
						string(tfe.SMTPAuthPlain),
						string(tfe.SMTPAuthLogin),
					),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username used to authenticate to the SMTP server. Required if auth is 'login' or 'plain'.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password used to authenticate to the SMTP server. Required if auth is 'login' or 'plain'. This value is write-only.",
				Optional:    true,
				Sensitive:   true,
			},
			"password_wo": schema.StringAttribute{
				Description: "**Deprecated** Use password instead. This attribute will be removed in a future version.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
				},
			},
			"test_email_address": schema.StringAttribute{
				Description: "The email address to send a test message to. This value is not persisted and is only used during testing.",
				Optional:    true,
			},
		},
	}
}

// Create implements resource.Resource
func (r *resourceTFESMTPSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTFESMTPSettings

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// SMTP settings always exist, so Create is the same as Update
	r.update(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements resource.Resource
func (r *resourceTFESMTPSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTFESMTPSettings

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading SMTP settings")
	settings, err := r.client.Admin.Settings.SMTP.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SMTP settings", err.Error())
		return
	}

	// Determine if we should use write-only pattern for password
	isWriteOnly := !state.PasswordWO.IsNull() && !state.PasswordWO.IsUnknown()

	result := modelFromTFEAdminSMTPSettings(settings, state.Password, isWriteOnly)

	// Preserve test_email_address from state (it's not returned by the API)
	result.TestEmailAddress = state.TestEmailAddress

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

// Update implements resource.Resource
func (r *resourceTFESMTPSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan modelTFESMTPSettings

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.update(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// update is a helper function that performs the actual update operation
func (r *resourceTFESMTPSettings) update(ctx context.Context, plan *modelTFESMTPSettings, diags *diag.Diagnostics) {
	options := tfe.AdminSMTPSettingsUpdateOptions{
		Enabled: tfe.Bool(plan.Enabled.ValueBool()),
	}

	if !plan.Host.IsNull() {
		options.Host = tfe.String(plan.Host.ValueString())
	}

	if !plan.Port.IsNull() {
		options.Port = tfe.Int(int(plan.Port.ValueInt64()))
	}

	if !plan.Sender.IsNull() {
		options.Sender = tfe.String(plan.Sender.ValueString())
	}

	if !plan.Auth.IsNull() {
		auth := tfe.SMTPAuthType(plan.Auth.ValueString())
		options.Auth = &auth
	}

	if !plan.Username.IsNull() {
		options.Username = tfe.String(plan.Username.ValueString())
	}

	// Handle password from either password or password_wo field
	password := plan.Password
	if !plan.PasswordWO.IsNull() && !plan.PasswordWO.IsUnknown() {
		password = plan.PasswordWO
	}

	if !password.IsNull() && !password.IsUnknown() {
		options.Password = tfe.String(password.ValueString())
	}

	if !plan.TestEmailAddress.IsNull() {
		options.TestEmailAddress = tfe.String(plan.TestEmailAddress.ValueString())
	}

	tflog.Debug(ctx, "Updating SMTP settings")
	settings, err := r.client.Admin.Settings.SMTP.Update(ctx, options)
	if err != nil {
		diags.AddError("Error updating SMTP settings", err.Error())
		return
	}

	// Determine if we should use write-only pattern for password
	isWriteOnly := !plan.PasswordWO.IsNull() && !plan.PasswordWO.IsUnknown()

	result := modelFromTFEAdminSMTPSettings(settings, password, isWriteOnly)

	// Preserve test_email_address in the result (it's not returned by the API)
	result.TestEmailAddress = plan.TestEmailAddress

	*plan = result
}

// Delete implements resource.Resource
func (r *resourceTFESMTPSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTFESMTPSettings

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// SMTP settings cannot be deleted, only disabled
	tflog.Debug(ctx, "Disabling SMTP settings")
	options := tfe.AdminSMTPSettingsUpdateOptions{
		Enabled: tfe.Bool(false),
	}

	_, err := r.client.Admin.Settings.SMTP.Update(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Error disabling SMTP settings", err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *resourceTFESMTPSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID is always "smtp" for SMTP settings
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Made with Bob
