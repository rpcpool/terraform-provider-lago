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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &subscriptionResource{}
	_ resource.ResourceWithConfigure   = &subscriptionResource{}
	_ resource.ResourceWithImportState = &subscriptionResource{}
)

func NewSubscriptionResource() resource.Resource {
	return &subscriptionResource{}
}

type subscriptionResource struct {
	client *lago.Client
}

type subscriptionResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	LagoID                  types.String `tfsdk:"lago_id"`
	ExternalID              types.String `tfsdk:"external_id"`
	ExternalCustomerID      types.String `tfsdk:"external_customer_id"`
	PlanCode                types.String `tfsdk:"plan_code"`
	Name                    types.String `tfsdk:"name"`
	BillingTime             types.String `tfsdk:"billing_time"`
	SubscriptionAt          types.String `tfsdk:"subscription_at"`
	EndingAt                types.String `tfsdk:"ending_at"`
	Status                  types.String `tfsdk:"status"`
	OnTerminationCreditNote types.String `tfsdk:"on_termination_credit_note"`
	OnTerminationInvoice    types.String `tfsdk:"on_termination_invoice"`
	CreatedAt               types.String `tfsdk:"created_at"`
	StartedAt               types.String `tfsdk:"started_at"`
}

func (r *subscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *subscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago subscription.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the subscription external ID.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the subscription.",
			},
			"external_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique external identifier for the subscription. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"external_customer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "External identifier of the customer to subscribe. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Code of the plan to subscribe the customer to. Can be changed in-place to trigger a plan upgrade or downgrade.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Display name for the subscription.",
			},
			"billing_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Billing time anchor. Allowed values: `anniversary`, `calendar`.",
				Validators: []validator.String{
					stringvalidator.OneOf("anniversary", "calendar"),
				},
			},
			"subscription_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Start date and time for the subscription (RFC3339). Defaults to the current date when not set.",
			},
			"ending_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "End date and time for the subscription (RFC3339). When set, the subscription will be automatically terminated at this date.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the subscription (`active`, `pending`, `terminated`, `canceled`).",
			},
			"on_termination_credit_note": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Credit note behaviour on termination. Allowed values: `credit`, `refund`, `offset`, `skip`.",
				Validators: []validator.String{
					stringvalidator.OneOf("credit", "refund", "offset", "skip"),
				},
			},
			"on_termination_invoice": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Invoice behaviour on termination. Allowed values: `generate`, `skip`.",
				Validators: []validator.String{
					stringvalidator.OneOf("generate", "skip"),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Subscription creation timestamp (RFC3339).",
			},
			"started_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Subscription start timestamp (RFC3339).",
			},
		},
	}
}

func (r *subscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subscriptionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandSubscriptionInput(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.Subscription().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Subscription", lagoErr.Error())
		return
	}

	state := mapSubscriptionToModel(created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subscriptionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscription, lagoErr := r.client.Subscription().Get(ctx, state.ExternalID.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Subscription", lagoErr.Error())
		return
	}

	// Treat out-of-band termination as resource destruction.
	if subscription.Status == lago.SubscriptionStatusTerminated {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := mapSubscriptionToModel(subscription, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *subscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subscriptionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandSubscriptionInput(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, lagoErr := r.client.Subscription().Update(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Subscription", lagoErr.Error())
		return
	}

	state := mapSubscriptionToModel(updated, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subscriptionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	terminateInput := lago.SubscriptionTerminateInput{
		ExternalID: state.ExternalID.ValueString(),
	}

	if !state.OnTerminationCreditNote.IsNull() && !state.OnTerminationCreditNote.IsUnknown() {
		terminateInput.OnTerminationCreditNote = lago.OnTerminationCreditNote(state.OnTerminationCreditNote.ValueString())
	}

	if !state.OnTerminationInvoice.IsNull() && !state.OnTerminationInvoice.IsUnknown() {
		terminateInput.OnTerminationInvoice = lago.OnTerminationInvoice(state.OnTerminationInvoice.ValueString())
	}

	_, lagoErr := r.client.Subscription().Terminate(ctx, terminateInput)
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Terminating Lago Subscription", lagoErr.Error())
	}
}

func (r *subscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("external_id"), req.ID)...)
}

func expandSubscriptionInput(model subscriptionResourceModel) (*lago.SubscriptionInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &lago.SubscriptionInput{
		ExternalID:         model.ExternalID.ValueString(),
		ExternalCustomerID: model.ExternalCustomerID.ValueString(),
		PlanCode:           model.PlanCode.ValueString(),
	}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		input.Name = model.Name.ValueString()
	}

	if !model.BillingTime.IsNull() && !model.BillingTime.IsUnknown() {
		input.BillingTime = lago.BillingTime(model.BillingTime.ValueString())
	}

	if !model.SubscriptionAt.IsNull() && !model.SubscriptionAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, model.SubscriptionAt.ValueString())
		if err != nil {
			diags.AddError(
				"Invalid subscription_at",
				fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
			)
			return nil, diags
		}
		input.SubscriptionAt = &t
	}

	if !model.EndingAt.IsNull() && !model.EndingAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, model.EndingAt.ValueString())
		if err != nil {
			diags.AddError(
				"Invalid ending_at",
				fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
			)
			return nil, diags
		}
		input.EndingAt = &t
	}

	return input, diags
}

func mapSubscriptionToModel(subscription *lago.Subscription, base subscriptionResourceModel) subscriptionResourceModel {
	state := base

	state.ID = types.StringValue(subscription.ExternalID)
	state.LagoID = types.StringValue(subscription.LagoID.String())
	state.ExternalID = types.StringValue(subscription.ExternalID)
	state.ExternalCustomerID = types.StringValue(subscription.ExternalCustomerID)
	state.PlanCode = types.StringValue(subscription.PlanCode)
	state.Status = types.StringValue(string(subscription.Status))

	state.Name = stringOrNull(subscription.Name)
	state.BillingTime = stringOrNull(string(subscription.BillingTime))

	if subscription.SubscriptionAt.IsZero() {
		state.SubscriptionAt = types.StringNull()
	} else {
		state.SubscriptionAt = types.StringValue(subscription.SubscriptionAt.Format(time.RFC3339))
	}

	state.EndingAt = timeOrNull(subscription.EndingAt)
	state.StartedAt = timeOrNull(subscription.StartedAt)

	if subscription.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(subscription.CreatedAt.Format(time.RFC3339))
	}

	// on_termination_credit_note and on_termination_invoice: read from API response
	// when non-empty; otherwise preserve from prior state to avoid perpetual drift.
	if subscription.OnTerminationCreditNote != "" {
		state.OnTerminationCreditNote = types.StringValue(string(subscription.OnTerminationCreditNote))
	}
	// If the API returned empty but base already had a value, keep it.

	if subscription.OnTerminationInvoice != "" {
		state.OnTerminationInvoice = types.StringValue(string(subscription.OnTerminationInvoice))
	}

	return state
}
