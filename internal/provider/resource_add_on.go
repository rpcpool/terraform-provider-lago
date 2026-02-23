package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &addOnResource{}
	_ resource.ResourceWithConfigure   = &addOnResource{}
	_ resource.ResourceWithImportState = &addOnResource{}
)

func NewAddOnResource() resource.Resource {
	return &addOnResource{}
}

type addOnResource struct {
	client *lago.Client
}

type addOnResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	LagoID             types.String `tfsdk:"lago_id"`
	Name               types.String `tfsdk:"name"`
	Code               types.String `tfsdk:"code"`
	Description        types.String `tfsdk:"description"`
	AmountCents        types.Int64  `tfsdk:"amount_cents"`
	AmountCurrency     types.String `tfsdk:"amount_currency"`
	InvoiceDisplayName types.String `tfsdk:"invoice_display_name"`
	TaxCodes           types.Set    `tfsdk:"tax_codes"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func (r *addOnResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_add_on"
}

func (r *addOnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago add-on.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the add-on code.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the add-on.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Add-on name.",
			},
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique add-on code. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Add-on description.",
			},
			"amount_cents": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Add-on amount in cents.",
			},
			"amount_currency": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Currency code for the add-on amount (e.g. `USD`).",
			},
			"invoice_display_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Display name shown on invoices.",
			},
			"tax_codes": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of tax codes to apply to the add-on.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Add-on creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *addOnResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *addOnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan addOnResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandAddOnInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.AddOn().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Add-On", lagoErr.Error())
		return
	}

	state, diags := mapAddOnToModel(created, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *addOnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state addOnResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addOn, lagoErr := r.client.AddOn().Get(ctx, state.Code.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Add-On", lagoErr.Error())
		return
	}

	newState, diags := mapAddOnToModel(addOn, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *addOnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan addOnResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandAddOnInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, lagoErr := r.client.AddOn().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Add-On", lagoErr.Error())
		return
	}

	state, diags := mapAddOnToModel(updated, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *addOnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state addOnResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.AddOn().Delete(ctx, state.Code.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Add-On", lagoErr.Error())
	}
}

func (r *addOnResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func expandAddOnInput(ctx context.Context, model addOnResourceModel) (*lago.AddOnInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &lago.AddOnInput{
		Name:           model.Name.ValueString(),
		Code:           model.Code.ValueString(),
		AmountCents:    int(model.AmountCents.ValueInt64()),
		AmountCurrency: lago.Currency(model.AmountCurrency.ValueString()),
	}

	if !model.Description.IsNull() {
		input.Description = model.Description.ValueString()
	}

	if !model.InvoiceDisplayName.IsNull() {
		input.InvoiceDisplayName = model.InvoiceDisplayName.ValueString()
	}

	taxCodes, taxDiags := expandStringSet(ctx, model.TaxCodes)
	diags.Append(taxDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.TaxCodes = taxCodes

	return input, diags
}

func mapAddOnToModel(addOn *lago.AddOn, base addOnResourceModel) (addOnResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := base

	state.ID = types.StringValue(addOn.Code)
	state.LagoID = types.StringValue(addOn.LagoID.String())
	state.Name = types.StringValue(addOn.Name)
	state.Code = types.StringValue(addOn.Code)
	state.AmountCents = types.Int64Value(int64(addOn.AmountCents))
	state.AmountCurrency = types.StringValue(string(addOn.AmountCurrency))
	state.Description = stringOrNull(addOn.Description)
	state.InvoiceDisplayName = stringOrNull(addOn.InvoiceDisplayName)

	taxCodes, taxDiags := flattenStringSet(taxCodesFromTaxes(addOn.Taxes))
	diags.Append(taxDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.TaxCodes = taxCodes

	if addOn.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(addOn.CreatedAt.Format(time.RFC3339))
	}

	return state, diags
}
