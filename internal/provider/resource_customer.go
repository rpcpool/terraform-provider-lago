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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &customerResource{}
	_ resource.ResourceWithConfigure   = &customerResource{}
	_ resource.ResourceWithImportState = &customerResource{}
)

func NewCustomerResource() resource.Resource {
	return &customerResource{}
}

type customerResource struct {
	client *lago.Client
}

type customerResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	LagoID                    types.String `tfsdk:"lago_id"`
	ExternalID                types.String `tfsdk:"external_id"`
	Name                      types.String `tfsdk:"name"`
	Email                     types.String `tfsdk:"email"`
	Phone                     types.String `tfsdk:"phone"`
	URL                       types.String `tfsdk:"url"`
	CustomerType              types.String `tfsdk:"customer_type"`
	Currency                  types.String `tfsdk:"currency"`
	Timezone                  types.String `tfsdk:"timezone"`
	AddressLine1              types.String `tfsdk:"address_line1"`
	AddressLine2              types.String `tfsdk:"address_line2"`
	City                      types.String `tfsdk:"city"`
	State                     types.String `tfsdk:"state"`
	Zipcode                   types.String `tfsdk:"zipcode"`
	Country                   types.String `tfsdk:"country"`
	LegalName                 types.String `tfsdk:"legal_name"`
	LegalNumber               types.String `tfsdk:"legal_number"`
	TaxIdentificationNumber   types.String `tfsdk:"tax_identification_number"`
	NetPaymentTerm            types.Int64  `tfsdk:"net_payment_term"`
	FinalizeZeroAmountInvoice types.String `tfsdk:"finalize_zero_amount_invoice"`
	TaxCodes                  types.Set    `tfsdk:"tax_codes"`
	BillingConfiguration      types.Object `tfsdk:"billing_configuration"`
	ShippingAddress           types.Object `tfsdk:"shipping_address"`
	Metadata                  types.List   `tfsdk:"metadata"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

type customerBillingConfigModel struct {
	InvoiceGracePeriod     types.Int64  `tfsdk:"invoice_grace_period"`
	PaymentProvider        types.String `tfsdk:"payment_provider"`
	PaymentProviderCode    types.String `tfsdk:"payment_provider_code"`
	ProviderCustomerID     types.String `tfsdk:"provider_customer_id"`
	SyncWithProvider       types.Bool   `tfsdk:"sync_with_provider"`
	DocumentLocale         types.String `tfsdk:"document_locale"`
	ProviderPaymentMethods types.Set    `tfsdk:"provider_payment_methods"`
}

type customerShippingAddressModel struct {
	AddressLine1 types.String `tfsdk:"address_line1"`
	AddressLine2 types.String `tfsdk:"address_line2"`
	City         types.String `tfsdk:"city"`
	State        types.String `tfsdk:"state"`
	Zipcode      types.String `tfsdk:"zipcode"`
	Country      types.String `tfsdk:"country"`
}

type customerMetadataModel struct {
	Key              types.String `tfsdk:"key"`
	Value            types.String `tfsdk:"value"`
	DisplayInInvoice types.Bool   `tfsdk:"display_in_invoice"`
}

func customerBillingConfigObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"invoice_grace_period":     types.Int64Type,
		"payment_provider":         types.StringType,
		"payment_provider_code":    types.StringType,
		"provider_customer_id":     types.StringType,
		"sync_with_provider":       types.BoolType,
		"document_locale":          types.StringType,
		"provider_payment_methods": types.SetType{ElemType: types.StringType},
	}}
}

func customerShippingAddressObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"address_line1": types.StringType,
		"address_line2": types.StringType,
		"city":          types.StringType,
		"state":         types.StringType,
		"zipcode":       types.StringType,
		"country":       types.StringType,
	}}
}

func customerMetadataObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":                types.StringType,
		"value":              types.StringType,
		"display_in_invoice": types.BoolType,
	}}
}

func (r *customerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer"
}

func (r *customerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago customer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the customer `external_id`.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the customer.",
			},
			"external_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Your application's unique identifier for the customer. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer full name.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer email address.",
			},
			"phone": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer phone number.",
			},
			"url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer website URL.",
			},
			"customer_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer type. Allowed values: `company`, `individual`.",
				Validators: []validator.String{
					stringvalidator.OneOf("company", "individual"),
				},
			},
			"currency": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Currency code for the customer (e.g. `USD`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer timezone (e.g. `America/New_York`).",
			},
			"address_line1": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "First line of the customer's billing address.",
			},
			"address_line2": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Second line of the customer's billing address.",
			},
			"city": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "City of the customer's billing address.",
			},
			"state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "State or region of the customer's billing address.",
			},
			"zipcode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Postal code of the customer's billing address.",
			},
			"country": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ISO 3166 country code of the customer's billing address.",
			},
			"legal_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer's legal entity name.",
			},
			"legal_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer's legal registration number.",
			},
			"tax_identification_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Customer's tax identification number (VAT, GST, etc.).",
			},
			"net_payment_term": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Net payment term in days.",
			},
			"finalize_zero_amount_invoice": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Controls zero-amount invoice behaviour. Allowed values: `finalize`, `skip`, `inherit`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("finalize", "skip", "inherit"),
				},
			},
			"tax_codes": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of tax codes to apply to the customer.",
			},
			"billing_configuration": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Customer billing configuration.",
				Attributes: map[string]schema.Attribute{
					"invoice_grace_period": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Grace period in days before an invoice is finalised.",
					},
					"payment_provider": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Payment provider identifier (e.g. `stripe`, `adyen`, `gocardless`).",
					},
					"payment_provider_code": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Code of the payment provider integration.",
					},
					"provider_customer_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Customer ID on the payment provider side.",
					},
					"sync_with_provider": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to sync the customer with the payment provider (write-only; not read back from API).",
					},
					"document_locale": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Locale for customer documents (e.g. `en`).",
					},
					"provider_payment_methods": schema.SetAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Payment methods enabled for the customer on the payment provider.",
					},
				},
			},
			"shipping_address": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Customer shipping address.",
				Attributes: map[string]schema.Attribute{
					"address_line1": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "First line of the shipping address.",
					},
					"address_line2": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Second line of the shipping address.",
					},
					"city": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "City of the shipping address.",
					},
					"state": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "State or region of the shipping address.",
					},
					"zipcode": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Postal code of the shipping address.",
					},
					"country": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "ISO 3166 country code of the shipping address.",
					},
				},
			},
			"metadata": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Custom metadata attached to the customer.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Metadata key.",
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Metadata value.",
						},
						"display_in_invoice": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Whether to display this metadata on invoices.",
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Customer creation timestamp (RFC3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Customer last update timestamp (RFC3339).",
			},
		},
	}
}

func (r *customerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandCustomerInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.Customer().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Customer", lagoErr.Error())
		return
	}

	state, diags := mapCustomerToModel(ctx, created, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	customer, lagoErr := r.client.Customer().Get(ctx, state.ExternalID.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Customer", lagoErr.Error())
		return
	}

	newState, diags := mapCustomerToModel(ctx, customer, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *customerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandCustomerInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Customer Update uses the same upsert endpoint as Create
	updated, lagoErr := r.client.Customer().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Customer", lagoErr.Error())
		return
	}

	state, diags := mapCustomerToModel(ctx, updated, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.Customer().Delete(ctx, state.ExternalID.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Customer", lagoErr.Error())
	}
}

func (r *customerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("external_id"), req.ID)...)
}

// expandCustomerInput converts the Terraform model to a Lago API input struct.
func expandCustomerInput(ctx context.Context, model customerResourceModel) (*lago.CustomerInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &lago.CustomerInput{
		ExternalID: model.ExternalID.ValueString(),
	}

	if !model.Name.IsNull() {
		input.Name = model.Name.ValueString()
	}
	if !model.Email.IsNull() {
		input.Email = model.Email.ValueString()
	}
	if !model.Phone.IsNull() {
		input.Phone = model.Phone.ValueString()
	}
	if !model.URL.IsNull() {
		input.URL = model.URL.ValueString()
	}
	if !model.CustomerType.IsNull() {
		input.CustomerType = lago.CustomerType(model.CustomerType.ValueString())
	}
	if !model.Currency.IsNull() {
		input.Currency = lago.Currency(model.Currency.ValueString())
	}
	if !model.Timezone.IsNull() {
		input.Timezone = model.Timezone.ValueString()
	}
	if !model.AddressLine1.IsNull() {
		input.AddressLine1 = model.AddressLine1.ValueString()
	}
	if !model.AddressLine2.IsNull() {
		input.AddressLine2 = model.AddressLine2.ValueString()
	}
	if !model.City.IsNull() {
		input.City = model.City.ValueString()
	}
	if !model.State.IsNull() {
		input.State = model.State.ValueString()
	}
	if !model.Zipcode.IsNull() {
		input.Zipcode = model.Zipcode.ValueString()
	}
	if !model.Country.IsNull() {
		input.Country = model.Country.ValueString()
	}
	if !model.LegalName.IsNull() {
		input.LegalName = model.LegalName.ValueString()
	}
	if !model.LegalNumber.IsNull() {
		input.LegalNumber = model.LegalNumber.ValueString()
	}
	if !model.TaxIdentificationNumber.IsNull() {
		input.TaxIdentificationNumber = model.TaxIdentificationNumber.ValueString()
	}
	if !model.NetPaymentTerm.IsNull() {
		input.NetPaymentTerm = int(model.NetPaymentTerm.ValueInt64())
	}
	if !model.FinalizeZeroAmountInvoice.IsNull() {
		input.FinalizeZeroAmountInvoice = lago.FinalizeZeroAmountInvoice(model.FinalizeZeroAmountInvoice.ValueString())
	}

	taxCodes, taxDiags := expandStringSet(ctx, model.TaxCodes)
	diags.Append(taxDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.TaxCodes = taxCodes

	billingCfg, bcDiags := expandCustomerBillingConfiguration(ctx, model.BillingConfiguration)
	diags.Append(bcDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.BillingConfiguration = billingCfg

	shippingAddr, saDiags := expandCustomerShippingAddress(ctx, model.ShippingAddress)
	diags.Append(saDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.ShippingAddress = shippingAddr

	metadata, mdDiags := expandCustomerMetadata(ctx, model.Metadata)
	diags.Append(mdDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.Metadata = metadata

	return input, diags
}

func expandCustomerBillingConfiguration(ctx context.Context, obj types.Object) (lago.CustomerBillingConfigurationInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out lago.CustomerBillingConfigurationInput

	if obj.IsNull() || obj.IsUnknown() {
		return out, diags
	}

	var model customerBillingConfigModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}

	if !model.InvoiceGracePeriod.IsNull() {
		out.InvoiceGracePeriod = int(model.InvoiceGracePeriod.ValueInt64())
	}
	if !model.PaymentProvider.IsNull() {
		out.PaymentProvider = lago.CustomerPaymentProvider(model.PaymentProvider.ValueString())
	}
	if !model.PaymentProviderCode.IsNull() {
		out.PaymentProviderCode = model.PaymentProviderCode.ValueString()
	}
	if !model.ProviderCustomerID.IsNull() {
		out.ProviderCustomerID = model.ProviderCustomerID.ValueString()
	}
	if !model.SyncWithProvider.IsNull() {
		out.SyncWithProvider = model.SyncWithProvider.ValueBool()
	}
	if !model.DocumentLocale.IsNull() {
		out.DocumentLocale = model.DocumentLocale.ValueString()
	}

	if !model.ProviderPaymentMethods.IsNull() && !model.ProviderPaymentMethods.IsUnknown() {
		var methods []string
		diags.Append(model.ProviderPaymentMethods.ElementsAs(ctx, &methods, false)...)
		if diags.HasError() {
			return out, diags
		}
		providerMethods := make([]lago.ProviderPaymentMethodType, 0, len(methods))
		for _, m := range methods {
			providerMethods = append(providerMethods, lago.ProviderPaymentMethodType(m))
		}
		out.ProviderPaymentMethods = providerMethods
	}

	return out, diags
}

func expandCustomerShippingAddress(ctx context.Context, obj types.Object) (lago.Address, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out lago.Address

	if obj.IsNull() || obj.IsUnknown() {
		return out, diags
	}

	var model customerShippingAddressModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}

	if !model.AddressLine1.IsNull() {
		out.AddressLine1 = model.AddressLine1.ValueString()
	}
	if !model.AddressLine2.IsNull() {
		out.AddressLine2 = model.AddressLine2.ValueString()
	}
	if !model.City.IsNull() {
		out.City = model.City.ValueString()
	}
	if !model.State.IsNull() {
		out.State = model.State.ValueString()
	}
	if !model.Zipcode.IsNull() {
		out.Zipcode = model.Zipcode.ValueString()
	}
	if !model.Country.IsNull() {
		out.Country = model.Country.ValueString()
	}

	return out, diags
}

func expandCustomerMetadata(ctx context.Context, list types.List) ([]lago.CustomerMetadataInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var models []customerMetadataModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]lago.CustomerMetadataInput, 0, len(models))
	for _, m := range models {
		meta := lago.CustomerMetadataInput{
			Key:   m.Key.ValueString(),
			Value: m.Value.ValueString(),
		}
		if !m.DisplayInInvoice.IsNull() {
			meta.DisplayInInvoice = m.DisplayInInvoice.ValueBool()
		}
		out = append(out, meta)
	}

	return out, diags
}

// mapCustomerToModel converts a Lago API Customer response into the Terraform state model.
// The base model is used to preserve write-only fields (e.g. sync_with_provider) that are
// not echoed back by the API.
func mapCustomerToModel(ctx context.Context, customer *lago.Customer, base customerResourceModel) (customerResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := base

	state.ID = types.StringValue(customer.ExternalID)
	state.LagoID = types.StringValue(customer.LagoID.String())
	state.ExternalID = types.StringValue(customer.ExternalID)

	state.Name = stringOrNull(customer.Name)
	state.Email = stringOrNull(customer.Email)
	state.Phone = stringOrNull(customer.Phone)
	state.URL = stringOrNull(customer.URL)
	state.CustomerType = stringOrNull(customer.CustomerType)
	state.Currency = stringOrNull(string(customer.Currency))
	state.Timezone = stringOrNull(customer.Timezone)

	state.AddressLine1 = stringOrNull(customer.AddressLine1)
	state.AddressLine2 = stringOrNull(customer.AddressLine2)
	state.City = stringOrNull(customer.City)
	state.State = stringOrNull(customer.State)
	state.Zipcode = stringOrNull(customer.Zipcode)
	state.Country = stringOrNull(customer.Country)

	state.LegalName = stringOrNull(customer.LegalName)
	state.LegalNumber = stringOrNull(customer.LegalNumber)
	state.TaxIdentificationNumber = stringOrNull(customer.TaxIdentificationNumber)

	if customer.NetPaymentTerm == 0 {
		state.NetPaymentTerm = types.Int64Null()
	} else {
		state.NetPaymentTerm = types.Int64Value(int64(customer.NetPaymentTerm))
	}

	if customer.FinalizeZeroAmountInvoice == "" {
		state.FinalizeZeroAmountInvoice = types.StringNull()
	} else {
		state.FinalizeZeroAmountInvoice = types.StringValue(string(customer.FinalizeZeroAmountInvoice))
	}

	taxCodes, taxDiags := flattenStringSet(taxCodesFromTaxes(customer.Taxes))
	diags.Append(taxDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.TaxCodes = taxCodes

	billingCfg, bcDiags := flattenCustomerBillingConfiguration(ctx, customer.BillingConfiguration, base.BillingConfiguration)
	diags.Append(bcDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.BillingConfiguration = billingCfg

	shippingAddr, saDiags := flattenCustomerShippingAddress(customer.ShippingAddress)
	diags.Append(saDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.ShippingAddress = shippingAddr

	metadata, mdDiags := flattenCustomerMetadata(customer.Metadata)
	diags.Append(mdDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.Metadata = metadata

	if customer.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(customer.CreatedAt.Format(time.RFC3339))
	}

	if customer.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringNull()
	} else {
		state.UpdatedAt = types.StringValue(customer.UpdatedAt.Format(time.RFC3339))
	}

	return state, diags
}

// flattenCustomerBillingConfiguration converts the API billing configuration response into a
// Terraform object. The baseBillingConfig is used to preserve write-only fields
// (sync_with_provider) that are not returned by the API.
func flattenCustomerBillingConfiguration(ctx context.Context, cfg lago.CustomerBillingConfiguration, baseBillingConfig types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// If there is no meaningful API data and no base config, leave as null.
	isEmpty := cfg.PaymentProvider == "" &&
		cfg.PaymentProviderCode == "" &&
		cfg.ProviderCustomerID == "" &&
		cfg.DocumentLocale == "" &&
		cfg.InvoiceGracePeriod == 0 &&
		len(cfg.ProviderPaymentMethods) == 0

	if isEmpty && baseBillingConfig.IsNull() {
		return types.ObjectNull(customerBillingConfigObjectType().AttrTypes), diags
	}

	// Preserve sync_with_provider from plan/state (write-only; not in API response).
	// Default to null (not set) rather than false so Terraform doesn't see spurious drift.
	syncWithProvider := types.BoolNull()
	if !baseBillingConfig.IsNull() && !baseBillingConfig.IsUnknown() {
		var baseModel customerBillingConfigModel
		diags.Append(baseBillingConfig.As(ctx, &baseModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return types.ObjectNull(customerBillingConfigObjectType().AttrTypes), diags
		}
		if !baseModel.SyncWithProvider.IsNull() {
			syncWithProvider = baseModel.SyncWithProvider
		}
	}

	var providerPaymentMethods types.Set
	if len(cfg.ProviderPaymentMethods) > 0 {
		methods := make([]attr.Value, 0, len(cfg.ProviderPaymentMethods))
		for _, m := range cfg.ProviderPaymentMethods {
			methods = append(methods, types.StringValue(string(m)))
		}
		set, setDiags := types.SetValue(types.StringType, methods)
		diags.Append(setDiags...)
		if diags.HasError() {
			return types.ObjectNull(customerBillingConfigObjectType().AttrTypes), diags
		}
		providerPaymentMethods = set
	} else {
		providerPaymentMethods = types.SetNull(types.StringType)
	}

	var invoiceGracePeriod types.Int64
	if cfg.InvoiceGracePeriod == 0 {
		invoiceGracePeriod = types.Int64Null()
	} else {
		invoiceGracePeriod = types.Int64Value(int64(cfg.InvoiceGracePeriod))
	}

	obj, objDiags := types.ObjectValue(customerBillingConfigObjectType().AttrTypes, map[string]attr.Value{
		"invoice_grace_period":     invoiceGracePeriod,
		"payment_provider":         stringOrNull(string(cfg.PaymentProvider)),
		"payment_provider_code":    stringOrNull(cfg.PaymentProviderCode),
		"provider_customer_id":     stringOrNull(cfg.ProviderCustomerID),
		"sync_with_provider":       syncWithProvider,
		"document_locale":          stringOrNull(cfg.DocumentLocale),
		"provider_payment_methods": providerPaymentMethods,
	})
	diags.Append(objDiags...)
	return obj, diags
}

func flattenCustomerShippingAddress(addr lago.Address) (types.Object, diag.Diagnostics) {
	isEmpty := addr.AddressLine1 == "" &&
		addr.AddressLine2 == "" &&
		addr.City == "" &&
		addr.State == "" &&
		addr.Zipcode == "" &&
		addr.Country == ""

	if isEmpty {
		return types.ObjectNull(customerShippingAddressObjectType().AttrTypes), nil
	}

	obj, diags := types.ObjectValue(customerShippingAddressObjectType().AttrTypes, map[string]attr.Value{
		"address_line1": stringOrNull(addr.AddressLine1),
		"address_line2": stringOrNull(addr.AddressLine2),
		"city":          stringOrNull(addr.City),
		"state":         stringOrNull(addr.State),
		"zipcode":       stringOrNull(addr.Zipcode),
		"country":       stringOrNull(addr.Country),
	})
	return obj, diags
}

func flattenCustomerMetadata(items []lago.MetadataResponse) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(customerMetadataObjectType()), nil
	}

	var diags diag.Diagnostics
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		obj, objDiags := types.ObjectValue(customerMetadataObjectType().AttrTypes, map[string]attr.Value{
			"key":                types.StringValue(item.Key),
			"value":              types.StringValue(item.Value),
			"display_in_invoice": types.BoolValue(item.DisplayInInvoice),
		})
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ListNull(customerMetadataObjectType()), diags
		}
		values = append(values, obj)
	}

	list, listDiags := types.ListValue(customerMetadataObjectType(), values)
	diags.Append(listDiags...)
	return list, diags
}
