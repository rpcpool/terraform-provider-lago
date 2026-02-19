package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/triton-one/terraform-provider-lago/internal/client"
)

var (
	_ resource.Resource                = &billableMetricResource{}
	_ resource.ResourceWithConfigure   = &billableMetricResource{}
	_ resource.ResourceWithImportState = &billableMetricResource{}
)

var allowedAggregationTypes = []string{
	"count_agg",
	"sum_agg",
	"max_agg",
	"latest_agg",
	"weighted_sum_agg",
	"unique_count_agg",
}

func NewBillableMetricResource() resource.Resource {
	return &billableMetricResource{}
}

type billableMetricResource struct {
	client *client.Client
}

type billableMetricResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Code             types.String `tfsdk:"code"`
	Description      types.String `tfsdk:"description"`
	AggregationType  types.String `tfsdk:"aggregation_type"`
	FieldName        types.String `tfsdk:"field_name"`
	Expression       types.String `tfsdk:"expression"`
	Recurring        types.Bool   `tfsdk:"recurring"`
	WeightedInterval types.String `tfsdk:"weighted_interval"`
	Filters          types.Set    `tfsdk:"filters"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *billableMetricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billable_metric"
}

func (r *billableMetricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago billable metric.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the metric code.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Billable metric name.",
			},
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique billable metric code.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Billable metric description.",
			},
			"aggregation_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Aggregation type.",
				Validators: []validator.String{
					stringvalidator.OneOf(allowedAggregationTypes...),
				},
			},
			"field_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Field to aggregate for non-count aggregations.",
			},
			"expression": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Expression used by custom metrics.",
			},
			"recurring": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the metric is recurring.",
			},
			"weighted_interval": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Interval used by weighted sum aggregation.",
			},
			"filters": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Optional metric filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Filter key.",
						},
						"values": schema.SetAttribute{
							Required:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Allowed values for the filter key.",
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Metric creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Metric last update timestamp.",
			},
		},
	}
}

func (r *billableMetricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *billableMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan billableMetricResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateWeightedInterval(plan.AggregationType, plan.WeightedInterval)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateBillableMetricInput{
		Name:            plan.Name.ValueString(),
		Code:            plan.Code.ValueString(),
		AggregationType: plan.AggregationType.ValueString(),
	}
	if !plan.Description.IsNull() {
		description := plan.Description.ValueString()
		input.Description = &description
	}
	if !plan.FieldName.IsNull() {
		fieldName := plan.FieldName.ValueString()
		input.FieldName = &fieldName
	}
	if !plan.Expression.IsNull() {
		expression := plan.Expression.ValueString()
		input.Expression = &expression
	}
	if !plan.Recurring.IsNull() {
		recurring := plan.Recurring.ValueBool()
		input.Recurring = &recurring
	}
	if !plan.WeightedInterval.IsNull() {
		weightedInterval := plan.WeightedInterval.ValueString()
		input.WeightedInterval = &weightedInterval
	}
	filters, diags := expandFilters(ctx, plan.Filters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Filters = filters

	created, err := r.client.CreateBillableMetric(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Lago Billable Metric", err.Error())
		return
	}

	state := mapBillableMetricToModel(created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *billableMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state billableMetricResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metric, err := r.client.GetBillableMetricByCode(ctx, state.Code.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error Reading Lago Billable Metric", err.Error())
		return
	}

	newState := mapBillableMetricToModel(metric, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *billableMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan billableMetricResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateWeightedInterval(plan.AggregationType, plan.WeightedInterval)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.UpdateBillableMetricInput{}

	name := plan.Name.ValueString()
	input.Name = &name

	aggregationType := plan.AggregationType.ValueString()
	input.AggregationType = &aggregationType

	if !plan.Description.IsNull() {
		description := plan.Description.ValueString()
		input.Description = &description
	}
	if !plan.FieldName.IsNull() {
		fieldName := plan.FieldName.ValueString()
		input.FieldName = &fieldName
	}
	if !plan.Expression.IsNull() {
		expression := plan.Expression.ValueString()
		input.Expression = &expression
	}
	if !plan.Recurring.IsNull() {
		recurring := plan.Recurring.ValueBool()
		input.Recurring = &recurring
	}
	if !plan.WeightedInterval.IsNull() {
		weightedInterval := plan.WeightedInterval.ValueString()
		input.WeightedInterval = &weightedInterval
	}
	filters, diags := expandFilters(ctx, plan.Filters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Filters = filters

	metric, err := r.client.UpdateBillableMetricByCode(ctx, plan.Code.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Lago Billable Metric", err.Error())
		return
	}

	state := mapBillableMetricToModel(metric, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *billableMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state billableMetricResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBillableMetricByCode(ctx, state.Code.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Lago Billable Metric", err.Error())
	}
}

func (r *billableMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func validateWeightedInterval(aggregationType types.String, weightedInterval types.String) (diags diag.Diagnostics) {
	if aggregationType.IsUnknown() || weightedInterval.IsUnknown() {
		return diags
	}

	if aggregationType.IsNull() || weightedInterval.IsNull() {
		return diags
	}

	if aggregationType.ValueString() != "weighted_sum_agg" && strings.TrimSpace(weightedInterval.ValueString()) != "" {
		diags.AddError(
			"Invalid weighted_interval usage",
			"`weighted_interval` can only be set when `aggregation_type` is `weighted_sum_agg`.",
		)
	}

	return diags
}

func mapBillableMetricToModel(metric *client.BillableMetric, base billableMetricResourceModel) billableMetricResourceModel {
	state := base

	state.ID = types.StringValue(metric.Code)
	state.Name = types.StringValue(metric.Name)
	state.Code = types.StringValue(metric.Code)
	state.AggregationType = types.StringValue(metric.AggregationType)
	state.Recurring = types.BoolValue(metric.Recurring)

	if metric.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(metric.Description)
	}

	if metric.FieldName == "" {
		state.FieldName = types.StringNull()
	} else {
		state.FieldName = types.StringValue(metric.FieldName)
	}

	if metric.Expression == "" {
		state.Expression = types.StringNull()
	} else {
		state.Expression = types.StringValue(metric.Expression)
	}

	if metric.WeightedInterval == "" {
		state.WeightedInterval = types.StringNull()
	} else {
		state.WeightedInterval = types.StringValue(metric.WeightedInterval)
	}

	filters, diags := flattenFilters(metric.Filters)
	if diags.HasError() {
		state.Filters = base.Filters
	} else {
		state.Filters = filters
	}

	if metric.CreatedAt == "" {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(metric.CreatedAt)
	}

	if metric.UpdatedAt == "" {
		state.UpdatedAt = types.StringNull()
	} else {
		state.UpdatedAt = types.StringValue(metric.UpdatedAt)
	}

	return state
}

type filterModel struct {
	Key    types.String `tfsdk:"key"`
	Values types.Set    `tfsdk:"values"`
}

func filterObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":    types.StringType,
			"values": types.SetType{ElemType: types.StringType},
		},
	}
}

func expandFilters(ctx context.Context, filters types.Set) ([]client.BillableMetricFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if filters.IsNull() {
		return nil, diags
	}
	if filters.IsUnknown() {
		diags.AddError("Invalid filters value", "`filters` contains unknown values.")
		return nil, diags
	}

	var models []filterModel
	diags.Append(filters.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]client.BillableMetricFilter, 0, len(models))
	for _, model := range models {
		if model.Key.IsNull() || model.Key.IsUnknown() {
			diags.AddError("Invalid filter key", "Each filter requires a known `key` value.")
			return nil, diags
		}
		if model.Values.IsNull() || model.Values.IsUnknown() {
			diags.AddError("Invalid filter values", "Each filter requires known `values`.")
			return nil, diags
		}

		var values []string
		diags.Append(model.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return nil, diags
		}

		out = append(out, client.BillableMetricFilter{
			Key:    model.Key.ValueString(),
			Values: values,
		})
	}

	return out, diags
}

func flattenFilters(filters []client.BillableMetricFilter) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(filters) == 0 {
		return types.SetNull(filterObjectType()), diags
	}

	values := make([]attr.Value, 0, len(filters))
	for _, filter := range filters {
		item, itemDiags := types.ObjectValue(
			filterObjectType().AttrTypes,
			map[string]attr.Value{
				"key":    types.StringValue(filter.Key),
				"values": stringSetValue(filter.Values),
			},
		)
		diags.Append(itemDiags...)
		values = append(values, item)
	}

	setValue, setDiags := types.SetValue(filterObjectType(), values)
	diags.Append(setDiags...)
	return setValue, diags
}

func stringSetValue(values []string) types.Set {
	if len(values) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}

	attrs := make([]attr.Value, 0, len(values))
	for _, v := range values {
		attrs = append(attrs, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, attrs)
}
