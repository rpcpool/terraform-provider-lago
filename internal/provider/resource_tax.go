package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &taxResource{}
	_ resource.ResourceWithConfigure   = &taxResource{}
	_ resource.ResourceWithImportState = &taxResource{}
)

func NewTaxResource() resource.Resource {
	return &taxResource{}
}

type taxResource struct {
	client *lago.Client
}

type taxResourceModel struct {
	ID                    types.String  `tfsdk:"id"`
	LagoID                types.String  `tfsdk:"lago_id"`
	Name                  types.String  `tfsdk:"name"`
	Code                  types.String  `tfsdk:"code"`
	Description           types.String  `tfsdk:"description"`
	Rate                  types.Float64 `tfsdk:"rate"`
	AppliedToOrganization types.Bool    `tfsdk:"applied_to_organization"`
	CreatedAt             types.String  `tfsdk:"created_at"`
}

func (r *taxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tax"
}

func (r *taxResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago tax.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the tax code.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the tax.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Tax name.",
			},
			"code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique tax code. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tax description.",
			},
			"rate": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Tax rate as a percentage (e.g. `20.0` for 20%).",
			},
			"applied_to_organization": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this tax is automatically applied to the organization.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Tax creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *taxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *taxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandTaxInput(plan)

	created, lagoErr := r.client.Tax().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Tax", lagoErr.Error())
		return
	}

	state := mapTaxToModel(created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *taxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state taxResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tax, lagoErr := r.client.Tax().Get(ctx, state.Code.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Tax", lagoErr.Error())
		return
	}

	newState := mapTaxToModel(tax, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *taxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan taxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandTaxInput(plan)

	updated, lagoErr := r.client.Tax().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Tax", lagoErr.Error())
		return
	}

	state := mapTaxToModel(updated, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *taxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state taxResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.Tax().Delete(ctx, state.Code.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Tax", lagoErr.Error())
	}
}

func (r *taxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("code"), req.ID)...)
}

func expandTaxInput(model taxResourceModel) *lago.TaxInput {
	rate := float32(model.Rate.ValueFloat64())

	input := &lago.TaxInput{
		Name:                  model.Name.ValueString(),
		Code:                  model.Code.ValueString(),
		Rate:                  &rate,
		AppliedToOrganization: model.AppliedToOrganization.ValueBool(),
	}

	if !model.Description.IsNull() {
		input.Description = model.Description.ValueString()
	}

	return input
}

func mapTaxToModel(tax *lago.Tax, base taxResourceModel) taxResourceModel {
	state := base

	state.ID = types.StringValue(tax.Code)
	state.LagoID = types.StringValue(tax.LagoID.String())
	state.Name = types.StringValue(tax.Name)
	state.Code = types.StringValue(tax.Code)
	state.Rate = types.Float64Value(float64(tax.Rate))
	state.AppliedToOrganization = types.BoolValue(tax.AppliedToOrganization)
	state.Description = stringOrNull(tax.Description)

	if tax.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(tax.CreatedAt.Format(time.RFC3339))
	}

	return state
}
