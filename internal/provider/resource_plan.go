package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/triton-one/terraform-provider-lago/internal/client"
)

var (
	_ resource.Resource                = &planResource{}
	_ resource.ResourceWithConfigure   = &planResource{}
	_ resource.ResourceWithImportState = &planResource{}
)

var allowedPlanIntervals = []string{
	"weekly",
	"monthly",
	"quarterly",
	"yearly",
}

func NewPlanResource() resource.Resource {
	return &planResource{}
}

type planResource struct {
	client *client.Client
}

type planResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	LagoID                  types.String `tfsdk:"lago_id"`
	Name                    types.String `tfsdk:"name"`
	Code                    types.String `tfsdk:"code"`
	Interval                types.String `tfsdk:"interval"`
	Description             types.String `tfsdk:"description"`
	AmountCents             types.Int64  `tfsdk:"amount_cents"`
	AmountCurrency          types.String `tfsdk:"amount_currency"`
	TrialPeriod             types.Int64  `tfsdk:"trial_period"`
	PayInAdvance            types.Bool   `tfsdk:"pay_in_advance"`
	BillChargesMonthly      types.Bool   `tfsdk:"bill_charges_monthly"`
	BillFixedChargesMonthly types.Bool   `tfsdk:"bill_fixed_charges_monthly"`
	InvoiceDisplayName      types.String `tfsdk:"invoice_display_name"`
	TaxCodes                types.Set    `tfsdk:"tax_codes"`
	Metadata                types.Map    `tfsdk:"metadata"`
	Charges                 types.List   `tfsdk:"charges"`
	MinimumCommitment       types.Object `tfsdk:"minimum_commitment"`
	FixedCharges            types.List   `tfsdk:"fixed_charges"`
	UsageThresholds         types.List   `tfsdk:"usage_thresholds"`
	Entitlements            types.List   `tfsdk:"entitlements"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

type planChargeModel struct {
	BillableMetricID   types.String `tfsdk:"billable_metric_id"`
	ChargeModel        types.String `tfsdk:"charge_model"`
	Invoiceable        types.Bool   `tfsdk:"invoiceable"`
	InvoiceDisplayName types.String `tfsdk:"invoice_display_name"`
	PayInAdvance       types.Bool   `tfsdk:"pay_in_advance"`
	RegroupPaidFees    types.Bool   `tfsdk:"regroup_paid_fees"`
	Prorated           types.Bool   `tfsdk:"prorated"`
	MinAmountCents     types.Int64  `tfsdk:"min_amount_cents"`
	PropertiesJSON     types.String `tfsdk:"properties_json"`
	FiltersJSON        types.String `tfsdk:"filters_json"`
	TaxCodes           types.Set    `tfsdk:"tax_codes"`
}

type planMinimumCommitmentModel struct {
	AmountCents        types.Int64  `tfsdk:"amount_cents"`
	InvoiceDisplayName types.String `tfsdk:"invoice_display_name"`
	TaxCodes           types.Set    `tfsdk:"tax_codes"`
}

type planFixedChargeModel struct {
	AddOnID            types.String `tfsdk:"add_on_id"`
	AddOnCode          types.String `tfsdk:"add_on_code"`
	ChargeModel        types.String `tfsdk:"charge_model"`
	InvoiceDisplayName types.String `tfsdk:"invoice_display_name"`
	PayInAdvance       types.Bool   `tfsdk:"pay_in_advance"`
	Prorated           types.Bool   `tfsdk:"prorated"`
	Units              types.Int64  `tfsdk:"units"`
	PropertiesJSON     types.String `tfsdk:"properties_json"`
	TaxCodes           types.Set    `tfsdk:"tax_codes"`
}

type planUsageThresholdModel struct {
	AmountCents          types.Int64  `tfsdk:"amount_cents"`
	ThresholdDisplayName types.String `tfsdk:"threshold_display_name"`
	Recurring            types.Bool   `tfsdk:"recurring"`
	PropertiesJSON       types.String `tfsdk:"properties_json"`
}

type planEntitlementModel struct {
	Code           types.String `tfsdk:"code"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Recurring      types.Bool   `tfsdk:"recurring"`
	PrivilegesJSON types.String `tfsdk:"privileges_json"`
}

func (r *planResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plan"
}

func (r *planResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago plan.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Terraform resource ID, set to the plan code."},
			"lago_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Lago internal identifier."},
			"name":    schema.StringAttribute{Required: true, MarkdownDescription: "Plan name."},
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique plan code.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interval": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Plan billing interval.",
				Validators:          []validator.String{stringvalidator.OneOf(allowedPlanIntervals...)},
			},
			"description":                schema.StringAttribute{Optional: true},
			"amount_cents":               schema.Int64Attribute{Required: true},
			"amount_currency":            schema.StringAttribute{Required: true},
			"trial_period":               schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
			"pay_in_advance":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"bill_charges_monthly":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"bill_fixed_charges_monthly": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"invoice_display_name":       schema.StringAttribute{Optional: true},
			"tax_codes":                  schema.SetAttribute{Optional: true, ElementType: types.StringType},
			"metadata":                   schema.MapAttribute{Optional: true, ElementType: types.StringType},
			"charges": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"billable_metric_id":   schema.StringAttribute{Required: true},
					"charge_model":         schema.StringAttribute{Required: true},
					"invoiceable":          schema.BoolAttribute{Optional: true},
					"invoice_display_name": schema.StringAttribute{Optional: true},
					"pay_in_advance":       schema.BoolAttribute{Optional: true},
					"regroup_paid_fees":    schema.BoolAttribute{Optional: true},
					"prorated":             schema.BoolAttribute{Optional: true},
					"min_amount_cents":     schema.Int64Attribute{Optional: true},
					"properties_json":      schema.StringAttribute{Optional: true},
					"filters_json":         schema.StringAttribute{Optional: true, MarkdownDescription: "JSON array of charge filters."},
					"tax_codes":            schema.SetAttribute{Optional: true, ElementType: types.StringType},
				}},
			},
			"minimum_commitment": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"amount_cents":         schema.Int64Attribute{Required: true},
					"invoice_display_name": schema.StringAttribute{Optional: true},
					"tax_codes":            schema.SetAttribute{Optional: true, ElementType: types.StringType},
				},
			},
			"fixed_charges": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"add_on_id":            schema.StringAttribute{Optional: true},
					"add_on_code":          schema.StringAttribute{Optional: true},
					"charge_model":         schema.StringAttribute{Optional: true},
					"invoice_display_name": schema.StringAttribute{Optional: true},
					"pay_in_advance":       schema.BoolAttribute{Optional: true},
					"prorated":             schema.BoolAttribute{Optional: true},
					"units":                schema.Int64Attribute{Optional: true},
					"properties_json":      schema.StringAttribute{Optional: true},
					"tax_codes":            schema.SetAttribute{Optional: true, ElementType: types.StringType},
				}},
			},
			"usage_thresholds": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"amount_cents":           schema.Int64Attribute{Optional: true},
					"threshold_display_name": schema.StringAttribute{Optional: true},
					"recurring":              schema.BoolAttribute{Optional: true},
					"properties_json":        schema.StringAttribute{Optional: true},
				}},
			},
			"entitlements": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"code":            schema.StringAttribute{Required: true},
					"name":            schema.StringAttribute{Optional: true},
					"description":     schema.StringAttribute{Optional: true},
					"recurring":       schema.BoolAttribute{Optional: true},
					"privileges_json": schema.StringAttribute{Optional: true},
				}},
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *planResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*lagoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *lagoProviderData, got: %T.", req.ProviderData))
		return
	}

	r.client = providerData.client
}

func (r *planResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan planResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandCreatePlanInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePlan(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Lago Plan", err.Error())
		return
	}

	state, mapDiags := mapPlanToModel(created, plan)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *planResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state planResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan, err := r.client.GetPlanByCode(ctx, state.Code.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Plan", err.Error())
		return
	}

	newState, mapDiags := mapPlanToModel(plan, state)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *planResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan planResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandUpdatePlanInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdatePlanByCode(ctx, plan.Code.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Lago Plan", err.Error())
		return
	}

	state, mapDiags := mapPlanToModel(updated, plan)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *planResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state planResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePlanByCode(ctx, state.Code.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Lago Plan", err.Error())
	}
}

func (r *planResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func expandCreatePlanInput(ctx context.Context, plan planResourceModel) (client.CreatePlanInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := client.CreatePlanInput{
		Name:           plan.Name.ValueString(),
		Code:           plan.Code.ValueString(),
		Interval:       plan.Interval.ValueString(),
		AmountCents:    plan.AmountCents.ValueInt64(),
		AmountCurrency: plan.AmountCurrency.ValueString(),
	}

	if !plan.Description.IsNull() {
		d := plan.Description.ValueString()
		input.Description = &d
	}
	if !plan.TrialPeriod.IsNull() {
		t := plan.TrialPeriod.ValueInt64()
		input.TrialPeriod = &t
	}
	if !plan.PayInAdvance.IsNull() {
		v := plan.PayInAdvance.ValueBool()
		input.PayInAdvance = &v
	}
	if !plan.BillChargesMonthly.IsNull() {
		v := plan.BillChargesMonthly.ValueBool()
		input.BillChargesMonthly = &v
	}
	if !plan.BillFixedChargesMonthly.IsNull() {
		v := plan.BillFixedChargesMonthly.ValueBool()
		input.BillFixedChargesMonthly = &v
	}
	if !plan.InvoiceDisplayName.IsNull() {
		v := plan.InvoiceDisplayName.ValueString()
		input.InvoiceDisplayName = &v
	}

	taxCodes, tdiags := expandStringSet(ctx, plan.TaxCodes)
	diags.Append(tdiags...)
	input.TaxCodes = taxCodes

	metadata, mdiags := expandStringMap(ctx, plan.Metadata)
	diags.Append(mdiags...)
	input.Metadata = metadata

	charges, cdiags := expandCharges(ctx, plan.Charges)
	diags.Append(cdiags...)
	input.Charges = charges

	minimumCommitment, mcdiags := expandMinimumCommitment(ctx, plan.MinimumCommitment)
	diags.Append(mcdiags...)
	input.MinimumCommitment = minimumCommitment

	fixedCharges, fcdiags := expandFixedCharges(ctx, plan.FixedCharges)
	diags.Append(fcdiags...)
	input.FixedCharges = fixedCharges

	usageThresholds, udiags := expandUsageThresholds(ctx, plan.UsageThresholds)
	diags.Append(udiags...)
	input.UsageThresholds = usageThresholds

	entitlements, ediags := expandEntitlements(ctx, plan.Entitlements)
	diags.Append(ediags...)
	input.Entitlements = entitlements

	return input, diags
}

func expandUpdatePlanInput(ctx context.Context, plan planResourceModel) (client.UpdatePlanInput, diag.Diagnostics) {
	createInput, diags := expandCreatePlanInput(ctx, plan)

	input := client.UpdatePlanInput{}
	name := createInput.Name
	input.Name = &name
	description := createInput.Description
	input.Description = description
	interval := createInput.Interval
	input.Interval = &interval
	amountCents := createInput.AmountCents
	input.AmountCents = &amountCents
	amountCurrency := createInput.AmountCurrency
	input.AmountCurrency = &amountCurrency
	input.TrialPeriod = createInput.TrialPeriod
	input.PayInAdvance = createInput.PayInAdvance
	input.BillChargesMonthly = createInput.BillChargesMonthly
	input.BillFixedChargesMonthly = createInput.BillFixedChargesMonthly
	input.InvoiceDisplayName = createInput.InvoiceDisplayName
	input.TaxCodes = createInput.TaxCodes
	input.Metadata = createInput.Metadata
	input.Charges = createInput.Charges
	input.MinimumCommitment = createInput.MinimumCommitment
	input.FixedCharges = createInput.FixedCharges
	input.UsageThresholds = createInput.UsageThresholds
	input.Entitlements = createInput.Entitlements

	return input, diags
}

func mapPlanToModel(plan *client.Plan, base planResourceModel) (planResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := base
	state.ID = types.StringValue(plan.Code)
	state.LagoID = stringOrNull(plan.LagoID)
	state.Name = types.StringValue(plan.Name)
	state.Code = types.StringValue(plan.Code)
	state.Interval = types.StringValue(plan.Interval)
	state.Description = stringOrNull(plan.Description)
	state.AmountCents = types.Int64Value(plan.AmountCents)
	state.AmountCurrency = stringOrNull(plan.AmountCurrency)
	state.TrialPeriod = types.Int64Value(int64(plan.TrialPeriod))
	state.PayInAdvance = types.BoolValue(plan.PayInAdvance)
	state.BillChargesMonthly = types.BoolValue(plan.BillChargesMonthly)
	state.BillFixedChargesMonthly = types.BoolValue(plan.BillFixedChargesMonthly)
	state.InvoiceDisplayName = stringOrNull(plan.InvoiceDisplayName)
	state.CreatedAt = stringOrNull(plan.CreatedAt)
	state.UpdatedAt = stringOrNull(plan.UpdatedAt)

	topLevelTaxCodes := plan.TaxCodes
	if len(topLevelTaxCodes) == 0 {
		topLevelTaxCodes = taxCodesFromTaxes(plan.Taxes)
	}
	taxCodes, tdiags := flattenStringSet(topLevelTaxCodes)
	diags.Append(tdiags...)
	state.TaxCodes = taxCodes

	metadata, mdiags := flattenStringMap(plan.Metadata)
	diags.Append(mdiags...)
	state.Metadata = metadata

	charges, cdiags := flattenCharges(plan.Charges)
	diags.Append(cdiags...)
	if !base.Charges.IsNull() && !base.Charges.IsUnknown() {
		// Lago may reorder/normalize charge payloads in create/update responses.
		// Preserve configured charge blocks in state to avoid inconsistent apply results.
		state.Charges = base.Charges
	} else {
		state.Charges = charges
	}

	minimumCommitment, mcdiags := flattenMinimumCommitment(plan.MinimumCommitment)
	diags.Append(mcdiags...)
	state.MinimumCommitment = minimumCommitment

	fixedCharges, fcdiags := flattenFixedCharges(plan.FixedCharges)
	diags.Append(fcdiags...)
	state.FixedCharges = fixedCharges

	usageThresholds, udiags := flattenUsageThresholds(plan.UsageThresholds)
	diags.Append(udiags...)
	state.UsageThresholds = usageThresholds

	entitlements, ediags := flattenEntitlements(plan.Entitlements)
	diags.Append(ediags...)
	state.Entitlements = entitlements

	return state, diags
}

func stringOrNull(value string) types.String {
	if strings.TrimSpace(value) == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func expandStringSet(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid set", "Set contains unknown values")
		return nil, diags
	}

	var values []string
	diags.Append(value.ElementsAs(ctx, &values, false)...)
	return values, diags
}

func flattenStringSet(values []string) (types.Set, diag.Diagnostics) {
	if len(values) == 0 {
		return types.SetNull(types.StringType), nil
	}

	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	set, diags := types.SetValue(types.StringType, elements)
	return set, diags
}

func expandStringMap(ctx context.Context, value types.Map) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid map", "Map contains unknown values")
		return nil, diags
	}

	out := map[string]string{}
	diags.Append(value.ElementsAs(ctx, &out, false)...)
	if len(out) == 0 {
		return nil, diags
	}
	return out, diags
}

func flattenStringMap(values map[string]string) (types.Map, diag.Diagnostics) {
	if len(values) == 0 {
		return types.MapNull(types.StringType), nil
	}

	attrs := make(map[string]attr.Value, len(values))
	for key, value := range values {
		attrs[key] = types.StringValue(value)
	}

	mapValue, diags := types.MapValue(types.StringType, attrs)
	return mapValue, diags
}

func parseJSONField(value types.String) (json.RawMessage, error) {
	if value.IsNull() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, nil
	}

	var out any
	if err := json.Unmarshal([]byte(value.ValueString()), &out); err != nil {
		return nil, err
	}
	return json.RawMessage([]byte(value.ValueString())), nil
}

func jsonFieldValue(value json.RawMessage) types.String {
	if len(value) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(value))
}

func chargeFiltersJSONValue(filters []client.PlanChargeFilter) types.String {
	if len(filters) == 0 {
		return types.StringNull()
	}
	raw, err := json.Marshal(filters)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func chargeObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"billable_metric_id":   types.StringType,
		"charge_model":         types.StringType,
		"invoiceable":          types.BoolType,
		"invoice_display_name": types.StringType,
		"pay_in_advance":       types.BoolType,
		"regroup_paid_fees":    types.BoolType,
		"prorated":             types.BoolType,
		"min_amount_cents":     types.Int64Type,
		"properties_json":      types.StringType,
		"filters_json":         types.StringType,
		"tax_codes":            types.SetType{ElemType: types.StringType},
	}}
}

func expandCharges(ctx context.Context, value types.List) ([]client.PlanCharge, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid charges value", "`charges` contains unknown values.")
		return nil, diags
	}

	var models []planChargeModel
	diags.Append(value.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]client.PlanCharge, 0, len(models))
	for _, m := range models {
		charge := client.PlanCharge{
			BillableMetricID: m.BillableMetricID.ValueString(),
			ChargeModel:      m.ChargeModel.ValueString(),
		}
		if !m.Invoiceable.IsNull() {
			v := m.Invoiceable.ValueBool()
			charge.Invoiceable = &v
		}
		if !m.InvoiceDisplayName.IsNull() {
			v := m.InvoiceDisplayName.ValueString()
			charge.InvoiceDisplayName = &v
		}
		if !m.PayInAdvance.IsNull() {
			v := m.PayInAdvance.ValueBool()
			charge.PayInAdvance = &v
		}
		if !m.RegroupPaidFees.IsNull() {
			v := m.RegroupPaidFees.ValueBool()
			charge.RegroupPaidFees = &v
		}
		if !m.Prorated.IsNull() {
			v := m.Prorated.ValueBool()
			charge.Prorated = &v
		}
		if !m.MinAmountCents.IsNull() {
			v := m.MinAmountCents.ValueInt64()
			charge.MinAmountCents = &v
		}
		if !m.PropertiesJSON.IsNull() {
			raw, err := parseJSONField(m.PropertiesJSON)
			if err != nil {
				diags.AddError("Invalid charge properties_json", err.Error())
				return nil, diags
			}
			charge.Properties = raw
		}
		if !m.FiltersJSON.IsNull() {
			var filters []client.PlanChargeFilter
			if err := json.Unmarshal([]byte(m.FiltersJSON.ValueString()), &filters); err != nil {
				diags.AddError("Invalid charge filters_json", err.Error())
				return nil, diags
			}
			charge.Filters = filters
		}
		taxCodes, tdiags := expandStringSet(ctx, m.TaxCodes)
		diags.Append(tdiags...)
		if diags.HasError() {
			return nil, diags
		}
		charge.TaxCodes = taxCodes
		out = append(out, charge)
	}

	return out, diags
}

func flattenCharges(charges []client.PlanCharge) (types.List, diag.Diagnostics) {
	if len(charges) == 0 {
		return types.ListNull(chargeObjectType()), nil
	}

	values := make([]attr.Value, 0, len(charges))
	var diags diag.Diagnostics
	for _, charge := range charges {
		chargeTaxCodes := charge.TaxCodes
		if len(chargeTaxCodes) == 0 {
			chargeTaxCodes = taxCodesFromTaxes(charge.Taxes)
		}
		taxCodes, tdiags := flattenStringSet(chargeTaxCodes)
		diags.Append(tdiags...)
		if diags.HasError() {
			return types.ListNull(chargeObjectType()), diags
		}
		billableMetricID := charge.BillableMetricID
		if strings.TrimSpace(billableMetricID) == "" {
			billableMetricID = charge.LagoBillableMetricID
		}
		attrs := map[string]attr.Value{
			"billable_metric_id":   stringOrNull(billableMetricID),
			"charge_model":         stringOrNull(charge.ChargeModel),
			"invoiceable":          boolPtrOrNull(charge.Invoiceable),
			"invoice_display_name": stringPtrOrNull(charge.InvoiceDisplayName),
			"pay_in_advance":       boolPtrOrNull(charge.PayInAdvance),
			"regroup_paid_fees":    boolPtrOrNull(charge.RegroupPaidFees),
			"prorated":             boolPtrOrNull(charge.Prorated),
			"min_amount_cents":     int64PtrOrNull(charge.MinAmountCents),
			"properties_json":      jsonFieldValue(charge.Properties),
			"filters_json":         chargeFiltersJSONValue(charge.Filters),
			"tax_codes":            taxCodes,
		}
		values = append(values, types.ObjectValueMust(chargeObjectType().AttrTypes, attrs))
	}

	list, ldiags := types.ListValue(chargeObjectType(), values)
	diags.Append(ldiags...)
	return list, diags
}

func minimumCommitmentObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"amount_cents":         types.Int64Type,
		"invoice_display_name": types.StringType,
		"tax_codes":            types.SetType{ElemType: types.StringType},
	}}
}

func expandMinimumCommitment(ctx context.Context, value types.Object) (*client.PlanMinimumCommitment, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid minimum_commitment", "`minimum_commitment` contains unknown values")
		return nil, diags
	}

	var model planMinimumCommitmentModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	out := &client.PlanMinimumCommitment{AmountCents: model.AmountCents.ValueInt64()}
	if !model.InvoiceDisplayName.IsNull() {
		v := model.InvoiceDisplayName.ValueString()
		out.InvoiceDisplayName = &v
	}
	taxCodes, tdiags := expandStringSet(ctx, model.TaxCodes)
	diags.Append(tdiags...)
	if diags.HasError() {
		return nil, diags
	}
	out.TaxCodes = taxCodes
	return out, diags
}

func flattenMinimumCommitment(value *client.PlanMinimumCommitment) (types.Object, diag.Diagnostics) {
	if value == nil {
		return types.ObjectNull(minimumCommitmentObjectType().AttrTypes), nil
	}

	minimumCommitmentTaxCodes := value.TaxCodes
	if len(minimumCommitmentTaxCodes) == 0 {
		minimumCommitmentTaxCodes = taxCodesFromTaxes(value.Taxes)
	}
	taxCodes, diags := flattenStringSet(minimumCommitmentTaxCodes)
	if diags.HasError() {
		return types.ObjectNull(minimumCommitmentObjectType().AttrTypes), diags
	}

	obj, odiags := types.ObjectValue(minimumCommitmentObjectType().AttrTypes, map[string]attr.Value{
		"amount_cents":         types.Int64Value(value.AmountCents),
		"invoice_display_name": stringPtrOrNull(value.InvoiceDisplayName),
		"tax_codes":            taxCodes,
	})
	diags.Append(odiags...)
	return obj, diags
}

func fixedChargeObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"add_on_id":            types.StringType,
		"add_on_code":          types.StringType,
		"charge_model":         types.StringType,
		"invoice_display_name": types.StringType,
		"pay_in_advance":       types.BoolType,
		"prorated":             types.BoolType,
		"units":                types.Int64Type,
		"properties_json":      types.StringType,
		"tax_codes":            types.SetType{ElemType: types.StringType},
	}}
}

func expandFixedCharges(ctx context.Context, value types.List) ([]client.PlanFixedCharge, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid fixed_charges", "`fixed_charges` contains unknown values")
		return nil, diags
	}

	var models []planFixedChargeModel
	diags.Append(value.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]client.PlanFixedCharge, 0, len(models))
	for _, m := range models {
		fc := client.PlanFixedCharge{}
		if !m.AddOnID.IsNull() {
			fc.AddOnID = m.AddOnID.ValueString()
		}
		if !m.AddOnCode.IsNull() {
			fc.AddOnCode = m.AddOnCode.ValueString()
		}
		if !m.ChargeModel.IsNull() {
			fc.ChargeModel = m.ChargeModel.ValueString()
		}
		if !m.InvoiceDisplayName.IsNull() {
			v := m.InvoiceDisplayName.ValueString()
			fc.InvoiceDisplayName = &v
		}
		if !m.PayInAdvance.IsNull() {
			v := m.PayInAdvance.ValueBool()
			fc.PayInAdvance = &v
		}
		if !m.Prorated.IsNull() {
			v := m.Prorated.ValueBool()
			fc.Prorated = &v
		}
		if !m.Units.IsNull() {
			v := m.Units.ValueInt64()
			fc.Units = &v
		}
		if !m.PropertiesJSON.IsNull() {
			raw, err := parseJSONField(m.PropertiesJSON)
			if err != nil {
				diags.AddError("Invalid fixed_charge properties_json", err.Error())
				return nil, diags
			}
			fc.Properties = raw
		}
		taxCodes, tdiags := expandStringSet(ctx, m.TaxCodes)
		diags.Append(tdiags...)
		if diags.HasError() {
			return nil, diags
		}
		fc.TaxCodes = taxCodes
		out = append(out, fc)
	}

	return out, diags
}

func flattenFixedCharges(values []client.PlanFixedCharge) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(fixedChargeObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	var diags diag.Diagnostics
	for _, v := range values {
		fixedChargeTaxCodes := v.TaxCodes
		if len(fixedChargeTaxCodes) == 0 {
			fixedChargeTaxCodes = taxCodesFromTaxes(v.Taxes)
		}
		taxCodes, tdiags := flattenStringSet(fixedChargeTaxCodes)
		diags.Append(tdiags...)
		if diags.HasError() {
			return types.ListNull(fixedChargeObjectType()), diags
		}
		addOnID := v.AddOnID
		if strings.TrimSpace(addOnID) == "" {
			addOnID = v.LagoAddOnID
		}
		out = append(out, types.ObjectValueMust(fixedChargeObjectType().AttrTypes, map[string]attr.Value{
			"add_on_id":            stringOrNull(addOnID),
			"add_on_code":          stringOrNull(v.AddOnCode),
			"charge_model":         stringOrNull(v.ChargeModel),
			"invoice_display_name": stringPtrOrNull(v.InvoiceDisplayName),
			"pay_in_advance":       boolPtrOrNull(v.PayInAdvance),
			"prorated":             boolPtrOrNull(v.Prorated),
			"units":                int64PtrOrNull(v.Units),
			"properties_json":      jsonFieldValue(v.Properties),
			"tax_codes":            taxCodes,
		}))
	}

	list, ldiags := types.ListValue(fixedChargeObjectType(), out)
	diags.Append(ldiags...)
	return list, diags
}

func usageThresholdObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"amount_cents":           types.Int64Type,
		"threshold_display_name": types.StringType,
		"recurring":              types.BoolType,
		"properties_json":        types.StringType,
	}}
}

func expandUsageThresholds(ctx context.Context, value types.List) ([]client.PlanUsageThreshold, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid usage_thresholds", "`usage_thresholds` contains unknown values")
		return nil, diags
	}

	var models []planUsageThresholdModel
	diags.Append(value.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]client.PlanUsageThreshold, 0, len(models))
	for _, m := range models {
		ut := client.PlanUsageThreshold{}
		if !m.AmountCents.IsNull() {
			v := m.AmountCents.ValueInt64()
			ut.AmountCents = &v
		}
		if !m.ThresholdDisplayName.IsNull() {
			v := m.ThresholdDisplayName.ValueString()
			ut.ThresholdDisplayName = &v
		}
		if !m.Recurring.IsNull() {
			v := m.Recurring.ValueBool()
			ut.Recurring = &v
		}
		if !m.PropertiesJSON.IsNull() {
			raw, err := parseJSONField(m.PropertiesJSON)
			if err != nil {
				diags.AddError("Invalid usage_threshold properties_json", err.Error())
				return nil, diags
			}
			ut.Properties = raw
		}
		out = append(out, ut)
	}

	return out, diags
}

func flattenUsageThresholds(values []client.PlanUsageThreshold) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(usageThresholdObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	for _, v := range values {
		out = append(out, types.ObjectValueMust(usageThresholdObjectType().AttrTypes, map[string]attr.Value{
			"amount_cents":           int64PtrOrNull(v.AmountCents),
			"threshold_display_name": stringPtrOrNull(v.ThresholdDisplayName),
			"recurring":              boolPtrOrNull(v.Recurring),
			"properties_json":        jsonFieldValue(v.Properties),
		}))
	}

	list, diags := types.ListValue(usageThresholdObjectType(), out)
	return list, diags
}

func entitlementObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"code":            types.StringType,
		"name":            types.StringType,
		"description":     types.StringType,
		"recurring":       types.BoolType,
		"privileges_json": types.StringType,
	}}
}

func expandEntitlements(ctx context.Context, value types.List) ([]client.PlanEntitlement, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() {
		return nil, diags
	}
	if value.IsUnknown() {
		diags.AddError("Invalid entitlements", "`entitlements` contains unknown values")
		return nil, diags
	}

	var models []planEntitlementModel
	diags.Append(value.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]client.PlanEntitlement, 0, len(models))
	for _, m := range models {
		e := client.PlanEntitlement{Code: m.Code.ValueString()}
		if !m.Name.IsNull() {
			v := m.Name.ValueString()
			e.Name = &v
		}
		if !m.Description.IsNull() {
			v := m.Description.ValueString()
			e.Description = &v
		}
		if !m.Recurring.IsNull() {
			v := m.Recurring.ValueBool()
			e.Recurring = &v
		}
		if !m.PrivilegesJSON.IsNull() {
			raw, err := parseJSONField(m.PrivilegesJSON)
			if err != nil {
				diags.AddError("Invalid entitlement privileges_json", err.Error())
				return nil, diags
			}
			e.Privileges = raw
		}
		out = append(out, e)
	}
	return out, diags
}

func flattenEntitlements(values []client.PlanEntitlement) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(entitlementObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	for _, v := range values {
		out = append(out, types.ObjectValueMust(entitlementObjectType().AttrTypes, map[string]attr.Value{
			"code":            stringOrNull(v.Code),
			"name":            stringPtrOrNull(v.Name),
			"description":     stringPtrOrNull(v.Description),
			"recurring":       boolPtrOrNull(v.Recurring),
			"privileges_json": jsonFieldValue(v.Privileges),
		}))
	}
	list, diags := types.ListValue(entitlementObjectType(), out)
	return list, diags
}

func boolPtrOrNull(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func stringPtrOrNull(value *string) types.String {
	if value == nil || strings.TrimSpace(*value) == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func int64PtrOrNull(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func taxCodesFromTaxes(taxes []client.Tax) []string {
	if len(taxes) == 0 {
		return nil
	}
	out := make([]string, 0, len(taxes))
	for _, tax := range taxes {
		if strings.TrimSpace(tax.Code) != "" {
			out = append(out, tax.Code)
		}
	}
	return out
}
