package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithConfigure   = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}
)

const organizationSingletonID = "organization"

// NewOrganizationResource returns a new lago_organization resource.
func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

type organizationResource struct {
	client *lago.Client
}

type organizationResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Email                     types.String `tfsdk:"email"`
	AddressLine1              types.String `tfsdk:"address_line1"`
	AddressLine2              types.String `tfsdk:"address_line2"`
	City                      types.String `tfsdk:"city"`
	State                     types.String `tfsdk:"state"`
	Zipcode                   types.String `tfsdk:"zipcode"`
	Country                   types.String `tfsdk:"country"`
	DefaultCurrency           types.String `tfsdk:"default_currency"`
	Timezone                  types.String `tfsdk:"timezone"`
	LegalName                 types.String `tfsdk:"legal_name"`
	LegalNumber               types.String `tfsdk:"legal_number"`
	TaxIdentificationNumber   types.String `tfsdk:"tax_identification_number"`
	NetPaymentTerm            types.Int64  `tfsdk:"net_payment_term"`
	DocumentNumbering         types.String `tfsdk:"document_numbering"`
	DocumentNumberPrefix      types.String `tfsdk:"document_number_prefix"`
	FinalizeZeroAmountInvoice types.Bool   `tfsdk:"finalize_zero_amount_invoice"`
	EmailSettings             types.Set    `tfsdk:"email_settings"`
	BillingConfiguration      types.Object `tfsdk:"billing_configuration"`
	CreatedAt                 types.String `tfsdk:"created_at"`
}

type organizationBillingConfigModel struct {
	InvoiceGracePeriod types.Int64  `tfsdk:"invoice_grace_period"`
	InvoiceFooter      types.String `tfsdk:"invoice_footer"`
	DocumentLocale     types.String `tfsdk:"document_locale"`
}

func organizationBillingConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"invoice_grace_period": types.Int64Type,
			"invoice_footer":       types.StringType,
			"document_locale":      types.StringType,
		},
	}
}

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Lago organization settings. This is a singleton resource — only one instance may exist per provider configuration. `Create` calls the organization update API (upsert); `Delete` is a no-op because the organization always exists in Lago.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID. Always set to `\"organization\"`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Organization display name.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization contact email address.",
			},
			"address_line1": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "First line of the organization's billing address.",
			},
			"address_line2": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Second line of the organization's billing address.",
			},
			"city": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "City of the organization's billing address.",
			},
			"state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "State or region of the organization's billing address.",
			},
			"zipcode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Postal code of the organization's billing address.",
			},
			"country": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Country of the organization's billing address (ISO 3166-1 alpha-2, e.g. `US`).",
			},
			"default_currency": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default currency for the organization (ISO 4217, e.g. `USD`).",
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default timezone for the organization (TZ database name, e.g. `Europe/Paris`).",
			},
			"legal_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Legal name of the organization.",
			},
			"legal_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Legal registration number of the organization.",
			},
			"tax_identification_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tax identification number (VAT number) of the organization.",
			},
			"net_payment_term": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default net payment term in days applied to invoices.",
			},
			"document_numbering": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Invoice document numbering scheme. Allowed values: `per_customer`, `per_organization`.",
				Validators: []validator.String{
					stringvalidator.OneOf("per_customer", "per_organization"),
				},
			},
			"document_number_prefix": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Prefix prepended to all generated document numbers.",
			},
			"finalize_zero_amount_invoice": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "When `true`, invoices with a zero total amount are finalized rather than voided. Note: the underlying API field uses `omitempty`, so setting this to `false` after it has been enabled may require manual adjustment in the Lago UI.",
			},
			"email_settings": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of event types that trigger email notifications (e.g. `invoice.finalized`, `credit_note.created`).",
			},
			"billing_configuration": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Organization-level billing configuration.",
				Attributes: map[string]schema.Attribute{
					"invoice_grace_period": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Grace period in days before an invoice is finalized after its billing period ends.",
					},
					"invoice_footer": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Footer text that appears on all generated invoices.",
					},
					"document_locale": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Locale used when generating invoice documents (e.g. `en`, `fr`, `de`).",
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Organization creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*lagoProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *lagoProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = providerData.client
}

// Create upserts the organization. The Lago organization always exists; Create simply
// applies the plan values via the Update API endpoint.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandOrganizationInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, lagoErr := r.client.Organization().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Configuring Lago Organization", lagoErr.Error())
		return
	}

	state, diags := mapOrganizationToModel(org, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes state by calling Update with a zero-value input. Because every field in
// OrganizationInput carries omitempty, passing an empty struct is a safe read-equivalent
// that returns the current organization state without modifying it.
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, lagoErr := r.client.Organization().Update(ctx, &lago.OrganizationInput{})
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Reading Lago Organization", lagoErr.Error())
		return
	}

	newState, diags := mapOrganizationToModel(org, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandOrganizationInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, lagoErr := r.client.Organization().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Organization", lagoErr.Error())
		return
	}

	state, diags := mapOrganizationToModel(org, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete is a no-op. The Lago organization cannot be deleted via the API. Terraform will
// simply remove the resource from its state.
func (r *organizationResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "lago_organization destroy called — the Lago organization itself is not deleted; only the Terraform state entry is removed.")
}

// ImportState allows importing by any ID string (conventionally "organization").
// The supplied ID is ignored and replaced with the fixed singleton value so that the
// subsequent Read can populate full state.
func (r *organizationResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), organizationSingletonID)...)
}

// expandOrganizationInput converts a Terraform plan model into a lago.OrganizationInput.
// Only non-null, non-unknown fields are populated so that omitempty preserves existing
// API values for any fields the user has not specified.
func expandOrganizationInput(ctx context.Context, model organizationResourceModel) (*lago.OrganizationInput, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	input := &lago.OrganizationInput{}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		input.Name = model.Name.ValueString()
	}
	if !model.Email.IsNull() && !model.Email.IsUnknown() {
		input.Email = model.Email.ValueString()
	}
	if !model.AddressLine1.IsNull() && !model.AddressLine1.IsUnknown() {
		input.AddressLine1 = model.AddressLine1.ValueString()
	}
	if !model.AddressLine2.IsNull() && !model.AddressLine2.IsUnknown() {
		input.AddressLine2 = model.AddressLine2.ValueString()
	}
	if !model.City.IsNull() && !model.City.IsUnknown() {
		input.City = model.City.ValueString()
	}
	if !model.State.IsNull() && !model.State.IsUnknown() {
		input.State = model.State.ValueString()
	}
	if !model.Zipcode.IsNull() && !model.Zipcode.IsUnknown() {
		input.Zipcode = model.Zipcode.ValueString()
	}
	if !model.Country.IsNull() && !model.Country.IsUnknown() {
		input.Country = model.Country.ValueString()
	}
	if !model.DefaultCurrency.IsNull() && !model.DefaultCurrency.IsUnknown() {
		input.DefaultCurrency = lago.Currency(model.DefaultCurrency.ValueString())
	}
	if !model.Timezone.IsNull() && !model.Timezone.IsUnknown() {
		input.Timezone = model.Timezone.ValueString()
	}
	if !model.LegalName.IsNull() && !model.LegalName.IsUnknown() {
		input.LegalName = model.LegalName.ValueString()
	}
	if !model.LegalNumber.IsNull() && !model.LegalNumber.IsUnknown() {
		input.LegalNumber = model.LegalNumber.ValueString()
	}
	if !model.TaxIdentificationNumber.IsNull() && !model.TaxIdentificationNumber.IsUnknown() {
		input.TaxIdentificationNumber = model.TaxIdentificationNumber.ValueString()
	}
	if !model.NetPaymentTerm.IsNull() && !model.NetPaymentTerm.IsUnknown() {
		input.NetPaymentTerm = int(model.NetPaymentTerm.ValueInt64())
	}
	if !model.DocumentNumbering.IsNull() && !model.DocumentNumbering.IsUnknown() {
		input.DocumentNumbering = lago.OrganizationDocumentNumbering(model.DocumentNumbering.ValueString())
	}
	if !model.DocumentNumberPrefix.IsNull() && !model.DocumentNumberPrefix.IsUnknown() {
		input.DocumentNumberPrefix = model.DocumentNumberPrefix.ValueString()
	}
	if !model.FinalizeZeroAmountInvoice.IsNull() && !model.FinalizeZeroAmountInvoice.IsUnknown() {
		input.FinalizeZeroAmountInvoice = model.FinalizeZeroAmountInvoice.ValueBool()
	}

	if !model.EmailSettings.IsNull() && !model.EmailSettings.IsUnknown() {
		var settings []string
		d := model.EmailSettings.ElementsAs(ctx, &settings, false)
		diagnostics.Append(d...)
		if !d.HasError() {
			input.EmailSettings = settings
		}
	}

	if !model.BillingConfiguration.IsNull() && !model.BillingConfiguration.IsUnknown() {
		var bc organizationBillingConfigModel
		d := model.BillingConfiguration.As(ctx, &bc, basetypes.ObjectAsOptions{})
		diagnostics.Append(d...)
		if !d.HasError() {
			billingInput := lago.OrganizationBillingConfigurationInput{}
			if !bc.InvoiceGracePeriod.IsNull() && !bc.InvoiceGracePeriod.IsUnknown() {
				billingInput.InvoiceGracePeriod = int(bc.InvoiceGracePeriod.ValueInt64())
			}
			if !bc.InvoiceFooter.IsNull() && !bc.InvoiceFooter.IsUnknown() {
				billingInput.InvoiceFooter = bc.InvoiceFooter.ValueString()
			}
			if !bc.DocumentLocale.IsNull() && !bc.DocumentLocale.IsUnknown() {
				billingInput.DocumentLocale = bc.DocumentLocale.ValueString()
			}
			input.BillingConfiguration = billingInput
		}
	}

	return input, diagnostics
}

// mapOrganizationToModel converts a lago.Organization API response into Terraform state.
// The base model is passed so that plan values for write-only or structurally-preserved
// fields can be carried forward where needed.
func mapOrganizationToModel(org *lago.Organization, base organizationResourceModel) (organizationResourceModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	state := base

	state.ID = types.StringValue(organizationSingletonID)
	state.Name = stringOrNull(org.Name)
	state.Email = stringOrNull(org.Email)
	state.AddressLine1 = stringOrNull(org.AddressLine1)
	state.AddressLine2 = stringOrNull(org.AddressLine2)
	state.City = stringOrNull(org.City)
	state.State = stringOrNull(org.State)
	state.Zipcode = stringOrNull(org.Zipcode)
	state.Country = stringOrNull(org.Country)
	state.DefaultCurrency = stringOrNull(string(org.DefaultCurrency))
	state.Timezone = stringOrNull(org.Timezone)
	state.LegalName = stringOrNull(org.LegalName)
	state.LegalNumber = stringOrNull(org.LegalNumber)
	state.TaxIdentificationNumber = stringOrNull(org.TaxIdentificationNumber)
	state.DocumentNumbering = stringOrNull(string(org.DocumentNumbering))
	state.DocumentNumberPrefix = stringOrNull(org.DocumentNumberPrefix)
	state.FinalizeZeroAmountInvoice = types.BoolValue(org.FinalizeZeroAmountInvoice)

	if org.NetPaymentTerm == 0 {
		state.NetPaymentTerm = types.Int64Null()
	} else {
		state.NetPaymentTerm = types.Int64Value(int64(org.NetPaymentTerm))
	}

	if len(org.EmailSettings) == 0 {
		state.EmailSettings = types.SetNull(types.StringType)
	} else {
		attrs := make([]attr.Value, 0, len(org.EmailSettings))
		for _, s := range org.EmailSettings {
			attrs = append(attrs, types.StringValue(s))
		}
		sv, d := types.SetValue(types.StringType, attrs)
		diagnostics.Append(d...)
		state.EmailSettings = sv
	}

	bc := org.BillingConfiguration
	billingIsEmpty := bc.InvoiceGracePeriod == 0 && bc.InvoiceFooter == "" && bc.DocumentLocale == ""

	if billingIsEmpty {
		// Only set billing_configuration to null when the plan also has it null/unknown,
		// to avoid clobbering an explicitly-configured block whose fields happen to be
		// at their default (zero) values.
		if base.BillingConfiguration.IsNull() || base.BillingConfiguration.IsUnknown() {
			state.BillingConfiguration = types.ObjectNull(organizationBillingConfigObjectType().AttrTypes)
		} else {
			bcObj, d := types.ObjectValue(
				organizationBillingConfigObjectType().AttrTypes,
				map[string]attr.Value{
					"invoice_grace_period": types.Int64Null(),
					"invoice_footer":       types.StringNull(),
					"document_locale":      types.StringNull(),
				},
			)
			diagnostics.Append(d...)
			state.BillingConfiguration = bcObj
		}
	} else {
		var invoiceGracePeriod types.Int64
		if bc.InvoiceGracePeriod == 0 {
			invoiceGracePeriod = types.Int64Null()
		} else {
			invoiceGracePeriod = types.Int64Value(int64(bc.InvoiceGracePeriod))
		}

		bcObj, d := types.ObjectValue(
			organizationBillingConfigObjectType().AttrTypes,
			map[string]attr.Value{
				"invoice_grace_period": invoiceGracePeriod,
				"invoice_footer":       stringOrNull(bc.InvoiceFooter),
				"document_locale":      stringOrNull(bc.DocumentLocale),
			},
		)
		diagnostics.Append(d...)
		state.BillingConfiguration = bcObj
	}

	if org.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(org.CreatedAt.Format(time.RFC3339))
	}

	return state, diagnostics
}
