package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	lago "github.com/getlago/lago-go-client"
	"github.com/google/uuid"
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
	client *lago.Client
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
					"billable_metric_id":   schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the billable metric."},
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
					"add_on_id":            schema.StringAttribute{Optional: true, MarkdownDescription: "UUID of the add-on."},
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
				}},
			},
			"entitlements": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"code":        schema.StringAttribute{Computed: true},
					"name":        schema.StringAttribute{Computed: true},
					"description": schema.StringAttribute{Computed: true},
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

	input, diags := expandPlanInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.Plan().Create(ctx, &input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Plan", lagoErr.Error())
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

	plan, lagoErr := r.client.Plan().Get(ctx, state.Code.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Plan", lagoErr.Error())
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

	input, diags := expandPlanInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, lagoErr := r.client.Plan().Update(ctx, &input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Plan", lagoErr.Error())
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

	_, lagoErr := r.client.Plan().Delete(ctx, state.Code.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Plan", lagoErr.Error())
	}
}

func (r *planResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func expandPlanInput(ctx context.Context, plan planResourceModel) (lago.PlanInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := lago.PlanInput{
		Name:               plan.Name.ValueString(),
		Code:               plan.Code.ValueString(),
		Interval:           lago.PlanInterval(plan.Interval.ValueString()),
		AmountCents:        int(plan.AmountCents.ValueInt64()),
		AmountCurrency:     lago.Currency(plan.AmountCurrency.ValueString()),
		PayInAdvance:       plan.PayInAdvance.ValueBool(),
		BillChargesMonthly: plan.BillChargesMonthly.ValueBool(),
		TrialPeriod:        float32(plan.TrialPeriod.ValueInt64()),
	}

	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.InvoiceDisplayName.IsNull() {
		input.InvoiceDisplayName = plan.InvoiceDisplayName.ValueString()
	}
	if !plan.BillFixedChargesMonthly.IsNull() {
		v := plan.BillFixedChargesMonthly.ValueBool()
		input.BillFixedChargesMonthly = &v
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

	return input, diags
}

func mapPlanToModel(p *lago.Plan, base planResourceModel) (planResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := base
	state.ID = types.StringValue(p.Code)
	state.LagoID = types.StringValue(p.LagoID.String())
	state.Name = types.StringValue(p.Name)
	state.Code = types.StringValue(p.Code)
	state.Interval = types.StringValue(string(p.Interval))
	state.Description = stringOrNull(p.Description)
	state.AmountCents = types.Int64Value(int64(p.AmountCents))
	state.AmountCurrency = types.StringValue(string(p.AmountCurrency))
	state.PayInAdvance = types.BoolValue(p.PayInAdvance)
	state.BillChargesMonthly = types.BoolValue(p.BillChargesMonthly)
	state.InvoiceDisplayName = stringOrNull(p.InvoiceDisplayName)

	if p.BillFixedChargesMonthly != nil {
		state.BillFixedChargesMonthly = types.BoolValue(*p.BillFixedChargesMonthly)
	} else {
		state.BillFixedChargesMonthly = types.BoolValue(false)
	}

	// lago.Plan has no TrialPeriod field — preserve from prior state.
	// On import the base is empty, so default to 0 (matching the schema default).
	if state.TrialPeriod.IsNull() || state.TrialPeriod.IsUnknown() {
		state.TrialPeriod = types.Int64Value(0)
	}

	// lago.Plan has no CreatedAt/UpdatedAt fields
	state.CreatedAt = types.StringNull()
	state.UpdatedAt = types.StringNull()

	taxCodes, tdiags := flattenStringSet(taxCodesFromTaxes(p.Taxes))
	diags.Append(tdiags...)
	state.TaxCodes = taxCodes

	metadata, mdiags := flattenStringMap(p.Metadata)
	diags.Append(mdiags...)
	state.Metadata = metadata

	charges, cdiags := flattenCharges(p.Charges)
	diags.Append(cdiags...)
	if !base.Charges.IsNull() && !base.Charges.IsUnknown() {
		// Lago may reorder/normalize charge payloads in create/update responses.
		// Preserve configured charge blocks in state to avoid inconsistent apply results.
		state.Charges = base.Charges
	} else {
		state.Charges = charges
	}

	minimumCommitment, mcdiags := flattenMinimumCommitment(p.MinimumCommitment)
	diags.Append(mcdiags...)
	state.MinimumCommitment = minimumCommitment

	fixedCharges, fcdiags := flattenFixedCharges(p.FixedCharges)
	diags.Append(fcdiags...)
	state.FixedCharges = fixedCharges

	usageThresholds, udiags := flattenUsageThresholds(p.UsageThresholds)
	diags.Append(udiags...)
	state.UsageThresholds = usageThresholds

	entitlements, ediags := flattenEntitlements(p.Entitlements)
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

func parseJSONField(value types.String) (map[string]interface{}, error) {
	if value.IsNull() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, nil
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(value.ValueString()), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseJSONFiltersField(value types.String) ([]lago.ChargeFilter, error) {
	if value.IsNull() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, nil
	}

	var out []lago.ChargeFilter
	if err := json.Unmarshal([]byte(value.ValueString()), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func jsonPropertiesValue(props map[string]interface{}) types.String {
	if len(props) == 0 {
		return types.StringNull()
	}
	raw, err := json.Marshal(props)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func chargeFiltersJSONValue(filters []lago.ChargeFilter) types.String {
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

func expandCharges(ctx context.Context, value types.List) ([]lago.PlanChargeInput, diag.Diagnostics) {
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

	out := make([]lago.PlanChargeInput, 0, len(models))
	for _, m := range models {
		bmID, err := uuid.Parse(m.BillableMetricID.ValueString())
		if err != nil {
			diags.AddError("Invalid billable_metric_id", fmt.Sprintf("Must be a valid UUID: %s", err))
			return nil, diags
		}
		charge := lago.PlanChargeInput{
			BillableMetricID: bmID,
			ChargeModel:      lago.ChargeModel(m.ChargeModel.ValueString()),
		}
		if !m.Invoiceable.IsNull() {
			charge.Invoiceable = m.Invoiceable.ValueBool()
		}
		if !m.PayInAdvance.IsNull() {
			charge.PayInAdvance = m.PayInAdvance.ValueBool()
		}
		if !m.RegroupPaidFees.IsNull() {
			if m.RegroupPaidFees.ValueBool() {
				charge.RegroupPaidFees = "included"
			} else {
				charge.RegroupPaidFees = ""
			}
		}
		if !m.Prorated.IsNull() {
			charge.Prorated = m.Prorated.ValueBool()
		}
		if !m.MinAmountCents.IsNull() {
			charge.MinAmountCents = int(m.MinAmountCents.ValueInt64())
		}
		if !m.PropertiesJSON.IsNull() {
			props, err := parseJSONField(m.PropertiesJSON)
			if err != nil {
				diags.AddError("Invalid charge properties_json", err.Error())
				return nil, diags
			}
			charge.Properties = props
		}
		if !m.FiltersJSON.IsNull() {
			filters, err := parseJSONFiltersField(m.FiltersJSON)
			if err != nil {
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

func flattenCharges(charges []lago.Charge) (types.List, diag.Diagnostics) {
	if len(charges) == 0 {
		return types.ListNull(chargeObjectType()), nil
	}

	values := make([]attr.Value, 0, len(charges))
	var diags diag.Diagnostics
	for _, charge := range charges {
		chargeTaxCodes := taxCodesFromTaxes(charge.Taxes)
		taxCodes, tdiags := flattenStringSet(chargeTaxCodes)
		diags.Append(tdiags...)
		if diags.HasError() {
			return types.ListNull(chargeObjectType()), diags
		}
		regroupPaidFees := charge.RegroupPaidFees == "included"
		attrs := map[string]attr.Value{
			"billable_metric_id":   types.StringValue(charge.LagoBillableMetricID.String()),
			"charge_model":         types.StringValue(string(charge.ChargeModel)),
			"invoiceable":          types.BoolValue(charge.Invoiceable),
			"invoice_display_name": stringOrNull(charge.InvoiceDisplayName),
			"pay_in_advance":       types.BoolValue(charge.PayInAdvance),
			"regroup_paid_fees":    types.BoolValue(regroupPaidFees),
			"prorated":             types.BoolValue(charge.Prorated),
			"min_amount_cents":     types.Int64Value(int64(charge.MinAmountCents)),
			"properties_json":      jsonPropertiesValue(charge.Properties),
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

func expandMinimumCommitment(ctx context.Context, value types.Object) (*lago.MinimumCommitmentInput, diag.Diagnostics) {
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

	out := &lago.MinimumCommitmentInput{AmountCents: int(model.AmountCents.ValueInt64())}
	if !model.InvoiceDisplayName.IsNull() {
		out.InvoiceDisplayName = model.InvoiceDisplayName.ValueString()
	}
	taxCodes, tdiags := expandStringSet(ctx, model.TaxCodes)
	diags.Append(tdiags...)
	if diags.HasError() {
		return nil, diags
	}
	out.TaxCodes = taxCodes
	return out, diags
}

func flattenMinimumCommitment(value *lago.MinimumCommitment) (types.Object, diag.Diagnostics) {
	if value == nil {
		return types.ObjectNull(minimumCommitmentObjectType().AttrTypes), nil
	}

	taxCodes, diags := flattenStringSet(taxCodesFromTaxes(value.Taxes))
	if diags.HasError() {
		return types.ObjectNull(minimumCommitmentObjectType().AttrTypes), diags
	}

	obj, odiags := types.ObjectValue(minimumCommitmentObjectType().AttrTypes, map[string]attr.Value{
		"amount_cents":         types.Int64Value(int64(value.AmountCents)),
		"invoice_display_name": stringOrNull(value.InvoiceDisplayName),
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

func expandFixedCharges(ctx context.Context, value types.List) ([]lago.FixedChargeInput, diag.Diagnostics) {
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

	out := make([]lago.FixedChargeInput, 0, len(models))
	for _, m := range models {
		fc := lago.FixedChargeInput{}
		if !m.AddOnID.IsNull() {
			addOnID, err := uuid.Parse(m.AddOnID.ValueString())
			if err != nil {
				diags.AddError("Invalid add_on_id", fmt.Sprintf("Must be a valid UUID: %s", err))
				return nil, diags
			}
			fc.AddOnID = addOnID
		}
		if !m.ChargeModel.IsNull() {
			fc.ChargeModel = lago.FixedChargeModel(m.ChargeModel.ValueString())
		}
		if !m.InvoiceDisplayName.IsNull() {
			fc.InvoiceDisplayName = m.InvoiceDisplayName.ValueString()
		}
		if !m.PayInAdvance.IsNull() {
			fc.PayInAdvance = m.PayInAdvance.ValueBool()
		}
		if !m.Prorated.IsNull() {
			fc.Prorated = m.Prorated.ValueBool()
		}
		if !m.Units.IsNull() {
			fc.Units = float64(m.Units.ValueInt64())
		}
		if !m.PropertiesJSON.IsNull() {
			props, err := parseJSONField(m.PropertiesJSON)
			if err != nil {
				diags.AddError("Invalid fixed_charge properties_json", err.Error())
				return nil, diags
			}
			if props != nil {
				raw, err := json.Marshal(props)
				if err != nil {
					diags.AddError("Invalid fixed_charge properties_json", err.Error())
					return nil, diags
				}
				var fcp lago.FixedChargeProperties
				if err := json.Unmarshal(raw, &fcp); err != nil {
					diags.AddError("Invalid fixed_charge properties_json", err.Error())
					return nil, diags
				}
				fc.Properties = &fcp
			}
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

func flattenFixedCharges(values []lago.FixedCharge) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(fixedChargeObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	var diags diag.Diagnostics
	for _, v := range values {
		taxCodes, tdiags := flattenStringSet(taxCodesFromTaxes(v.Taxes))
		diags.Append(tdiags...)
		if diags.HasError() {
			return types.ListNull(fixedChargeObjectType()), diags
		}
		var propertiesJSON types.String
		if v.Properties != nil {
			raw, err := json.Marshal(v.Properties)
			if err == nil {
				propertiesJSON = types.StringValue(string(raw))
			} else {
				propertiesJSON = types.StringNull()
			}
		} else {
			propertiesJSON = types.StringNull()
		}
		out = append(out, types.ObjectValueMust(fixedChargeObjectType().AttrTypes, map[string]attr.Value{
			"add_on_id":            types.StringValue(v.LagoAddOnID.String()),
			"add_on_code":          stringOrNull(v.AddOnCode),
			"charge_model":         types.StringValue(string(v.ChargeModel)),
			"invoice_display_name": stringOrNull(v.InvoiceDisplayName),
			"pay_in_advance":       types.BoolValue(v.PayInAdvance),
			"prorated":             types.BoolValue(v.Prorated),
			"units":                types.Int64Value(int64(v.Units)),
			"properties_json":      propertiesJSON,
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
	}}
}

func expandUsageThresholds(ctx context.Context, value types.List) ([]lago.UsageThresholdInput, diag.Diagnostics) {
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

	out := make([]lago.UsageThresholdInput, 0, len(models))
	for _, m := range models {
		ut := lago.UsageThresholdInput{}
		if !m.AmountCents.IsNull() {
			ut.AmountCents = int(m.AmountCents.ValueInt64())
		}
		if !m.ThresholdDisplayName.IsNull() {
			ut.ThresholdDisplayName = m.ThresholdDisplayName.ValueString()
		}
		if !m.Recurring.IsNull() {
			ut.Recurring = m.Recurring.ValueBool()
		}
		out = append(out, ut)
	}

	return out, diags
}

func flattenUsageThresholds(values []lago.UsageThreshold) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(usageThresholdObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	for _, v := range values {
		out = append(out, types.ObjectValueMust(usageThresholdObjectType().AttrTypes, map[string]attr.Value{
			"amount_cents":           types.Int64Value(int64(v.AmountCents)),
			"threshold_display_name": stringOrNull(v.ThresholdDisplayName),
			"recurring":              types.BoolValue(v.Recurring),
		}))
	}

	list, diags := types.ListValue(usageThresholdObjectType(), out)
	return list, diags
}

func entitlementObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"code":        types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	}}
}

func flattenEntitlements(values []lago.PlanEntitlement) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(entitlementObjectType()), nil
	}

	out := make([]attr.Value, 0, len(values))
	for _, v := range values {
		out = append(out, types.ObjectValueMust(entitlementObjectType().AttrTypes, map[string]attr.Value{
			"code":        stringOrNull(v.Code),
			"name":        stringOrNull(v.Name),
			"description": stringOrNull(v.Description),
		}))
	}
	list, diags := types.ListValue(entitlementObjectType(), out)
	return list, diags
}

func taxCodesFromTaxes(taxes []lago.Tax) []string {
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
