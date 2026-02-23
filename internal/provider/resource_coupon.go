package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &couponResource{}
	_ resource.ResourceWithConfigure   = &couponResource{}
	_ resource.ResourceWithImportState = &couponResource{}
)

func NewCouponResource() resource.Resource {
	return &couponResource{}
}

type couponResource struct {
	client *lago.Client
}

type couponResourceModel struct {
	ID                  types.String  `tfsdk:"id"`
	LagoID              types.String  `tfsdk:"lago_id"`
	Name                types.String  `tfsdk:"name"`
	Code                types.String  `tfsdk:"code"`
	Description         types.String  `tfsdk:"description"`
	CouponType          types.String  `tfsdk:"coupon_type"`
	AmountCents         types.Int64   `tfsdk:"amount_cents"`
	AmountCurrency      types.String  `tfsdk:"amount_currency"`
	PercentageRate      types.Float64 `tfsdk:"percentage_rate"`
	Expiration          types.String  `tfsdk:"expiration"`
	ExpirationAt        types.String  `tfsdk:"expiration_at"`
	Frequency           types.String  `tfsdk:"frequency"`
	FrequencyDuration   types.Int64   `tfsdk:"frequency_duration"`
	Reusable            types.Bool    `tfsdk:"reusable"`
	PlanCodes           types.Set     `tfsdk:"plan_codes"`
	BillableMetricCodes types.Set     `tfsdk:"billable_metric_codes"`
	CreatedAt           types.String  `tfsdk:"created_at"`
}

func (r *couponResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_coupon"
}

func (r *couponResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago coupon.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the coupon code.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the coupon.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Coupon name.",
			},
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique coupon code. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Coupon description.",
			},
			"coupon_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Coupon type. Allowed values: `fixed_amount`, `percentage`.",
				Validators: []validator.String{
					stringvalidator.OneOf("fixed_amount", "percentage"),
				},
			},
			"amount_cents": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Discount amount in cents. Required when `coupon_type` is `fixed_amount`.",
			},
			"amount_currency": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Currency code for the discount amount (e.g. `USD`). Required when `coupon_type` is `fixed_amount`.",
			},
			"percentage_rate": schema.Float64Attribute{
				Optional:            true,
				MarkdownDescription: "Discount percentage rate (e.g. `15.0` for 15%). Required when `coupon_type` is `percentage`.",
			},
			"expiration": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Expiration strategy. Allowed values: `time_limit`, `no_expiration`.",
				Validators: []validator.String{
					stringvalidator.OneOf("time_limit", "no_expiration"),
				},
			},
			"expiration_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Expiration date and time (RFC3339). Required when `expiration` is `time_limit`.",
			},
			"frequency": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Frequency at which the coupon is applied. Allowed values: `once`, `recurring`.",
				Validators: []validator.String{
					stringvalidator.OneOf("once", "recurring"),
				},
			},
			"frequency_duration": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of billing periods the coupon applies to. Required when `frequency` is `recurring`.",
			},
			"reusable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the coupon can be applied multiple times to the same customer.",
			},
			"plan_codes": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of plan codes this coupon is limited to. If empty, the coupon applies to all plans.",
			},
			"billable_metric_codes": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of billable metric codes this coupon is limited to.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Coupon creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *couponResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *couponResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan couponResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandCouponInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.Coupon().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Coupon", lagoErr.Error())
		return
	}

	state, diags := mapCouponToModel(created, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *couponResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state couponResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	coupon, lagoErr := r.client.Coupon().Get(ctx, state.Code.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Coupon", lagoErr.Error())
		return
	}

	newState, diags := mapCouponToModel(coupon, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *couponResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan couponResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandCouponInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, lagoErr := r.client.Coupon().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Coupon", lagoErr.Error())
		return
	}

	state, diags := mapCouponToModel(updated, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *couponResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state couponResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.Coupon().Delete(ctx, state.Code.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Coupon", lagoErr.Error())
	}
}

func (r *couponResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func expandCouponInput(ctx context.Context, model couponResourceModel) (*lago.CouponInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	couponType := lago.CouponCalculationType(model.CouponType.ValueString())

	// Cross-field validation for union type fields.
	if couponType == lago.CouponTypeFixedAmount {
		if model.AmountCents.IsNull() {
			diags.AddError(
				"Missing amount_cents",
				"`amount_cents` is required when `coupon_type` is `fixed_amount`.",
			)
		}
		if model.AmountCurrency.IsNull() {
			diags.AddError(
				"Missing amount_currency",
				"`amount_currency` is required when `coupon_type` is `fixed_amount`.",
			)
		}
	}

	if couponType == lago.CouponTypePercentage {
		if model.PercentageRate.IsNull() {
			diags.AddError(
				"Missing percentage_rate",
				"`percentage_rate` is required when `coupon_type` is `percentage`.",
			)
		}
	}

	expiration := lago.CouponExpiration(model.Expiration.ValueString())
	if expiration == lago.CouponExpirationTimeLimit && model.ExpirationAt.IsNull() {
		diags.AddError(
			"Missing expiration_at",
			"`expiration_at` is required when `expiration` is `time_limit`.",
		)
	}

	frequency := lago.CouponFrequency(model.Frequency.ValueString())
	if frequency == lago.CouponFrequencyRecurring && model.FrequencyDuration.IsNull() {
		diags.AddError(
			"Missing frequency_duration",
			"`frequency_duration` is required when `frequency` is `recurring`.",
		)
	}

	if diags.HasError() {
		return nil, diags
	}

	input := &lago.CouponInput{
		Name:       model.Name.ValueString(),
		Code:       model.Code.ValueString(),
		CouponType: couponType,
		Expiration: expiration,
		Frequency:  frequency,
		Reusable:   model.Reusable.ValueBool(),
	}

	if !model.Description.IsNull() {
		input.Description = model.Description.ValueString()
	}

	if !model.AmountCents.IsNull() {
		input.AmountCents = int(model.AmountCents.ValueInt64())
	}

	if !model.AmountCurrency.IsNull() {
		input.AmountCurrency = lago.Currency(model.AmountCurrency.ValueString())
	}

	if !model.PercentageRate.IsNull() {
		input.PercentageRate = model.PercentageRate.ValueFloat64()
	}

	if !model.ExpirationAt.IsNull() {
		t, err := time.Parse(time.RFC3339, model.ExpirationAt.ValueString())
		if err != nil {
			diags.AddError(
				"Invalid expiration_at",
				fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
			)
			return nil, diags
		}
		input.ExpirationAt = &t
	}

	if !model.FrequencyDuration.IsNull() {
		input.FrequencyDuration = int(model.FrequencyDuration.ValueInt64())
	}

	planCodes, pdiags := expandStringSet(ctx, model.PlanCodes)
	diags.Append(pdiags...)
	if diags.HasError() {
		return nil, diags
	}

	bmCodes, bmdiags := expandStringSet(ctx, model.BillableMetricCodes)
	diags.Append(bmdiags...)
	if diags.HasError() {
		return nil, diags
	}

	if len(planCodes) > 0 || len(bmCodes) > 0 {
		input.AppliesTo = lago.LimitationInput{
			PlanCodes:           planCodes,
			BillableMetricCodes: bmCodes,
		}
	}

	return input, diags
}

func mapCouponToModel(coupon *lago.Coupon, base couponResourceModel) (couponResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := base

	state.ID = types.StringValue(coupon.Code)
	state.LagoID = types.StringValue(coupon.LagoID.String())
	state.Name = types.StringValue(coupon.Name)
	state.Code = types.StringValue(coupon.Code)
	state.CouponType = types.StringValue(string(coupon.CouponType))
	state.Expiration = types.StringValue(string(coupon.Expiration))
	state.Frequency = types.StringValue(string(coupon.Frequency))
	state.Reusable = types.BoolValue(coupon.Reusable)

	state.Description = stringOrNull(coupon.Description)

	// Map union type fields — set unused type's field to null.
	if coupon.CouponType == lago.CouponTypeFixedAmount {
		state.AmountCents = types.Int64Value(int64(coupon.AmountCents))
		state.AmountCurrency = types.StringValue(string(coupon.AmountCurrency))
		state.PercentageRate = types.Float64Null()
	} else {
		state.AmountCents = types.Int64Null()
		state.AmountCurrency = types.StringNull()
		state.PercentageRate = types.Float64Value(coupon.PercentageRate)
	}

	if coupon.ExpirationAt == nil {
		state.ExpirationAt = types.StringNull()
	} else {
		state.ExpirationAt = types.StringValue(coupon.ExpirationAt.Format(time.RFC3339))
	}

	if coupon.Frequency == lago.CouponFrequencyRecurring {
		state.FrequencyDuration = types.Int64Value(int64(coupon.FrequencyDuration))
	} else {
		state.FrequencyDuration = types.Int64Null()
	}

	// Plan codes — use what the API returned; fall back to null set if empty.
	planCodes, pdiags := flattenStringSet(coupon.PlanCodes)
	diags.Append(pdiags...)
	if diags.HasError() {
		return state, diags
	}
	// Preserve null vs empty: if user didn't configure plan_codes, keep null.
	if planCodes.IsNull() && !base.PlanCodes.IsNull() {
		planCodes, pdiags = flattenStringSet(nil)
		diags.Append(pdiags...)
	}
	state.PlanCodes = planCodes

	bmCodes, bmdiags := flattenStringSet(coupon.BillableMetricCodes)
	diags.Append(bmdiags...)
	if diags.HasError() {
		return state, diags
	}
	if bmCodes.IsNull() && !base.BillableMetricCodes.IsNull() {
		bmCodes, bmdiags = flattenStringSet(nil)
		diags.Append(bmdiags...)
	}
	state.BillableMetricCodes = bmCodes

	if coupon.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(coupon.CreatedAt.Format(time.RFC3339))
	}

	return state, diags
}
